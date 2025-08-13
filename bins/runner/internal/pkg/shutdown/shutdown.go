package shutdown

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var shutdownDbusArgs = []dbusArgs{
	{
		destination: "org.gnome.SessionManager",
		path:        "/org/gnome/SessionManager",
		iface:       "org.gnome.SessionManager",
		method:      "Shutdown",
		body:        []any{},
	},
	{
		destination: "org.kde.ksmserver",
		path:        "/KSMServer",
		iface:       "org.kde.KSMServerInterface",
		method:      "logout",
		body:        []any{-1, 2, 2},
	},
	{
		destination: "org.xfce.SessionManager",
		path:        "/org/xfce/SessionManager",
		iface:       "org.xfce.SessionManager",
		method:      "Shutdown",
		body:        []any{true},
	},
	{
		destination: "org.freedesktop.login1",
		path:        "/org/freedesktop/login1",
		iface:       "org.freedesktop.login1.Manager",
		method:      "PowerOff",
		body:        []any{true},
	},
	{
		destination: "org.freedesktop.PowerManagement",
		path:        "/org/freedesktop/PowerManagement",
		iface:       "org.freedesktop.PowerManagement",
		method:      "Shutdown",
		body:        []any{},
	},
	{
		destination: "org.freedesktop.SessionManagement",
		path:        "/org/freedesktop/SessionManagement",
		iface:       "org.freedesktop.SessionManagement",
		method:      "Shutdown",
		body:        []any{},
	},
	{
		destination: "org.freedesktop.ConsoleKit",
		path:        "/org/freedesktop/ConsoleKit/Manager",
		iface:       "org.freedesktop.ConsoleKit.Manager",
		method:      "Stop",
		body:        []any{},
	},
	{
		destination: "org.freedesktop.Hal",
		path:        "/org/freedesktop/Hal/devices/computer",
		iface:       "org.freedesktop.Hal.Device.SystemPowerManagement",
		method:      "Shutdown",
		body:        []any{},
	},
	{
		destination: "org.freedesktop.systemd1",
		path:        "/org/freedesktop/systemd1",
		iface:       "org.freedesktop.systemd1.Manager",
		method:      "PowerOff",
		body:        []any{},
	},
}

func Shutdown(ctx context.Context, l *zap.Logger, v *validator.Validate) (err error) {
	l.Info("preparing to shut down")
	for _, v := range shutdownDbusArgs {
		l.Debug(fmt.Sprintf("trying destination: %s", v.destination), zap.String("path", v.path), zap.String("iface", v.iface))
		if reply, err := dbusSend(
			v.destination,
			v.path,
			v.iface,
			v.method,
			v.body,
		); reply && err == nil {
			return nil
		}
	}

	l.Error("none of the methods worked. falling back to shell.")
	err = runCommand(ctx, l, v, "shutdown", "-h", "now")

	l.Error("shell fallback failed - executing shutdown with sudo")
	err = runCommand(ctx, l, v, "sudo", "shutdown", "-h", "now")
	return err
}
