package config

import (
	"log"
	"os"
	"strings"

	"github.com/danztran/telescope/pkg/collector"
	"github.com/danztran/telescope/pkg/mapnode"
	"github.com/danztran/telescope/pkg/promscope"
	"github.com/danztran/telescope/pkg/scope"
	"github.com/danztran/telescope/pkg/server"
	"github.com/spf13/viper"
)

var Values Config

type Config struct {
	Server    server.Config    `mapstructure:"server"`
	Scope     scope.Config     `mapstructure:"scope"`
	Collector collector.Config `mapstructure:"collector"`
	Promscope promscope.Config `mapstructure:"promscope"`
	Mapnode   mapnode.Config   `mapstructure:"mapnode"`
}

func init() {
	config := viper.New()
	config.SetConfigName("telescope-config") // config file name
	if configPath, ok := os.LookupEnv("TELESCOPE_CONFIG"); ok {
		config.AddConfigPath(configPath)
	}
	config.AddConfigPath(".")
	config.AddConfigPath("./config/")
	config.AddConfigPath("../config/")
	config.AddConfigPath("../../config/")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	config.AutomaticEnv()

	err := config.ReadInConfig()
	if err != nil {
		log.Fatalf("error read config / %s", err)
	}

	err = config.Unmarshal(&Values)
	if err != nil {
		log.Fatalf("error parse config / %s", err)
	}
}
