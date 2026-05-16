package tuya

import (
	"encoding/json"

	"github.com/skel2007/smart-bridge/internal/devices"
)

type tuyaTokenResult struct {
	AccessToken string `json:"access_token"`
}

type tuyaDevice struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CustomName string `json:"customName"`
	Category   string `json:"category"`
	IsOnline   bool   `json:"isOnline"`
}

type tuyaDeviceSpecifications struct {
	Category  string             `json:"category"`
	Functions []tuyaFunctionSpec `json:"functions"`
	Status    []tuyaStatusSpec   `json:"status"`
}

type tuyaFunctionSpec struct {
	Code   string          `json:"code"`
	Type   string          `json:"type"`
	Values json.RawMessage `json:"values"`
}

type tuyaStatusSpec struct {
	Code   string          `json:"code"`
	Type   string          `json:"type"`
	Values json.RawMessage `json:"values"`
}

type tuyaDeviceStatus struct {
	Code  string          `json:"code"`
	Value json.RawMessage `json:"value"`
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
