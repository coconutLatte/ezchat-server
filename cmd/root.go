package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coconutLatte/ezchat-server/config"
	"github.com/coconutLatte/ezchat-server/db"
	"github.com/coconutLatte/ezchat-server/logger"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ezchat-server",
	Short: "ezchat-server",
	Long:  "ezchat-server",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := config.Init(cfgFile)
		if err != nil {
			fmt.Println("failed to initialize config")
			return err
		}

		err = logger.Init(logger.Options{
			Level:      config.Get().Log.Level,
			Encoding:   config.Get().Log.Encoding,
			File:       config.Get().Log.File,
			MaxSizeMB:  config.Get().Log.MaxSizeMB,
			MaxBackups: config.Get().Log.MaxBackups,
			MaxAgeDays: config.Get().Log.MaxAgeDays,
			Compress:   config.Get().Log.Compress,
			Console:    config.Get().Log.Console,
		})
		if err != nil {
			fmt.Println("failed to initialize logger")
			return err
		}
		logger.S().Info("logger initialized")

		err = db.Init(db.Options{
			DSN:            config.Get().DB.DSN,
			MaxOpenConns:   config.Get().DB.MaxOpenConns,
			MaxIdleConns:   config.Get().DB.MaxIdleConns,
			ConnMaxIdleSec: config.Get().DB.ConnMaxIdleSec,
			ConnMaxLifeSec: config.Get().DB.ConnMaxLifeSec,
		})

		if err != nil {
			logger.S().Error("failed to initialize database")
			return err
		}
		logger.S().Info("database initialized")

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		if cfg == nil {
			logger.S().Error("no config loaded")
			return errors.New("no config loaded")
		}

		// Setup Gin router
		r := gin.New()
		r.Use(gin.Recovery())
		r.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
		srv := &http.Server{Addr: addr, Handler: r}

		// start server
		go func() {
			logger.S().Infow("http server starting", "addr", addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.S().Errorw("http server error", "error", err)
			}
		}()

		// graceful shutdown
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logger.S().Info("shutting down http server")

		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.S().Errorw("http server shutdown error", "error", err)
			return err
		}
		logger.S().Info("http server exited gracefully")
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Persistent flag for config file
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path (supports .yaml/.yml/.json/.toml)")

	// Environment variable support
	// keep here for backward compatibility if other code relied on it
	viper.SetEnvPrefix("EZCHAT")
	viper.AutomaticEnv()
}
