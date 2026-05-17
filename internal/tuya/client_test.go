package tuya

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/stretchr/testify/require"
)

func TestListDevices(t *testing.T) {
	client, api := newTestClient(t, map[testRoute]testResponse{
		getRoute(tokenURI):   tuyaResult(`{"access_token":"access-token"}`),
		getRoute(devicesURI): tuyaResult(`[{"id":"dev-1","name":"Lamp","customName":"Desk lamp","category":"dj","isOnline":true}]`),
	})

	deviceList, err := client.ListDevices(context.Background())
	require.NoError(t, err)
	require.Equal(t, []devices.Device{
		{ID: "dev-1", Name: "Desk lamp", Type: devices.DeviceTypeLight, Online: true},
	}, deviceList)
	require.Equal(t, []string{tokenURI, devicesURI}, api.requestURIs())
}

func TestListDevicesFetchesAllPages(t *testing.T) {
	nextPageURI := "/v2.0/cloud/thing/device?last_id=dev-20&page_size=20"
	client, api := newTestClient(t, map[testRoute]testResponse{
		getRoute(tokenURI):    tuyaResult(`{"access_token":"access-token"}`),
		getRoute(devicesURI):  tuyaResult(tuyaDevicesJSON("dev-", 1, 20)),
		getRoute(nextPageURI): tuyaResult(`[{"id":"dev-21","name":"Device 21","category":"dj","isOnline":true}]`),
	})

	deviceList, err := client.ListDevices(context.Background())
	require.NoError(t, err)
	require.Len(t, deviceList, 21)
	require.Equal(t, "dev-1", deviceList[0].ID)
	require.Equal(t, "dev-21", deviceList[20].ID)
	require.Equal(t, []string{tokenURI, devicesURI, nextPageURI}, api.requestURIs())
}

func TestListDevicesReturnsErrorWhenTokenIsMissing(t *testing.T) {
	client, _ := newTestClient(t, map[testRoute]testResponse{
		getRoute(tokenURI): tuyaResult(`{}`),
	})

	_, err := client.ListDevices(context.Background())

	require.EqualError(t, err, "tuya token response missing access_token")
}

func TestListDevicesReturnsErrorWhenPageCursorIsMissing(t *testing.T) {
	client, _ := newTestClient(t, map[testRoute]testResponse{
		getRoute(tokenURI):   tuyaResult(`{"access_token":"access-token"}`),
		getRoute(devicesURI): tuyaResult(tuyaDevicesWithMissingLastIDJSON()),
	})

	_, err := client.ListDevices(context.Background())

	require.EqualError(t, err, "tuya device list response missing id for pagination")
}

func TestListDevicesDoesNotExposeSecretsInErrors(t *testing.T) {
	client, _ := newTestClient(t, map[testRoute]testResponse{
		getRoute(tokenURI):   tuyaResult(`{"access_token":"access-secret"}`),
		getRoute(devicesURI): tuyaError(http.StatusOK, "1106", "permission denied"),
	})

	_, err := client.ListDevices(context.Background())
	require.Error(t, err)
	require.NotContains(t, err.Error(), "super-secret")
	require.NotContains(t, err.Error(), "access-secret")
}

func TestListCapabilities(t *testing.T) {
	client, api := newTestClient(t, map[testRoute]testResponse{
		getRoute(tokenURI): tuyaResult(`{"access_token":"access-token"}`),
		getRoute(deviceSpecificationsURI): tuyaResult(`{
			"functions": [
				{"code":"switch_led","type":"Boolean","values":"{}"},
				{"code":"bright_value_v2","type":"Integer","values":"{\"min\":10,\"max\":1000,\"scale\":0,\"step\":1}"}
			]
		}`),
		getRoute(deviceStatusURI): tuyaResult(`[
			{"code":"switch_led","value":true},
			{"code":"bright_value_v2","value":1000}
		]`),
	})

	capabilities, err := client.ListCapabilities(context.Background(), "device-id")

	require.NoError(t, err)
	require.Equal(t, []devices.Capability{
		devices.NewOnOffCapability(devices.CapabilityInstancePower, true),
		devices.NewRangeCapability(
			devices.CapabilityInstanceBrightness,
			100,
			devices.RangeParameters{Min: 0, Max: 100, Precision: 1},
		),
	}, capabilities)
	require.Equal(t, []string{tokenURI, deviceSpecificationsURI, deviceStatusURI}, api.requestURIs())
}

func TestListCapabilitiesEscapesDeviceID(t *testing.T) {
	specificationsURI := "/v1.0/devices/device%2Fid%20with%20space/specifications"
	statusURI := "/v1.0/devices/device%2Fid%20with%20space/status"
	client, api := newTestClient(t, map[testRoute]testResponse{
		getRoute(tokenURI):          tuyaResult(`{"access_token":"access-token"}`),
		getRoute(specificationsURI): tuyaResult(`{"functions":[]}`),
		getRoute(statusURI):         tuyaResult(`[]`),
	})

	_, err := client.ListCapabilities(context.Background(), "device/id with space")

	require.NoError(t, err)
	require.Equal(t, []string{tokenURI, specificationsURI, statusURI}, api.requestURIs())
}

