package tuya

import (
	"testing"

	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/skel2007/smart-bridge/internal/tuya/internal/cloud"
	"github.com/stretchr/testify/require"
)

func TestMapCapabilityCommand(t *testing.T) {
	tests := []struct {
		name          string
		command       devices.CapabilityCommand
		specification cloud.DeviceSpecifications
		want          cloud.Command
	}{
		{
			name:    "power",
			command: devices.NewOnOffCommand(devices.CapabilityInstancePower, true),
			specification: cloud.DeviceSpecifications{Functions: []cloud.FunctionSpec{
				{Code: "switch_led"},
			}},
			want: cloud.Command{Code: "switch_led", Value: true},
		},
		{
			name:    "brightness",
			command: devices.NewRangeCommand(devices.CapabilityInstanceBrightness, 50),
			specification: cloud.DeviceSpecifications{Functions: []cloud.FunctionSpec{
				{
					Code:   "bright_value_v2",
					Values: []byte(`{"min":10,"max":1000,"scale":0,"step":1}`),
				},
			}},
			want: cloud.Command{Code: "bright_value_v2", Value: 505},
		},
		{
			name:    "brightness snaps to tuya step",
			command: devices.NewRangeCommand(devices.CapabilityInstanceBrightness, 50),
			specification: cloud.DeviceSpecifications{Functions: []cloud.FunctionSpec{
				{
					Code:   "bright_value_v2",
					Values: []byte(`{"min":10,"max":1000,"scale":0,"step":10}`),
				},
			}},
			want: cloud.Command{Code: "bright_value_v2", Value: 510},
		},
		{
			name:    "color temperature level",
			command: devices.NewRangeCommand(devices.CapabilityInstanceColorTemperatureLevel, 75),
			specification: cloud.DeviceSpecifications{Functions: []cloud.FunctionSpec{
				{
					Code:   "temp_value_v2",
					Values: []byte(`"{\"min\":0,\"max\":1000,\"scale\":0,\"step\":1}"`),
				},
			}},
			want: cloud.Command{Code: "temp_value_v2", Value: 750},
		},
		{
			name: "color",
			command: devices.NewColorCommand(devices.CapabilityInstanceColor, devices.HSVColor{
				Hue:        120,
				Saturation: 80,
				Value:      90,
			}),
			specification: cloud.DeviceSpecifications{Functions: []cloud.FunctionSpec{
				{Code: "colour_data_v2"},
			}},
			want: cloud.Command{
				Code: "colour_data_v2",
				Value: tuyaHSVValue{
					Hue:        120,
					Saturation: 800,
					Value:      900,
				},
			},
		},
		{
			name: "legacy color",
			command: devices.NewColorCommand(devices.CapabilityInstanceColor, devices.HSVColor{
				Hue:        37,
				Saturation: 100,
				Value:      50,
			}),
			specification: cloud.DeviceSpecifications{Functions: []cloud.FunctionSpec{
				{Code: "colour_data"},
			}},
			want: cloud.Command{
				Code: "colour_data",
				Value: tuyaHSVValue{
					Hue:        37,
					Saturation: 255,
					Value:      128,
				},
			},
		},
		{
			name:    "mode",
			command: devices.NewModeCommand(devices.CapabilityInstanceWorkMode, "white"),
			specification: cloud.DeviceSpecifications{Functions: []cloud.FunctionSpec{
				{
					Code:   "work_mode",
					Values: []byte(`{"range":["white","colour"]}`),
				},
			}},
			want: cloud.Command{Code: "work_mode", Value: "white"},
		},
		{
			name:    "prefers v2 range function",
			command: devices.NewRangeCommand(devices.CapabilityInstanceBrightness, 50),
			specification: cloud.DeviceSpecifications{Functions: []cloud.FunctionSpec{
				{
					Code:   "bright_value",
					Values: []byte(`{"min":0,"max":255,"scale":0,"step":1}`),
				},
				{
					Code:   "bright_value_v2",
					Values: []byte(`{"min":10,"max":1000,"scale":0,"step":1}`),
				},
			}},
			want: cloud.Command{Code: "bright_value_v2", Value: 505},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapCapabilityCommand(tt.command, tt.specification)

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestMapCapabilityCommandErrors(t *testing.T) {
	tests := []struct {
		name          string
		command       devices.CapabilityCommand
		specification cloud.DeviceSpecifications
		wantErr       string
	}{
		{
			name:          "invalid domain command",
			command:       devices.NewRangeCommand(devices.CapabilityInstanceBrightness, 101),
			specification: cloud.DeviceSpecifications{},
			wantErr:       "range command state must be between 0 and 100: 101",
		},
		{
			name:          "missing tuya function",
			command:       devices.NewOnOffCommand(devices.CapabilityInstancePower, true),
			specification: cloud.DeviceSpecifications{},
			wantErr:       "tuya function not found for capability instance: power",
		},
		{
			name:    "missing range values",
			command: devices.NewRangeCommand(devices.CapabilityInstanceBrightness, 50),
			specification: cloud.DeviceSpecifications{Functions: []cloud.FunctionSpec{
				{Code: "bright_value_v2"},
			}},
			wantErr: "tuya range values are missing or invalid",
		},
		{
			name:    "invalid range values",
			command: devices.NewRangeCommand(devices.CapabilityInstanceBrightness, 50),
			specification: cloud.DeviceSpecifications{Functions: []cloud.FunctionSpec{
				{
					Code:   "bright_value_v2",
					Values: []byte(`{"min":1000,"max":10,"scale":0,"step":1}`),
				},
			}},
			wantErr: "tuya range values are missing or invalid",
		},
		{
			name:    "missing mode values",
			command: devices.NewModeCommand(devices.CapabilityInstanceWorkMode, "white"),
			specification: cloud.DeviceSpecifications{Functions: []cloud.FunctionSpec{
				{Code: "work_mode"},
			}},
			wantErr: "tuya mode values are missing or invalid",
		},
		{
			name:    "unsupported mode",
			command: devices.NewModeCommand(devices.CapabilityInstanceWorkMode, "scene"),
			specification: cloud.DeviceSpecifications{Functions: []cloud.FunctionSpec{
				{
					Code:   "work_mode",
					Values: []byte(`{"range":["white","colour"]}`),
				},
			}},
			wantErr: "tuya mode value is not supported: scene",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mapCapabilityCommand(tt.command, tt.specification)

			require.EqualError(t, err, tt.wantErr)
		})
	}
}
