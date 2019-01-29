package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	squashHomeDirName    = ".squash"
	squashConfigFileName = "config.yaml"
)

var defaultConfigYaml = []byte(`# Squash configuration file
# The specification can be found at https://squash.solo.io
secure_mode: false
verbose: true
log_commands: false
createdby: squash-initialization
`)

func writeDefaultConfigFile(fp string) error {
	fmt.Printf("Squash config file not found. Writing default config to %v.\n", fp)
	if err := ioutil.WriteFile(fp, defaultConfigYaml, 0644); err != nil {
		return err
	}
	return nil
}

// Note on use of config file:
// - should implement a consistent way of reading all values
// and harmonizing them with flags
func (top *Options) readConfigValues(c *Config) error {
	// Only read the config once
	if top.Internal.ConfigLoaded {
		return nil
	}

	if err := top.prepareViperConfig(); err != nil {
		return err
	}

	// TODO(mitchdraft) - allow config to be overriden by flags
	c.verbose = viper.GetBool("verbose")
	c.secureMode = viper.GetBool("secure_mode")
	c.logCmds = viper.GetBool("log_commands")

	top.Internal.ConfigRead = true
	return nil
}

// This needs to be called before viper can read any config values
func (top *Options) prepareViperConfig() error {
	// only load the config once
	if top.Internal.ConfigLoaded {
		return nil
	}
	// read config file
	// TODO(mitchdraft) - get this from an optional flag
	cfgFile := ""
	if cfgFile != "" {
		// Use config file from the flag.
		top.printVerbosef("Reading squash config from %v\n", cfgFile)
		viper.SetConfigFile(cfgFile)
	} else {
		squashDir, err := squashDir()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(squashDir, 0755); err != nil {
			return err
		}
		squashConfigFile := filepath.Join(squashDir, squashConfigFileName)
		if _, err := os.Stat(squashConfigFile); err == nil {
			// path exists
			top.printVerbosef("Reading squash config from %v\n", squashConfigFile)
		} else {
			if err := writeDefaultConfigFile(squashConfigFile); err != nil {
				return err
			}
		}

		viper.SetConfigFile(squashConfigFile)
	}
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("Can't read config: %v", err)
	}
	top.Internal.ConfigLoaded = true
	return nil
}

func squashDir() (string, error) {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, squashHomeDirName), nil
}
