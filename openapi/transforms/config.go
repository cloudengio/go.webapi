// Copyright 2022 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package transforms

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the loaded transformer configuration.
type Config struct {
	Configs    []yaml.Node `yaml:"configs"`
	Transforms []string
}

// ConfigureAll configures all of the transformers currently registered.
func (c Config) ConfigureAll() error {
	for i, name := range c.Transforms {
		tfr, ok := installed[name]
		if !ok {
			return fmt.Errorf("transformer %v not installed", name)
		}
		if err := tfr.Configure(c.Configs[i]); err != nil {
			return err
		}
	}
	return nil
}

// LoadConfigFile loads the transform configuration from the
// specified YAML file.
func LoadConfigFile(filename string) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}
	return ParseConfig(data)
}

// ParseConfig parses the supplied YAML data to create an instance
// of Config.
func ParseConfig(data []byte) (Config, error) {
	cfg := Config{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	cfg.Transforms = make([]string, len(cfg.Configs))
	for i, c := range cfg.Configs {
		var tmp map[string]yaml.Node
		if err := c.Decode(&tmp); err != nil {
			return Config{}, err
		}
		for name, config := range tmp {
			cfg.Configs[i] = config
			cfg.Transforms[i] = name
			break
		}
	}
	return cfg, nil
}
