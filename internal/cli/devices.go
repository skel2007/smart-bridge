package cli

import (
	"fmt"

	"github.com/spf13/cobra"
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

	cmd.Flags().BoolVar(&opts.outputJSON, "json", false, "print raw JSON response")
	return cmd
}

func runDevicesList(cmd *cobra.Command, rootOpts *options, opts *devicesListOptions) error {
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "devices list is not implemented yet (config=%s, json=%t)\n", rootOpts.configPath, opts.outputJSON)
	return nil
}
