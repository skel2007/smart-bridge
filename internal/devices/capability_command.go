package devices

import (
	"errors"
	"fmt"
)

type CapabilityCommand struct {
	Instance CapabilityInstance `json:"instance"`

	// Tagged variant payloads keep JSON flat; Validate enforces one payload matching Instance.
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

	expected, ok := CapabilityTypeForInstance(command.Instance)
	if !ok {
		return fmt.Errorf("unknown capability command instance: %s", command.Instance)
	}

	actual, count := command.capabilityType()
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
	case CapabilityTypeRange:
		return validateRangeCommand(command.Range)
	case CapabilityTypeColor:
		return validateColorCommand(command.Color)
	case CapabilityTypeMode:
		return validateModeCommand(command.Mode)
	default:
		return nil
	}
}

func (command CapabilityCommand) capabilityType() (CapabilityType, int) {
	var capabilityType CapabilityType
	var count int

	if command.OnOff != nil {
		capabilityType = CapabilityTypeOnOff
		count++
	}
	if command.Range != nil {
		capabilityType = CapabilityTypeRange
		count++
	}
	if command.Color != nil {
		capabilityType = CapabilityTypeColor
		count++
	}
	if command.Mode != nil {
		capabilityType = CapabilityTypeMode
		count++
	}

	return capabilityType, count
}

func validateRangeCommand(command *RangeCommand) error {
	if command.State < PercentMin || command.State > PercentMax {
		return fmt.Errorf("range command state must be between %d and %d: %v", PercentMin, PercentMax, command.State)
	}

	return nil
}

func validateColorCommand(command *ColorCommand) error {
	if command.State.Hue < 0 || command.State.Hue > 360 {
		return fmt.Errorf("color command hue must be between 0 and 360: %v", command.State.Hue)
	}
	if command.State.Saturation < PercentMin || command.State.Saturation > PercentMax {
		return fmt.Errorf("color command saturation must be between %d and %d: %v", PercentMin, PercentMax, command.State.Saturation)
	}
	if command.State.Value < PercentMin || command.State.Value > PercentMax {
		return fmt.Errorf("color command value must be between %d and %d: %v", PercentMin, PercentMax, command.State.Value)
	}

	return nil
}

func validateModeCommand(command *ModeCommand) error {
	if command.State == "" {
		return errors.New("mode command state is required")
	}

	return nil
}
