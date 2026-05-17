package devices

import "context"

type DeviceGateway interface {
	ListDevices(ctx context.Context) ([]Device, error)
	ListCapabilities(ctx context.Context, deviceID string) ([]Capability, error)
	SendCommand(ctx context.Context, deviceID string, command CapabilityCommand) error
}
