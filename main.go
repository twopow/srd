package main

import (
	"os"

	"github.com/spf13/cobra"

	"srd/internal/cache"
	"srd/internal/config"
	"srd/internal/log"
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
	cobra.OnInitialize(initConfig)
	config.SetupFlags(rootCmd)
}

func initLogger(debug bool, logCfg config.LogConfig) {
	if debug {
		log.SetGlobalLevel(log.DebugLevel)
	} else {
		log.SetGlobalLevel(log.InfoLevel)
	}

	if logCfg.Pretty {
		log.SetPrettyOutput(os.Stderr)
		log.Info().Msg("pretty log output enabled")
	}
}

func initConfig() {
	// Get the config file path from the command
	cfgFile, _ := rootCmd.PersistentFlags().GetString("config")
	config.InitConfig(cfgFile)

	cfg := config.GetConfig()

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
