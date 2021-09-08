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
2021/09/08 10:18:26 dbus-tool
========================================
2021/09/08 10:18:26 sender= :1.149 uid= 1000
↑_ process:4800 -- /usr/lib/polkit-1-dde/dde-polkit-agent
  ↑_ process:4633 -- /usr/bin/startdde
    ↑_ process:4560 -- /usr/bin/kwin_wayland-platformdde-kwin-wayland--xwayland--drm--no-lockscreenstartdde-wayland
      ↑_ process:4537 -- lightdm--session-child111819
        ↑_ process:4389 -- /usr/sbin/lightdm
          ↑_ process:1 -- /sbin/initsplashnokaslr
```