package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr        string
	LogLevel        string
	ShutdownTimeout time.Duration

	DatabaseURL              string
	DatabaseMaxConns         int32
	DatabaseMinConns         int32
	DatabaseMaxConnLifetime  time.Duration
	DatabaseConnectTimeout   time.Duration
	DatabaseQueryTimeout     time.Duration
	ServiceVersion           string
	OTELExporterOTLPEndpoint string
	OTELExporterOTLPProtocol string
	OTELServiceName          string
	OTELResourceAttributes   string
	OTELTracesSampler        string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	var cfg Config

	cfg.HTTPAddr = os.Getenv("HTTP_ADDR")
	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR is required")
	}

	cfg.LogLevel = strings.ToLower(os.Getenv("LOG_LEVEL"))
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return Config{}, fmt.Errorf(
			"invalid LOG_LEVEL %q: must be debug, info, warn or error",
			cfg.LogLevel,
		)
	}

	shutdownTimeout, err := time.ParseDuration(os.Getenv("SHUTDOWN_TIMEOUT"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid SHUTDOWN_TIMEOUT: %w", err)
	}
	if shutdownTimeout <= 0 {
		return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT must be greater than 0")
	}
	cfg.ShutdownTimeout = shutdownTimeout

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	maxConns, err := strconv.Atoi(os.Getenv("DATABASE_MAX_CONNS"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid DATABASE_MAX_CONNS: %w", err)
	}
	if maxConns <= 0 {
		return Config{}, fmt.Errorf("DATABASE_MAX_CONNS must be greater than 0")
	}
	cfg.DatabaseMaxConns = int32(maxConns) //nolint:gosec // maxConns is validated to be > 0 and fits in int32

	minConns, err := strconv.Atoi(os.Getenv("DATABASE_MIN_CONNS"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid DATABASE_MIN_CONNS: %w", err)
	}
	if minConns < 0 {
		return Config{}, fmt.Errorf("DATABASE_MIN_CONNS must be greater than or equal to 0")
	}
	cfg.DatabaseMinConns = int32(minConns) //nolint:gosec // minConns is validated to be >= 0 and fits in int32

	if cfg.DatabaseMinConns > cfg.DatabaseMaxConns {
		return Config{}, fmt.Errorf(
			"DATABASE_MIN_CONNS must not be greater than DATABASE_MAX_CONNS",
		)
	}

	maxConnLifetime, err := time.ParseDuration(os.Getenv("DATABASE_MAX_CONN_LIFETIME"))
	if err != nil {
		return Config{}, fmt.Errorf(
			"invalid DATABASE_MAX_CONN_LIFETIME: %w",
			err,
		)
	}
	if maxConnLifetime <= 0 {
		return Config{}, fmt.Errorf(
			"DATABASE_MAX_CONN_LIFETIME must be greater than 0",
		)
	}
	cfg.DatabaseMaxConnLifetime = maxConnLifetime

	connectTimeout, err := time.ParseDuration(os.Getenv("DATABASE_CONNECT_TIMEOUT"))
	if err != nil {
		return Config{}, fmt.Errorf(
			"invalid DATABASE_CONNECT_TIMEOUT: %w",
			err,
		)
	}
	if connectTimeout <= 0 {
		return Config{}, fmt.Errorf(
			"DATABASE_CONNECT_TIMEOUT must be greater than 0",
		)
	}
	cfg.DatabaseConnectTimeout = connectTimeout

	queryTimeout, err := time.ParseDuration(os.Getenv("DATABASE_QUERY_TIMEOUT"))
	if err != nil {
		return Config{}, fmt.Errorf(
			"invalid DATABASE_QUERY_TIMEOUT: %w",
			err,
		)
	}
	if queryTimeout <= 0 {
		return Config{}, fmt.Errorf(
			"DATABASE_QUERY_TIMEOUT must be greater than 0",
		)
	}
	cfg.DatabaseQueryTimeout = queryTimeout
	cfg.ServiceVersion = os.Getenv("SERVICE_VERSION")
	if cfg.ServiceVersion == "" {
		cfg.ServiceVersion = "dev"
	}

	cfg.OTELServiceName = os.Getenv("OTEL_SERVICE_NAME")
	if cfg.OTELServiceName == "" {
		cfg.OTELServiceName = "trip-service"
	}

	cfg.OTELExporterOTLPEndpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")

	cfg.OTELExporterOTLPProtocol = os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	if cfg.OTELExporterOTLPProtocol == "" {
		cfg.OTELExporterOTLPProtocol = "http/protobuf"
	}
	switch cfg.OTELExporterOTLPProtocol {
	case "http/protobuf", "grpc":
	default:
		return Config{}, fmt.Errorf(
			"invalid OTEL_EXPORTER_OTLP_PROTOCOL %q: must be http/protobuf or grpc",
			cfg.OTELExporterOTLPProtocol,
		)
	}

	cfg.OTELResourceAttributes = os.Getenv("OTEL_RESOURCE_ATTRIBUTES")

	cfg.OTELTracesSampler = os.Getenv("OTEL_TRACES_SAMPLER")
	if cfg.OTELTracesSampler == "" {
		cfg.OTELTracesSampler = "parentbased_always_on"
	}

	return cfg, nil
}
