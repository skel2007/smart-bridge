package yandex

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/stretchr/testify/require"
)

func TestMapDeviceActionCommands(t *testing.T) {
	action := DeviceAction{
		ID: "light-1",
		Capabilities: []CapabilityAction{
			newTestCapabilityAction(capabilityTypeOnOff, capabilityInstanceOn, true),
			newTestCapabilityAction(capabilityTypeRange, capabilityInstanceBrightness, 42),
		},
	}

	commands, err := mapDeviceActionCommands(action)

	require.NoError(t, err)
	require.Equal(t, []devices.CapabilityCommand{
		devices.NewOnOffCommand(devices.CapabilityInstancePower, true),
		devices.NewRangeCommand(devices.CapabilityInstanceBrightness, 42),
	}, commands)
}

func TestMapCapabilityActionCommand(t *testing.T) {
	tests := []struct {
		name   string
		action CapabilityAction
		want   devices.CapabilityCommand
	}{
		{
			name:   "power",
			action: newTestCapabilityAction(capabilityTypeOnOff, capabilityInstanceOn, true),
			want:   devices.NewOnOffCommand(devices.CapabilityInstancePower, true),
		},
		{
			name:   "brightness",
			action: newTestCapabilityAction(capabilityTypeRange, capabilityInstanceBrightness, 42),
			want:   devices.NewRangeCommand(devices.CapabilityInstanceBrightness, 42),
		},
		{
			name: "hsv",
			action: newTestCapabilityAction(capabilityTypeColorSetting, capabilityInstanceHSV, HSVValue{
				H: 120,
				S: 80,
				V: 90,
			}),
			want: devices.NewColorCommand(devices.CapabilityInstanceColor, devices.HSVColor{
				Hue:        120,
				Saturation: 80,
				Value:      90,
			}),
		},
		{
			name:   "temperature k",
			action: newTestCapabilityAction(capabilityTypeColorSetting, capabilityInstanceTemperatureK, 4600),
			want:   devices.NewRangeCommand(devices.CapabilityInstanceColorTemperatureLevel, 50),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := mapCapabilityActionCommand(tt.action)

			require.NoError(t, err)
			require.Equal(t, tt.want, command)
		})
	}
}

func TestMapCapabilityActionCommandInvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		action CapabilityAction
	}{
		{
			name:   "power value type",
			action: newTestCapabilityAction(capabilityTypeOnOff, capabilityInstanceOn, "true"),
		},
		{
			name:   "brightness outside domain range",
			action: newTestCapabilityAction(capabilityTypeRange, capabilityInstanceBrightness, 101),
		},
		{
			name: "hsv outside domain range",
			action: newTestCapabilityAction(capabilityTypeColorSetting, capabilityInstanceHSV, HSVValue{
				H: 120,
				S: 101,
				V: 90,
			}),
		},
		{
			name:   "hsv missing component",
			action: newTestCapabilityAction(capabilityTypeColorSetting, capabilityInstanceHSV, map[string]int{"h": 120, "s": 80}),
		},
		{
			name:   "temperature outside yandex range",
			action: newTestCapabilityAction(capabilityTypeColorSetting, capabilityInstanceTemperatureK, 7000),
		},
		{
			name:   "null value",
			action: newTestCapabilityAction(capabilityTypeOnOff, capabilityInstanceOn, nil),
		},
		{
			name: "missing value",
			action: CapabilityAction{
				Type: capabilityTypeRange,
				State: CapabilityActionState{
					Instance: capabilityInstanceBrightness,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mapCapabilityActionCommand(tt.action)

			mappingErr, ok := errors.AsType[actionMappingError](err)
			require.True(t, ok)
			require.Equal(t, errorCodeInvalidValue, mappingErr.Code)
		})
	}
}

func TestMapCapabilityActionCommandUnsupportedActions(t *testing.T) {
	tests := []struct {
		name   string
		action CapabilityAction
	}{
		{
			name:   "unsupported type",
			action: newTestCapabilityAction("devices.capabilities.mode", "cleanup_mode", "auto"),
		},
		{
			name:   "unsupported on off instance",
			action: newTestCapabilityAction(capabilityTypeOnOff, "mute", true),
		},
		{
			name:   "unsupported range instance",
			action: newTestCapabilityAction(capabilityTypeRange, "volume", 42),
		},
		{
			name: "relative brightness",
			action: CapabilityAction{
				Type: capabilityTypeRange,
				State: CapabilityActionState{
					Instance: capabilityInstanceBrightness,
					Value:    json.RawMessage(`10`),
					Relative: new(true),
				},
			},
		},
		{
			name:   "unsupported color setting instance",
			action: newTestCapabilityAction(capabilityTypeColorSetting, "rgb", map[string]int{"r": 255}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mapCapabilityActionCommand(tt.action)

			mappingErr, ok := errors.AsType[actionMappingError](err)
			require.True(t, ok)
			require.Equal(t, errorCodeNotSupportedInCurrentMode, mappingErr.Code)
		})
	}
}

func TestMapDeviceActionCommandsStopsAtFirstMappingError(t *testing.T) {
	action := DeviceAction{
		ID: "light-1",
		Capabilities: []CapabilityAction{
			newTestCapabilityAction(capabilityTypeOnOff, capabilityInstanceOn, true),
			newTestCapabilityAction(capabilityTypeRange, "volume", 42),
		},
	}

	commands, err := mapDeviceActionCommands(action)

	require.Nil(t, commands)
	mappingErr, ok := errors.AsType[actionMappingError](err)
	require.True(t, ok)
	require.Equal(t, errorCodeNotSupportedInCurrentMode, mappingErr.Code)
}

func newTestCapabilityAction(capabilityType string, instance string, value any) CapabilityAction {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	return CapabilityAction{
		Type: capabilityType,
		State: CapabilityActionState{
			Instance: instance,
			Value:    data,
		},
	}
}
