package tuya

import (
	"context"
	"testing"

	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/skel2007/smart-bridge/internal/tuya/internal/cloud"
	"github.com/stretchr/testify/require"
)

func TestGatewayDoesNotCacheSpecificationsByDefault(t *testing.T) {
	api := newGatewayCacheCloudAPI()
	gateway := newGateway(api)

	_, err := gateway.ListCapabilities(context.Background(), "device-id")
	require.NoError(t, err)
	_, err = gateway.ListCapabilities(context.Background(), "device-id")
	require.NoError(t, err)

	require.Equal(t, 2, api.specificationCalls)
	require.Equal(t, 2, api.statusCalls)
}

func TestGatewaySpecificationCacheKeepsStatusFresh(t *testing.T) {
	api := newGatewayCacheCloudAPI()
	gateway := newGateway(api, WithSpecificationCache())

	_, err := gateway.ListCapabilities(context.Background(), "device-id")
	require.NoError(t, err)
	_, err = gateway.ListCapabilities(context.Background(), "device-id")
	require.NoError(t, err)

	require.Equal(t, 1, api.specificationCalls)
	require.Equal(t, 2, api.statusCalls)
}

func TestGatewaySpecificationCacheIsSharedByReadAndWrite(t *testing.T) {
	api := newGatewayCacheCloudAPI()
	gateway := newGateway(api, WithSpecificationCache())

	_, err := gateway.ListCapabilities(context.Background(), "device-id")
	require.NoError(t, err)
	err = gateway.SendCommands(context.Background(), "device-id", []devices.CapabilityCommand{
		devices.NewOnOffCommand(devices.CapabilityInstancePower, true),
	})
	require.NoError(t, err)

	require.Equal(t, 1, api.specificationCalls)
	require.Equal(t, 1, api.statusCalls)
	require.Equal(t, []cloud.Command{{Code: "switch_led", Value: true}}, api.commands)
}

type gatewayCacheCloudAPI struct {
	specifications     cloud.DeviceSpecifications
	specificationCalls int

	status      []cloud.DeviceStatus
	statusCalls int

	commands []cloud.Command
}

func newGatewayCacheCloudAPI() *gatewayCacheCloudAPI {
	return &gatewayCacheCloudAPI{
		specifications: powerAndBrightnessSpecifications(),
		status: []cloud.DeviceStatus{
			{Code: "switch_led", Value: []byte(`true`)},
			{Code: "bright_value_v2", Value: []byte(`1000`)},
		},
	}
}

func (api *gatewayCacheCloudAPI) ListProjectDevices(context.Context) ([]cloud.Device, error) {
	return nil, nil
}

func (api *gatewayCacheCloudAPI) GetDeviceSpecifications(context.Context, string) (cloud.DeviceSpecifications, error) {
	api.specificationCalls++

	return api.specifications, nil
}

func (api *gatewayCacheCloudAPI) GetDeviceStatus(context.Context, string) ([]cloud.DeviceStatus, error) {
	api.statusCalls++
	return append([]cloud.DeviceStatus(nil), api.status...), nil
}

func (api *gatewayCacheCloudAPI) SendCommands(_ context.Context, _ string, commands []cloud.Command) error {
	api.commands = append([]cloud.Command(nil), commands...)
	return nil
}
