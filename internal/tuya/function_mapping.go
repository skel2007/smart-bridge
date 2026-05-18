package tuya

import (
	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/skel2007/smart-bridge/internal/tuya/internal/cloud"
)

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

func selectFunctionsByInstance(functions []cloud.FunctionSpec) map[devices.CapabilityInstance]cloud.FunctionSpec {
	functionByCode := make(map[string]cloud.FunctionSpec, len(functions))
	for _, function := range functions {
		functionByCode[function.Code] = function
	}

	selected := make(map[devices.CapabilityInstance]cloud.FunctionSpec)
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
