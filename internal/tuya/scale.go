package tuya

import (
	"math"

	"github.com/skel2007/smart-bridge/internal/devices"
)

func scaleTuyaRangePercent(value float64, minValue float64, maxValue float64) float64 {
	return devices.ScaleRangeToPercent(value, minValue, maxValue)
}

func scaleDomainPercentToTuyaRange(value float64, minValue float64, maxValue float64) float64 {
	return devices.ScalePercentToRange(value, minValue, maxValue)
}

func scaleTuyaColorPercent(value float64, maxValue float64) float64 {
	if maxValue <= 0 {
		return value
	}

	return devices.ScaleRangeToPercent(value, 0, maxValue)
}

func scaleDomainPercentToTuyaColor(value float64, maxValue float64) float64 {
	if maxValue <= 0 {
		return value
	}

	return math.Round(devices.ScalePercentToRange(value, 0, maxValue))
}

func roundTuyaIntegerStep(value float64, values tuyaIntegerValues) float64 {
	step := values.Step
	if step <= 0 {
		step = 1
	}

	rounded := values.Min + math.Round((value-values.Min)/step)*step
	if rounded < values.Min {
		return values.Min
	}
	if rounded > values.Max {
		return values.Max
	}

	return math.Round(rounded)
}

func roundToPrecision(value float64, precision float64) float64 {
	if precision <= 0 {
		return value
	}

	return math.Round(value/precision) * precision
}
