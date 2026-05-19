package yandex

import (
	"encoding/json"
	"testing"

	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/stretchr/testify/require"
)

func TestMapCapabilityDescriptionsMergesColorSettingParameters(t *testing.T) {
	descriptions := mapCapabilityDescriptions([]devices.Capability{
		devices.NewColorCapabilityWithoutState(devices.CapabilityInstanceColor),
		devices.NewRangeCapabilityWithoutState(devices.CapabilityInstanceColorTemperatureLevel, devices.RangeParameters{
			Min:       0,
			Max:       100,
			Precision: 1,
		}),
	})

	data, err := json.Marshal(descriptions)

	require.NoError(t, err)
	require.JSONEq(t, `[
		{
			"type": "devices.capabilities.color_setting",
			"retrievable": true,
			"reportable": false,
			"parameters": {
				"color_model": "hsv",
				"temperature_k": {
					"min": 2700,
					"max": 6500
				}
			}
		}
	]`, string(data))
}

func TestMapCapabilityDescriptionsSkipsUnsupportedCapabilities(t *testing.T) {
	descriptions := mapCapabilityDescriptions([]devices.Capability{
		devices.NewModeCapabilityWithoutState(devices.CapabilityInstanceWorkMode, devices.ModeParameters{
			Modes: []string{"white", "colour"},
		}),
		{
			Type:     devices.CapabilityTypeRange,
			Instance: "unknown",
		},
	})

	require.Empty(t, descriptions)
}

func TestMapDeviceState(t *testing.T) {
	state := mapDeviceState("light-1", []devices.Capability{
		devices.NewOnOffCapability(devices.CapabilityInstancePower, true),
		devices.NewRangeCapability(devices.CapabilityInstanceBrightness, 42, devices.RangeParameters{
			Min:       0,
			Max:       100,
			Precision: 1,
		}),
		devices.NewColorCapability(devices.CapabilityInstanceColor, devices.HSVColor{
			Hue:        120.4,
			Saturation: 80.5,
			Value:      90,
		}),
		devices.NewRangeCapability(devices.CapabilityInstanceColorTemperatureLevel, 50, devices.RangeParameters{
			Min:       0,
			Max:       100,
			Precision: 1,
		}),
	})

	data, err := json.Marshal(state)

	require.NoError(t, err)
	require.JSONEq(t, `{
		"id": "light-1",
		"capabilities": [
			{
				"type": "devices.capabilities.on_off",
				"state": {
					"instance": "on",
					"value": true
				}
			},
			{
				"type": "devices.capabilities.range",
				"state": {
					"instance": "brightness",
					"value": 42
				}
			},
			{
				"type": "devices.capabilities.color_setting",
				"state": {
					"instance": "hsv",
					"value": {
						"h": 120,
						"s": 81,
						"v": 90
					}
				}
			},
			{
				"type": "devices.capabilities.color_setting",
				"state": {
					"instance": "temperature_k",
					"value": 4600
				}
			}
		]
	}`, string(data))
}

func TestMapCapabilityStatesSkipsMissingAndUnsupportedStates(t *testing.T) {
	states := mapCapabilityStates([]devices.Capability{
		devices.NewOnOffCapabilityWithoutState(devices.CapabilityInstancePower),
		devices.NewRangeCapabilityWithoutState(devices.CapabilityInstanceBrightness, devices.RangeParameters{
			Min:       0,
			Max:       100,
			Precision: 1,
		}),
		devices.NewColorCapabilityWithoutState(devices.CapabilityInstanceColor),
		devices.NewModeCapability(devices.CapabilityInstanceWorkMode, "white", devices.ModeParameters{
			Modes: []string{"white"},
		}),
	})

	require.Empty(t, states)
}
