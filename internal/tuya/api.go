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
)

const (
	tokenPath                = "/v1.0/token"
	projectDevices           = "/v2.0/cloud/thing/device"
	deviceSpecificationsPath = "/v1.0/devices/%s/specifications"
	deviceStatusPath         = "/v1.0/devices/%s/status"
	deviceCommandsPath       = "/v1.0/devices/%s/commands"
	listPageSize             = 20
)

type api struct {
	endpoint     string
	clientID     string
	clientSecret string
	httpClient   *http.Client
	now          func() time.Time
	nonce        func() (string, error)
	accessToken  string
}

func newAPI(credentials Credentials) *api {
	return &api{
		endpoint:     strings.TrimRight(credentials.Endpoint, "/"),
		clientID:     credentials.ClientID,
		clientSecret: credentials.ClientSecret,
		httpClient:   http.DefaultClient,
		now:          time.Now,
		nonce:        randomNonce,
	}
}

func (api *api) listProjectDevices(ctx context.Context, lastID string) ([]tuyaDevice, error) {
	if err := api.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("page_size", strconv.Itoa(listPageSize))
	if lastID != "" {
		query.Set("last_id", lastID)
	}

	var result []tuyaDevice
	if err := api.do(ctx, http.MethodGet, projectDevices, query, nil, api.accessToken, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (api *api) getDeviceSpecifications(ctx context.Context, deviceID string) (tuyaDeviceSpecifications, error) {
	if err := api.ensureAccessToken(ctx); err != nil {
		return tuyaDeviceSpecifications{}, err
	}

	path := fmt.Sprintf(deviceSpecificationsPath, url.PathEscape(deviceID))

	var result tuyaDeviceSpecifications
	if err := api.do(ctx, http.MethodGet, path, nil, nil, api.accessToken, &result); err != nil {
		return tuyaDeviceSpecifications{}, err
	}

	return result, nil
}

func (api *api) getDeviceStatus(ctx context.Context, deviceID string) ([]tuyaDeviceStatus, error) {
	if err := api.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(deviceStatusPath, url.PathEscape(deviceID))

	var result []tuyaDeviceStatus
	if err := api.do(ctx, http.MethodGet, path, nil, nil, api.accessToken, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (api *api) sendCommands(ctx context.Context, deviceID string, commands []tuyaCommand) error {
	if err := api.ensureAccessToken(ctx); err != nil {
		return err
	}

	body, err := json.Marshal(tuyaCommandsRequest{Commands: commands})
	if err != nil {
		return fmt.Errorf("encode tuya commands request: %w", err)
	}

	path := fmt.Sprintf(deviceCommandsPath, url.PathEscape(deviceID))

	return api.do(ctx, http.MethodPost, path, nil, body, api.accessToken, nil)
}

func (api *api) ensureAccessToken(ctx context.Context) error {
	if api.accessToken != "" {
		return nil
	}

	query := url.Values{}
	query.Set("grant_type", "1")

	var result tuyaTokenResult
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
