package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/godbus/dbus"
	"github.com/sqp/pulseaudio"
)

func (cl *Client) notify(message string, hints map[string]interface{}) {
	obj := cl.conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")
	out := obj.Call(
		"org.freedesktop.Notifications.Notify", 0, "Volume",
		cl.notificationReplacementID, "volume", "Volume", message, []string{}, hints, 500)
	out.Store(&cl.notificationReplacementID)
}

type Client struct {
	mtx                       sync.Mutex
	notificationReplacementID uint32
	conn                      *dbus.Conn
}

func (cl *Client) DeviceVolumeUpdated(path dbus.ObjectPath, values []uint32) {
	cl.mtx.Lock()
	defer cl.mtx.Unlock()
	if values[0] == values[1] {
		percent := values[0] / 655
		cl.notify(fmt.Sprintf("%d%%", percent), map[string]interface{}{
			"value": int(percent),
		})
	}
}
func (cl *Client) DeviceMuteUpdated(path dbus.ObjectPath, value bool) {
	cl.mtx.Lock()
	defer cl.mtx.Unlock()
	if value {
		cl.notify("muted", map[string]interface{}{})
	} else {
		cl.notify("unmuted", map[string]interface{}{})
	}
}

func main() {

	pulse, e := pulseaudio.New()
	if e != nil {
		log.Panicln("failed to connect to pulse", e)
	}
	sessionBus, err := dbus.SessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
	}

	client := &Client{conn: sessionBus}
	pulse.Register(client)
	log.Printf("listening for pulse events")
	pulse.Listen()
}
