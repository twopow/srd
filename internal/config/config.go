package config

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Host     string
	Port     int
	Debug    bool
	Resolver ResolverConfig
	Log      LogConfig
	Cache    CacheConfig
}

type LogConfig struct {
	Pretty bool
}

type ResolverConfig struct {
	RecordPrefix string
}

type CacheConfig struct {
	TTL             time.Duration
	CleanupInterval time.Duration
}

var (
	cfgFile string
)

// InitConfig initializes the configuration
func InitConfig(cfgFilePath string) {
	cfgFile = cfgFilePath

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Read environment variables
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err == nil {
		log.Info().Str("config", viper.ConfigFileUsed()).Msg("using config file")
	}
}

// SetupFlags sets up the command flags for configuration
func SetupFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	cmd.PersistentFlags().String("host", "localhost", "server host")
	cmd.PersistentFlags().Int("port", 8080, "server port")
	cmd.PersistentFlags().Bool("debug", false, "enable debug mode")
	cmd.PersistentFlags().String("resolver-record-prefix", "_srd", "dns record prefix")
	cmd.PersistentFlags().Bool("log-pretty", false, "enable pretty log output")
	cmd.PersistentFlags().Duration("cache-ttl", 300*time.Second, "cache ttl")
	cmd.PersistentFlags().Duration("cache-cleanup-interval", 600*time.Second, "cache cleanup interval")

	// Bind flags to viper
	viper.BindPFlag("host", cmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("port", cmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("resolver-record-prefix", cmd.PersistentFlags().Lookup("resolver-record-prefix"))
	viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("log-pretty", cmd.PersistentFlags().Lookup("log-pretty"))
	viper.BindPFlag("cache-ttl", cmd.PersistentFlags().Lookup("cache-ttl"))
	viper.BindPFlag("cache-cleanup-interval", cmd.PersistentFlags().Lookup("cache-cleanup-interval"))
}

// GetConfig returns the application configuration
func GetConfig() Config {
	return Config{
		Host:  viper.GetString("host"),
		Port:  viper.GetInt("port"),
		Debug: viper.GetBool("debug"),
		Resolver: ResolverConfig{
			RecordPrefix: viper.GetString("resolver-record-prefix"),
		},
		Log: LogConfig{
			Pretty: viper.GetBool("log-pretty"),
		},
		Cache: CacheConfig{
			TTL:             viper.GetDuration("cache-ttl"),
			CleanupInterval: viper.GetDuration("cache-cleanup-interval"),
		},
	}
}
