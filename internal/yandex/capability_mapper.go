package yandex

import (
	"math"

	"github.com/skel2007/smart-bridge/internal/devices"
)

const (
	capabilityTypeOnOff        = "devices.capabilities.on_off"
	capabilityTypeRange        = "devices.capabilities.range"
	capabilityTypeColorSetting = "devices.capabilities.color_setting"

	capabilityInstanceOn           = "on"
	capabilityInstanceBrightness   = "brightness"
	capabilityInstanceHSV          = "hsv"
	capabilityInstanceTemperatureK = "temperature_k"

	unitPercent = "unit.percent"
)

func mapCapabilityDescriptions(capabilities []devices.Capability) []CapabilityDescription {
	descriptions := make([]CapabilityDescription, 0, len(capabilities))
	colorDescription := ColorSettingParameters{}

	for _, capability := range capabilities {
		switch capability.Instance {
		case devices.CapabilityInstancePower:
			descriptions = append(descriptions, newCapabilityDescription(capabilityTypeOnOff, nil))
		case devices.CapabilityInstanceBrightness:
			descriptions = append(descriptions, newCapabilityDescription(
				capabilityTypeRange,
				RangeParameters{
					Instance: capabilityInstanceBrightness,
					Unit:     unitPercent,
					Range:    mapRange(capability.Range),
				},
			))
		case devices.CapabilityInstanceColor:
			colorDescription.ColorModel = capabilityInstanceHSV
		case devices.CapabilityInstanceColorTemperatureLevel:
			colorDescription.Temperature = &TemperatureKRange{
				Min: colorTemperatureMinK,
				Max: colorTemperatureMaxK,
			}
		}
	}

	if colorDescription.ColorModel != "" || colorDescription.Temperature != nil {
		descriptions = append(descriptions, newCapabilityDescription(capabilityTypeColorSetting, colorDescription))
	}

	return descriptions
}

func mapCapabilityStates(capabilities []devices.Capability) []CapabilityState {
	states := make([]CapabilityState, 0, len(capabilities))

	for _, capability := range capabilities {
		switch capability.Instance {
		case devices.CapabilityInstancePower:
			if capability.OnOff != nil && capability.OnOff.State != nil {
				states = append(states, CapabilityState{
					Type: capabilityTypeOnOff,
					State: CapabilityStateValue{
						Instance: capabilityInstanceOn,
						Value:    *capability.OnOff.State,
					},
				})
			}
		case devices.CapabilityInstanceBrightness:
			if capability.Range != nil && capability.Range.State != nil {
				states = append(states, CapabilityState{
					Type: capabilityTypeRange,
					State: CapabilityStateValue{
						Instance: capabilityInstanceBrightness,
						Value:    *capability.Range.State,
					},
				})
			}
		case devices.CapabilityInstanceColor:
			if capability.Color != nil && capability.Color.State != nil {
				states = append(states, CapabilityState{
					Type: capabilityTypeColorSetting,
					State: CapabilityStateValue{
						Instance: capabilityInstanceHSV,
						Value: HSVValue{
							H: int(math.Round(capability.Color.State.Hue)),
							S: int(math.Round(capability.Color.State.Saturation)),
							V: int(math.Round(capability.Color.State.Value)),
						},
					},
				})
			}
		case devices.CapabilityInstanceColorTemperatureLevel:
			if capability.Range != nil && capability.Range.State != nil {
				states = append(states, CapabilityState{
					Type: capabilityTypeColorSetting,
					State: CapabilityStateValue{
						Instance: capabilityInstanceTemperatureK,
						Value:    mapColorTemperatureLevelToKelvin(*capability.Range.State),
					},
				})
			}
		}
	}

	return states
}

func newCapabilityDescription(capabilityType string, parameters any) CapabilityDescription {
	return CapabilityDescription{
		Type:        capabilityType,
		Retrievable: true,
		Reportable:  false,
		Parameters:  parameters,
	}
}

func mapRange(capability *devices.RangeCapability) *ValueRange {
	if capability == nil {
		return nil
	}

	return &ValueRange{
		Min:       capability.Parameters.Min,
		Max:       capability.Parameters.Max,
		Precision: capability.Parameters.Precision,
	}
}
