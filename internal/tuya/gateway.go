package tuya

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/skel2007/smart-bridge/internal/devices"
)

type Credentials struct {
	Endpoint     string
	ClientID     string
	ClientSecret string
}

type Gateway struct {
	api *api
}

type Option func(*Gateway)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(gateway *Gateway) {
		if httpClient != nil {
			gateway.api.httpClient = httpClient
		}
	}
}

func WithNowFunc(now func() time.Time) Option {
	return func(gateway *Gateway) {
		if now != nil {
			gateway.api.now = now
		}
	}
}

func WithNonceFunc(nonce func() (string, error)) Option {
	return func(gateway *Gateway) {
		if nonce != nil {
			gateway.api.nonce = nonce
		}
	}
}

func NewGateway(credentials Credentials, options ...Option) *Gateway {
	gateway := &Gateway{api: newAPI(credentials)}
	for _, option := range options {
		option(gateway)
	}

	return gateway
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
