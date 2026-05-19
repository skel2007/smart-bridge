package tuya

import (
	"context"
	"errors"
	"log/slog"

	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/skel2007/smart-bridge/internal/tuya/internal/cloud"
)

type Credentials struct {
	Endpoint     string
	ClientID     string
	ClientSecret string
}

type Option func(*gatewayOptions)

type gatewayOptions struct {
	logger *slog.Logger
}

func WithLogger(logger *slog.Logger) Option {
	return func(options *gatewayOptions) {
		if logger != nil {
			options.logger = logger
		}
	}
}

type Gateway struct {
	api cloudAPI
}

type cloudAPI interface {
	ListProjectDevices(ctx context.Context) ([]cloud.Device, error)
	GetDeviceSpecifications(ctx context.Context, deviceID string) (cloud.DeviceSpecifications, error)
	GetDeviceStatus(ctx context.Context, deviceID string) ([]cloud.DeviceStatus, error)
	SendCommands(ctx context.Context, deviceID string, commands []cloud.Command) error
}

func NewGateway(credentials Credentials, options ...Option) *Gateway {
	var gatewayOptions gatewayOptions
	for _, option := range options {
		option(&gatewayOptions)
	}

	return newGateway(cloud.NewAPI(cloud.Credentials{
		Endpoint:     credentials.Endpoint,
		ClientID:     credentials.ClientID,
		ClientSecret: credentials.ClientSecret,
	}, gatewayOptions.logger))
}

func newGateway(api cloudAPI) *Gateway {
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

	mappedCommands := make([]cloud.Command, 0, len(commands))
	for _, command := range commands {
		mappedCommand, err := mapCapabilityCommand(command, specifications)
		if err != nil {
			return err
		}

		mappedCommands = append(mappedCommands, mappedCommand)
	}

	return gateway.api.SendCommands(ctx, deviceID, mappedCommands)
}
