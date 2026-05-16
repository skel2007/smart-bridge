package cli

import (
	"github.com/spf13/cobra"

	"github.com/skel2007/smart-bridge/internal/config"
)

type options struct {
	configPath string
	config     config.Config
}

func NewRootCommand() *cobra.Command {
	opts := &options{}

	cmd := &cobra.Command{
		Use:           "smart-bridge",
		Short:         "Tools for inspecting and bridging smart home devices",
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	cmd.PersistentFlags().StringVar(&opts.configPath, "config", "config.yaml", "path to config file")
	cmd.AddCommand(newDevicesCommand(opts))

	return cmd
}

func (opts *options) loadConfig() error {
	cfg, err := config.Load(opts.configPath)
	if err != nil {
		return err
	}

	opts.config = cfg

	return nil
}
