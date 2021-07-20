package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	linuxproc "github.com/c9s/goprocinfo/linux"
	"github.com/godbus/dbus"
)

func main() {
	log.Println("vim-go")
	dbus_init()

	for {
		time.Sleep(time.Second)
	}
}

func dbus_init() {
	conn, err := dbus.SessionBus()
	if err != nil {
		return
	}

	obj := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus").(*dbus.Object)

	err = obj.AddMatchSignal("org.freedesktop.DBus", "NameOwnerChanged").Err
	if err != nil {
		return
	}

	ch := make(chan *dbus.Signal, 100)
	conn.Signal(ch)

	go func() {
		for signal := range ch {
			signalProcess(conn, signal)
		}
	}()
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
			var pid uint32
			obj := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus")
			err := obj.Call("org.freedesktop.DBus.GetConnectionUnixProcessID", 0, newName).Store(&pid)
			if err != nil {
				return err
			}

			fmt.Println("========================================")
			log.Println("sender=", newName)

			prefix := ""
			for pid != 0 {
				cmdline, stat, err := GetPidInfo(pid)
				if err != nil {
					break
				}

				fmt.Printf("%sprocess:%d -- %s\n", prefix, pid, cmdline)
				//log.Println("ppid=", stat.Ppid)

				pid = uint32(stat.Ppid)
				prefix = prefix + "  "
			}
			fmt.Println("")

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
