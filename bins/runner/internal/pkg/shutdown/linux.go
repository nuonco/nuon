package shutdown

import (
	"fmt"
	"os"

	"github.com/godbus/dbus/v5"
)

type dbusArgs struct {
	destination string
	path        string
	iface       string
	method      string
	body        []any
}

func nameHasOwner(name string) (reply bool, err error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return false, err
	}
	defer func(conn *dbus.Conn) {
		err = conn.Close()
	}(conn)

	obj := conn.Object("org.freedesktop.DBus", "/")
	err = obj.Call("org.freedesktop.DBus.NameHasOwner", 0, name).Store(&reply)
	if err != nil {
		return false, err
	}

	return reply, nil
}

func dbusSend(
	destination string,
	path string,
	iface string,
	method string,
	body []any,
) (reply bool, err error) {
	hasOwner, err := nameHasOwner(destination)
	if err != nil {
		return false, err
	}

	if hasOwner {
		conn, err := dbus.ConnectSessionBus()
		if err != nil {
			return false, err
		}
		defer func(conn *dbus.Conn) {
			err = conn.Close()
		}(conn)

		obj := conn.Object(destination, dbus.ObjectPath(path))
		ifaceMethod := fmt.Sprintf("%s.%s", iface, method)
		err = obj.Call(ifaceMethod, 0, body...).Store(&reply)
		if err != nil {
			return false, err
		}

		return reply, nil
	}

	return false, nil
}

func getSessionId() string {
	session, ok := os.LookupEnv("XDG_SESSION_ID")
	if ok && len(session) == 0 {
		sessionBytes, err := os.ReadFile("/proc/self/sessionid")
		if err != nil {
			session = ""
		} else {
			session = string(sessionBytes)
		}
	}

	return session
}
