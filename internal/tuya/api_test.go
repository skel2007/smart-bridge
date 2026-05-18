package tuya

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIListProjectDevicesFetchesAllPages(t *testing.T) {
	nextPageURI := "/v2.0/cloud/thing/device?last_id=dev-20&page_size=20"
	gateway, api := newTestGateway(t, map[testRoute]testResponse{
		getRoute(tokenURI):    tuyaResult(`{"access_token":"access-token"}`),
		getRoute(devicesURI):  tuyaResult(tuyaDevicesJSON("dev-", 1, 20)),
		getRoute(nextPageURI): tuyaResult(`[{"id":"dev-21","name":"Device 21","category":"dj","isOnline":true}]`),
	})

	deviceList, err := gateway.api.listProjectDevices(context.Background())

	require.NoError(t, err)
	require.Len(t, deviceList, 21)
	require.Equal(t, "dev-1", deviceList[0].ID)
	require.Equal(t, "dev-21", deviceList[20].ID)
	require.Equal(t, []string{tokenURI, devicesURI, nextPageURI}, api.requestURIs())
}

func TestAPIListProjectDevicesReturnsErrorWhenPageCursorIsMissing(t *testing.T) {
	gateway, _ := newTestGateway(t, map[testRoute]testResponse{
		getRoute(tokenURI):   tuyaResult(`{"access_token":"access-token"}`),
		getRoute(devicesURI): tuyaResult(tuyaDevicesWithMissingLastIDJSON()),
	})

	_, err := gateway.api.listProjectDevices(context.Background())

	require.EqualError(t, err, "tuya device list response missing id for pagination")
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
