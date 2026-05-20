package yandex

import (
	"errors"
	"strings"

	"github.com/skel2007/smart-bridge/internal/devices"
)

const (
	actionStatusDone  = "DONE"
	actionStatusError = "ERROR"
)

func newDeviceStateError(deviceID string, code string, message string) DeviceState {
	return DeviceState{
		ID:           deviceID,
		ErrorCode:    code,
		ErrorMessage: message,
	}
}

func newDeviceStateErrors(devices []DeviceRequest, code string, message string) []DeviceState {
	states := make([]DeviceState, 0, len(devices))
	for _, device := range devices {
		states = append(states, newDeviceStateError(device.ID, code, message))
	}

	return states
}

func newDeviceActionError(deviceID string, code string, message string) DeviceActionResult {
	return DeviceActionResult{
		ID:           deviceID,
		ActionResult: new(newActionError(code, message)),
	}
}

func newDeviceActionErrors(actions []DeviceAction, code string, message string) []DeviceActionResult {
	results := make([]DeviceActionResult, 0, len(actions))
	actionError := newActionError(code, message)
	for _, action := range actions {
		results = append(results, newDeviceCapabilityResults(action, actionError))
	}

	return results
}

func newDeviceCapabilityResults(action DeviceAction, result ActionResult) DeviceActionResult {
	if len(action.Capabilities) == 0 {
		return DeviceActionResult{
			ID:           action.ID,
			ActionResult: new(result),
		}
	}

	capabilities := make([]CapabilityActionResult, 0, len(action.Capabilities))
	for _, capability := range action.Capabilities {
		capabilities = append(capabilities, CapabilityActionResult{
			Type: capability.Type,
			State: CapabilityActionResultState{
				Instance:     capability.State.Instance,
				ActionResult: result,
			},
		})
	}

	return DeviceActionResult{
		ID:           action.ID,
		Capabilities: capabilities,
	}
}

func mapActionSendError(err error) (code string, message string) {
	if errors.Is(err, devices.ErrCapabilityNotSupported) {
		return errorCodeNotSupportedInCurrentMode, "action is not supported in current mode"
	}

	return errorCodeDeviceUnreachable, "device is unreachable"
}

func newActionError(code string, message string) ActionResult {
	return ActionResult{
		Status:       actionStatusError,
		ErrorCode:    code,
		ErrorMessage: strings.TrimSpace(message),
	}
}
