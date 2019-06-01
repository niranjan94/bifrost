package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// LoadDefaults replaces the defaults with their resolved values so that they can be used in other parts directly
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
	viper.Set("dryRun", viper.GetBool("dry-run"))
	viper.Set("filter", viper.GetString("only"))
	viper.SetDefault("region", "ap-southeast-1")
	viper.SetDefault("serverless.rootDir", ".")
	if viper.GetBool("dryRun") {
		logrus.Warn("running in DRY RUN mode. No changes will be persisted to cloud service.")
	}

}
