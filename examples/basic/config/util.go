package config

import (
	"os"
	"strconv"
)

func getEnv(key string) string {
	value := os.Getenv(key)
	return value
}

func getEnvInt(key string) int {
	value := os.Getenv(key)
	if value == "" {
		return 0
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return intValue
}
