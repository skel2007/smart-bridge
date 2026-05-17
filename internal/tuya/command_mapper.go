package tuya

import (
	"fmt"
	"math"

	"github.com/skel2007/smart-bridge/internal/devices"
)

func mapCapabilityCommand(command devices.CapabilityCommand, specifications tuyaDeviceSpecifications) (tuyaCommand, error) {
	if err := command.Validate(); err != nil {
		return tuyaCommand{}, err
	}

	function, ok := findCommandFunction(command.Instance, specifications.Functions)
	if !ok {
		return tuyaCommand{}, fmt.Errorf("tuya function not found for capability instance: %s", command.Instance)
	}

	value, err := mapCommandValue(command, function)
	if err != nil {
		return tuyaCommand{}, err
	}

	return tuyaCommand{
		Code:  function.Code,
		Value: value,
	}, nil
}

func findCommandFunction(instance devices.CapabilityInstance, functions []tuyaFunctionSpec) (tuyaFunctionSpec, bool) {
	for _, function := range functions {
		if functionMatchesInstance(function.Code, instance) {
			return function, true
		}
	}

	return tuyaFunctionSpec{}, false
}

func functionMatchesInstance(code string, instance devices.CapabilityInstance) bool {
	switch instance {
	case devices.CapabilityInstancePower:
		return code == "switch" || code == "switch_led"
	case devices.CapabilityInstanceBrightness:
		return code == "bright_value" || code == "bright_value_v2"
	case devices.CapabilityInstanceColorTemperatureLevel:
		return code == "temp_value" || code == "temp_value_v2"
	case devices.CapabilityInstanceColor:
		return code == "colour_data" || code == "colour_data_v2"
	case devices.CapabilityInstanceWorkMode:
		return code == "work_mode"
	default:
		return false
	}
}

func mapCommandValue(command devices.CapabilityCommand, function tuyaFunctionSpec) (any, error) {
	switch command.Instance {
	case devices.CapabilityInstancePower:
		return command.OnOff.State, nil
	case devices.CapabilityInstanceBrightness, devices.CapabilityInstanceColorTemperatureLevel:
		return mapRangeCommandValue(command.Range.State, function.Values)
	case devices.CapabilityInstanceColor:
		return mapColorCommandValue(command.Color.State, function.Code), nil
	case devices.CapabilityInstanceWorkMode:
		return mapModeCommandValue(command.Mode.State, function.Values)
	default:
		return nil, fmt.Errorf("unsupported capability command instance: %s", command.Instance)
	}
}

func mapRangeCommandValue(value float64, rawValues []byte) (int, error) {
	var tuyaValues tuyaIntegerValues
	if !decodeTuyaValues(rawValues, &tuyaValues) || !tuyaValues.validRange() {
		return 0, fmt.Errorf("tuya range values are missing or invalid")
	}

	return int(math.Round(scaleDomainPercentToTuyaRange(value, tuyaValues.Min, tuyaValues.Max))), nil
}

func scaleDomainPercentToTuyaRange(value float64, minValue float64, maxValue float64) float64 {
	if maxValue <= minValue {
		return value
	}

	return minValue + value/100*(maxValue-minValue)
}

func mapColorCommandValue(value devices.HSVColor, code string) tuyaHSVValue {
	maxSaturationValue := 255.0
	if code == "colour_data_v2" {
		maxSaturationValue = 1000
	}

	return tuyaHSVValue{
		Hue:        value.Hue,
		Saturation: scaleDomainPercentToTuyaColor(value.Saturation, maxSaturationValue),
		Value:      scaleDomainPercentToTuyaColor(value.Value, maxSaturationValue),
	}
}

func scaleDomainPercentToTuyaColor(value float64, maxValue float64) float64 {
	if maxValue <= 0 {
		return value
	}

	return math.Round(value / 100 * maxValue)
}

func mapModeCommandValue(value string, rawValues []byte) (string, error) {
	var tuyaValues tuyaEnumValues
	if !decodeTuyaValues(rawValues, &tuyaValues) {
		return "", fmt.Errorf("tuya mode values are missing or invalid")
	}

	for _, mode := range tuyaValues.Range {
		if mode == value {
			return value, nil
		}
	}

	return "", fmt.Errorf("tuya mode value is not supported: %s", value)
}
