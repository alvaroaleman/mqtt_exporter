package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/alvaroaleman/mqtt_exporter/internal"
)

func main() {
	o := internal.Opts{}
	var logLevel string
	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&o.ServerAddress, "server-address", "localhost", "MQTT server address")
	cmd.Flags().IntVar(&o.ServerPort, "server-port", 1883, "MQTT server port")
	cmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	cmd.Flags().StringSliceVar(&o.Topics, "topic", []string{"#"}, "MQTT topics to subscribe to, '#' for all opics")
	cmd.Flags().StringVar(&o.ConfigFile, "config-file", "", "Path to configuration file")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		log, err := setupLogger(logLevel)
		if err != nil {
			return fmt.Errorf("failed to configure logger: %w", err)
		}
		return internal.Run(o, log)
	}
	if err := cmd.Execute(); err != nil {
		fmt.Printf("error executing command: %v", err)
		os.Exit(1)
	}
}

func setupLogger(logLevel string) (*zap.Logger, error) {
	var level zapcore.Level
	switch logLevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "dpanic":
		level = zap.DPanicLevel
	case "panic":
		level = zap.PanicLevel
	case "fatal":
		level = zap.FatalLevel
	default:
		return nil, fmt.Errorf("unknown log level: %s", logLevel)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	return cfg.Build()
}
