package config

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"strings"
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
		conversionErr := newConversionError(envKey, "duration", defaultValue)
		log.Warn().Msg(fmt.Sprintf("%v: %v", conversionErr, err))
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
		conversionErr := newConversionError(envKey, "duration", defaultValue)
		log.Warn().Msg(fmt.Sprintf("%v: %v", conversionErr, err))
		return time.Duration(defaultValue) * time.Second
	}

	return parsed
}

func readEnvStringList(envKey string, defaultValue []string) []string {
	envValue := os.Getenv(envKey)
	if envValue == "" {
		log.Warn().Msg(newNotFoundError(envKey, defaultValue).Error())
		return defaultValue
	}

	values := strings.Split(envValue, ",")
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}

	return values
}
