package app

import (
	"context"
	"log"

	"github.com/spf13/viper"
)

var config *viper.Viper

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
		log.Println("No config loaded")
		return nil
	}
	config = viper.New()
	config.SetConfigFile(file)
	return config.ReadInConfig()
}

func contextWithConfig(ctx context.Context) context.Context {
	return context.WithValue(ctx, configContextKey{}, config)
}

func ConfigFromContext(ctx context.Context) *viper.Viper {
	cfg, _ := ctx.Value(configContextKey{}).(*viper.Viper)
	return cfg
}
