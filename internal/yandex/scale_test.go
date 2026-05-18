package yandex

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapColorTemperatureLevelToKelvin(t *testing.T) {
	tests := []struct {
		name  string
		level float64
		want  int
	}{
		{
			name:  "minimum",
			level: 0,
			want:  2700,
		},
		{
			name:  "middle",
			level: 50,
			want:  4600,
		},
		{
			name:  "maximum",
			level: 100,
			want:  6500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, mapColorTemperatureLevelToKelvin(tt.level))
		})
	}
}
