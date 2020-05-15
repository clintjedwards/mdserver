//go:generate go run mdserver/generate.go

package main

import (
	"github.com/clintjedwards/mdserver/cli"
	"github.com/clintjedwards/mdserver/config"
	"github.com/rs/zerolog/log"
)

func main() {

	config, err := config.FromEnv()
	if err != nil {
		log.Fatal().Err(err).Msg("could not load env config")
	}

	setupLogging(config.LogLevel, config.Debug)

	cli.RootCmd.Execute()
}
