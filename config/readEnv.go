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
		log.Warn().Msg(newNotFoundError(envKey, defaultValue).Error())
		return defaultValue
	}

	return envValue
}

func readEnvUint(envKey string, defaultValue uint) uint {
	envValueString := os.Getenv(envKey)
	if envValueString == "" {
		log.Warn().Msg(newNotFoundError(envKey, defaultValue).Error())
		return defaultValue
	}

	parsed, err := strconv.ParseUint(envValueString, 10, 64)
	if err != nil {
		log.Warn().Msg(newConversionError(envKey, "uint", defaultValue).Error())
		return defaultValue
	}

	return uint(parsed)
}

func readEnvDuration(envKey string, defaultValue uint) time.Duration {
	envValueString := os.Getenv(envKey)
	if envValueString == "" {
		log.Warn().Msg(newNotFoundError(envKey, defaultValue).Error())
		return time.Duration(defaultValue) * time.Second
	}

	parsed, err := time.ParseDuration(envValueString)
	if err != nil {
		log.Warn().Msg(newConversionError(envKey, "duration", defaultValue).Error())
		return time.Duration(defaultValue) * time.Second
	}

	return parsed * time.Second
}
