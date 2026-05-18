package tuya

import (
	"fmt"

	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/skel2007/smart-bridge/internal/tuya/internal/cloud"
)

func mapCapabilityCommand(command devices.CapabilityCommand, specifications cloud.DeviceSpecifications) (cloud.Command, error) {
	if err := command.Validate(); err != nil {
		return cloud.Command{}, err
	}

	function, ok := findCommandFunction(command.Instance, specifications.Functions)
	if !ok {
		return cloud.Command{}, fmt.Errorf("tuya function not found for capability instance: %s", command.Instance)
	}

	value, err := mapCommandValue(command, function)
	if err != nil {
		return cloud.Command{}, err
	}

	return cloud.Command{
		Code:  function.Code,
		Value: value,
	}, nil
}

func findCommandFunction(instance devices.CapabilityInstance, functions []cloud.FunctionSpec) (cloud.FunctionSpec, bool) {
	functionsByInstance := selectCapabilityFunctionsByInstance(functions)
	function, ok := functionsByInstance[instance]
	if !ok {
		return cloud.FunctionSpec{}, false
	}

	return function, true
}

func mapCommandValue(command devices.CapabilityCommand, function cloud.FunctionSpec) (any, error) {
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

	return int(roundToPrecision(scaleDomainPercentToTuyaRange(value, tuyaValues.Min, tuyaValues.Max), 1)), nil
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
