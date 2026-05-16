package tuya

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/skel2007/smart-bridge/internal/devices"
)

const (
	tokenPath                = "/v1.0/token"
	projectDevices           = "/v2.0/cloud/thing/device"
	deviceSpecificationsPath = "/v1.0/devices/%s/specifications"
	deviceStatusPath         = "/v1.0/devices/%s/status"
	deviceCommandsPath       = "/v1.0/devices/%s/commands"
	listPageSize             = 20
)

type Credentials struct {
	Endpoint     string
	ClientID     string
	ClientSecret string
}

type Client struct {
	endpoint     string
	clientID     string
	clientSecret string
	httpClient   *http.Client
	now          func() time.Time
	nonce        func() (string, error)
	accessToken  string
}

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		if httpClient != nil {
			client.httpClient = httpClient
		}
	}
}

func WithNowFunc(now func() time.Time) Option {
	return func(client *Client) {
		if now != nil {
			client.now = now
		}
	}
}

func WithNonceFunc(nonce func() (string, error)) Option {
	return func(client *Client) {
		if nonce != nil {
			client.nonce = nonce
		}
	}
}

func NewClient(credentials Credentials, options ...Option) *Client {
	client := &Client{
		endpoint:     strings.TrimRight(credentials.Endpoint, "/"),
		clientID:     credentials.ClientID,
		clientSecret: credentials.ClientSecret,
		httpClient:   http.DefaultClient,
		now:          time.Now,
		nonce:        randomNonce,
	}

	for _, option := range options {
		option(client)
	}

	return client
}

func (client *Client) ListDevices(ctx context.Context) ([]devices.Device, error) {
	if err := client.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	var out []devices.Device
	var lastID string

	for {
		query := url.Values{}
		query.Set("page_size", strconv.Itoa(listPageSize))
		if lastID != "" {
			query.Set("last_id", lastID)
		}

		var result []tuyaDevice
		if err := client.do(ctx, http.MethodGet, projectDevices, query, nil, client.accessToken, &result); err != nil {
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
	specifications, err := client.getDeviceSpecifications(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	status, err := client.getDeviceStatus(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	return mapCapabilities(specifications, status), nil
}

func (client *Client) SendCommand(ctx context.Context, deviceID string, command devices.CapabilityCommand) error {
	specifications, err := client.getDeviceSpecifications(ctx, deviceID)
	if err != nil {
		return err
	}

	mappedCommand, err := mapCapabilityCommand(command, specifications)
	if err != nil {
		return err
	}

	body, err := json.Marshal(tuyaCommandsRequest{Commands: []tuyaCommand{mappedCommand}})
	if err != nil {
		return fmt.Errorf("encode tuya commands request: %w", err)
	}

	path := fmt.Sprintf(deviceCommandsPath, url.PathEscape(deviceID))

	return client.do(ctx, http.MethodPost, path, nil, body, client.accessToken, nil)
}

func (client *Client) ensureAccessToken(ctx context.Context) error {
	if client.accessToken != "" {
		return nil
	}

	query := url.Values{}
	query.Set("grant_type", "1")

	var result tuyaTokenResult
	if err := client.do(ctx, http.MethodGet, tokenPath, query, nil, "", &result); err != nil {
		return err
	}
	if result.AccessToken == "" {
		return fmt.Errorf("tuya token response missing access_token")
	}

	client.accessToken = result.AccessToken

	return nil
}

func (client *Client) getDeviceSpecifications(ctx context.Context, deviceID string) (tuyaDeviceSpecifications, error) {
	if err := client.ensureAccessToken(ctx); err != nil {
		return tuyaDeviceSpecifications{}, err
	}

	path := fmt.Sprintf(deviceSpecificationsPath, url.PathEscape(deviceID))

	var result tuyaDeviceSpecifications
	if err := client.do(ctx, http.MethodGet, path, nil, nil, client.accessToken, &result); err != nil {
		return tuyaDeviceSpecifications{}, err
	}

	return result, nil
}

func (client *Client) getDeviceStatus(ctx context.Context, deviceID string) ([]tuyaDeviceStatus, error) {
	if err := client.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(deviceStatusPath, url.PathEscape(deviceID))

	var result []tuyaDeviceStatus
	if err := client.do(ctx, http.MethodGet, path, nil, nil, client.accessToken, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func randomNonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
