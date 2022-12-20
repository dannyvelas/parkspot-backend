package config

import (
	"fmt"
	"os"
	"time"
)

type HttpConfig struct {
	domain             string
	port               string
	readTimeout        time.Duration
	writeTimeout       time.Duration
	idleTimeout        time.Duration
	corsAllowedOrigins []string
}

func newHttpConfig() (HttpConfig, error) {
	var domain string
	if domain = os.Getenv("DOMAIN"); domain == "" {
		return HttpConfig{}, fmt.Errorf("DOMAIN is required.")
	}

	return HttpConfig{
		domain:             domain,
		port:               readEnvString("PORT", "5000"),
		readTimeout:        readEnvDuration("READTIMEOUT", 5),
		writeTimeout:       readEnvDuration("WRITETIMEOUT", 10),
		idleTimeout:        readEnvDuration("IDLETIMEOUT", 120),
		corsAllowedOrigins: readEnvStringList("CORSALLOWEDORIGINS", []string{"http://*"}),
	}, nil
}

func (httpConfig HttpConfig) Domain() string {
	return httpConfig.domain
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

func (httpConfig HttpConfig) IDleTimeout() time.Duration {
	return httpConfig.idleTimeout
}

func (httpConfig HttpConfig) CORSAllowedOrigins() []string {
	return httpConfig.corsAllowedOrigins
}
