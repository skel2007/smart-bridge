package tuya

import (
	"encoding/json"
	"math"

	"github.com/skel2007/smart-bridge/internal/devices"
)

func mapDevice(device tuyaDevice) devices.Device {
	name := device.Name
	if device.CustomName != "" {
		name = device.CustomName
	}

	return devices.Device{
		ID:     device.ID,
		Name:   name,
		Type:   mapDeviceType(device.Category),
		Online: device.IsOnline,
	}
}

func mapDeviceType(category string) devices.DeviceType {
	switch category {
	case "dj", "xdd", "fwd", "dc", "dd", "gyd", "fsd", "tyndj", "tgq":
		return devices.DeviceTypeLight
	case "cz", "pc":
		return devices.DeviceTypeSocket
	case "kg", "cjkg", "ckqdkg", "clkg", "tgkg":
		return devices.DeviceTypeSwitch
	default:
		return devices.DeviceTypeOther
	}
}

func mapCapabilities(specifications tuyaDeviceSpecifications, status []tuyaDeviceStatus) []devices.Capability {
	statusByCode := make(map[string]json.RawMessage, len(status))
	for _, item := range status {
		statusByCode[item.Code] = item.Value
	}

	capabilities := make([]devices.Capability, 0, len(specifications.Functions))
	seen := make(map[devices.CapabilityInstance]bool)
	for _, function := range specifications.Functions {
		capability, ok := mapCapability(function, statusByCode[function.Code])
		if !ok || seen[capability.Instance] {
			continue
		}

		seen[capability.Instance] = true
		capabilities = append(capabilities, capability)
	}

	return capabilities
}

func mapCapability(function tuyaFunctionSpec, state json.RawMessage) (devices.Capability, bool) {
	switch function.Code {
	case "switch", "switch_led":
		return mapOnOffCapability(state), true
	case "bright_value", "bright_value_v2":
		return mapRangeCapability(devices.CapabilityInstanceBrightness, function.Values, state), true
	case "temp_value", "temp_value_v2":
		return mapRangeCapability(devices.CapabilityInstanceColorTemperatureLevel, function.Values, state), true
	case "colour_data":
		return mapColorCapability(state, 255), true
	case "colour_data_v2":
		return mapColorCapability(state, 1000), true
	case "work_mode":
		return mapModeCapability(function.Values, state), true
	default:
		return devices.Capability{}, false
	}
}

func mapOnOffCapability(state json.RawMessage) devices.Capability {
	var value bool
	if decodeRawJSON(state, &value) {
		return devices.NewOnOffCapability(devices.CapabilityInstancePower, value)
	}

	return devices.NewOnOffCapabilityWithoutState(devices.CapabilityInstancePower)
}

func mapRangeCapability(instance devices.CapabilityInstance, values json.RawMessage, state json.RawMessage) devices.Capability {
	var tuyaValues tuyaIntegerValues
	decodeTuyaValues(values, &tuyaValues)

	parameters := devices.RangeParameters{
		Min:       scaleTuyaNumber(tuyaValues.Min, tuyaValues.Scale),
		Max:       scaleTuyaNumber(tuyaValues.Max, tuyaValues.Scale),
		Precision: scaleTuyaNumber(tuyaValues.Step, tuyaValues.Scale),
	}

	var value float64
	if decodeRawJSON(state, &value) {
		return devices.NewRangeCapability(instance, scaleTuyaNumber(value, tuyaValues.Scale), parameters)
	}

	return devices.NewRangeCapabilityWithoutState(instance, parameters)
}

func mapColorCapability(state json.RawMessage, maxSaturationValue float64) devices.Capability {
	var value tuyaHSVValue
	if decodeTuyaValues(state, &value) {
		return devices.NewColorCapability(devices.CapabilityInstanceColor, devices.HSVColor{
			Hue:        value.Hue,
			Saturation: scaleTuyaColorPercent(value.Saturation, maxSaturationValue),
			Value:      scaleTuyaColorPercent(value.Value, maxSaturationValue),
		})
	}

	return devices.NewColorCapabilityWithoutState(devices.CapabilityInstanceColor)
}

func mapModeCapability(values json.RawMessage, state json.RawMessage) devices.Capability {
	var tuyaValues tuyaEnumValues
	decodeTuyaValues(values, &tuyaValues)

	parameters := devices.ModeParameters{Modes: tuyaValues.Range}

	var value string
	if decodeRawJSON(state, &value) {
		return devices.NewModeCapability(devices.CapabilityInstanceWorkMode, value, parameters)
	}

	return devices.NewModeCapabilityWithoutState(devices.CapabilityInstanceWorkMode, parameters)
}

func decodeTuyaValues(raw json.RawMessage, out any) bool {
	if decodeRawJSON(raw, out) {
		return true
	}

	var text string
	if !decodeRawJSON(raw, &text) || text == "" {
		return false
	}

	return json.Unmarshal([]byte(text), out) == nil
}

func decodeRawJSON(raw json.RawMessage, out any) bool {
	if len(raw) == 0 {
		return false
	}

	return json.Unmarshal(raw, out) == nil
}

func scaleTuyaNumber(value float64, scale float64) float64 {
	if scale <= 0 {
		return value
	}

	return value / math.Pow(10, scale)
}

func scaleTuyaColorPercent(value float64, maxValue float64) float64 {
	if maxValue <= 0 {
		return value
	}

	return value / maxValue * 100
}

type tuyaIntegerValues struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Scale float64 `json:"scale"`
	Step  float64 `json:"step"`
}

type tuyaEnumValues struct {
	Range []string `json:"range"`
}

type tuyaHSVValue struct {
	Hue        float64 `json:"h"`
	Saturation float64 `json:"s"`
	Value      float64 `json:"v"`
}
