package config

type Config struct {
	GeoCoder GeoCoderConfig
	Server   ServerConfig
}

type GeoCoderConfig struct {
	GeoIPDbPath string `env:"GEOIP_DATABASE_PATH" default:"db/RU-GeoIP-Country.mmdb"`
	Debug       bool   `env:"GEOCODER_DEBUG" default:"false"`
}

type ServerConfig struct {
	Host        string `env:"GEOCODER_SERVER_HOST" validate:"required,host"`
	Port        int    `env:"GEOCODER_SERVER_PORT" validate:"required,port"`
	Swagger     bool   `env:"GEOCODER_SERVER_SWAGGER" default:"false"`
	CorsEnabled bool   `env:"GEOCODER_CORS_ENABLED" default:"false"`
}