func TestListCapabilitiesReturnsSpecificationsError(t *testing.T) {
	client, _ := newTestClient(t, map[testRoute]testResponse{
		getRoute(tokenURI):                tuyaResult(`{"access_token":"access-token"}`),
		getRoute(deviceSpecificationsURI): tuyaError(http.StatusOK, "1106", "permission denied"),
	})

	_, err := client.ListCapabilities(context.Background(), "device-id")

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, "1106", apiErr.Code)
}

func TestListCapabilitiesReturnsStatusError(t *testing.T) {
	client, api := newTestClient(t, map[testRoute]testResponse{
		getRoute(tokenURI):                tuyaResult(`{"access_token":"access-token"}`),
		getRoute(deviceSpecificationsURI): tuyaResult(`{"functions":[]}`),
		getRoute(deviceStatusURI):         tuyaError(http.StatusOK, "1107", "status denied"),
	})

	_, err := client.ListCapabilities(context.Background(), "device-id")

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, "1107", apiErr.Code)
	require.Equal(t, []string{tokenURI, deviceSpecificationsURI, deviceStatusURI}, api.requestURIs())
}

func TestSendCommands(t *testing.T) {
	client, api := newTestClient(t, map[testRoute]testResponse{
		getRoute(tokenURI): tuyaResult(`{"access_token":"access-token"}`),
		getRoute(deviceSpecificationsURI): tuyaResult(`{
			"functions": [
				{"code":"switch_led","type":"Boolean","values":"{}"},
				{"code":"bright_value_v2","type":"Integer","values":"{\"min\":10,\"max\":1000,\"scale\":0,\"step\":1}"}
			]
		}`),
		postRoute(deviceCommandsURI): tuyaResult(`true`),
	})

	err := client.SendCommands(context.Background(), "device-id", []devices.CapabilityCommand{
		devices.NewOnOffCommand(devices.CapabilityInstancePower, true),
		devices.NewRangeCommand(devices.CapabilityInstanceBrightness, 50),
	})

	require.NoError(t, err)
	require.Equal(t, []string{tokenURI, deviceSpecificationsURI, deviceCommandsURI}, api.requestURIs())
	require.Equal(t, http.MethodPost, api.requests[2].Method)
	require.Equal(t, "application/json", api.requests[2].Header.Get("Content-Type"))
	require.Equal(t, "access-token", api.requests[2].Header.Get("access_token"))
	require.JSONEq(t, `{
		"commands": [
			{"code": "switch_led", "value": true},
			{"code": "bright_value_v2", "value": 505}
		]
	}`, api.bodies[2])
}

func TestSendCommandsEscapesDeviceID(t *testing.T) {
	specificationsURI := "/v1.0/devices/device%2Fid%20with%20space/specifications"
	commandsURI := "/v1.0/devices/device%2Fid%20with%20space/commands"
	client, api := newTestClient(t, map[testRoute]testResponse{
		getRoute(tokenURI): tuyaResult(`{"access_token":"access-token"}`),
		getRoute(specificationsURI): tuyaResult(`{
			"functions": [
				{"code":"switch_led","type":"Boolean","values":"{}"}
			]
		}`),
		postRoute(commandsURI): tuyaResult(`true`),
	})

	err := client.SendCommands(context.Background(), "device/id with space", []devices.CapabilityCommand{
		devices.NewOnOffCommand(devices.CapabilityInstancePower, true),
	})

	require.NoError(t, err)
	require.Equal(t, []string{tokenURI, specificationsURI, commandsURI}, api.requestURIs())
}

func TestSendCommandsReturnsErrorWhenEmpty(t *testing.T) {
	client, api := newTestClient(t, map[testRoute]testResponse{})

	err := client.SendCommands(context.Background(), "device-id", nil)

	require.EqualError(t, err, "capability commands are required")
	require.Empty(t, api.requestURIs())
}

func TestSendCommandsReturnsMappingError(t *testing.T) {
	client, api := newTestClient(t, map[testRoute]testResponse{
		getRoute(tokenURI):                tuyaResult(`{"access_token":"access-token"}`),
		getRoute(deviceSpecificationsURI): tuyaResult(`{"functions":[]}`),
	})

	err := client.SendCommands(context.Background(), "device-id", []devices.CapabilityCommand{
		devices.NewOnOffCommand(devices.CapabilityInstancePower, true),
	})

	require.EqualError(t, err, "tuya function not found for capability instance: power")
	require.Equal(t, []string{tokenURI, deviceSpecificationsURI}, api.requestURIs())
}

func TestSendCommandsReturnsTuyaError(t *testing.T) {
	client, api := newTestClient(t, map[testRoute]testResponse{
		getRoute(tokenURI): tuyaResult(`{"access_token":"access-token"}`),
		getRoute(deviceSpecificationsURI): tuyaResult(`{
			"functions": [
				{"code":"switch_led","type":"Boolean","values":"{}"}
			]
		}`),
		postRoute(deviceCommandsURI): tuyaError(http.StatusOK, "1108", "command denied"),
	})

	err := client.SendCommands(context.Background(), "device-id", []devices.CapabilityCommand{
		devices.NewOnOffCommand(devices.CapabilityInstancePower, true),
	})

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, "1108", apiErr.Code)
	require.Equal(t, []string{tokenURI, deviceSpecificationsURI, deviceCommandsURI}, api.requestURIs())
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
