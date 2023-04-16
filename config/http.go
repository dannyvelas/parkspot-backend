package config

import (
	"fmt"
	"os"
	"time"
)

type HttpConfig struct {
	Domain             string
	Port               string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	CORSAllowedOrigins []string
}

func newHttpConfig() (HttpConfig, error) {
	var domain string
	if domain = os.Getenv("DOMAIN"); domain == "" {
		return HttpConfig{}, fmt.Errorf("DOMAIN is required.")
	}

	return HttpConfig{
		Domain:             domain,
		Port:               readEnvString("PORT", "5000"),
		ReadTimeout:        readEnvDuration("READTIMEOUT", 5),
		WriteTimeout:       readEnvDuration("WRITETIMEOUT", 10),
		IdleTimeout:        readEnvDuration("IDLETIMEOUT", 120),
		CORSAllowedOrigins: readEnvStringList("CORSALLOWEDORIGINS", []string{"http://*"}),
	}, nil
}
