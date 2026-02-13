package app

import (
	"context"
	"log/slog"

	"github.com/yeka-go/app/config"
)

var configuration config.Config

var configFile string

type configContextKey struct{}

func SetConfigFile(file string) {
	configFile = file
}

func initConfig(cfgFile string) error {
	file := configFile
	if cfgFile != "" {
		file = cfgFile
	}
	if file == "" {
		slog.Debug("No config loaded")
		return nil
	}

	cfg, err := config.FromFile(cfgFile)
	configuration = cfg
	return err
}

func contextWithConfig(ctx context.Context, cfg config.Config) context.Context {
	return context.WithValue(ctx, configContextKey{}, cfg)
}

func ConfigFromContext(ctx context.Context) config.Config {
	cfg, _ := ctx.Value(configContextKey{}).(config.Config)
	return cfg
}
