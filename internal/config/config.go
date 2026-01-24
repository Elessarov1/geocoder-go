package config

import "github.com/Elessarov1/service-kit/component/server"
import "github.com/Elessarov1/service-kit/component/grpc"

type Config struct {
	GeoCoder GeoCoderConfig
	Server   server.Config
	GRPC     grpc.Config
}

type GeoCoderConfig struct {
	GeoIPDbPath string `env:"GEOIP_DATABASE_PATH" default:"db/RU-GeoIP-Country.mmdb"`
	Debug       bool   `env:"GEOCODER_DEBUG" default:"false"`
}
