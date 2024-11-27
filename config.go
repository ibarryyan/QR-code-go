package main

import (
	"github.com/spf13/viper"
)

var globalConfig *GlobalConfig

type GlobalConfig struct {
	Port     int
	TmpPath  string
	Domain   string
	TaskCron string
}

func GetGlobalConfig() *GlobalConfig {
	return globalConfig
}

func InitConfig() {
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	globalConfig = &GlobalConfig{
		Port:     viper.GetInt("server.port"),
		TmpPath:  viper.GetString("server.tmp-path"),
		Domain:   viper.GetString("server.domain"),
		TaskCron: viper.GetString("task.cron"),
	}
}
