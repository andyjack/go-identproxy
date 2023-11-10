#!/bin/sh
# This file was automatically generated
# by the pfSense service handler.

rc_start() {
	echo starting go-identproxy
    /usr/sbin/daemon /usr/local/bin/go-identproxy 8113
	return 0
}

rc_stop() {
	echo stopping go-identproxy
    /usr/bin/killall -q go-identproxy
	return 0
}

rc_restart() {
	rc_stop
	rc_start

}

case $1 in
	start)
		rc_start
		;;
	stop)
		rc_stop
		;;
	restart)
		rc_restart
		;;
esac

