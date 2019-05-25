package config

import (
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"time"
)

// Sub returns new Viper instance representing a sub tree of this instance.
// Sub is case-insensitive for a key.
func GetSub(key string) *viper.Viper {
	return viper.Sub(key)
}

// Get can retrieve any value given the key to use.
// Get is case-insensitive for a key.
// Get has the behavior of returning the value associated with the first
// place from where it is set. Viper will check in the following order:
// override, flag, env, config file, key/value store, default
//
// Get returns an interface. For a specific value use one of the Get____ methods.
func Get(key string) interface{} {
	return getResolvedValue(key)
}

// GetString returns the value associated with the key as a string.
func GetString(key string) string {
	return cast.ToString(getResolvedValue(key))
}

// GetBool returns the value associated with the key as a boolean.
func GetBool(key string) bool {
	return cast.ToBool(getResolvedValue(key))
}

// GetFloat64 returns the value associated with the key as a float64.
func GetFloat64(key string) float64 {
	return cast.ToFloat64(getResolvedValue(key))
}

// GetInt returns the value associated with the key as an integer.
func GetInt(key string) int {
	return cast.ToInt(getResolvedValue(key))
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func GetStringMap(key string) map[string]interface{} {
	return cast.ToStringMap(
		MemoizedFn(key, "getResolvedStringMapValue", getResolvedStringMapValue),
	)
}

// GetStringMapString returns the value associated with the key as a map of strings.
func GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(
		MemoizedFn(key, "getResolvedStringMapStringValue", getResolvedStringMapStringValue),
	)
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func GetStringSlice(key string) []string {
	return cast.ToStringSlice(getResolvedValue(key))
}

// GetTime returns the value associated with the key as time.
func GetTime(key string) time.Time {
	return cast.ToTime(getResolvedValue(key))
}

// GetDuration returns the value associated with the key as a duration.
func GetDuration(key string) time.Duration {
	return cast.ToDuration(getResolvedValue(key))
}

// IsSet checks to see if the key has been set in any of the data locations.
// IsSet is case-insensitive for a key.
func IsSet(key string) bool {
	return viper.IsSet(key)
}

// All merges all settings and returns them as a map[string]interface{}.
func All() map[string]interface{} {
	return viper.AllSettings()
}
