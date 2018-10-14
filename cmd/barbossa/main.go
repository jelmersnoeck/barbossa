package main

import (
	"github.com/jelmersnoeck/barbossa/internal/webhook/cmd"
	"github.com/spf13/viper"
)

func main() {
	viper.SetEnvPrefix("barbossa")

	cmd.Execute()
}
