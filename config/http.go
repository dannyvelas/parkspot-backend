package config

import (
	"time"
)

type HttpConfig struct {
	FrontendURL        string
	CookieDomain       string
	Port               string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	CORSAllowedOrigins []string
}

func newHttpConfig() (HttpConfig, error) {
	return HttpConfig{
		FrontendURL:        readEnvString("FRONTENDURL", "http://localhost:5173"),
		CookieDomain:       readEnvString("COOKIEDOMAIN", "localhost"),
		Port:               readEnvString("PORT", "5000"),
		ReadTimeout:        readEnvDuration("READTIMEOUT", 5),
		WriteTimeout:       readEnvDuration("WRITETIMEOUT", 10),
		IdleTimeout:        readEnvDuration("IDLETIMEOUT", 120),
		CORSAllowedOrigins: readEnvStringList("CORSALLOWEDORIGINS", []string{"http://*"}),
	}, nil
}
