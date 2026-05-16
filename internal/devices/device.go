package devices

type Device struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	Type   DeviceType `json:"type"`
	Online bool       `json:"online"`
}

type DeviceType string

const (
	DeviceTypeLight  DeviceType = "light"
	DeviceTypeSocket DeviceType = "socket"
	DeviceTypeSwitch DeviceType = "switch"
	DeviceTypeOther  DeviceType = "other"
)
