package config

import (
	"github.com/spf13/viper"
)

func LoadDefaults() {
	stringMap := viper.GetStringMap("defaults")
	newMap := map[string]interface{}{}
	for k, v := range stringMap {
		switch v.(type) {
		case string:
			newMap[k] = getResolvedValue("defaults." + k)
		case map[string]string:
			newMap[k] = getResolvedStringMapStringValue("defaults." + k)
		case map[string]interface{}:
			newMap[k] = getResolvedStringMapValue("defaults." + k)
		default:
			newMap[k] = v
		}
	}
	viper.Set("defaults", newMap)
}
