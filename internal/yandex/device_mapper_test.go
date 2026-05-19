package yandex

import (
	"encoding/json"
	"testing"

	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/stretchr/testify/require"
)

func TestMapDeviceDescription(t *testing.T) {
	description := mapDeviceDescription(
		devices.Device{
			ID:     "light-1",
			Name:   "Desk light",
			Type:   devices.DeviceTypeLight,
			Online: true,
		},
		[]devices.Capability{
			devices.NewOnOffCapability(devices.CapabilityInstancePower, true),
			devices.NewRangeCapability(devices.CapabilityInstanceBrightness, 42, devices.RangeParameters{
				Min:       0,
				Max:       100,
				Precision: 1,
			}),
		},
	)

	data, err := json.Marshal(description)

	require.NoError(t, err)
	require.JSONEq(t, `{
		"id": "light-1",
		"name": "Desk light",
		"status_info": {
			"reportable": false
		},
		"type": "devices.types.light",
		"capabilities": [
			{
				"type": "devices.capabilities.on_off",
				"retrievable": true,
				"reportable": false
			},
			{
				"type": "devices.capabilities.range",
				"retrievable": true,
				"reportable": false,
				"parameters": {
					"instance": "brightness",
					"unit": "unit.percent",
					"range": {
						"min": 0,
						"max": 100,
						"precision": 1
					}
				}
			}
		]
	}`, string(data))
}

func TestMapDeviceDescriptionMapsKnownDeviceTypes(t *testing.T) {
	tests := []struct {
		name       string
		deviceType devices.DeviceType
		want       string
	}{
		{
			name:       "light",
			deviceType: devices.DeviceTypeLight,
			want:       "devices.types.light",
		},
		{
			name:       "socket",
			deviceType: devices.DeviceTypeSocket,
			want:       "devices.types.socket",
		},
		{
			name:       "switch",
			deviceType: devices.DeviceTypeSwitch,
			want:       "devices.types.switch",
		},
		{
			name:       "other",
			deviceType: devices.DeviceTypeOther,
			want:       "devices.types.other",
		},
		{
			name:       "unknown",
			deviceType: "unsupported",
			want:       "devices.types.other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			description := mapDeviceDescription(
				devices.Device{ID: "device-1", Name: "Device", Type: tt.deviceType},
				nil,
			)

			require.Equal(t, tt.want, description.Type)
		})
	}
}
