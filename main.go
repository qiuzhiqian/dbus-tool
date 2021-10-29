package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	linuxproc "github.com/c9s/goprocinfo/linux"
	"github.com/godbus/dbus"
	"golang.org/x/sys/unix"

	"github.com/urfave/cli/v2"
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
	app := &cli.App{
		Name:    "dbus-tool",
		Version: "2.0.0",
		Usage:   "dbus connection debug tool!",
		Action: func(c *cli.Context) error {
			cli.ShowAppHelpAndExit(c, 0)
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "monitor",
				Aliases: []string{"m"},
				Usage:   "monitor dbus connection",
				Action: func(c *cli.Context) error {
					address := c.String("address")
					progress := c.String("progress")

					var conn *dbus.Conn
					var err error
					if address == "system" {
						conn, err = dbus.SystemBus()
						if err != nil {
							return err
						}
					} else {
						conn, err = dbus.SessionBus()
						if err != nil {
							return err
						}
					}

					dbusNamesInfo(conn, progress)
					dbusMonitor(conn, progress)
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "address",
						Aliases: []string{"a"},
						Value:   "session",
						Usage:   "set dbus address(session/system)",
					},
					&cli.StringFlag{
						Name:    "progress",
						Aliases: []string{"p"},
						Usage:   "set which progess to filter",
					},
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list dbus connection",
				Action: func(c *cli.Context) error {
					address := c.String("address")
					progress := c.String("progress")
					sender := c.String("sender")

					var conn *dbus.Conn
					var err error
					if address == "system" {
						conn, err = dbus.SystemBus()
						if err != nil {
							return err
						}
					} else {
						conn, err = dbus.SessionBus()
						if err != nil {
							return err
						}
					}

					defer conn.Close()
					if sender != "" {
						p, err := GetPidTreeBySender(conn, sender, "")
						if err != nil {
							return err
						}
						fmt.Println(p.Display())
					} else if progress != "" {
						dbusNamesInfo(conn, progress)
					} else if progress == "" && sender == "" {
						dbusNamesInfo(conn, "")
					}
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "address",
						Aliases: []string{"a"},
						Value:   "session",
						Usage:   "set dbus address(session/system)",
					},
					&cli.StringFlag{
						Name:    "progress",
						Aliases: []string{"p"},
						Usage:   "set which progess to filter",
					},
					&cli.StringFlag{
						Name:    "sender",
						Aliases: []string{"s"},
						Usage:   "set which sender to filter",
					},
				},
			},
		},
	}

	app.EnableBashCompletion = true

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func dbusNamesInfo(conn *dbus.Conn, rule string) {
	names := listNames(conn)
	if len(names) == 0 {
		return
	}

	for _, name := range names {
		//fmt.Println("name:", name)
		p, err := GetPidTreeBySender(conn, name, rule)
		if err != nil {
			continue
		}

		fmt.Println(time.Now(), "=======================")
		fmt.Println(p.Display())
	}
}

func dbusMonitor(conn *dbus.Conn, rule string) {

	obj := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus").(*dbus.Object)

	err := obj.AddMatchSignal("org.freedesktop.DBus", "NameOwnerChanged").Err
	if err != nil {
		return
	}

	ch := make(chan *dbus.Signal, 100)
	conn.Signal(ch)

	for signal := range ch {
		signalProcess(conn, signal, rule)
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

func signalProcess(conn *dbus.Conn, sig *dbus.Signal, rule string) error {
	if len(sig.Body) > 1 {
		name := sig.Body[0].(string)
		oldName := sig.Body[1].(string)
		newName := sig.Body[2].(string)

		if name == "" {
			return fmt.Errorf("name is nil")
		}

		if oldName == "" && newName != "" {
			p, err := GetPidTreeBySender(conn, newName, rule)
			if err != nil {
				if err != fmt.Errorf("match error.") {
					return err
				}
				return nil
			}
			fmt.Println(p.Display())
		}
	}

	return nil
}

func GetPidInfo(pid uint32) (string, *linuxproc.ProcessStat, uint32, error) {
	pidFile := fmt.Sprintf("/proc/%d", pid)
	var fileStat unix.Stat_t
	//获取/proc/$pid文件夹stat，该文件夹stat的uid权限可以代表进程uid
	err := unix.Stat(pidFile, &fileStat)
	if err != nil {
		return "", nil, 0, err
	}

	fileName := fmt.Sprintf("/proc/%d/cmdline", pid)

	fileData, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println("read file error =", err)
		return "", nil, 0, err
	}

	//通过hexdump发现，读出来的数据使用\x0分割，为了正常显示，需要使用\x20替换一下。
	cmdline := bytes.ReplaceAll(fileData, []byte{0x00}, []byte{0x20})

	statName := fmt.Sprintf("/proc/%d/stat", pid)
	stat, err := linuxproc.ReadProcessStat(statName)
	if err != nil {
		return "", nil, 0, err
	}

	return string(cmdline), stat, fileStat.Uid, nil
}

func GetPidTreeBySender(conn *dbus.Conn, name string, progress string) (*ProcessInfo, error) {
	pid, err := getConnectionUnixProcessID(conn, name)
	if err != nil {
		return nil, err
	}

	/*uid, err := getConnectionUnixUser(conn, name)
	if err != nil {
		return nil, err
	}*/

	var node *ProcessInfo

	hasMatch := false
	for pid != 0 {
		cmdline, stat, uid, err := GetPidInfo(pid)
		if err != nil {
			break
		}

		if progress != "" && !hasMatch && strings.Contains(cmdline, progress) && name != "" {
			hasMatch = true
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

		name = ""
		pid = uint32(stat.Ppid)
	}

	if !hasMatch && progress != "" {
		return nil, fmt.Errorf("match error.")
	}

	if node == nil {
		return nil, fmt.Errorf("find node info error")
	}
	return node, nil
}

func displayPidTree(pid uint32, prefix string) {
	for pid != 0 {
		cmdline, stat, _, err := GetPidInfo(pid)
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
