package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"srd/internal/cache"
	"srd/internal/config"
	"srd/internal/resolver"
	"srd/internal/server"
)

var (
	rootCmd = &cobra.Command{
		Use:   "srd",
		Short: "SRD application",
		Run:   run,
	}
)

func init() {
	// proactivly set the time field format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	cobra.OnInitialize(initConfig)
	config.SetupFlags(rootCmd)
}

func initLogger(debug bool, logCfg config.LogConfig) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if logCfg.Pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		log.Info().Msg("pretty log output enabled")
	}

	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug mode enabled")
	}
}

func initConfig() {
	// Get the config file path from the command
	cfgFile, _ := rootCmd.PersistentFlags().GetString("config")
	config.InitConfig(cfgFile)

	cfg := config.GetConfig()

	// TODO: remove
	fmt.Printf("cfg: %+v\n", cfg)

	initLogger(cfg.Debug, cfg.Log)
}

func run(cmd *cobra.Command, args []string) {
	cfg := config.GetConfig()
	cache := cache.New(cfg.Cache)

	// pass config to resolver
	resolver.Init(cfg.Resolver, cache)

	if err := server.StartServer(cfg.Host, cfg.Port); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}
