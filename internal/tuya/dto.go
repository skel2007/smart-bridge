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
		ID:       device.ID,
		Name:     name,
		Category: device.Category,
		Online:   device.IsOnline,
	}
}
