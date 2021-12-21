package config

import (
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"time"
)

type HttpConfig struct {
	port         uint
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
}

const (
	defaultReadTimeout  = 5 * time.Second
	defaultWriteTimeout = 10 * time.Second
	defaultIdleTimeout  = 120 * time.Second
)

func newHttpConfig() (HttpConfig, error) {
	var httpConfig HttpConfig

	if portString := os.Getenv("HTTP_PORT"); portString == "" {
		return HttpConfig{}, NotFoundError{"HTTP_PORT"}
	} else if parsed, err := strconv.ParseUint(portString, 10, 64); err != nil {
		return HttpConfig{}, ConversionError{"HTTP_PORT", "uint"}
	} else {
		httpConfig.port = uint(parsed)
	}

	if readTimeoutString := os.Getenv("HTTP_READTIMEOUT"); readTimeoutString == "" {
		notFoundError := NotFoundError{"HTTP_READTIMEOUT"}.Error()
		log.Warn().Msgf("%s. Using Default of: %s seconds.", notFoundError, defaultReadTimeout)
		httpConfig.readTimeout = defaultReadTimeout
	} else if parsed, err := time.ParseDuration(readTimeoutString); err != nil {
		conversionError := ConversionError{"HTTP_READTIMEOUT", "duration"}.Error()
		log.Warn().Msgf("%s. Using Default of: %s seconds.", conversionError, defaultReadTimeout)
		httpConfig.readTimeout = defaultReadTimeout
	} else {
		httpConfig.readTimeout = parsed
	}

	if writeTimeoutString := os.Getenv("HTTP_WRITETIMEOUT"); writeTimeoutString == "" {
		notFoundError := NotFoundError{"HTTP_WRITETIMEOUT"}.Error()
		log.Warn().Msgf("%s. Using Default of: %s seconds.", notFoundError, defaultWriteTimeout)
		httpConfig.writeTimeout = defaultWriteTimeout
	} else if parsed, err := time.ParseDuration(writeTimeoutString); err != nil {
		conversionError := ConversionError{"HTTP_WRITETIMEOUT", "duration"}.Error()
		log.Warn().Msgf("%s. Using Default of: %s seconds.", conversionError, defaultWriteTimeout)
		httpConfig.writeTimeout = defaultWriteTimeout
	} else {
		httpConfig.writeTimeout = parsed
	}

	if idleTimeoutString := os.Getenv("HTTP_IDLETIMEOUT"); idleTimeoutString == "" {
		notFoundError := NotFoundError{"HTTP_IDLETIMEOUT"}.Error()
		log.Warn().Msgf("%s. Using Default of: %s seconds.", notFoundError, defaultIdleTimeout)
		httpConfig.idleTimeout = defaultIdleTimeout
	} else if parsed, err := time.ParseDuration(idleTimeoutString); err != nil {
		conversionError := ConversionError{"HTTP_IDLETIMEOUT", "duration"}.Error()
		log.Warn().Msgf("%s. Using Default of: %s seconds.", conversionError, defaultIdleTimeout)
		httpConfig.idleTimeout = defaultIdleTimeout
	} else {
		httpConfig.idleTimeout = parsed
	}

	return httpConfig, nil
}

func (httpConfig HttpConfig) Port() uint {
	return httpConfig.port
}

func (httpConfig HttpConfig) ReadTimeout() time.Duration {
	return httpConfig.readTimeout
}

func (httpConfig HttpConfig) WriteTimeout() time.Duration {
	return httpConfig.writeTimeout
}

func (httpConfig HttpConfig) IdleTimeout() time.Duration {
	return httpConfig.idleTimeout
}
