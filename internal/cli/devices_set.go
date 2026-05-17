package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/skel2007/smart-bridge/internal/devices"
)

type devicesSetResult struct {
	Sent bool `json:"sent"`
}

type devicesSetColorOptions struct {
	hue        float64
	saturation float64
	value      float64
}

type devicesSetCommandSpec struct {
	instance devices.CapabilityInstance
	alias    string
	short    string
}

var devicesSetCommandSpecs = []devicesSetCommandSpec{
	{
		instance: devices.CapabilityInstancePower,
		short:    "Set device power",
	},
	{
		instance: devices.CapabilityInstanceBrightness,
		short:    "Set device brightness",
	},
	{
		instance: devices.CapabilityInstanceColorTemperatureLevel,
		alias:    "color-temperature",
		short:    "Set device color temperature level",
	},
	{
		instance: devices.CapabilityInstanceColor,
		short:    "Set device color",
	},
	{
		instance: devices.CapabilityInstanceWorkMode,
		alias:    "mode",
		short:    "Set device mode",
	},
}

func newDevicesSetCommand(rootOpts *options, devicesOpts *devicesOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set a device capability through Tuya Cloud",
	}

	for _, spec := range devicesSetCommandSpecs {
		cmd.AddCommand(newDevicesSetCapabilityCommand(rootOpts, devicesOpts, spec))
	}

	return cmd
}

func newDevicesSetCapabilityCommand(rootOpts *options, devicesOpts *devicesOptions, spec devicesSetCommandSpec) *cobra.Command {
	capabilityType, ok := devices.CapabilityTypeForInstance(spec.instance)
	if !ok {
		panic(fmt.Sprintf("unsupported devices set capability instance: %s", spec.instance))
	}

	switch capabilityType {
	case devices.CapabilityTypeOnOff:
		return newDevicesSetOnOffCommand(rootOpts, devicesOpts, spec)
	case devices.CapabilityTypeRange:
		return newDevicesSetRangeCommand(rootOpts, devicesOpts, spec)
	case devices.CapabilityTypeColor:
		return newDevicesSetColorCommand(rootOpts, devicesOpts, spec)
	case devices.CapabilityTypeMode:
		return newDevicesSetModeCommand(rootOpts, devicesOpts, spec)
	default:
		panic(fmt.Sprintf("unsupported devices set capability type: %s", capabilityType))
	}
}

func newDevicesSetOnOffCommand(rootOpts *options, devicesOpts *devicesOptions, spec devicesSetCommandSpec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     spec.name() + " <device-id> on|off",
		Short:   spec.short,
		Example: "  smart-bridge devices set " + spec.name() + " <device-id> on",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			command, err := parseOnOffCommand(spec.instance, args[1])
			if err != nil {
				return err
			}

			return runDevicesSet(cmd, rootOpts, devicesOpts, args[0], command)
		},
	}

	return cmd
}

func newDevicesSetRangeCommand(rootOpts *options, devicesOpts *devicesOptions, spec devicesSetCommandSpec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     spec.name() + " <device-id> <0-100>",
		Short:   spec.short,
		Example: "  smart-bridge devices set " + spec.name() + " <device-id> 50",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			command, err := parseRangeCommand(spec.name(), spec.instance, args[1])
			if err != nil {
				return err
			}

			return runDevicesSet(cmd, rootOpts, devicesOpts, args[0], command)
		},
	}

	return cmd
}

func newDevicesSetColorCommand(rootOpts *options, devicesOpts *devicesOptions, spec devicesSetCommandSpec) *cobra.Command {
	opts := &devicesSetColorOptions{}

	cmd := &cobra.Command{
		Use:   spec.name() + " <device-id>",
		Short: spec.short,
		Example: "  smart-bridge devices set " + spec.name() + ` <device-id> \
    --hue 120 \
    --saturation 80 \
    --value 90`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			command, err := parseColorCommand(cmd, spec.instance, opts)
			if err != nil {
				return err
			}

			return runDevicesSet(cmd, rootOpts, devicesOpts, args[0], command)
		},
	}

	cmd.Flags().Float64Var(&opts.hue, "hue", 0, "color hue from 0 to 360")
	cmd.Flags().Float64Var(&opts.saturation, "saturation", 0, "color saturation from 0 to 100")
	cmd.Flags().Float64Var(&opts.value, "value", 0, "color value from 0 to 100")

	return cmd
}

func newDevicesSetModeCommand(rootOpts *options, devicesOpts *devicesOptions, spec devicesSetCommandSpec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     spec.name() + " <device-id> <mode>",
		Short:   spec.short,
		Example: "  smart-bridge devices set " + spec.name() + " <device-id> white",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			command, err := parseModeCommand(spec.instance, args[1])
			if err != nil {
				return err
			}

			return runDevicesSet(cmd, rootOpts, devicesOpts, args[0], command)
		},
	}

	return cmd
}

func (spec devicesSetCommandSpec) name() string {
	if spec.alias != "" {
		return spec.alias
	}

	return strings.ReplaceAll(string(spec.instance), "_", "-")
}

func runDevicesSet(cmd *cobra.Command, rootOpts *options, devicesOpts *devicesOptions, deviceID string, command devices.CapabilityCommand) error {
	gateway, err := loadDeviceGateway(rootOpts.configPath)
	if err != nil {
		return err
	}

	if err := gateway.SendCommand(cmd.Context(), deviceID, command); err != nil {
		return err
	}

	if devicesOpts.outputJSON {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")

		return encoder.Encode(devicesSetResult{Sent: true})
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Command sent")

	return nil
}

func parseOnOffCommand(instance devices.CapabilityInstance, state string) (devices.CapabilityCommand, error) {
	switch state {
	case "on":
		return devices.NewOnOffCommand(instance, true), nil
	case "off":
		return devices.NewOnOffCommand(instance, false), nil
	default:
		return devices.CapabilityCommand{}, fmt.Errorf("%s state must be on or off", instance)
	}
}

func parseRangeCommand(name string, instance devices.CapabilityInstance, state string) (devices.CapabilityCommand, error) {
	value, err := strconv.ParseFloat(state, 64)
	if err != nil {
		return devices.CapabilityCommand{}, fmt.Errorf("%s state must be a number: %w", name, err)
	}

	command := devices.NewRangeCommand(instance, value)
	if err := command.Validate(); err != nil {
		return devices.CapabilityCommand{}, err
	}

	return command, nil
}

func parseColorCommand(cmd *cobra.Command, instance devices.CapabilityInstance, opts *devicesSetColorOptions) (devices.CapabilityCommand, error) {
	for _, name := range []string{"hue", "saturation", "value"} {
		if !cmd.Flags().Changed(name) {
			return devices.CapabilityCommand{}, fmt.Errorf("color requires --%s", name)
		}
	}

	command := devices.NewColorCommand(instance, devices.HSVColor{
		Hue:        opts.hue,
		Saturation: opts.saturation,
		Value:      opts.value,
	})
	if err := command.Validate(); err != nil {
		return devices.CapabilityCommand{}, err
	}

	return command, nil
}

func parseModeCommand(instance devices.CapabilityInstance, state string) (devices.CapabilityCommand, error) {
	command := devices.NewModeCommand(instance, state)
	if err := command.Validate(); err != nil {
		return devices.CapabilityCommand{}, err
	}

	return command, nil
}
