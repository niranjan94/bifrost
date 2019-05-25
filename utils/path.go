package utils

import "os"

// GetCwd returns the current working directory and panics on errors.
func GetCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return cwd
}