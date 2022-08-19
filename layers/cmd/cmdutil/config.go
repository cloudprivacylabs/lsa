package cmdutil

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

type Config struct {
	IndexedProperties []string `json:"indexedProperties" yaml:"indexedProperties"`
	SourceProperty    string   `json:"sourceProperty" yaml:"sourceProperty"`
	SourceRowProperty string   `json:"sourceRowProperty" yaml:"sourceRowProperty"`
}

var config Config

func InitConfig(cfg Config) {
	config = cfg
}

func GetConfig() Config {
	return config
}

func LoadConfig(cmd *cobra.Command) (Config, error) {
	var cfg Config
	// parse provided config file
	if cfgfile, _ := cmd.Flags().GetString("cfg"); len(cfgfile) > 0 {
		if err := ReadJSONOrYAML(cfgfile, &cfg); err != nil {
			return Config{}, err
		}
	} else {
		// search current directory
		err := ReadJSONOrYAML(".layers.config", &cfg)
		if err != nil {
			// otherwise search home directory
			homeDir, err := homedir.Dir()
			if err != nil {
				return Config{}, err
			}
			err = ReadJSONOrYAML(filepath.Join(homeDir, ".layers.config"), &cfg)
			if err != nil && errors.Is(err, os.ErrNotExist) {
				return Config{}, err
			}
		}
	}
	InitConfig(cfg)
	return cfg, nil
}
