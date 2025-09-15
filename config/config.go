package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/coconutLatte/ezchat-server/db"
	"github.com/coconutLatte/ezchat-server/logger"
	"github.com/spf13/viper"
)

// Config mirrors the layout of config/config.example.yaml
type Config struct {
	HTTP HTTPConfig `mapstructure:"http" json:"http" yaml:"http"`
	Log  LogConfig  `mapstructure:"log" json:"log" yaml:"log"`
	DB   DBConfig   `mapstructure:"db" json:"db" yaml:"db"`
}

type HTTPConfig struct {
	Host string `mapstructure:"host" json:"host" yaml:"host"`
	Port int    `mapstructure:"port" json:"port" yaml:"port"`
}

type LogConfig struct {
	Level      string `mapstructure:"level" json:"level" yaml:"level"`
	Encoding   string `mapstructure:"encoding" json:"encoding" yaml:"encoding"`
	File       string `mapstructure:"file" json:"file" yaml:"file"`
	MaxSizeMB  int    `mapstructure:"max_size_mb" json:"max_size_mb" yaml:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups" json:"max_backups" yaml:"max_backups"`
	MaxAgeDays int    `mapstructure:"max_age_days" json:"max_age_days" yaml:"max_age_days"`
	Compress   bool   `mapstructure:"compress" json:"compress" yaml:"compress"`
	Console    bool   `mapstructure:"console" json:"console" yaml:"console"`
}

type DBConfig struct {
	DSN            string `mapstructure:"dsn" json:"dsn" yaml:"dsn"`
	MaxOpenConns   int    `mapstructure:"max_open_conns" json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns   int    `mapstructure:"max_idle_conns" json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxIdleSec int    `mapstructure:"conn_max_idle_sec" json:"conn_max_idle_sec" yaml:"conn_max_idle_sec"`
	ConnMaxLifeSec int    `mapstructure:"conn_max_life_sec" json:"conn_max_life_sec" yaml:"conn_max_life_sec"`
}

var current *Config

// InitFromViper reads the current Viper state into the package singleton.
func InitFromViper() error {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	current = &cfg
	return nil
}

// InitWithFile configures Viper (search paths, env), reads config, and unmarshals.
func InitWithFile(cfgFile string) error {
	viper.SetEnvPrefix("EZCHAT")
	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("./config")
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		if cfgFile != "" {
			return fmt.Errorf("failed to read config file %s: %w", cfgFile, err)
		}
		var perr viper.ConfigFileNotFoundError
		if !os.IsNotExist(err) && !errors.As(err, &perr) {
			return err
		}
		// allow empty
	}

	if err := InitFromViper(); err != nil {
		return err
	}

	// initialize logger using config via options (avoid import cycle)
	lopts := logger.Options{
		Level:      current.Log.Level,
		Encoding:   current.Log.Encoding,
		File:       current.Log.File,
		MaxSizeMB:  current.Log.MaxSizeMB,
		MaxBackups: current.Log.MaxBackups,
		MaxAgeDays: current.Log.MaxAgeDays,
		Compress:   current.Log.Compress,
		Console:    current.Log.Console,
	}
	if err := logger.Init(lopts); err != nil {
		return err
	}

	// initialize database and ping
	if current.DB.DSN != "" {
		dopts := db.Options{
			DSN:            current.DB.DSN,
			MaxOpenConns:   current.DB.MaxOpenConns,
			MaxIdleConns:   current.DB.MaxIdleConns,
			ConnMaxIdleSec: current.DB.ConnMaxIdleSec,
			ConnMaxLifeSec: current.DB.ConnMaxLifeSec,
		}
		if err := db.Init(dopts); err != nil {
			return err
		}
	}

	if path := viper.ConfigFileUsed(); path != "" {
		abs, _ := filepath.Abs(path)
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", abs)
	}
	return nil
}

// Init is the public initialization entrypoint used by the CLI.
// It accepts an optional file path and delegates to InitWithFile.
func Init(cfgFile string) error {
	return InitWithFile(cfgFile)
}

// Get returns the last initialized configuration.
// Call after InitFromViper has succeeded.
func Get() *Config {
	return current
}
