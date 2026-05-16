package devices

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCapabilityCommandConstructors(t *testing.T) {
	tests := []struct {
		name    string
		command CapabilityCommand
		check   func(t *testing.T, command CapabilityCommand)
	}{
		{
			name:    "on off",
			command: NewOnOffCommand(CapabilityInstancePower, true),
			check: func(t *testing.T, command CapabilityCommand) {
				t.Helper()

				require.Equal(t, CapabilityInstancePower, command.Instance)
				require.NotNil(t, command.OnOff)
				require.True(t, command.OnOff.State)
				require.Nil(t, command.Range)
				require.Nil(t, command.Color)
				require.Nil(t, command.Mode)
			},
		},
		{
			name:    "range",
			command: NewRangeCommand(CapabilityInstanceBrightness, 42),
			check: func(t *testing.T, command CapabilityCommand) {
				t.Helper()

				require.Equal(t, CapabilityInstanceBrightness, command.Instance)
				require.NotNil(t, command.Range)
				require.Equal(t, 42.0, command.Range.State)
				require.Nil(t, command.OnOff)
				require.Nil(t, command.Color)
				require.Nil(t, command.Mode)
			},
		},
		{
			name:    "color",
			command: NewColorCommand(CapabilityInstanceColor, HSVColor{Hue: 120, Saturation: 80, Value: 90}),
			check: func(t *testing.T, command CapabilityCommand) {
				t.Helper()

				require.Equal(t, CapabilityInstanceColor, command.Instance)
				require.NotNil(t, command.Color)
				require.Equal(t, HSVColor{Hue: 120, Saturation: 80, Value: 90}, command.Color.State)
				require.Nil(t, command.OnOff)
				require.Nil(t, command.Range)
				require.Nil(t, command.Mode)
			},
		},
		{
			name:    "mode",
			command: NewModeCommand(CapabilityInstanceWorkMode, "white"),
			check: func(t *testing.T, command CapabilityCommand) {
				t.Helper()

				require.Equal(t, CapabilityInstanceWorkMode, command.Instance)
				require.NotNil(t, command.Mode)
				require.Equal(t, "white", command.Mode.State)
				require.Nil(t, command.OnOff)
				require.Nil(t, command.Range)
				require.Nil(t, command.Color)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, tt.command)
			require.NoError(t, tt.command.Validate())
		})
	}
}

func TestCapabilityCommandJSON(t *testing.T) {
	command := NewRangeCommand(CapabilityInstanceBrightness, 42)

	data, err := json.Marshal(command)

	require.NoError(t, err)
	require.JSONEq(t, `{
		"instance": "brightness",
		"range": {
			"state": 42
		}
	}`, string(data))
}

func TestCapabilityCommandValidate(t *testing.T) {
	tests := []struct {
		name    string
		command CapabilityCommand
	}{
		{
			name:    "power",
			command: NewOnOffCommand(CapabilityInstancePower, false),
		},
		{
			name:    "brightness lower bound",
			command: NewRangeCommand(CapabilityInstanceBrightness, 0),
		},
		{
			name:    "brightness upper bound",
			command: NewRangeCommand(CapabilityInstanceBrightness, 100),
		},
		{
			name:    "color temperature level",
			command: NewRangeCommand(CapabilityInstanceColorTemperatureLevel, 50),
		},
		{
			name:    "color lower bounds",
			command: NewColorCommand(CapabilityInstanceColor, HSVColor{Hue: 0, Saturation: 0, Value: 0}),
		},
		{
			name:    "color upper bounds",
			command: NewColorCommand(CapabilityInstanceColor, HSVColor{Hue: 360, Saturation: 100, Value: 100}),
		},
		{
			name:    "mode",
			command: NewModeCommand(CapabilityInstanceWorkMode, "white"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, tt.command.Validate())
		})
	}
}

func TestCapabilityCommandValidateErrors(t *testing.T) {
	tests := []struct {
		name    string
		command CapabilityCommand
		wantErr string
	}{
		{
			name:    "empty instance",
			command: CapabilityCommand{OnOff: &OnOffCommand{State: true}},
			wantErr: "capability command instance is required",
		},
		{
			name:    "unknown instance",
			command: NewRangeCommand("unsupported", 50),
			wantErr: "unknown capability command instance: unsupported",
		},
		{
			name:    "missing payload",
			command: CapabilityCommand{Instance: CapabilityInstancePower},
			wantErr: "capability command payload is required",
		},
		{
			name: "multiple payloads",
			command: CapabilityCommand{
				Instance: CapabilityInstancePower,
				OnOff:    &OnOffCommand{State: true},
				Range:    &RangeCommand{State: 50},
			},
			wantErr: "capability command must have exactly one payload",
		},
		{
			name:    "payload does not match instance",
			command: NewOnOffCommand(CapabilityInstanceBrightness, true),
			wantErr: "capability command payload \"on_off\" does not match instance \"brightness\"",
		},
		{
			name:    "range below lower bound",
			command: NewRangeCommand(CapabilityInstanceBrightness, -1),
			wantErr: "range command state must be between 0 and 100: -1",
		},
		{
			name:    "range above upper bound",
			command: NewRangeCommand(CapabilityInstanceBrightness, 101),
			wantErr: "range command state must be between 0 and 100: 101",
		},
		{
			name:    "hue below lower bound",
			command: NewColorCommand(CapabilityInstanceColor, HSVColor{Hue: -1, Saturation: 50, Value: 50}),
			wantErr: "color command hue must be between 0 and 360: -1",
		},
		{
			name:    "hue above upper bound",
			command: NewColorCommand(CapabilityInstanceColor, HSVColor{Hue: 361, Saturation: 50, Value: 50}),
			wantErr: "color command hue must be between 0 and 360: 361",
		},
		{
			name:    "saturation below lower bound",
			command: NewColorCommand(CapabilityInstanceColor, HSVColor{Hue: 120, Saturation: -1, Value: 50}),
			wantErr: "color command saturation must be between 0 and 100: -1",
		},
		{
			name:    "saturation above upper bound",
			command: NewColorCommand(CapabilityInstanceColor, HSVColor{Hue: 120, Saturation: 101, Value: 50}),
			wantErr: "color command saturation must be between 0 and 100: 101",
		},
		{
			name:    "value below lower bound",
			command: NewColorCommand(CapabilityInstanceColor, HSVColor{Hue: 120, Saturation: 50, Value: -1}),
			wantErr: "color command value must be between 0 and 100: -1",
		},
		{
			name:    "value above upper bound",
			command: NewColorCommand(CapabilityInstanceColor, HSVColor{Hue: 120, Saturation: 50, Value: 101}),
			wantErr: "color command value must be between 0 and 100: 101",
		},
		{
			name:    "empty mode",
			command: NewModeCommand(CapabilityInstanceWorkMode, ""),
			wantErr: "mode command state is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.EqualError(t, tt.command.Validate(), tt.wantErr)
		})
	}
}
