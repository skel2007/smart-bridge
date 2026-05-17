package tuya

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScaleTuyaRangePercent(t *testing.T) {
	require.Equal(t, 0.0, scaleTuyaRangePercent(10, 10, 1000))
	require.Equal(t, 50.0, scaleTuyaRangePercent(505, 10, 1000))
	require.Equal(t, 100.0, scaleTuyaRangePercent(1000, 10, 1000))
	require.Equal(t, 50.0, scaleTuyaRangePercent(50, 1000, 10))
}

func TestScaleDomainPercentToTuyaRange(t *testing.T) {
	require.Equal(t, 10.0, scaleDomainPercentToTuyaRange(0, 10, 1000))
	require.Equal(t, 505.0, scaleDomainPercentToTuyaRange(50, 10, 1000))
	require.Equal(t, 1000.0, scaleDomainPercentToTuyaRange(100, 10, 1000))
	require.Equal(t, 50.0, scaleDomainPercentToTuyaRange(50, 1000, 10))
}

func TestScaleTuyaColorPercent(t *testing.T) {
	require.Equal(t, 80.0, scaleTuyaColorPercent(800, 1000))
	require.InDelta(t, 50.196, scaleTuyaColorPercent(128, 255), 0.001)
	require.Equal(t, 50.0, scaleTuyaColorPercent(50, 0))
}

func TestScaleDomainPercentToTuyaColor(t *testing.T) {
	require.Equal(t, 800.0, scaleDomainPercentToTuyaColor(80, 1000))
	require.Equal(t, 128.0, scaleDomainPercentToTuyaColor(50, 255))
	require.Equal(t, 50.0, scaleDomainPercentToTuyaColor(50, 0))
}

func TestRoundToPrecision(t *testing.T) {
	require.Equal(t, 1.0, roundToPrecision(1.0101, 1))
	require.Equal(t, 1.5, roundToPrecision(1.49, 0.5))
	require.Equal(t, 1.49, roundToPrecision(1.49, 0))
}
