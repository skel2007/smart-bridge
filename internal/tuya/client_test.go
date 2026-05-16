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
	client, api := newTestClient(t, map[string]testResponse{
		tokenURI:   tuyaResult(`{"access_token":"access-token"}`),
		devicesURI: tuyaResult(`[{"id":"dev-1","name":"Lamp","customName":"Desk lamp","category":"dj","isOnline":true}]`),
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
	client, api := newTestClient(t, map[string]testResponse{
		tokenURI:    tuyaResult(`{"access_token":"access-token"}`),
		devicesURI:  tuyaResult(tuyaDevicesJSON("dev-", 1, 20)),
		nextPageURI: tuyaResult(`[{"id":"dev-21","name":"Device 21","category":"dj","isOnline":true}]`),
	})

	deviceList, err := client.ListDevices(context.Background())
	require.NoError(t, err)
	require.Len(t, deviceList, 21)
	require.Equal(t, "dev-1", deviceList[0].ID)
	require.Equal(t, "dev-21", deviceList[20].ID)
	require.Equal(t, []string{tokenURI, devicesURI, nextPageURI}, api.requestURIs())
}

func TestListDevicesReturnsErrorWhenTokenIsMissing(t *testing.T) {
	client, _ := newTestClient(t, map[string]testResponse{
		tokenURI: tuyaResult(`{}`),
	})

	_, err := client.ListDevices(context.Background())

	require.EqualError(t, err, "tuya token response missing access_token")
}

func TestListDevicesReturnsErrorWhenPageCursorIsMissing(t *testing.T) {
	client, _ := newTestClient(t, map[string]testResponse{
		tokenURI:   tuyaResult(`{"access_token":"access-token"}`),
		devicesURI: tuyaResult(tuyaDevicesWithMissingLastIDJSON()),
	})

	_, err := client.ListDevices(context.Background())

	require.EqualError(t, err, "tuya device list response missing id for pagination")
}

func TestListDevicesDoesNotExposeSecretsInErrors(t *testing.T) {
	client, _ := newTestClient(t, map[string]testResponse{
		tokenURI:   tuyaResult(`{"access_token":"access-secret"}`),
		devicesURI: tuyaError(http.StatusOK, "1106", "permission denied"),
	})

	_, err := client.ListDevices(context.Background())
	require.Error(t, err)
	require.NotContains(t, err.Error(), "super-secret")
	require.NotContains(t, err.Error(), "access-secret")
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
