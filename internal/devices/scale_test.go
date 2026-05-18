package devices

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScaleRangeToPercent(t *testing.T) {
	require.Equal(t, 0.0, ScaleRangeToPercent(10, 10, 1000))
	require.Equal(t, 50.0, ScaleRangeToPercent(505, 10, 1000))
	require.Equal(t, 100.0, ScaleRangeToPercent(1000, 10, 1000))
	require.Equal(t, 50.0, ScaleRangeToPercent(50, 1000, 10))
}

func TestScalePercentToRange(t *testing.T) {
	require.Equal(t, 10.0, ScalePercentToRange(0, 10, 1000))
	require.Equal(t, 505.0, ScalePercentToRange(50, 10, 1000))
	require.Equal(t, 1000.0, ScalePercentToRange(100, 10, 1000))
	require.Equal(t, 50.0, ScalePercentToRange(50, 1000, 10))
}
