package cli

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/skel2007/smart-bridge/internal/config"
	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/skel2007/smart-bridge/internal/tuya"
)

type devicesListOptions struct {
	outputJSON bool
}

func newDevicesCommand(rootOpts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "devices",
		Short: "Inspect Tuya devices",
	}

	cmd.AddCommand(newDevicesListCommand(rootOpts))

	return cmd
}

func newDevicesListCommand(rootOpts *options) *cobra.Command {
	opts := &devicesListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List devices from Tuya Cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDevicesList(cmd, rootOpts, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.outputJSON, "json", false, "print devices as JSON")

	return cmd
}

func runDevicesList(cmd *cobra.Command, rootOpts *options, opts *devicesListOptions) error {
	cfg, err := config.Load(rootOpts.configPath)
	if err != nil {
		return err
	}

	client := tuya.NewClient(tuya.Credentials{
		Endpoint:     cfg.Tuya.Endpoint,
		ClientID:     cfg.Tuya.ClientID,
		ClientSecret: cfg.Tuya.ClientSecret,
	})

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
