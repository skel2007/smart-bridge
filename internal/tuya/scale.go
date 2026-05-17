package tuya

import "math"

const (
	domainPercentMin = 0
	domainPercentMax = 100
)

func scaleTuyaRangePercent(value float64, minValue float64, maxValue float64) float64 {
	if maxValue <= minValue {
		return value
	}

	return scaleRange(value, minValue, maxValue, domainPercentMin, domainPercentMax)
}

func scaleDomainPercentToTuyaRange(value float64, minValue float64, maxValue float64) float64 {
	if maxValue <= minValue {
		return value
	}

	return scaleRange(value, domainPercentMin, domainPercentMax, minValue, maxValue)
}

func scaleTuyaColorPercent(value float64, maxValue float64) float64 {
	if maxValue <= 0 {
		return value
	}

	return scaleRange(value, 0, maxValue, domainPercentMin, domainPercentMax)
}

func scaleDomainPercentToTuyaColor(value float64, maxValue float64) float64 {
	if maxValue <= 0 {
		return value
	}

	return math.Round(scaleRange(value, domainPercentMin, domainPercentMax, 0, maxValue))
}

func scaleRange(value float64, inputMin float64, inputMax float64, outputMin float64, outputMax float64) float64 {
	if inputMax <= inputMin {
		return value
	}

	return outputMin + (value-inputMin)/(inputMax-inputMin)*(outputMax-outputMin)
}

func roundToPrecision(value float64, precision float64) float64 {
	if precision <= 0 {
		return value
	}

	return math.Round(value/precision) * precision
}
