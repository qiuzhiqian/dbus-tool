export GO111MODULE=on

all:build

build:
	go mod tidy
	go build

install:
	install -Dm755 dbus-tool ${DESTDIR}/usr/sbin/dbus-tool

	install -Dm755 misc/dbus-tool ${DESTDIR}/etc/bash_completion.d/dbus-tool

clean:
	rm -f dbus-tool
