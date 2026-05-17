package tuya

import (
	"testing"

	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/stretchr/testify/require"
)

func TestMapDeviceType(t *testing.T) {
	tests := []struct {
		name     string
		category string
		want     devices.DeviceType
	}{
		{name: "light", category: "dj", want: devices.DeviceTypeLight},
		{name: "socket", category: "cz", want: devices.DeviceTypeSocket},
		{name: "power strip", category: "pc", want: devices.DeviceTypeSocket},
		{name: "switch", category: "kg", want: devices.DeviceTypeSwitch},
		{name: "unknown", category: "unknown", want: devices.DeviceTypeOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, mapDeviceType(tt.category))
		})
	}
}
