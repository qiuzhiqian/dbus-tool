package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	linuxproc "github.com/c9s/goprocinfo/linux"
	"github.com/godbus/dbus"
)

type DBUS_TYPE int

const (
	SESSION DBUS_TYPE = iota
	SYSTEM
)

type ProcessInfo struct {
	pid    uint32
	cmd    string
	sender string
	uid    uint32
	child  *ProcessInfo
}

func (p *ProcessInfo) Display() string {
	prefix := "↳"
	tab := ""
	output := ""
	current := p
	for {
		output = fmt.Sprintf("%s%s%s[%d][%d][%s] %s\n", output, tab, prefix, current.pid, current.uid, current.sender, current.cmd)
		tab = tab + " "
		if current.child == nil {
			break
		}
		current = current.child
	}

	return output
}

func main() {
	log.Println("dbus-tool")

	//sessionType := "session"
	dbusType := SESSION
	isSession := true
	isSystem := false
	isFollow := false
	name := ""

	flag.BoolVar(&isSession, "session", true, "use session dbus")
	flag.BoolVar(&isSystem, "system", false, "use system dbus")
	flag.BoolVar(&isFollow, "f", false, "follow monitor")
	flag.StringVar(&name, "name", "", "display sender info with name")
	flag.Parse()

	if isSession && isSystem {
		dbusType = SYSTEM
	} else if !isSession && !isSystem {
		dbusType = SESSION
	} else if isSystem {
		dbusType = SYSTEM
	} else {
		dbusType = SESSION
	}

	var conn *dbus.Conn
	var err error
	if dbusType != SYSTEM {
		conn, err = dbus.SessionBus()
		if err != nil {
			return
		}
	} else {
		conn, err = dbus.SystemBus()
		if err != nil {
			return
		}
	}

	if len(name) != 0 {
		p, err := GetPidTreeBySender(conn, name)
		if err != nil {
			return
		}
		fmt.Println(p.Display())
		return
	}

	dbusNamesInfo(conn)
	// will block
	if isFollow {
		dbusMonitor(conn)
	}

	defer conn.Close()
}

func dbusNamesInfo(conn *dbus.Conn) {
	names := listNames(conn)
	if len(names) == 0 {
		return
	}

	for _, name := range names {
		//fmt.Println("name:", name)
		p, err := GetPidTreeBySender(conn, name)
		if err != nil {
			continue
		}
		fmt.Println(p.Display())
		//fmt.Println("=======")
	}
}

func dbusMonitor(conn *dbus.Conn) {

	obj := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus").(*dbus.Object)

	err := obj.AddMatchSignal("org.freedesktop.DBus", "NameOwnerChanged").Err
	if err != nil {
		return
	}

	ch := make(chan *dbus.Signal, 100)
	conn.Signal(ch)

	for signal := range ch {
		signalProcess(conn, signal)
	}
}

func getConnectionUnixProcessID(conn *dbus.Conn, name string) (uint32, error) {
	var pid uint32
	obj := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus")
	err := obj.Call("org.freedesktop.DBus.GetConnectionUnixProcessID", 0, name).Store(&pid)
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func getConnectionUnixUser(conn *dbus.Conn, name string) (uint32, error) {
	var uid uint32
	obj := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus")
	err := obj.Call("org.freedesktop.DBus.GetConnectionUnixUser", 0, name).Store(&uid)
	if err != nil {
		return 0, err
	}
	return uid, nil
}

func listNames(conn *dbus.Conn) []string {
	names := make([]string, 0)
	obj := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus")
	err := obj.Call("org.freedesktop.DBus.ListNames", 0).Store(&names)
	if err != nil {
		return nil
	}
	return names
}

func signalProcess(conn *dbus.Conn, sig *dbus.Signal) error {
	if len(sig.Body) > 1 {
		name := sig.Body[0].(string)
		oldName := sig.Body[1].(string)
		newName := sig.Body[2].(string)

		if name == "" {
			return fmt.Errorf("name is nil")
		}

		if oldName == "" && newName != "" {
			p, err := GetPidTreeBySender(conn, newName)
			if err != nil {
				return err
			}
			fmt.Println(p.Display())
		}
	}

	return nil
}

func GetPidInfo(pid uint32) (string, *linuxproc.ProcessStat, error) {
	fileName := fmt.Sprintf("/proc/%d/cmdline", pid)

	fileData, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println("read file error =", err)
		return "", nil, err
	}

	statName := fmt.Sprintf("/proc/%d/stat", pid)
	stat, err := linuxproc.ReadProcessStat(statName)
	if err != nil {
		return "", nil, err
	}

	return string(fileData), stat, nil
}

func GetPidTreeBySender(conn *dbus.Conn, name string) (*ProcessInfo, error) {
	pid, err := getConnectionUnixProcessID(conn, name)
	if err != nil {
		return nil, err
	}

	uid, err := getConnectionUnixUser(conn, name)
	if err != nil {
		return nil, err
	}

	var node *ProcessInfo

	for pid != 0 {
		cmdline, stat, err := GetPidInfo(pid)
		if err != nil {
			break
		}

		if node != nil {
			parent := &ProcessInfo{
				pid:    pid,
				uid:    uid,
				cmd:    cmdline,
				sender: name,
				child:  node,
			}
			node = parent
		} else {
			node = &ProcessInfo{
				pid:    pid,
				uid:    uid,
				cmd:    cmdline,
				sender: name,
			}
		}

		uid = 0
		name = ""
		pid = uint32(stat.Ppid)
	}

	if node == nil {
		return nil, fmt.Errorf("find node info error")
	}
	return node, nil
}

func displayPidTree(pid uint32, prefix string) {
	for pid != 0 {
		cmdline, stat, err := GetPidInfo(pid)
		if err != nil {
			break
		}

		fmt.Printf("%sprocess:%d -- %s\n", prefix, pid, cmdline)
		//log.Println("ppid=", stat.Ppid)

		pid = uint32(stat.Ppid)
		prefix = "  " + prefix
	}
	fmt.Println("")
}
