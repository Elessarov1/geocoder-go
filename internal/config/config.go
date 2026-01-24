package config

import "github.com/Elessarov1/service-kit/component/server"

type Config struct {
	GeoCoder GeoCoderConfig
	Server   server.Config
	GRPC     GRPCConfig
}

type GeoCoderConfig struct {
	GeoIPDbPath string `env:"GEOIP_DATABASE_PATH" default:"db/RU-GeoIP-Country.mmdb"`
	Debug       bool   `env:"GEOCODER_DEBUG" default:"false"`
}

type GRPCConfig struct {
	Host       string `env:"GEOCODER_GRPC_HOST" validate:"required,host"`
	Port       int    `env:"GEOCODER_GRPC_PORT" validate:"required,port"`
	Reflection bool   `env:"GEOCODER_GRPC_REFLECTION" default:"true"`
}
