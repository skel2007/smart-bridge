package devices

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewOnOffCapability(t *testing.T) {
	capability := NewOnOffCapability(CapabilityInstancePower, true)

	require.Equal(t, CapabilityTypeOnOff, capability.Type)
	require.Equal(t, CapabilityInstancePower, capability.Instance)
	require.NotNil(t, capability.OnOff)
	require.NotNil(t, capability.OnOff.State)
	require.True(t, *capability.OnOff.State)
	require.Nil(t, capability.Range)
	require.Nil(t, capability.Color)
	require.Nil(t, capability.Mode)
}

func TestNewOnOffCapabilityWithoutState(t *testing.T) {
	capability := NewOnOffCapabilityWithoutState(CapabilityInstancePower)

	require.Equal(t, CapabilityTypeOnOff, capability.Type)
	require.Equal(t, CapabilityInstancePower, capability.Instance)
	require.NotNil(t, capability.OnOff)
	require.Nil(t, capability.OnOff.State)
	require.Nil(t, capability.Range)
	require.Nil(t, capability.Color)
	require.Nil(t, capability.Mode)
}

func TestNewRangeCapability(t *testing.T) {
	parameters := RangeParameters{Min: 0, Max: 100, Precision: 1}

	capability := NewRangeCapability(CapabilityInstanceBrightness, 75, parameters)

	require.Equal(t, CapabilityTypeRange, capability.Type)
	require.Equal(t, CapabilityInstanceBrightness, capability.Instance)
	require.NotNil(t, capability.Range)
	require.NotNil(t, capability.Range.State)
	require.Equal(t, 75.0, *capability.Range.State)
	require.Equal(t, parameters, capability.Range.Parameters)
	require.Nil(t, capability.OnOff)
	require.Nil(t, capability.Color)
	require.Nil(t, capability.Mode)
}

func TestNewRangeCapabilityWithoutState(t *testing.T) {
	parameters := RangeParameters{Min: 0, Max: 100, Precision: 1}

	capability := NewRangeCapabilityWithoutState(CapabilityInstanceBrightness, parameters)

	require.Equal(t, CapabilityTypeRange, capability.Type)
	require.Equal(t, CapabilityInstanceBrightness, capability.Instance)
	require.NotNil(t, capability.Range)
	require.Nil(t, capability.Range.State)
	require.Equal(t, parameters, capability.Range.Parameters)
	require.Nil(t, capability.OnOff)
	require.Nil(t, capability.Color)
	require.Nil(t, capability.Mode)
}

func TestNewColorCapability(t *testing.T) {
	color := HSVColor{Hue: 120, Saturation: 80, Value: 90}

	capability := NewColorCapability(CapabilityInstanceColor, color)
	color.Hue = 240

	require.Equal(t, CapabilityTypeColor, capability.Type)
	require.Equal(t, CapabilityInstanceColor, capability.Instance)
	require.NotNil(t, capability.Color)
	require.NotNil(t, capability.Color.State)
	require.Equal(t, HSVColor{Hue: 120, Saturation: 80, Value: 90}, *capability.Color.State)
	require.Nil(t, capability.OnOff)
	require.Nil(t, capability.Range)
	require.Nil(t, capability.Mode)
}

func TestNewColorCapabilityWithoutState(t *testing.T) {
	capability := NewColorCapabilityWithoutState(CapabilityInstanceColor)

	require.Equal(t, CapabilityTypeColor, capability.Type)
	require.Equal(t, CapabilityInstanceColor, capability.Instance)
	require.NotNil(t, capability.Color)
	require.Nil(t, capability.Color.State)
	require.Nil(t, capability.OnOff)
	require.Nil(t, capability.Range)
	require.Nil(t, capability.Mode)
}

func TestNewModeCapability(t *testing.T) {
	parameters := ModeParameters{Modes: []string{"white", "colour"}}

	capability := NewModeCapability(CapabilityInstanceWorkMode, "white", parameters)
	parameters.Modes[0] = "mutated"

	require.Equal(t, CapabilityTypeMode, capability.Type)
	require.Equal(t, CapabilityInstanceWorkMode, capability.Instance)
	require.NotNil(t, capability.Mode)
	require.NotNil(t, capability.Mode.State)
	require.Equal(t, "white", *capability.Mode.State)
	require.Equal(t, ModeParameters{Modes: []string{"white", "colour"}}, capability.Mode.Parameters)
	require.Nil(t, capability.OnOff)
	require.Nil(t, capability.Range)
	require.Nil(t, capability.Color)
}

func TestNewModeCapabilityWithoutState(t *testing.T) {
	parameters := ModeParameters{Modes: []string{"white", "colour"}}

	capability := NewModeCapabilityWithoutState(CapabilityInstanceWorkMode, parameters)
	parameters.Modes[0] = "mutated"

	require.Equal(t, CapabilityTypeMode, capability.Type)
	require.Equal(t, CapabilityInstanceWorkMode, capability.Instance)
	require.NotNil(t, capability.Mode)
	require.Nil(t, capability.Mode.State)
	require.Equal(t, ModeParameters{Modes: []string{"white", "colour"}}, capability.Mode.Parameters)
	require.Nil(t, capability.OnOff)
	require.Nil(t, capability.Range)
	require.Nil(t, capability.Color)
}

func TestCapabilityJSON(t *testing.T) {
	capability := NewRangeCapability(
		CapabilityInstanceBrightness,
		75,
		RangeParameters{Min: 0, Max: 100, Precision: 1},
	)

	data, err := json.Marshal(capability)

	require.NoError(t, err)
	require.JSONEq(t, `{
		"type": "range",
		"instance": "brightness",
		"range": {
			"state": 75,
			"parameters": {
				"min": 0,
				"max": 100,
				"precision": 1
			}
		}
	}`, string(data))
}

func TestCapabilityJSONOmitsMissingState(t *testing.T) {
	capability := NewOnOffCapabilityWithoutState(CapabilityInstancePower)

	data, err := json.Marshal(capability)

	require.NoError(t, err)
	require.JSONEq(t, `{
		"type": "on_off",
		"instance": "power",
		"on_off": {}
	}`, string(data))
}
