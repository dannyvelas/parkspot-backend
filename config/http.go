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
	defaultPort         = 5000
	defaultReadTimeout  = 5
	defaultWriteTimeout = 10
	defaultIdleTimeout  = 120
)

func newHttpConfig() HttpConfig {
	var httpConfig HttpConfig

	if portString := os.Getenv("HTTP_PORT"); portString == "" {
		log.Warn().Msg(notFoundError{"HTTP_PORT"}.ErrorUsingDefault(defaultPort))
		httpConfig.port = defaultPort
	} else if parsed, err := strconv.ParseUint(portString, 10, 64); err != nil {
		log.Warn().Msg(conversionError{"HTTP_PORT", "uint"}.ErrorUsingDefault(defaultPort))
		httpConfig.port = defaultPort
	} else if parsed < 80 {
		log.Warn().Msg(invalidError{"PG_PORT", "< 80"}.ErrorUsingDefault(defaultPort))
		httpConfig.port = defaultPort
	} else {
		httpConfig.port = uint(parsed)
	}

	if readTimeoutString := os.Getenv("HTTP_READTIMEOUT"); readTimeoutString == "" {
		log.Warn().Msg(notFoundError{"HTTP_READTIMEOUT"}.ErrorUsingDefault(defaultReadTimeout))
		httpConfig.readTimeout = defaultReadTimeout * time.Second
	} else if parsed, err := time.ParseDuration(readTimeoutString); err != nil {
		log.Warn().Msg(conversionError{"HTTP_READTIMEOUT", "duration"}.ErrorUsingDefault(defaultReadTimeout))
		httpConfig.readTimeout = defaultReadTimeout * time.Second
	} else {
		httpConfig.readTimeout = parsed * time.Second
	}

	if writeTimeoutString := os.Getenv("HTTP_WRITETIMEOUT"); writeTimeoutString == "" {
		log.Warn().Msg(notFoundError{"HTTP_WRITETIMEOUT"}.ErrorUsingDefault(defaultWriteTimeout))
		httpConfig.writeTimeout = defaultWriteTimeout * time.Second
	} else if parsed, err := time.ParseDuration(writeTimeoutString); err != nil {
		log.Warn().Msg(conversionError{"HTTP_WRITETIMEOUT", "duration"}.ErrorUsingDefault(defaultWriteTimeout))
		httpConfig.writeTimeout = defaultWriteTimeout * time.Second
	} else {
		httpConfig.writeTimeout = parsed * time.Second
	}

	if idleTimeoutString := os.Getenv("HTTP_IDLETIMEOUT"); idleTimeoutString == "" {
		log.Warn().Msg(notFoundError{"HTTP_IDLETIMEOUT"}.ErrorUsingDefault(defaultIdleTimeout))
		httpConfig.idleTimeout = defaultIdleTimeout * time.Second
	} else if parsed, err := time.ParseDuration(idleTimeoutString); err != nil {
		log.Warn().Msg(conversionError{"HTTP_IDLETIMEOUT", "duration"}.ErrorUsingDefault(defaultIdleTimeout))
		httpConfig.idleTimeout = defaultIdleTimeout * time.Second
	} else {
		httpConfig.idleTimeout = parsed * time.Second
	}

	return httpConfig
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
