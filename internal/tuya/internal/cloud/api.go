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
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	tokenPath                = "/v1.0/token"
	refreshTokenPath         = "/v1.0/token/%s"
	projectDevices           = "/v2.0/cloud/thing/device"
	deviceSpecificationsPath = "/v1.0/devices/%s/specifications"
	deviceStatusPath         = "/v1.0/devices/%s/status"
	deviceCommandsPath       = "/v1.0/devices/%s/commands"
	listPageSize             = 20
)

const defaultHTTPTimeout = 10 * time.Second
const tokenRefreshMargin = time.Minute

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
	tokenMu     sync.Mutex
	token       tokenState
}

type tokenState struct {
	accessToken  string
	refreshToken string
	expiresAt    time.Time
}

func (token tokenState) valid(now time.Time) bool {
	return token.accessToken != "" && now.Add(tokenRefreshMargin).Before(token.expiresAt)
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
	accessToken, err := api.ensureAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("page_size", strconv.Itoa(listPageSize))
	if lastID != "" {
		query.Set("last_id", lastID)
	}

	var result []Device
	if err := api.do(ctx, http.MethodGet, projectDevices, query, nil, accessToken, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (api *API) GetDeviceSpecifications(ctx context.Context, deviceID string) (DeviceSpecifications, error) {
	accessToken, err := api.ensureAccessToken(ctx)
	if err != nil {
		return DeviceSpecifications{}, err
	}

	path := fmt.Sprintf(deviceSpecificationsPath, url.PathEscape(deviceID))

	var result DeviceSpecifications
	if err := api.do(ctx, http.MethodGet, path, nil, nil, accessToken, &result); err != nil {
		return DeviceSpecifications{}, err
	}

	return result, nil
}

func (api *API) GetDeviceStatus(ctx context.Context, deviceID string) ([]DeviceStatus, error) {
	accessToken, err := api.ensureAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf(deviceStatusPath, url.PathEscape(deviceID))

	var result []DeviceStatus
	if err := api.do(ctx, http.MethodGet, path, nil, nil, accessToken, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (api *API) SendCommands(ctx context.Context, deviceID string, commands []Command) error {
	accessToken, err := api.ensureAccessToken(ctx)
	if err != nil {
		return err
	}

	body, err := json.Marshal(commandsRequest{Commands: commands})
	if err != nil {
		return fmt.Errorf("encode tuya commands request: %w", err)
	}

	path := fmt.Sprintf(deviceCommandsPath, url.PathEscape(deviceID))

	return api.do(ctx, http.MethodPost, path, nil, body, accessToken, nil)
}

func (api *API) ensureAccessToken(ctx context.Context) (string, error) {
	api.tokenMu.Lock()
	defer api.tokenMu.Unlock()

	now := api.now()
	if api.token.valid(now) {
		return api.token.accessToken, nil
	}

	if api.token.refreshToken != "" {
		// If refresh fails but the context is still active, fall back to a fresh token request below.
		if err := api.refreshAccessToken(ctx, now); err == nil {
			return api.token.accessToken, nil
		} else if ctx.Err() != nil {
			return "", err
		}
	}

	if err := api.requestAccessToken(ctx, now); err != nil {
		return "", err
	}

	return api.token.accessToken, nil
}

func (api *API) requestAccessToken(ctx context.Context, now time.Time) error {
	query := url.Values{}
	query.Set("grant_type", "1")

	var result tokenResult
	if err := api.do(ctx, http.MethodGet, tokenPath, query, nil, "", &result); err != nil {
		return err
	}

	return api.storeToken(result, now)
}

func (api *API) refreshAccessToken(ctx context.Context, now time.Time) error {
	path := fmt.Sprintf(refreshTokenPath, url.PathEscape(api.token.refreshToken))

	var result tokenResult
	if err := api.do(ctx, http.MethodGet, path, nil, nil, "", &result); err != nil {
		return err
	}

	return api.storeToken(result, now)
}

func (api *API) storeToken(result tokenResult, now time.Time) error {
	if result.AccessToken == "" {
		return fmt.Errorf("tuya token response missing access_token")
	}
	if result.ExpireTime <= 0 {
		return fmt.Errorf("tuya token response missing expire_time")
	}

	api.token = tokenState{
		accessToken:  result.AccessToken,
		refreshToken: result.RefreshToken,
		expiresAt:    now.Add(time.Duration(result.ExpireTime) * time.Second),
	}

	return nil
}

func randomNonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
