package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/AnarkisGaming/isabelle/types"
)

var (
	// Config defines the configuration in use by this Isabelle instance
	Config *types.Configuration
)

// Init reads and parses the config file into a config object accessible via isabelle/config
func Init() error {
	bytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		return err
	}

	if err = json.Unmarshal(bytes, &Config); err != nil {
		return err
	}

	return nil
}
