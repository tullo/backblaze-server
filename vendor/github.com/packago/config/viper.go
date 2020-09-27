package config

import (
	"github.com/spf13/viper"
)

var configs = make(map[string]*viper.Viper)

func File(names ...string) *viper.Viper {
	name := "config"
	if len(names) > 0 {
		name = names[0]
	}
	if configs[name] == nil {
		configs[name] = viper.New()
		configs[name].SetConfigName(name)
		configs[name].SetConfigType("json")
		configs[name].AddConfigPath(".")
		err := configs[name].ReadInConfig()
		if err != nil {
			panic(err)
		}
	}
	return configs[name]
}
