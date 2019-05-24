package config

import (
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"regexp"
	"strings"
)

// varExpressionMatchRegex matches a valid reference string
var varExpressionMatchRegex = regexp.MustCompile(`(?i)\${([a-z:.]+)(,(\s+)?'(.+?)')?}`)

// getValueFromReference retrieves the value from config from a reference string
func getValueFromReference(reference string) interface{} {
	matches := varExpressionMatchRegex.FindStringSubmatch(reference)
	typeKeyPair := strings.Split(matches[1], ":")
	defaultValue := matches[4]
	if val := viper.GetString(typeKeyPair[1]); val != "" {
		return val
	}
	return defaultValue
}

// isReference checks if the string value looks like a reference string
func isReference(value string) bool {
	return value != "" && strings.HasPrefix(value, "${") && varExpressionMatchRegex.MatchString(value)
}

// getResolvedValue gets the value at a given key and also recursively resolves references as needed
func getResolvedValue(key string) interface{} {
	currentValue := viper.Get(key)
	currentValueString, isString := currentValue.(string)
	if !isString || !isReference(currentValueString) {
		return currentValue
	}
	return MemoizedFn(currentValueString, "getResolvedValue", getValueFromReference)
}

func getResolvedStringMapStringValue(key string) interface{} {
	return cast.ToStringMapString(getResolvedStringMapValue(key))
}

func getResolvedStringMapValue(key string) interface{} {
	mapValue := cast.ToStringMap(getResolvedValue(key))
	for k, v := range mapValue {
		stringValue, isString := v.(string)
		if isString && isReference(stringValue) {
			mapValue[k] = MemoizedFn(stringValue, "getResolvedValue", getValueFromReference)
		}
	}
	return mapValue
}
