package config

import "github.com/kelseyhightower/envconfig"

// All env vars are prefixed with "mdserver"
// example: MDSERVER_DEBUG

// Config refers to general application configuration
type Config struct {
	Debug    bool   `envconfig:"debug" default:"false"`
	LogLevel string `envconfig:"loglevel" default:"info"`
	Addr     string `envconfig:"addr" default:"localhost:8080"`
	Dir      string `envconfig:"dir" default:"."`
	Theme    string `envconfig:"theme" default:"dark"`
}

// FromEnv pulls configration from environment variables
func FromEnv() (*Config, error) {
	var config Config
	err := envconfig.Process("mdserver", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
