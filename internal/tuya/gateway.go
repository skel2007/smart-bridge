package tuya

import (
	"context"
	"errors"

	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/skel2007/smart-bridge/internal/tuya/internal/cloud"
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
	ListProjectDevices(ctx context.Context) ([]tuyaDevice, error)
	GetDeviceSpecifications(ctx context.Context, deviceID string) (tuyaDeviceSpecifications, error)
	GetDeviceStatus(ctx context.Context, deviceID string) ([]tuyaDeviceStatus, error)
	SendCommands(ctx context.Context, deviceID string, commands []tuyaCommand) error
}

func NewGateway(credentials Credentials) *Gateway {
	return newGateway(cloud.NewAPI(cloud.Credentials{
		Endpoint:     credentials.Endpoint,
		ClientID:     credentials.ClientID,
		ClientSecret: credentials.ClientSecret,
	}))
}

func newGateway(api tuyaAPI) *Gateway {
	return &Gateway{api: api}
}

func (gateway *Gateway) ListDevices(ctx context.Context) ([]devices.Device, error) {
	result, err := gateway.api.ListProjectDevices(ctx)
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
	specifications, err := gateway.api.GetDeviceSpecifications(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	status, err := gateway.api.GetDeviceStatus(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	return mapCapabilities(specifications, status), nil
}

func (gateway *Gateway) SendCommands(ctx context.Context, deviceID string, commands []devices.CapabilityCommand) error {
	if len(commands) == 0 {
		return errors.New("capability commands are required")
	}

	specifications, err := gateway.api.GetDeviceSpecifications(ctx, deviceID)
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

	return gateway.api.SendCommands(ctx, deviceID, mappedCommands)
}
