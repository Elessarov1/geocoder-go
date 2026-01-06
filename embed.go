package Geocoder

import (
	"embed"
)

//go:embed _openapi/openapi.yaml
var Swagger []byte

//go:embed _openapi/swaggerui/*
var SwaggerUI embed.FS
