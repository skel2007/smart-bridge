package tuya

import (
	"context"
	"errors"

	"github.com/skel2007/smart-bridge/internal/devices"
)

type Credentials struct {
	Endpoint     string
	ClientID     string
	ClientSecret string
}

type Gateway struct {
	api tuyaAPI
}

type tuyaAPI interface {
	listProjectDevices(ctx context.Context) ([]tuyaDevice, error)
	getDeviceSpecifications(ctx context.Context, deviceID string) (tuyaDeviceSpecifications, error)
	getDeviceStatus(ctx context.Context, deviceID string) ([]tuyaDeviceStatus, error)
	sendCommands(ctx context.Context, deviceID string, commands []tuyaCommand) error
}

func NewGateway(credentials Credentials) *Gateway {
	return newGateway(newAPI(credentials))
}

func newGateway(api tuyaAPI) *Gateway {
	return &Gateway{api: api}
}

func (gateway *Gateway) ListDevices(ctx context.Context) ([]devices.Device, error) {
	result, err := gateway.api.listProjectDevices(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]devices.Device, 0, len(result))
	for _, device := range result {
		out = append(out, mapDevice(device))
	}

	return out, nil
}

func (gateway *Gateway) ListCapabilities(ctx context.Context, deviceID string) ([]devices.Capability, error) {
	specifications, err := gateway.api.getDeviceSpecifications(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	status, err := gateway.api.getDeviceStatus(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	return mapCapabilities(specifications, status), nil
}

func (gateway *Gateway) SendCommands(ctx context.Context, deviceID string, commands []devices.CapabilityCommand) error {
	if len(commands) == 0 {
		return errors.New("capability commands are required")
	}

	specifications, err := gateway.api.getDeviceSpecifications(ctx, deviceID)
	if err != nil {
		return err
	}

	mappedCommands := make([]tuyaCommand, 0, len(commands))
	for _, command := range commands {
		mappedCommand, err := mapCapabilityCommand(command, specifications)
		if err != nil {
			return err
		}

		mappedCommands = append(mappedCommands, mappedCommand)
	}

	return gateway.api.sendCommands(ctx, deviceID, mappedCommands)
}
