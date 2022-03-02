package config

import (
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"time"
)

func readEnvString(envKey string, defaultValue string) (envValue string) {
	envValue = os.Getenv(envKey)
	if envValue == "" {
		log.Warn().Msg(notFoundError{envKey}.ErrorUsingDefault(defaultValue))
		envValue = defaultValue
	}
	return
}

func readEnvUint(envKey string, defaultValue uint) (envValue uint) {
	if envValueString := os.Getenv(envKey); envValueString == "" {
		log.Warn().Msg(notFoundError{envKey}.ErrorUsingDefault(defaultValue))
		envValue = defaultValue
	} else if parsed, err := strconv.ParseUint(envValueString, 10, 64); err != nil {
		log.Warn().Msg(conversionError{envKey, "uint"}.ErrorUsingDefault(defaultValue))
		envValue = defaultValue
	} else {
		envValue = uint(parsed)
	}
	return
}

func readEnvDuration(envKey string, defaultValue uint) (envValue time.Duration) {
	if envValueString := os.Getenv(envKey); envValueString == "" {
		log.Warn().Msg(notFoundError{envKey}.ErrorUsingDefault(defaultValue))
		envValue = time.Duration(defaultValue) * time.Second
	} else if parsed, err := time.ParseDuration(envValueString); err != nil {
		log.Warn().Msg(conversionError{envKey, "duration"}.ErrorUsingDefault(defaultValue))
		envValue = time.Duration(defaultValue) * time.Second
	} else {
		envValue = parsed * time.Second
	}
	return
}
