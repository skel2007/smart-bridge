package tuya

import "github.com/skel2007/smart-bridge/internal/devices"

type functionMapping struct {
	instance devices.CapabilityInstance
	codes    []string
}

var functionMappings = []functionMapping{
	{instance: devices.CapabilityInstancePower, codes: []string{"switch_led", "switch"}},
	{instance: devices.CapabilityInstanceWorkMode, codes: []string{"work_mode"}},
	{instance: devices.CapabilityInstanceBrightness, codes: []string{"bright_value_v2", "bright_value"}},
	{instance: devices.CapabilityInstanceColorTemperatureLevel, codes: []string{"temp_value_v2", "temp_value"}},
	{instance: devices.CapabilityInstanceColor, codes: []string{"colour_data_v2", "colour_data"}},
}

func selectFunctionsByInstance(functions []tuyaFunctionSpec) map[devices.CapabilityInstance]tuyaFunctionSpec {
	functionByCode := make(map[string]tuyaFunctionSpec, len(functions))
	for _, function := range functions {
		functionByCode[function.Code] = function
	}

	selected := make(map[devices.CapabilityInstance]tuyaFunctionSpec)
	for _, mapping := range functionMappings {
		for _, code := range mapping.codes {
			function, ok := functionByCode[code]
			if !ok {
				continue
			}

			selected[mapping.instance] = function
			break
		}
	}

	return selected
}
