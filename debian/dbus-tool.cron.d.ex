#
# Regular cron jobs for the dbus-tool package
#
0 4	* * *	root	[ -x /usr/bin/dbus-tool_maintenance ] && /usr/bin/dbus-tool_maintenance
