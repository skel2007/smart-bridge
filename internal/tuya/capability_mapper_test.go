package tuya

import (
	"testing"

	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/stretchr/testify/require"
)

func TestMapCapability(t *testing.T) {
	tests := []struct {
		name     string
		function tuyaFunctionSpec
		state    []byte
		want     devices.Capability
		wantOK   bool
	}{
		{
			name:     "on off",
			function: tuyaFunctionSpec{Code: "switch_led"},
			state:    []byte(`true`),
			want:     devices.NewOnOffCapability(devices.CapabilityInstancePower, true),
			wantOK:   true,
		},
		{
			name: "brightness",
			function: tuyaFunctionSpec{
				Code:   "bright_value_v2",
				Values: []byte(`"{\"min\":10,\"max\":1000,\"scale\":0,\"step\":1}"`),
			},
			state: []byte(`1000`),
			want: devices.NewRangeCapability(
				devices.CapabilityInstanceBrightness,
				100,
				devices.RangeParameters{Min: 0, Max: 100, Precision: 1},
			),
			wantOK: true,
		},
		{
			name: "brightness rounded to precision",
			function: tuyaFunctionSpec{
				Code:   "bright_value_v2",
				Values: []byte(`"{\"min\":10,\"max\":1000,\"scale\":0,\"step\":1}"`),
			},
			state: []byte(`20`),
			want: devices.NewRangeCapability(
				devices.CapabilityInstanceBrightness,
				1,
				devices.RangeParameters{Min: 0, Max: 100, Precision: 1},
			),
			wantOK: true,
		},
		{
			name: "brightness middle",
			function: tuyaFunctionSpec{
				Code:   "bright_value_v2",
				Values: []byte(`"{\"min\":10,\"max\":1000,\"scale\":0,\"step\":1}"`),
			},
			state: []byte(`505`),
			want: devices.NewRangeCapability(
				devices.CapabilityInstanceBrightness,
				50,
				devices.RangeParameters{Min: 0, Max: 100, Precision: 1},
			),
			wantOK: true,
		},
		{
			name: "color temperature level",
			function: tuyaFunctionSpec{
				Code:   "temp_value_v2",
				Values: []byte(`"{\"min\":0,\"max\":1000,\"scale\":0,\"step\":1}"`),
			},
			state: []byte(`500`),
			want: devices.NewRangeCapability(
				devices.CapabilityInstanceColorTemperatureLevel,
				50,
				devices.RangeParameters{Min: 0, Max: 100, Precision: 1},
			),
			wantOK: true,
		},
		{
			name: "range with invalid tuya values",
			function: tuyaFunctionSpec{
				Code: "bright_value_v2",
			},
			state: []byte(`1000`),
			want: devices.NewRangeCapabilityWithoutState(
				devices.CapabilityInstanceBrightness,
				devices.RangeParameters{Min: 0, Max: 100, Precision: 1},
			),
			wantOK: true,
		},
		{
			name:     "color",
			function: tuyaFunctionSpec{Code: "colour_data_v2"},
			state:    []byte(`"{\"h\":120,\"s\":800,\"v\":900}"`),
			want: devices.NewColorCapability(devices.CapabilityInstanceColor, devices.HSVColor{
				Hue:        120,
				Saturation: 80,
				Value:      90,
			}),
			wantOK: true,
		},
		{
			name: "mode",
			function: tuyaFunctionSpec{
				Code:   "work_mode",
				Values: []byte(`{"range":["white","colour"]}`),
			},
			state: []byte(`"white"`),
			want: devices.NewModeCapability(
				devices.CapabilityInstanceWorkMode,
				"white",
				devices.ModeParameters{Modes: []string{"white", "colour"}},
			),
			wantOK: true,
		},
		{
			name:     "known function without state",
			function: tuyaFunctionSpec{Code: "switch"},
			want:     devices.NewOnOffCapabilityWithoutState(devices.CapabilityInstancePower),
			wantOK:   true,
		},
		{
			name:     "unknown",
			function: tuyaFunctionSpec{Code: "unsupported_code"},
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capability, ok := mapCapability(tt.function, tt.state)

			require.Equal(t, tt.wantOK, ok)
			require.Equal(t, tt.want, capability)
		})
	}
}

func TestMapCapabilitiesUsesPreferredFunctions(t *testing.T) {
	specifications := tuyaDeviceSpecifications{Functions: []tuyaFunctionSpec{
		{
			Code:   "bright_value",
			Values: []byte(`{"min":0,"max":255,"scale":0,"step":1}`),
		},
		{
			Code:   "bright_value_v2",
			Values: []byte(`{"min":10,"max":1000,"scale":0,"step":1}`),
		},
	}}
	status := []tuyaDeviceStatus{
		{Code: "bright_value", Value: []byte(`255`)},
		{Code: "bright_value_v2", Value: []byte(`505`)},
	}

	require.Equal(t, []devices.Capability{
		devices.NewRangeCapability(
			devices.CapabilityInstanceBrightness,
			50,
			devices.RangeParameters{Min: 0, Max: 100, Precision: 1},
		),
	}, mapCapabilities(specifications, status))
}
