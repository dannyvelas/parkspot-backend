package config

import (
	"time"
)

type HttpConfig struct {
	host               string
	port               string
	readTimeout        time.Duration
	writeTimeout       time.Duration
	idleTimeout        time.Duration
	corsAllowedOrigins []string
}

const (
	defaultHttpHost         = "127.0.0.1"
	defaultHttpPort         = "5000"
	defaultHttpReadTimeout  = 5
	defaultHttpWriteTimeout = 10
	defaultHttpIdleTimeout  = 120
)

var (
	defaultCORSAllowedOrigins = []string{"http://*"}
)

func newHttpConfig() HttpConfig {
	return HttpConfig{
		host:               readEnvString("HOST", defaultHttpHost),
		port:               readEnvString("PORT", defaultHttpPort),
		readTimeout:        readEnvDuration("READTIMEOUT", defaultHttpReadTimeout),
		writeTimeout:       readEnvDuration("WRITETIMEOUT", defaultHttpWriteTimeout),
		idleTimeout:        readEnvDuration("IDLETIMEOUT", defaultHttpIdleTimeout),
		corsAllowedOrigins: readEnvStringList("CORSALLOWEDORIGINS", defaultCORSAllowedOrigins),
	}
}

func (httpConfig HttpConfig) Host() string {
	return httpConfig.host
}

func (httpConfig HttpConfig) Port() string {
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

func (httpConfig HttpConfig) CORSAllowedOrigins() []string {
	return httpConfig.corsAllowedOrigins
}
