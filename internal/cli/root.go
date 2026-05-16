package cli

import (
	"github.com/spf13/cobra"
)

type options struct {
	configPath string
}

func NewRootCommand() *cobra.Command {
	opts := &options{}

	cmd := &cobra.Command{
		Use:           "smart-bridge",
		Short:         "Tools for inspecting and bridging smart home devices",
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	cmd.PersistentFlags().StringVar(&opts.configPath, "config", "", "path to config file")
	cmd.AddCommand(newDevicesCommand(opts))
	return cmd
}
