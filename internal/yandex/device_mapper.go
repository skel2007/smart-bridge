package yandex

import "github.com/skel2007/smart-bridge/internal/devices"

const (
	deviceTypeLight  = "devices.types.light"
	deviceTypeSocket = "devices.types.socket"
	deviceTypeSwitch = "devices.types.switch"
	deviceTypeOther  = "devices.types.other"
)

func mapDeviceDescription(device devices.Device, capabilities []devices.Capability) DeviceDescription {
	return DeviceDescription{
		ID:           device.ID,
		Name:         device.Name,
		StatusInfo:   StatusInfo{Reportable: false},
		Type:         mapDeviceType(device.Type),
		Capabilities: mapCapabilityDescriptions(capabilities),
	}
}

func mapDeviceState(deviceID string, capabilities []devices.Capability) DeviceState {
	return DeviceState{
		ID:           deviceID,
		Capabilities: mapCapabilityStates(capabilities),
	}
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
