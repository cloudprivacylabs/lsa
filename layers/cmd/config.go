package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/layers/cmd/config"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

func loadConfig(cmd *cobra.Command) (config.Config, error) {
	var cfg config.Config
	// parse provided config file
	if cfgfile, _ := cmd.Flags().GetString("cfg"); len(cfgfile) > 0 {
		if err := cmdutil.ReadJSONOrYAML(cfgfile, &cfg); err != nil {
			return config.Config{}, err
		}
	} else {
		// search current directory
		err := cmdutil.ReadJSONOrYAML("/.layers.config", &cfg)
		if err != nil {
			// otherwise search home directory
			homeDir, err := homedir.Dir()
			if err != nil {
				return config.Config{}, err
			}
			err = cmdutil.ReadJSONOrYAML(fmt.Sprintf(homeDir+"./layers.config"), &cfg)
			if err != nil && errors.Is(err, os.ErrNotExist) {
				return config.Config{}, err
			}
		}
	}
	return cfg, nil
}

var configCmd = &cobra.Command{
	Use:   "cfg",
	Short: "Load config",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig(cmd)
		if err != nil {
			return err
		}
		config.InitConfig(&cfg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.Flags().String("cfg", "", "configuration spec for node properties and labels (default: layers.config.yaml)")
}
