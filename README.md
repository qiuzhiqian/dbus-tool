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
```
$ ./dbus-tool --session
$ ./dbus-tool --session -f

$ sudo ./dbus-tool --system
$ sudo ./dbus-tool --system --name ":1.149"
```

显示格式如下：
```
2021-09-13 10:55:53.320891785 +0800 CST m=+0.084775664 =======================
↳[1][0][] /sbin/init splash nokaslr 
 ↳[5193][0][] /usr/sbin/lightdm 
  ↳[5367][0][] lightdm --session-child 9 18 19 
   ↳[5458][1000][] /usr/bin/kwin_wayland -platform dde-kwin-wayland --xwayland --drm --no-lockscreen startdde-wayland 
    ↳[5539][1000][] /usr/bin/startdde 
     ↳[5728][1000][:1.37] /usr/lib/polkit-1-dde/dde-polkit-agent
```