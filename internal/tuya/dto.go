package tuya

import "github.com/skel2007/smart-bridge/internal/devices"

type tokenResult struct {
	AccessToken string `json:"access_token"`
}

type tuyaDevice struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CustomName string `json:"customName"`
	Category   string `json:"category"`
	IsOnline   bool   `json:"isOnline"`
}

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
