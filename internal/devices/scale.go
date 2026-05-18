package devices

const (
	PercentMin = 0
	PercentMax = 100
)

func ScaleRangeToPercent(value float64, minValue float64, maxValue float64) float64 {
	if maxValue <= minValue {
		return value
	}

	return scaleRange(value, minValue, maxValue, PercentMin, PercentMax)
}

func ScalePercentToRange(value float64, minValue float64, maxValue float64) float64 {
	if maxValue <= minValue {
		return value
	}

	return scaleRange(value, PercentMin, PercentMax, minValue, maxValue)
}

func scaleRange(value float64, inputMin float64, inputMax float64, outputMin float64, outputMax float64) float64 {
	return outputMin + (value-inputMin)/(inputMax-inputMin)*(outputMax-outputMin)
}
