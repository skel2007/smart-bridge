package yandex

import (
	"math"

	"github.com/skel2007/smart-bridge/internal/devices"
)

const (
	deviceTypeLight  = "devices.types.light"
	deviceTypeSocket = "devices.types.socket"
	deviceTypeSwitch = "devices.types.switch"
	deviceTypeOther  = "devices.types.other"

	capabilityTypeOnOff        = "devices.capabilities.on_off"
	capabilityTypeRange        = "devices.capabilities.range"
	capabilityTypeColorSetting = "devices.capabilities.color_setting"

	capabilityInstanceOn           = "on"
	capabilityInstanceBrightness   = "brightness"
	capabilityInstanceHSV          = "hsv"
	capabilityInstanceTemperatureK = "temperature_k"

	unitPercent = "unit.percent"
)

func MapDeviceDescription(device devices.Device, capabilities []devices.Capability) DeviceDescription {
	return DeviceDescription{
		ID:           device.ID,
		Name:         device.Name,
		StatusInfo:   StatusInfo{Reportable: false},
		Type:         mapDeviceType(device.Type),
		Capabilities: MapCapabilityDescriptions(capabilities),
	}
}

func MapDeviceState(deviceID string, capabilities []devices.Capability) DeviceState {
	return DeviceState{
		ID:           deviceID,
		Capabilities: MapCapabilityStates(capabilities),
	}
}

func MapCapabilityDescriptions(capabilities []devices.Capability) []CapabilityDescription {
	descriptions := make([]CapabilityDescription, 0, len(capabilities))
	colorDescription := ColorSettingParameters{}

	for _, capability := range capabilities {
		switch capability.Instance {
		case devices.CapabilityInstancePower:
			descriptions = append(descriptions, CapabilityDescription{
				Type:        capabilityTypeOnOff,
				Retrievable: true,
				Reportable:  false,
			})
		case devices.CapabilityInstanceBrightness:
			descriptions = append(descriptions, CapabilityDescription{
				Type:        capabilityTypeRange,
				Retrievable: true,
				Reportable:  false,
				Parameters: RangeParameters{
					Instance: capabilityInstanceBrightness,
					Unit:     unitPercent,
					Range:    mapRange(capability.Range),
				},
			})
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
		descriptions = append(descriptions, CapabilityDescription{
			Type:        capabilityTypeColorSetting,
			Retrievable: true,
			Reportable:  false,
			Parameters:  colorDescription,
		})
	}

	return descriptions
}

func MapCapabilityStates(capabilities []devices.Capability) []CapabilityState {
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

func mapDeviceType(deviceType devices.DeviceType) string {
	switch deviceType {
	case devices.DeviceTypeLight:
		return deviceTypeLight
	case devices.DeviceTypeSocket:
		return deviceTypeSocket
	case devices.DeviceTypeSwitch:
		return deviceTypeSwitch
	default:
		return deviceTypeOther
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
