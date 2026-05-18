package cloud

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

	"github.com/hashicorp/go-retryablehttp"
)

const (
	tokenPath                = "/v1.0/token"
	projectDevices           = "/v2.0/cloud/thing/device"
	deviceSpecificationsPath = "/v1.0/devices/%s/specifications"
	deviceStatusPath         = "/v1.0/devices/%s/status"
	deviceCommandsPath       = "/v1.0/devices/%s/commands"
	listPageSize             = 20
)

const defaultHTTPTimeout = 10 * time.Second

type Credentials struct {
	Endpoint     string
	ClientID     string
	ClientSecret string
}

type API struct {
	credentials Credentials
	transport   *retryablehttp.Client
	now         func() time.Time
	nonce       func() (string, error)
	accessToken string
}

func NewAPI(credentials Credentials) *API {
	api := &API{
		credentials: credentials,
		transport:   retryablehttp.NewClient(),
		now:         time.Now,
		nonce:       randomNonce,
	}

	api.credentials.Endpoint = strings.TrimRight(api.credentials.Endpoint, "/")
	api.transport.HTTPClient.Timeout = defaultHTTPTimeout
	api.transport.Logger = nil
	api.transport.ErrorHandler = retryablehttp.PassthroughErrorHandler
	api.transport.PrepareRetry = api.signRequest

	return api
}

func (api *API) ListProjectDevices(ctx context.Context) ([]Device, error) {
	var out []Device
	var lastID string

	for {
		result, err := api.listProjectDevicesPage(ctx, lastID)
		if err != nil {
			return nil, err
		}

		if len(result) == 0 {
			return out, nil
		}

		out = append(out, result...)

		if len(result) < listPageSize {
			return out, nil
		}

		lastID = result[len(result)-1].ID
		if lastID == "" {
			return nil, fmt.Errorf("tuya device list response missing id for pagination")
		}
	}
}

func (api *API) listProjectDevicesPage(ctx context.Context, lastID string) ([]Device, error) {
	if err := api.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("page_size", strconv.Itoa(listPageSize))
	if lastID != "" {
		query.Set("last_id", lastID)
	}

	var result []Device
	if err := api.do(ctx, http.MethodGet, projectDevices, query, nil, api.accessToken, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (api *API) GetDeviceSpecifications(ctx context.Context, deviceID string) (DeviceSpecifications, error) {
	if err := api.ensureAccessToken(ctx); err != nil {
		return DeviceSpecifications{}, err
	}

	path := fmt.Sprintf(deviceSpecificationsPath, url.PathEscape(deviceID))

	var result DeviceSpecifications
	if err := api.do(ctx, http.MethodGet, path, nil, nil, api.accessToken, &result); err != nil {
		return DeviceSpecifications{}, err
	}

	return result, nil
}

func (api *API) GetDeviceStatus(ctx context.Context, deviceID string) ([]DeviceStatus, error) {
	if err := api.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(deviceStatusPath, url.PathEscape(deviceID))

	var result []DeviceStatus
	if err := api.do(ctx, http.MethodGet, path, nil, nil, api.accessToken, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (api *API) SendCommands(ctx context.Context, deviceID string, commands []Command) error {
	if err := api.ensureAccessToken(ctx); err != nil {
		return err
	}

	body, err := json.Marshal(commandsRequest{Commands: commands})
	if err != nil {
		return fmt.Errorf("encode tuya commands request: %w", err)
	}

	path := fmt.Sprintf(deviceCommandsPath, url.PathEscape(deviceID))

	return api.do(ctx, http.MethodPost, path, nil, body, api.accessToken, nil)
}

func (api *API) ensureAccessToken(ctx context.Context) error {
	if api.accessToken != "" {
		return nil
	}

	query := url.Values{}
	query.Set("grant_type", "1")

	var result tokenResult
	if err := api.do(ctx, http.MethodGet, tokenPath, query, nil, "", &result); err != nil {
		return err
	}
	if result.AccessToken == "" {
		return fmt.Errorf("tuya token response missing access_token")
	}

	api.accessToken = result.AccessToken

	return nil
}

func randomNonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
