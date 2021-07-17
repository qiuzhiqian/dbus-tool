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
			//log.Println(signal)
			signalProcess(conn, signal)
		}
	}()
}

func signalProcess(conn *dbus.Conn, sig *dbus.Signal) error {
	if len(sig.Body) > 1 {
		name := sig.Body[0].(string)
		oldName := sig.Body[1].(string)
		newName := sig.Body[2].(string)

		//log.Println(name)
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

			fileName := fmt.Sprintf("/proc/%d/cmdline", pid)

			fileData, err := ioutil.ReadFile(fileName)
			if err != nil {
				log.Println("read file error =", err)
				return err
			}

			statName := fmt.Sprintf("/proc/%d/stat", pid)
			stat, err := linuxproc.ReadProcessStat(statName)
			if err != nil {
				return err
			}

			fmt.Println("========================================")
			log.Println("sender=", newName)
			log.Println("fileName=", fileName)
			log.Println("process:", string(fileData))
			log.Println("ppid=", stat.Ppid)
			fmt.Println("")
		}
	}

	return nil
}
