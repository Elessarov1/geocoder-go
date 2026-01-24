package keys

// common keys for all components
const (
	Enabled   = "enabled"
	DependsOn = "depends_on"
)

const (
	Server = "server"
	GRPC   = "grpc"
	// Kafka = "kafka"
	// Postgres = "postgres"
	// ClickHouse = "clickhouse"

)

// http/grpc server keys
const (
	Host            = "host"
	Port            = "port"
	ShutdownTimeout = "shutdown_timeout"
	Metrics         = "metrics"
	Swagger         = "swagger"
	CORS            = "cors"
	Path            = "path"
	YAMLPath        = "yaml_path"
	UIPath          = "ui_path"
	Reflection      = "reflection"
)
