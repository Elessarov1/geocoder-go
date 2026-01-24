package grpc

import "time"

type Config struct {
	Host            string
	Port            int
	ShutdownTimeout time.Duration

	Reflection ReflectionConfig
}

type ReflectionConfig struct {
	Enabled bool
}
