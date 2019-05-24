package config

import (
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

func GetStringMapSub(key string, withDefaultsApplied bool) map[string]*viper.Viper {
	stringMap := cast.ToStringMap(getResolvedStringMapValue(key))
	newMap := map[string]*viper.Viper{}
	for k := range stringMap {
		sub := viper.Sub(key + "." + k)
		if withDefaultsApplied {
			for k, v := range viper.GetStringMap("defaults") {
				sub.SetDefault(k, v)
			}
		}
		newMap[k] = sub
	}
	return newMap
}