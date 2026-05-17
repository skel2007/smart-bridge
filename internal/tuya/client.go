package tuya

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/skel2007/smart-bridge/internal/devices"
)

type Credentials struct {
	Endpoint     string
	ClientID     string
	ClientSecret string
}

type Client struct {
	api *api
}

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		if httpClient != nil {
			client.api.httpClient = httpClient
		}
	}
}

func WithNowFunc(now func() time.Time) Option {
	return func(client *Client) {
		if now != nil {
			client.api.now = now
		}
	}
}

func WithNonceFunc(nonce func() (string, error)) Option {
	return func(client *Client) {
		if nonce != nil {
			client.api.nonce = nonce
		}
	}
}

func NewClient(credentials Credentials, options ...Option) *Client {
	client := &Client{api: newAPI(credentials)}
	for _, option := range options {
		option(client)
	}

	return client
}

func (client *Client) ListDevices(ctx context.Context) ([]devices.Device, error) {
	var out []devices.Device
	var lastID string

	for {
		result, err := client.api.listProjectDevices(ctx, lastID)
		if err != nil {
			return nil, err
		}

		if len(result) == 0 {
			return out, nil
		}

		for _, device := range result {
			out = append(out, mapDevice(device))
		}

		if len(result) < listPageSize {
			return out, nil
		}

		lastID = result[len(result)-1].ID
		if lastID == "" {
			return nil, fmt.Errorf("tuya device list response missing id for pagination")
		}
	}
}

func (client *Client) ListCapabilities(ctx context.Context, deviceID string) ([]devices.Capability, error) {
	specifications, err := client.api.getDeviceSpecifications(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	status, err := client.api.getDeviceStatus(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	return mapCapabilities(specifications, status), nil
}

func (client *Client) SendCommands(ctx context.Context, deviceID string, commands []devices.CapabilityCommand) error {
	if len(commands) == 0 {
		return errors.New("capability commands are required")
	}

	specifications, err := client.api.getDeviceSpecifications(ctx, deviceID)
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

	return client.api.sendCommands(ctx, deviceID, mappedCommands)
}
