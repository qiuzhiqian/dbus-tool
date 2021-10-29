# 概述
这是一个dbus调试工具，主要是根据dbus连接name来显示对应的pid进程信息

# 编译
第一次编译的时候，可能需要自动安装一下以来
```
$ go mod tidy
```
之后就只需要正常编译即可
```
go build
```

# 使用
USEAGE:
```
$ dbus-tool -h
NAME:
   dbus-tool - dbus connection debug tool!

USAGE:
   dbus-tool [global options] command [command options] [arguments...]

VERSION:
   2.0.0

COMMANDS:
   monitor, m  monitor dbus connection
   list, l     list dbus connection
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

monitor实时监控当前及后续的连接
```
$ dbus-tool monitor --help
NAME:
   dbus-tool monitor - monitor dbus connection

USAGE:
   dbus-tool monitor [command options] [arguments...]

OPTIONS:
   --address value, -a value   set dbus address(session/system) (default: "session")
   --progress value, -p value  set which progess to filter
   --help, -h                  show help (default: false)
```
比如：
```
$ dbus-tool monitor --address system --progress lightdm
2021-10-29 16:48:14.905830491 +0800 CST m=+0.012836395 =======================
↳[1][0][] /sbin/init splash nokaslr 
 ↳[3511][0][] /usr/sbin/lightdm 
  ↳[5280][0][:1.109] lightdm --session-child 11 18 19 

2021-10-29 16:48:14.933106918 +0800 CST m=+0.040112822 =======================
↳[1][0][] /sbin/init splash nokaslr 
 ↳[3511][0][org.freedesktop.DisplayManager] /usr/sbin/lightdm 

2021-10-29 16:48:14.944928213 +0800 CST m=+0.051934117 =======================
↳[1][0][] /sbin/init splash nokaslr 
 ↳[3511][0][:1.36] /usr/sbin/lightdm 

```
