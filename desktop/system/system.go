package system

import "tractor.dev/toolkit-go/desktop"

type Display struct {
	Name        string
	Size        desktop.Size
	Position    desktop.Position
	ScaleFactor float64
}

type PowerInfo struct {
	IsOnBattery    bool
	IsCharging     bool
	BatteryPercent float64
}

func Displays() []Display {
	return displays()
}

func Power() PowerInfo {
	return power()
}
