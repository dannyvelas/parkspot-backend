package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type HttpConfig struct {
	port         uint
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func newHttpConfig() (HttpConfig, error) {
	var httpConfig HttpConfig

	if portString := os.Getenv("HTTP_PORT"); portString == "" {
		return HttpConfig{}, varNotFoundError("HTTP_PORT")
	} else if parsed, err := strconv.ParseUint(portString, 10, 64); err != nil {
		return HttpConfig{}, errors.New("HTTP_PORT could not be converted to an integer.")
	} else {
		httpConfig.port = uint(parsed)
	}

	if readTimeoutString := os.Getenv("HTTP_READTIMEOUT"); readTimeoutString == "" {
		return HttpConfig{}, varNotFoundError("HTTP_READTIMEOUT")
	} else if parsed, err := time.ParseDuration(readTimeoutString); err != nil {
		return HttpConfig{}, errors.New("HTTP_READTIMEOUT could not be converted to an integer.")
	} else {
		httpConfig.readTimeout = parsed
	}

	if writeTimeoutString := os.Getenv("HTTP_WRITETIMEOUT"); writeTimeoutString == "" {
		return HttpConfig{}, varNotFoundError("HTTP_WRITETIMEOUT")
	} else if parsed, err := time.ParseDuration(writeTimeoutString); err != nil {
		return HttpConfig{}, errors.New("HTTP_WRITETIMEOUT could not be converted to an integer.")
	} else {
		httpConfig.writeTimeout = parsed
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
