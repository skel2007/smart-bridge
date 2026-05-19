package yandex

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/skel2007/smart-bridge/internal/devices"
)

func mapDeviceActionCommands(action DeviceAction) ([]devices.CapabilityCommand, error) {
	commands := make([]devices.CapabilityCommand, 0, len(action.Capabilities))

	for _, capability := range action.Capabilities {
		command, err := mapCapabilityActionCommand(capability)
		if err != nil {
			return nil, err
		}

		commands = append(commands, command)
	}

	return commands, nil
}

func mapCapabilityActionCommand(action CapabilityAction) (devices.CapabilityCommand, error) {
	switch action.Type {
	case capabilityTypeOnOff:
		return mapOnOffActionCommand(action)
	case capabilityTypeRange:
		return mapRangeActionCommand(action)
	case capabilityTypeColorSetting:
		return mapColorSettingActionCommand(action)
	default:
		return devices.CapabilityCommand{}, unsupportedAction(action)
	}
}

func mapOnOffActionCommand(action CapabilityAction) (devices.CapabilityCommand, error) {
	if action.State.Instance != capabilityInstanceOn {
		return devices.CapabilityCommand{}, unsupportedAction(action)
	}

	var state bool
	if err := decodeActionValue(action, &state); err != nil {
		return devices.CapabilityCommand{}, err
	}

	return devices.NewOnOffCommand(devices.CapabilityInstancePower, state), nil
}

func mapRangeActionCommand(action CapabilityAction) (devices.CapabilityCommand, error) {
	if action.State.Instance != capabilityInstanceBrightness {
		return devices.CapabilityCommand{}, unsupportedAction(action)
	}
	if action.State.Relative != nil && *action.State.Relative {
		return devices.CapabilityCommand{}, unsupportedAction(action)
	}

	var state float64
	if err := decodeActionValue(action, &state); err != nil {
		return devices.CapabilityCommand{}, err
	}

	command := devices.NewRangeCommand(devices.CapabilityInstanceBrightness, state)
	if err := command.Validate(); err != nil {
		return devices.CapabilityCommand{}, invalidActionValue(action, err)
	}

	return command, nil
}

func mapColorSettingActionCommand(action CapabilityAction) (devices.CapabilityCommand, error) {
	switch action.State.Instance {
	case capabilityInstanceHSV:
		return mapHSVActionCommand(action)
	case capabilityInstanceTemperatureK:
		return mapTemperatureKActionCommand(action)
	default:
		return devices.CapabilityCommand{}, unsupportedAction(action)
	}
}

func mapHSVActionCommand(action CapabilityAction) (devices.CapabilityCommand, error) {
	var state actionHSVValue
	if err := decodeActionValue(action, &state); err != nil {
		return devices.CapabilityCommand{}, err
	}
	if state.H == nil || state.S == nil || state.V == nil {
		return devices.CapabilityCommand{}, invalidActionValue(action, errors.New("hsv value must include h, s, and v"))
	}

	command := devices.NewColorCommand(devices.CapabilityInstanceColor, devices.HSVColor{
		Hue:        float64(*state.H),
		Saturation: float64(*state.S),
		Value:      float64(*state.V),
	})
	if err := command.Validate(); err != nil {
		return devices.CapabilityCommand{}, invalidActionValue(action, err)
	}

	return command, nil
}

func mapTemperatureKActionCommand(action CapabilityAction) (devices.CapabilityCommand, error) {
	var state float64
	if err := decodeActionValue(action, &state); err != nil {
		return devices.CapabilityCommand{}, err
	}

	command := devices.NewRangeCommand(
		devices.CapabilityInstanceColorTemperatureLevel,
		mapKelvinToColorTemperatureLevel(state),
	)
	if err := command.Validate(); err != nil {
		return devices.CapabilityCommand{}, invalidActionValue(action, err)
	}

	return command, nil
}

func decodeActionValue(action CapabilityAction, out any) error {
	if len(action.State.Value) == 0 {
		return invalidActionValue(action, errors.New("action value is required"))
	}
	if bytes.Equal(bytes.TrimSpace(action.State.Value), []byte("null")) {
		return invalidActionValue(action, errors.New("action value is required"))
	}
	if err := json.Unmarshal(action.State.Value, out); err != nil {
		return invalidActionValue(action, err)
	}

	return nil
}

func unsupportedAction(action CapabilityAction) actionMappingError {
	return actionMappingError{
		Code:    errorCodeNotSupportedInCurrentMode,
		Message: fmt.Sprintf("unsupported action: type %q instance %q", action.Type, action.State.Instance),
	}
}

func invalidActionValue(action CapabilityAction, err error) actionMappingError {
	return actionMappingError{
		Code:    errorCodeInvalidValue,
		Message: fmt.Sprintf("invalid action value for type %q instance %q: %v", action.Type, action.State.Instance, err),
		Cause:   err,
	}
}

type actionHSVValue struct {
	H *int `json:"h"`
	S *int `json:"s"`
	V *int `json:"v"`
}
