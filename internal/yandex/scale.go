package yandex

import "github.com/skel2007/smart-bridge/internal/devices"

const (
	colorTemperatureMinK = 2700
	colorTemperatureMaxK = 6500
)

func mapColorTemperatureLevelToKelvin(level float64) int {
	return int(devices.ScalePercentToRange(level, colorTemperatureMinK, colorTemperatureMaxK))
}
