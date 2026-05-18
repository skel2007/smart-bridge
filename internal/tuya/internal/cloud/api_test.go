package cloud

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIListProjectDevicesFetchesAllPages(t *testing.T) {
	nextPageURI := "/v2.0/cloud/thing/device?last_id=dev-20&page_size=20"
	api, recorder := newTestAPI(t,
		get(tokenURI, tuyaResult(`{"access_token":"access-token"}`)),
		get(devicesURI, tuyaResult(tuyaDevicesJSON("dev-", 1, 20))),
		get(nextPageURI, tuyaResult(`[{"id":"dev-21","name":"Device 21","category":"dj","isOnline":true}]`)),
	)

	deviceList, err := api.ListProjectDevices(context.Background())

	require.NoError(t, err)
	require.Len(t, deviceList, 21)
	require.Equal(t, "dev-1", deviceList[0].ID)
	require.Equal(t, "dev-21", deviceList[20].ID)
	require.Equal(t, tokenURI, recorder.request(0).URL.RequestURI())
	require.Equal(t, devicesURI, recorder.request(1).URL.RequestURI())
	require.Equal(t, nextPageURI, recorder.request(2).URL.RequestURI())
}

func TestAPIListProjectDevicesReturnsErrorWhenPageCursorIsMissing(t *testing.T) {
	api, _ := newTestAPI(t,
		get(tokenURI, tuyaResult(`{"access_token":"access-token"}`)),
		get(devicesURI, tuyaResult(tuyaDevicesWithMissingLastIDJSON())),
	)

	_, err := api.ListProjectDevices(context.Background())

	require.EqualError(t, err, "tuya device list response missing id for pagination")
}

func TestAPIListProjectDevicesReturnsErrorWhenTokenIsMissing(t *testing.T) {
	api, _ := newTestAPI(t,
		get(tokenURI, tuyaResult(`{}`)),
	)

	_, err := api.ListProjectDevices(context.Background())

	require.EqualError(t, err, "tuya token response missing access_token")
}

func TestAPIListProjectDevicesDoesNotExposeSecretsInErrors(t *testing.T) {
	api, _ := newTestAPI(t,
		get(tokenURI, tuyaResult(`{"access_token":"access-secret"}`)),
		get(devicesURI, tuyaError(http.StatusOK, "1106", "permission denied")),
	)

	_, err := api.ListProjectDevices(context.Background())

	require.Error(t, err)
	require.NotContains(t, err.Error(), "super-secret")
	require.NotContains(t, err.Error(), "access-secret")
}

func TestAPIGetDeviceSpecificationsEscapesDeviceID(t *testing.T) {
	specificationsURI := "/v1.0/devices/device%2Fid%20with%20space/specifications"
	api, recorder := newTestAPI(t,
		get(tokenURI, tuyaResult(`{"access_token":"access-token"}`)),
		get(specificationsURI, tuyaResult(`{"functions":[]}`)),
	)

	_, err := api.GetDeviceSpecifications(context.Background(), "device/id with space")

	require.NoError(t, err)
	require.Equal(t, tokenURI, recorder.request(0).URL.RequestURI())
	require.Equal(t, specificationsURI, recorder.request(1).URL.RequestURI())
}

func TestAPIGetDeviceStatusEscapesDeviceID(t *testing.T) {
	statusURI := "/v1.0/devices/device%2Fid%20with%20space/status"
	api, recorder := newTestAPI(t,
		get(tokenURI, tuyaResult(`{"access_token":"access-token"}`)),
		get(statusURI, tuyaResult(`[]`)),
	)

	_, err := api.GetDeviceStatus(context.Background(), "device/id with space")

	require.NoError(t, err)
	require.Equal(t, tokenURI, recorder.request(0).URL.RequestURI())
	require.Equal(t, statusURI, recorder.request(1).URL.RequestURI())
}

func TestAPISendCommands(t *testing.T) {
	api, recorder := newTestAPI(t,
		get(tokenURI, tuyaResult(`{"access_token":"access-token"}`)),
		post(deviceCommandsURI, tuyaResult(`true`)),
	)

	err := api.SendCommands(context.Background(), "device-id", []Command{
		{Code: "switch_led", Value: true},
		{Code: "bright_value_v2", Value: 505},
	})

	require.NoError(t, err)
	req := recorder.request(1)
	require.Equal(t, tokenURI, recorder.request(0).URL.RequestURI())
	require.Equal(t, deviceCommandsURI, req.URL.RequestURI())
	require.Equal(t, http.MethodPost, req.Method)
	require.Equal(t, "application/json", req.Header.Get("Content-Type"))
	require.Equal(t, "access-token", req.Header.Get("access_token"))
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.JSONEq(t, `{
		"commands": [
			{"code": "switch_led", "value": true},
			{"code": "bright_value_v2", "value": 505}
		]
	}`, string(body))
}

func TestAPISendCommandsEscapesDeviceID(t *testing.T) {
	commandsURI := "/v1.0/devices/device%2Fid%20with%20space/commands"
	api, recorder := newTestAPI(t,
		get(tokenURI, tuyaResult(`{"access_token":"access-token"}`)),
		post(commandsURI, tuyaResult(`true`)),
	)

	err := api.SendCommands(context.Background(), "device/id with space", []Command{
		{Code: "switch_led", Value: true},
	})

	require.NoError(t, err)
	require.Equal(t, tokenURI, recorder.request(0).URL.RequestURI())
	require.Equal(t, commandsURI, recorder.request(1).URL.RequestURI())
}

func tuyaDevicesJSON(prefix string, start, count int) string {
	items := make([]string, 0, count)
	for i := start; i < start+count; i++ {
		id := prefix + strconv.Itoa(i)
		name := "Device " + strconv.Itoa(i)
		items = append(items, `{"id":"`+id+`","name":"`+name+`","category":"dj","isOnline":true}`)
	}

	return `[` + strings.Join(items, ",") + `]`
}

func tuyaDevicesWithMissingLastIDJSON() string {
	items := make([]string, 0, listPageSize)
	for i := 1; i < listPageSize; i++ {
		id := "dev-" + strconv.Itoa(i)
		name := "Device " + strconv.Itoa(i)
		items = append(items, `{"id":"`+id+`","name":"`+name+`","category":"dj","isOnline":true}`)
	}
	items = append(items, `{"id":"","name":"Device without id","category":"dj","isOnline":true}`)

	return `[` + strings.Join(items, ",") + `]`
}
