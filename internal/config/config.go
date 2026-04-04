package config

import (
	"fmt"
	"log/slog"
)

type Config struct {
	LogLevel slog.Level

	HTTPAddress string
	HTTPPort    int

	MetricsPath      string
	MetricsGoBuiltin bool
}

func GetConfig() (Config, error) {
	logLevel, err := parseEnvLogLevel("LOG_LEVEL", slog.LevelInfo)
	if err != nil {
		return Config{}, fmt.Errorf("Unable to parse environment variable '%s%s': %w", ENVIRONMENT_VARIABLE_PREFIX, "LOG_LEVEL", err)
	}

	httpPort, err := parseEnvNumber("HTTP_PORT", 8080)
	if err != nil {
		return Config{}, fmt.Errorf("Unable to parse environment variable '%s%s': %w", ENVIRONMENT_VARIABLE_PREFIX, "HTTP_PORT", err)
	}

	metricsGoBuiltin, err := parseEnvBool("METRICS_GO_BUILTIN", false)
	if err != nil {
		return Config{}, fmt.Errorf("Unable to parse environment variable '%s%s': %w", ENVIRONMENT_VARIABLE_PREFIX, "METRICS_GO_BUILTIN", err)
	}

	config := Config{
		LogLevel: logLevel,

		HTTPAddress: parseEnvString("HTTP_ADDRESS", "127.0.0.1"),
		HTTPPort:    httpPort,

		MetricsPath:      parseEnvString("METRICS_PATH", "/metrics"),
		MetricsGoBuiltin: metricsGoBuiltin,
	}

	slog.Debug("Loaded configuration",
		"log_level", config.LogLevel,

		"http_address", config.HTTPAddress,
		"http_port", config.HTTPPort,

		"metrics_path", config.MetricsPath,
		"metrics_go_builtin", config.MetricsGoBuiltin,
	)

	return config, nil
}
