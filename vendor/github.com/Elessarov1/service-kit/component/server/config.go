package server

import "time"

type Config struct {
	Host            string
	Port            int
	ShutdownTimeout time.Duration
	Metrics         MetricsConfig
	Swagger         SwaggerConfig
	CORS            CORSConfig
}

type MetricsConfig struct {
	Enabled bool
	Path    string
}

type SwaggerConfig struct {
	Enabled  bool
	YAMLPath string
	UIPath   string
}

type CORSConfig struct {
	Enabled bool
}
