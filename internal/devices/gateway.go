package devices

import "context"

type DeviceGateway interface {
	ListDevices(ctx context.Context) ([]Device, error)
	ListCapabilities(ctx context.Context, deviceID string) ([]Capability, error)
	SendCommands(ctx context.Context, deviceID string, commands []CapabilityCommand) error
}
