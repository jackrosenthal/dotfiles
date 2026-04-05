package main

import (
	"fmt"
	"log"
	"math"

	"github.com/godbus/dbus/v5"
)

const (
	batteryObjPath = dbus.ObjectPath("/org/freedesktop/UPower/devices/battery_BAT0")
	upowerDest     = "org.freedesktop.UPower"
	deviceIface    = "org.freedesktop.UPower.Device"
	propsIface     = "org.freedesktop.DBus.Properties"
	notifyDest     = "org.freedesktop.Notifications"
	notifyObjPath  = dbus.ObjectPath("/org/freedesktop/Notifications")
)

const stateDischarging uint32 = 2

type batteryState struct {
	percentage    float64
	state         uint32
	notifiedLevel int
	lastNotifyID  uint32
}

func (bs *batteryState) readProps(conn *dbus.Conn) error {
	obj := conn.Object(upowerDest, batteryObjPath)

	var pct float64
	if err := obj.Call(propsIface+".Get", 0, deviceIface, "Percentage").Store(&pct); err != nil {
		return fmt.Errorf("get Percentage: %w", err)
	}
	bs.percentage = pct

	var state uint32
	if err := obj.Call(propsIface+".Get", 0, deviceIface, "State").Store(&state); err != nil {
		return fmt.Errorf("get State: %w", err)
	}
	bs.state = state
	return nil
}

func (bs *batteryState) check(notifyObj dbus.BusObject) {
	if bs.state != stateDischarging {
		bs.notifiedLevel = 100
		return
	}

	pct := int(math.Round(bs.percentage))

	type threshold struct {
		level   int
		urgency uint8
		summary string
		body    string
		timeout int32
	}

	thresholds := []threshold{
		{5, 2, "Battery Critical", fmt.Sprintf("%d%% — plug in now!", pct), 0},
		{10, 2, "Battery Low", fmt.Sprintf("%d%% remaining", pct), 0},
		{20, 1, "Battery Low", fmt.Sprintf("%d%% remaining", pct), -1},
	}

	for _, t := range thresholds {
		if pct <= t.level && bs.notifiedLevel > t.level {
			hints := map[string]dbus.Variant{
				"urgency": dbus.MakeVariant(t.urgency),
			}
			var newID uint32
			err := notifyObj.Call(notifyDest+".Notify", 0,
				"battery-notify",
				bs.lastNotifyID,
				"battery-low",
				t.summary,
				t.body,
				[]string{},
				hints,
				t.timeout,
			).Store(&newID)
			if err != nil {
				log.Printf("notify: %v", err)
				return
			}
			bs.lastNotifyID = newID
			bs.notifiedLevel = t.level
			return
		}
	}
}

func main() {
	sysBus, err := dbus.ConnectSystemBus()
	if err != nil {
		log.Fatalf("connect system bus: %v", err)
	}
	defer sysBus.Close()

	sessBus, err := dbus.ConnectSessionBus()
	if err != nil {
		log.Fatalf("connect session bus: %v", err)
	}
	defer sessBus.Close()

	notifyObj := sessBus.Object(notifyDest, notifyObjPath)

	if err := sysBus.AddMatchSignal(
		dbus.WithMatchObjectPath(batteryObjPath),
		dbus.WithMatchInterface(propsIface),
		dbus.WithMatchMember("PropertiesChanged"),
	); err != nil {
		log.Fatalf("add match: %v", err)
	}

	signals := make(chan *dbus.Signal, 10)
	sysBus.Signal(signals)

	bs := &batteryState{notifiedLevel: 100}
	if err := bs.readProps(sysBus); err != nil {
		log.Printf("initial read: %v", err)
	}
	bs.check(notifyObj)

	for sig := range signals {
		if len(sig.Body) < 2 {
			continue
		}
		changedProps, ok := sig.Body[1].(map[string]dbus.Variant)
		if !ok {
			continue
		}

		updated := false
		if v, ok := changedProps["Percentage"]; ok {
			if p, ok := v.Value().(float64); ok {
				bs.percentage = p
				updated = true
			}
		}
		if v, ok := changedProps["State"]; ok {
			if s, ok := v.Value().(uint32); ok {
				bs.state = s
				updated = true
			}
		}
		if updated {
			bs.check(notifyObj)
		}
	}
}
