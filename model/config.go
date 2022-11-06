package model

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port     int    `split_words:"true" default:"8080"`
	Method   string `split_words:"true" default:"html"`
	Endpoint string `split_words:"true" default:"http://wesense:88"`
	Interval int    `split_words:"true" default:"60"`
	Version  string `default:"0.1.0" envconfig:"SERVICE_VERSION"`
	Instance string `default:"wesense_exporter" envconfig:"SERVICE_INSTANCE"`
}

func LoadConfig(appName string) Config {
	var config Config
	err := envconfig.Process(appName, &config)
	if err != nil {
		panic(err)
	}
	return config
}
