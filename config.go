package app

import (
	"log"

	"github.com/spf13/viper"
)

var config *viper.Viper

var configFile string

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
