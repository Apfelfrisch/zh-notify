package main

import (
	"fmt"

	"github.com/apfelfrisch/zh-notify/cmd"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

func initConfig() {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Could not read config file: %s\n", err))
	}
}

func main() {
	initConfig()
	cmd.Execute()
}
