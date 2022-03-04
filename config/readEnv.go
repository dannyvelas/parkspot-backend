package config

import (
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"time"
)

func readEnvString(envKey string, defaultValue string) string {
	envValue := os.Getenv(envKey)
	if envValue == "" {
		log.Warn().Msg(notFoundError{envKey}.ErrorUsingDefault(defaultValue))
		return defaultValue
	}

	return envValue
}

func readEnvUint(envKey string, defaultValue uint) uint {
	envValueString := os.Getenv(envKey)
	if envValueString == "" {
		log.Warn().Msg(notFoundError{envKey}.ErrorUsingDefault(defaultValue))
		return defaultValue
	}

	parsed, err := strconv.ParseUint(envValueString, 10, 64)
	if err != nil {
		log.Warn().Msg(conversionError{envKey, "uint"}.ErrorUsingDefault(defaultValue))
		return defaultValue
	}

	return uint(parsed)
}

func readEnvDuration(envKey string, defaultValue uint) time.Duration {
	envValueString := os.Getenv(envKey)
	if envValueString == "" {
		log.Warn().Msg(notFoundError{envKey}.ErrorUsingDefault(defaultValue))
		return time.Duration(defaultValue) * time.Second
	}

	parsed, err := time.ParseDuration(envValueString)
	if err != nil {
		log.Warn().Msg(conversionError{envKey, "duration"}.ErrorUsingDefault(defaultValue))
		return time.Duration(defaultValue) * time.Second
	}

	return parsed * time.Second
}
