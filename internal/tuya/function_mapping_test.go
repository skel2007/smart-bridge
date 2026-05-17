package tuya

import (
	"testing"

	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/stretchr/testify/require"
)

func TestSelectFunctionsByInstance(t *testing.T) {
	tests := []struct {
		name      string
		functions []tuyaFunctionSpec
		want      map[devices.CapabilityInstance]tuyaFunctionSpec
	}{
		{
			name: "prefers v2 over legacy",
			functions: []tuyaFunctionSpec{
				{Code: "bright_value"},
				{Code: "bright_value_v2"},
			},
			want: map[devices.CapabilityInstance]tuyaFunctionSpec{
				devices.CapabilityInstanceBrightness: {Code: "bright_value_v2"},
			},
		},
		{
			name: "falls back to legacy",
			functions: []tuyaFunctionSpec{
				{Code: "temp_value"},
			},
			want: map[devices.CapabilityInstance]tuyaFunctionSpec{
				devices.CapabilityInstanceColorTemperatureLevel: {Code: "temp_value"},
			},
		},
		{
			name: "selects multiple instances",
			functions: []tuyaFunctionSpec{
				{Code: "switch"},
				{Code: "work_mode"},
				{Code: "colour_data"},
				{Code: "colour_data_v2"},
			},
			want: map[devices.CapabilityInstance]tuyaFunctionSpec{
				devices.CapabilityInstancePower:    {Code: "switch"},
				devices.CapabilityInstanceWorkMode: {Code: "work_mode"},
				devices.CapabilityInstanceColor:    {Code: "colour_data_v2"},
			},
		},
		{
			name: "ignores unknown functions",
			functions: []tuyaFunctionSpec{
				{Code: "unknown"},
			},
			want: map[devices.CapabilityInstance]tuyaFunctionSpec{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, selectFunctionsByInstance(tt.functions))
		})
	}
}
