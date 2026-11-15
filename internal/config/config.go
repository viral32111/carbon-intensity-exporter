package config

import (
	"fmt"
	"log/slog"
	"strings"
	"unicode"
)

const (
	DEFAULT_LOG_LEVEL = slog.LevelInfo

	DEFAULT_HTTP_ADDRESS = "127.0.0.1"
	DEFAULT_HTTP_PORT    = 5000

	DEFAULT_METRICS_PATH       = "/metrics"
	DEFAULT_METRICS_GO_BUILTIN = false
)

type Config struct {
	LogLevel slog.Level

	HTTPAddress string
	HTTPPort    int

	MetricsPath      string
	MetricsGoBuiltin bool
}

func GetConfig() (Config, error) {
	logLevel, err := parseEnvLogLevel("LOG_LEVEL", DEFAULT_LOG_LEVEL)
	if err != nil {
		return Config{}, fmt.Errorf("Unable to parse environment variable '%s%s': %w", ENVIRONMENT_VARIABLE_PREFIX, "LOG_LEVEL", err)
	}

	httpPort, err := parseEnvNumber("HTTP_PORT", DEFAULT_HTTP_PORT)
	if err != nil {
		return Config{}, fmt.Errorf("Unable to parse environment variable '%s%s': %w", ENVIRONMENT_VARIABLE_PREFIX, "HTTP_PORT", err)
	}

	if httpPort < 1 || httpPort > 65535 {
		return Config{}, fmt.Errorf("Invalid value for environment variable '%s%s': Port %d must be between 1 and 65535", ENVIRONMENT_VARIABLE_PREFIX, "HTTP_PORT", httpPort)
	}

	metricsGoBuiltin, err := parseEnvBool("METRICS_GO_BUILTIN", DEFAULT_METRICS_GO_BUILTIN)
	if err != nil {
		return Config{}, fmt.Errorf("Unable to parse environment variable '%s%s': %w", ENVIRONMENT_VARIABLE_PREFIX, "METRICS_GO_BUILTIN", err)
	}

	metricsPath := parseEnvString("METRICS_PATH", DEFAULT_METRICS_PATH)

	if !strings.HasPrefix(metricsPath, "/") {
		return Config{}, fmt.Errorf("Invalid value for environment variable '%s%s': Path %q must start with a slash", ENVIRONMENT_VARIABLE_PREFIX, "METRICS_PATH", metricsPath)
	}

	if len(metricsPath) > 1 && strings.HasSuffix(metricsPath, "/") {
		return Config{}, fmt.Errorf("Invalid value for environment variable '%s%s': Path %q must not end with a slash", ENVIRONMENT_VARIABLE_PREFIX, "METRICS_PATH", metricsPath)
	}

	if strings.ContainsFunc(metricsPath, unicode.IsSpace) {
		return Config{}, fmt.Errorf("Invalid value for environment variable '%s%s': Path %q must not contain whitespace", ENVIRONMENT_VARIABLE_PREFIX, "METRICS_PATH", metricsPath)
	}

	config := Config{
		LogLevel: logLevel,

		HTTPAddress: parseEnvString("HTTP_ADDRESS", DEFAULT_HTTP_ADDRESS),
		HTTPPort:    httpPort,

		MetricsPath:      metricsPath,
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
