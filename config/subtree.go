package config

import (
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// GetStringMapSub returns a map of subtrees at a key
// can also apply defaults if required
func GetStringMapSub(key string, withDefaultsApplied bool) map[string]*viper.Viper {
	stringMap := cast.ToStringMap(getResolvedStringMapValue(key))
	newMap := map[string]*viper.Viper{}
	defaults := viper.GetStringMap("defaults")
	for k := range stringMap {
		sub := viper.Sub(key + "." + k)
		if withDefaultsApplied {
			for k, v := range defaults {
				switch cv := v.(type) {
				case map[string]string:
					defaultsMap := cv
					for k, v := range sub.GetStringMapString(k) {
						defaultsMap[k] = v
					}
					sub.Set(k, v)
				case map[string]interface{}:
					defaultsMap := cv
					for k, v := range sub.GetStringMap(k) {
						defaultsMap[k] = v
					}
					sub.Set(k, v)
				default:
					sub.SetDefault(k, v)
				}
			}
		}
		newMap[k] = sub
	}
	return newMap
}