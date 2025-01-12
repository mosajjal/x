package main

import (
	"fmt"
)

// Config is the configuration for the application.
type Config struct {
	LogLevel string `hcl:"log_level"`

	Endpoints []EndpointCfg `hcl:"endpoint,block"`
}

// EndpointCfg is the configuration for an endpoint.
type EndpointCfg struct {
	Type           string   `hcl:"type,label"`
	Name           string   `hcl:"name,label"`
	Modes          []string `hcl:"modes"`
	HTTPListener   string   `hcl:"http_listener"`
	SocketListener string   `hcl:"socket_listener"`
	HTTPBasePath   string   `hcl:"http_base_path"`
	File           string   `hcl:"file"`
	FileFormat     string   `hcl:"file_format"`
	Inverted       bool     `hcl:"inverted"`
	AutoReload     uint     `hcl:"auto_reload"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	for _, e := range c.Endpoints {
		if e.File == "" {
			return fmt.Errorf("file is required for endpoint %s", e.Name)
		}
		if e.FileFormat == "" {
			return fmt.Errorf("file_format is required for endpoint %s", e.Name)
		}
		if e.AutoReload < 0 {
			return fmt.Errorf("auto_reload must be >= 0 for endpoint %s", e.Name)
		}
		for _, m := range e.Modes {
			if m != "http" && m != "socket" {
				return fmt.Errorf("invalid mode %s for endpoint %s", m, e.Name)
			}
		}
	}
	return nil
}
