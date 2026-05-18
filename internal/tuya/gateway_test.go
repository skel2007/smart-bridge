package tuya

import (
	"context"
	"errors"
	"testing"

	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/stretchr/testify/require"
)

func TestListDevices(t *testing.T) {
	api := &recordingTuyaAPI{
		devices: []tuyaDevice{
			{ID: "dev-1", Name: "Lamp", CustomName: "Desk lamp", Category: "dj", IsOnline: true},
		},
	}

	deviceList, err := newGateway(api).ListDevices(context.Background())

	require.NoError(t, err)
	require.Equal(t, []devices.Device{
		{ID: "dev-1", Name: "Desk lamp", Type: devices.DeviceTypeLight, Online: true},
	}, deviceList)
}

func TestListDevicesReturnsAPIError(t *testing.T) {
	api := &recordingTuyaAPI{listDevicesErr: errors.New("list devices failed")}

	_, err := newGateway(api).ListDevices(context.Background())

	require.EqualError(t, err, "list devices failed")
}

func TestListCapabilities(t *testing.T) {
	api := &recordingTuyaAPI{
		specifications: powerAndBrightnessSpecifications(),
		status: []tuyaDeviceStatus{
			{Code: "switch_led", Value: []byte(`true`)},
			{Code: "bright_value_v2", Value: []byte(`1000`)},
		},
	}

	capabilities, err := newGateway(api).ListCapabilities(context.Background(), "device-id")

	require.NoError(t, err)
	require.Equal(t, []devices.Capability{
		devices.NewOnOffCapability(devices.CapabilityInstancePower, true),
		devices.NewRangeCapability(
			devices.CapabilityInstanceBrightness,
			100,
			devices.RangeParameters{Min: 0, Max: 100, Precision: 1},
		),
	}, capabilities)
	require.Equal(t, "device-id", api.specificationsDevice)
	require.Equal(t, "device-id", api.statusDevice)
}

func TestListCapabilitiesReturnsSpecificationsError(t *testing.T) {
	api := &recordingTuyaAPI{specificationsErr: errors.New("specifications failed")}

	_, err := newGateway(api).ListCapabilities(context.Background(), "device-id")

	require.EqualError(t, err, "specifications failed")
	require.Empty(t, api.statusDevice)
}

func TestListCapabilitiesReturnsStatusError(t *testing.T) {
	api := &recordingTuyaAPI{statusErr: errors.New("status failed")}

	_, err := newGateway(api).ListCapabilities(context.Background(), "device-id")

	require.EqualError(t, err, "status failed")
}

func TestSendCommands(t *testing.T) {
	api := &recordingTuyaAPI{
		specifications: powerAndBrightnessSpecifications(),
	}

	err := newGateway(api).SendCommands(context.Background(), "device-id", []devices.CapabilityCommand{
		devices.NewOnOffCommand(devices.CapabilityInstancePower, true),
		devices.NewRangeCommand(devices.CapabilityInstanceBrightness, 50),
	})

	require.NoError(t, err)
	require.Equal(t, "device-id", api.specificationsDevice)
	require.Equal(t, "device-id", api.commandsDevice)
	require.Equal(t, []tuyaCommand{
		{Code: "switch_led", Value: true},
		{Code: "bright_value_v2", Value: 505},
	}, api.sentCommands)
}

func TestSendCommandsReturnsErrorWhenEmpty(t *testing.T) {
	api := &recordingTuyaAPI{}

	err := newGateway(api).SendCommands(context.Background(), "device-id", nil)

	require.EqualError(t, err, "capability commands are required")
	require.Empty(t, api.specificationsDevice)
	require.Empty(t, api.commandsDevice)
}

func TestSendCommandsReturnsMappingError(t *testing.T) {
	api := &recordingTuyaAPI{specifications: tuyaDeviceSpecifications{Functions: []tuyaFunctionSpec{}}}

	err := newGateway(api).SendCommands(context.Background(), "device-id", []devices.CapabilityCommand{
		devices.NewOnOffCommand(devices.CapabilityInstancePower, true),
	})

	require.EqualError(t, err, "tuya function not found for capability instance: power")
	require.Equal(t, "device-id", api.specificationsDevice)
	require.Empty(t, api.commandsDevice)
}

func TestSendCommandsReturnsAPIError(t *testing.T) {
	api := &recordingTuyaAPI{
		specifications: tuyaDeviceSpecifications{
			Functions: []tuyaFunctionSpec{
				{Code: "switch_led", Type: "Boolean", Values: []byte(`{}`)},
			},
		},
		commandsErr: errors.New("send commands failed"),
	}

	err := newGateway(api).SendCommands(context.Background(), "device-id", []devices.CapabilityCommand{
		devices.NewOnOffCommand(devices.CapabilityInstancePower, true),
	})

	require.EqualError(t, err, "send commands failed")
}

func powerAndBrightnessSpecifications() tuyaDeviceSpecifications {
	return tuyaDeviceSpecifications{
		Functions: []tuyaFunctionSpec{
			{Code: "switch_led", Type: "Boolean", Values: []byte(`{}`)},
			{Code: "bright_value_v2", Type: "Integer", Values: []byte(`{"min":10,"max":1000,"scale":0,"step":1}`)},
		},
	}
}

type recordingTuyaAPI struct {
	devices        []tuyaDevice
	listDevicesErr error

	specifications       tuyaDeviceSpecifications
	specificationsErr    error
	specificationsDevice string

	status       []tuyaDeviceStatus
	statusErr    error
	statusDevice string

	commandsErr    error
	commandsDevice string
	sentCommands   []tuyaCommand
}

func (api *recordingTuyaAPI) listProjectDevices(context.Context) ([]tuyaDevice, error) {
	return api.devices, api.listDevicesErr
}

func (api *recordingTuyaAPI) getDeviceSpecifications(_ context.Context, deviceID string) (tuyaDeviceSpecifications, error) {
	api.specificationsDevice = deviceID

	return api.specifications, api.specificationsErr
}

func (api *recordingTuyaAPI) getDeviceStatus(_ context.Context, deviceID string) ([]tuyaDeviceStatus, error) {
	api.statusDevice = deviceID

	return api.status, api.statusErr
}

func (api *recordingTuyaAPI) sendCommands(_ context.Context, deviceID string, commands []tuyaCommand) error {
	api.commandsDevice = deviceID
	api.sentCommands = commands

	return api.commandsErr
}
