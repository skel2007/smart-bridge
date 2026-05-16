package devices

type DeviceDetails struct {
	Device
	Capabilities []Capability `json:"capabilities"`
}

type Capability struct {
	Type     CapabilityType     `json:"type"`
	Instance CapabilityInstance `json:"instance"`

	OnOff *OnOffCapability `json:"on_off,omitempty"`
	Range *RangeCapability `json:"range,omitempty"`
	Color *ColorCapability `json:"color,omitempty"`
	Mode  *ModeCapability  `json:"mode,omitempty"`
}

type CapabilityType string

const (
	CapabilityTypeOnOff CapabilityType = "on_off"
	CapabilityTypeRange CapabilityType = "range"
	CapabilityTypeColor CapabilityType = "color"
	CapabilityTypeMode  CapabilityType = "mode"
)

type CapabilityInstance string

const (
	CapabilityInstancePower                 CapabilityInstance = "power"
	CapabilityInstanceBrightness            CapabilityInstance = "brightness"
	CapabilityInstanceColorTemperature      CapabilityInstance = "color_temperature"
	CapabilityInstanceColorTemperatureLevel CapabilityInstance = "color_temperature_level"
	CapabilityInstanceColor                 CapabilityInstance = "color"
	CapabilityInstanceWorkMode              CapabilityInstance = "work_mode"
)

type OnOffCapability struct {
	State *bool `json:"state,omitempty"`
}

type RangeCapability struct {
	State      *float64        `json:"state,omitempty"`
	Parameters RangeParameters `json:"parameters"`
}

type RangeParameters struct {
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Precision float64 `json:"precision"`
}

type ColorCapability struct {
	State *HSVColor `json:"state,omitempty"`
}

type HSVColor struct {
	Hue        float64 `json:"hue"`
	Saturation float64 `json:"saturation"`
	Value      float64 `json:"value"`
}

type ModeCapability struct {
	State      *string        `json:"state,omitempty"`
	Parameters ModeParameters `json:"parameters"`
}

type ModeParameters struct {
	Modes []string `json:"modes"`
}

func NewOnOffCapability(instance CapabilityInstance, state bool) Capability {
	return Capability{
		Type:     CapabilityTypeOnOff,
		Instance: instance,
		OnOff:    &OnOffCapability{State: &state},
	}
}

func NewOnOffCapabilityWithoutState(instance CapabilityInstance) Capability {
	return Capability{
		Type:     CapabilityTypeOnOff,
		Instance: instance,
		OnOff:    &OnOffCapability{},
	}
}

func NewRangeCapability(instance CapabilityInstance, state float64, parameters RangeParameters) Capability {
	return Capability{
		Type:     CapabilityTypeRange,
		Instance: instance,
		Range: &RangeCapability{
			State:      &state,
			Parameters: parameters,
		},
	}
}

func NewRangeCapabilityWithoutState(instance CapabilityInstance, parameters RangeParameters) Capability {
	return Capability{
		Type:     CapabilityTypeRange,
		Instance: instance,
		Range: &RangeCapability{
			Parameters: parameters,
		},
	}
}

func NewColorCapability(instance CapabilityInstance, state HSVColor) Capability {
	return Capability{
		Type:     CapabilityTypeColor,
		Instance: instance,
		Color:    &ColorCapability{State: &state},
	}
}

func NewColorCapabilityWithoutState(instance CapabilityInstance) Capability {
	return Capability{
		Type:     CapabilityTypeColor,
		Instance: instance,
		Color:    &ColorCapability{},
	}
}

func NewModeCapability(instance CapabilityInstance, state string, parameters ModeParameters) Capability {
	return Capability{
		Type:     CapabilityTypeMode,
		Instance: instance,
		Mode: &ModeCapability{
			State:      &state,
			Parameters: cloneModeParameters(parameters),
		},
	}
}

func NewModeCapabilityWithoutState(instance CapabilityInstance, parameters ModeParameters) Capability {
	return Capability{
		Type:     CapabilityTypeMode,
		Instance: instance,
		Mode: &ModeCapability{
			Parameters: cloneModeParameters(parameters),
		},
	}
}

// cloneModeParameters copies Modes so callers cannot mutate the backing slice after construction.
func cloneModeParameters(parameters ModeParameters) ModeParameters {
	return ModeParameters{Modes: append([]string(nil), parameters.Modes...)}
}
