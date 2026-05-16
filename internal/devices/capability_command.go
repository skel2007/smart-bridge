package devices

import (
	"errors"
	"fmt"
)

type CapabilityCommand struct {
	Instance CapabilityInstance `json:"instance"`

	OnOff *OnOffCommand `json:"on_off,omitempty"`
	Range *RangeCommand `json:"range,omitempty"`
	Color *ColorCommand `json:"color,omitempty"`
	Mode  *ModeCommand  `json:"mode,omitempty"`
}

type OnOffCommand struct {
	State bool `json:"state"`
}

type RangeCommand struct {
	State float64 `json:"state"`
}

type ColorCommand struct {
	State HSVColor `json:"state"`
}

type ModeCommand struct {
	State string `json:"state"`
}

func NewOnOffCommand(instance CapabilityInstance, state bool) CapabilityCommand {
	return CapabilityCommand{
		Instance: instance,
		OnOff:    &OnOffCommand{State: state},
	}
}

func NewRangeCommand(instance CapabilityInstance, state float64) CapabilityCommand {
	return CapabilityCommand{
		Instance: instance,
		Range:    &RangeCommand{State: state},
	}
}

func NewColorCommand(instance CapabilityInstance, state HSVColor) CapabilityCommand {
	return CapabilityCommand{
		Instance: instance,
		Color:    &ColorCommand{State: state},
	}
}

func NewModeCommand(instance CapabilityInstance, state string) CapabilityCommand {
	return CapabilityCommand{
		Instance: instance,
		Mode:     &ModeCommand{State: state},
	}
}

func (command CapabilityCommand) Validate() error {
	if command.Instance == "" {
		return errors.New("capability command instance is required")
	}

	expected, ok := commandPayloadKindForInstance(command.Instance)
	if !ok {
		return fmt.Errorf("unknown capability command instance: %s", command.Instance)
	}

	actual, count := command.commandPayloadKind()
	if count == 0 {
		return errors.New("capability command payload is required")
	}
	if count > 1 {
		return errors.New("capability command must have exactly one payload")
	}
	if actual != expected {
		return fmt.Errorf("capability command payload %q does not match instance %q", actual, command.Instance)
	}

	switch actual {
	case commandPayloadKindRange:
		return validateRangeCommand(command.Range)
	case commandPayloadKindColor:
		return validateColorCommand(command.Color)
	case commandPayloadKindMode:
		return validateModeCommand(command.Mode)
	default:
		return nil
	}
}

type commandPayloadKind string

const (
	commandPayloadKindOnOff commandPayloadKind = "on_off"
	commandPayloadKindRange commandPayloadKind = "range"
	commandPayloadKindColor commandPayloadKind = "color"
	commandPayloadKindMode  commandPayloadKind = "mode"
)

func commandPayloadKindForInstance(instance CapabilityInstance) (commandPayloadKind, bool) {
	switch instance {
	case CapabilityInstancePower:
		return commandPayloadKindOnOff, true
	case CapabilityInstanceBrightness, CapabilityInstanceColorTemperatureLevel:
		return commandPayloadKindRange, true
	case CapabilityInstanceColor:
		return commandPayloadKindColor, true
	case CapabilityInstanceWorkMode:
		return commandPayloadKindMode, true
	default:
		return "", false
	}
}

func (command CapabilityCommand) commandPayloadKind() (commandPayloadKind, int) {
	var kind commandPayloadKind
	var count int

	if command.OnOff != nil {
		kind = commandPayloadKindOnOff
		count++
	}
	if command.Range != nil {
		kind = commandPayloadKindRange
		count++
	}
	if command.Color != nil {
		kind = commandPayloadKindColor
		count++
	}
	if command.Mode != nil {
		kind = commandPayloadKindMode
		count++
	}

	return kind, count
}

func validateRangeCommand(command *RangeCommand) error {
	if command.State < 0 || command.State > 100 {
		return fmt.Errorf("range command state must be between 0 and 100: %v", command.State)
	}

	return nil
}

func validateColorCommand(command *ColorCommand) error {
	if command.State.Hue < 0 || command.State.Hue > 360 {
		return fmt.Errorf("color command hue must be between 0 and 360: %v", command.State.Hue)
	}
	if command.State.Saturation < 0 || command.State.Saturation > 100 {
		return fmt.Errorf("color command saturation must be between 0 and 100: %v", command.State.Saturation)
	}
	if command.State.Value < 0 || command.State.Value > 100 {
		return fmt.Errorf("color command value must be between 0 and 100: %v", command.State.Value)
	}

	return nil
}

func validateModeCommand(command *ModeCommand) error {
	if command.State == "" {
		return errors.New("mode command state is required")
	}

	return nil
}
