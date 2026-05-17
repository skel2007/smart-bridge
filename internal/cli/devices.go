package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/skel2007/smart-bridge/internal/config"
	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/skel2007/smart-bridge/internal/tuya"
)

type devicesOptions struct {
	outputJSON bool
}

func newDevicesCommand(rootOpts *options) *cobra.Command {
	opts := &devicesOptions{}

	cmd := &cobra.Command{
		Use:   "devices",
		Short: "Inspect Tuya devices",
	}

	cmd.PersistentFlags().BoolVar(&opts.outputJSON, "json", false, "print output as JSON")
	cmd.AddCommand(newDevicesListCommand(rootOpts, opts))
	cmd.AddCommand(newDevicesCapabilitiesCommand(rootOpts, opts))
	cmd.AddCommand(newDevicesSetCommand(rootOpts, opts))

	return cmd
}

func newDevicesListCommand(rootOpts *options, opts *devicesOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List devices from Tuya Cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDevicesList(cmd, rootOpts, opts)
		},
	}

	return cmd
}

func newDevicesCapabilitiesCommand(rootOpts *options, opts *devicesOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "capabilities <device-id>",
		Short: "List device capabilities from Tuya Cloud",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDevicesCapabilities(cmd, rootOpts, opts, args[0])
		},
	}

	return cmd
}

func runDevicesList(cmd *cobra.Command, rootOpts *options, opts *devicesOptions) error {
	client, err := newTuyaClient(rootOpts.configPath)
	if err != nil {
		return err
	}

	deviceList, err := client.ListDevices(cmd.Context())
	if err != nil {
		return err
	}

	if opts.outputJSON {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")

		return encoder.Encode(deviceList)
	}

	printDevicesTable(cmd, deviceList)

	return nil
}

func runDevicesCapabilities(cmd *cobra.Command, rootOpts *options, opts *devicesOptions, deviceID string) error {
	client, err := newTuyaClient(rootOpts.configPath)
	if err != nil {
		return err
	}

	capabilities, err := client.ListCapabilities(cmd.Context(), deviceID)
	if err != nil {
		return err
	}

	if opts.outputJSON {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")

		return encoder.Encode(capabilities)
	}

	printCapabilitiesTable(cmd, capabilities)

	return nil
}

func newTuyaClient(configPath string) (*tuya.Client, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	return tuya.NewClient(tuya.Credentials{
		Endpoint:     cfg.Tuya.Endpoint,
		ClientID:     cfg.Tuya.ClientID,
		ClientSecret: cfg.Tuya.ClientSecret,
	}), nil
}

func printDevicesTable(cmd *cobra.Command, deviceList []devices.Device) {
	if len(deviceList) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No devices found")
		return
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(writer, "ID\tNAME\tTYPE\tONLINE")
	for _, device := range deviceList {
		_, _ = fmt.Fprintf(writer, "%s\t%s\t%s\t%t\n", device.ID, device.Name, device.Type, device.Online)
	}
	_ = writer.Flush()
}

func printCapabilitiesTable(cmd *cobra.Command, capabilities []devices.Capability) {
	if len(capabilities) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No capabilities found")
		return
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(writer, "TYPE\tINSTANCE\tSTATE\tPARAMETERS")
	for _, capability := range capabilities {
		_, _ = fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\n",
			capability.Type,
			capability.Instance,
			capabilityState(capability),
			capabilityParameters(capability),
		)
	}
	_ = writer.Flush()
}

func capabilityState(capability devices.Capability) string {
	switch {
	case capability.OnOff != nil && capability.OnOff.State != nil:
		return fmt.Sprintf("%t", *capability.OnOff.State)
	case capability.Range != nil && capability.Range.State != nil:
		return fmt.Sprintf("%g", *capability.Range.State)
	case capability.Color != nil && capability.Color.State != nil:
		color := capability.Color.State
		return fmt.Sprintf("h=%g s=%g v=%g", color.Hue, color.Saturation, color.Value)
	case capability.Mode != nil && capability.Mode.State != nil:
		return *capability.Mode.State
	default:
		return "-"
	}
}

func capabilityParameters(capability devices.Capability) string {
	switch {
	case capability.Range != nil:
		parameters := capability.Range.Parameters
		return fmt.Sprintf("min=%g max=%g step=%g", parameters.Min, parameters.Max, parameters.Precision)
	case capability.Mode != nil && len(capability.Mode.Parameters.Modes) > 0:
		return "modes=" + strings.Join(capability.Mode.Parameters.Modes, ",")
	default:
		return "-"
	}
}
