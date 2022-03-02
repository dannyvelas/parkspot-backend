package config

import (
	"time"
)

type HttpConfig struct {
	port         uint
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
}

const (
	defaultHttpPort         = 5000
	defaultHttpReadTimeout  = 5
	defaultHttpWriteTimeout = 10
	defaultHttpIdleTimeout  = 120
)

func newHttpConfig() HttpConfig {
	return HttpConfig{
		port:         readEnvUint("HTTP_PORT", defaultHttpPort),
		readTimeout:  readEnvDuration("HTTP_READTIMEOUT", defaultHttpReadTimeout),
		writeTimeout: readEnvDuration("HTTP_WRITETIMEOUT", defaultHttpWriteTimeout),
		idleTimeout:  readEnvDuration("HTTP_IDLETIMEOUT", defaultHttpIdleTimeout),
	}
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
