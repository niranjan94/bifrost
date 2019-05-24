package config

import (
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"time"
)

func GetSub(key string) *viper.Viper {
	return viper.Sub(key)
}

func Get(key string) interface{} {
	return getResolvedValue(key)
}

func GetString(key string) string {
	return cast.ToString(getResolvedValue(key))
}

func GetBool(key string) bool {
	return cast.ToBool(getResolvedValue(key))
}

func GetFloat64(key string) float64 {
	return cast.ToFloat64(getResolvedValue(key))
}

func GetInt(key string) int {
	return cast.ToInt(getResolvedValue(key))
}

func GetStringMap(key string) map[string]interface{} {
	return cast.ToStringMap(
		MemoizedFn(key, "getResolvedStringMapValue", getResolvedStringMapValue),
	)
}

func GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(
		MemoizedFn(key, "getResolvedStringMapStringValue", getResolvedStringMapStringValue),
	)
}

func GetStringSlice(key string) []string {
	return cast.ToStringSlice(getResolvedValue(key))
}

func GetTime(key string) time.Time {
	return cast.ToTime(getResolvedValue(key))
}

func GetDuration(key string) time.Duration {
	return cast.ToDuration(getResolvedValue(key))
}

func IsSet(key string) bool {
	return viper.IsSet(key)
}

func All() map[string]interface{} {
	return viper.AllSettings()
}
