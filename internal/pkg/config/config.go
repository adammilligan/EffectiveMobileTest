package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig
	DB        DBConfig
	Log       LogConfig
	RateLimit RateLimitConfig
}

func (c Config) HTTPAddr() string {
	return c.Server.Host + ":" + c.Server.Port
}

func (c Config) PostgresDSN() string {
	sslMode := c.DB.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DB.User,
		c.DB.Password,
		c.DB.Host,
		c.DB.Port,
		c.DB.Name,
		sslMode,
	)
}

func Load() (Config, error) {
	cfg := defaultConfig()

	cfgFromYaml, err := loadFromYaml(envOrDefault("CONFIG_PATH", "configs/config.yaml"))
	if err != nil {
		return Config{}, err
	}

	cfg = merge(cfg, cfgFromYaml)
	cfg = applyEnvOverrides(cfg)

	if cfg.Server.Host == "" {
		return Config{}, errors.New("server.host is empty")
	}

	if cfg.Server.Port == "" {
		return Config{}, errors.New("server.port is empty")
	}

	if cfg.DB.Host == "" {
		return Config{}, errors.New("db.host is empty")
	}

	if cfg.DB.Port == "" {
		return Config{}, errors.New("db.port is empty")
	}

	if cfg.DB.Name == "" {
		return Config{}, errors.New("db.name is empty")
	}

	if cfg.DB.User == "" {
		return Config{}, errors.New("db.user is empty")
	}

	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: "8080",
		},
		Log: LogConfig{
			Level: "info",
		},
		DB: DBConfig{
			Host:     "postgres",
			Port:     "5432",
			Name:     "subscriptions",
			User:     "subscriptions",
			Password: "subscriptions",
			SSLMode:  "disable",
		},
		RateLimit: RateLimitConfig{
			IsEnabled:         true,
			RequestsPerMinute: 120,
			Burst:             20,
		},
	}
}

func loadFromYaml(path string) (Config, error) {
	if path == "" {
		return Config{}, errors.New("CONFIG_PATH is empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config yaml: %w", err)
	}

	var yc yamlConfig
	if unmarshalErr := yaml.Unmarshal(data, &yc); unmarshalErr != nil {
		return Config{}, fmt.Errorf("parse config yaml: %w", unmarshalErr)
	}

	var rateLimit RateLimitConfig
	if yc.RateLimit.IsEnabled != nil {
		rateLimit.IsEnabled = *yc.RateLimit.IsEnabled
		rateLimit.IsEnabledIsSet = true
	}
	if yc.RateLimit.RequestsPerMinute != nil {
		rateLimit.RequestsPerMinute = *yc.RateLimit.RequestsPerMinute
		rateLimit.IsRequestsPerMinuteIsSet = true
	}
	if yc.RateLimit.Burst != nil {
		rateLimit.Burst = *yc.RateLimit.Burst
		rateLimit.IsBurstIsSet = true
	}

	return Config{
		Server:    yc.Server,
		Log:       yc.Log,
		DB:        yc.DB,
		RateLimit: rateLimit,
	}, nil
}

func merge(base Config, override Config) Config {
	if override.Server.Host != "" {
		base.Server.Host = override.Server.Host
	}

	if override.Server.Port != "" {
		base.Server.Port = override.Server.Port
	}

	if override.Log.Level != "" {
		base.Log.Level = override.Log.Level
	}

	if override.DB.Host != "" {
		base.DB.Host = override.DB.Host
	}

	if override.DB.Port != "" {
		base.DB.Port = override.DB.Port
	}

	if override.DB.Name != "" {
		base.DB.Name = override.DB.Name
	}

	if override.DB.User != "" {
		base.DB.User = override.DB.User
	}

	if override.DB.Password != "" {
		base.DB.Password = override.DB.Password
	}

	if override.DB.SSLMode != "" {
		base.DB.SSLMode = override.DB.SSLMode
	}

	if override.RateLimit.IsEnabledIsSet {
		base.RateLimit.IsEnabled = override.RateLimit.IsEnabled
	}

	if override.RateLimit.IsRequestsPerMinuteIsSet {
		base.RateLimit.RequestsPerMinute = override.RateLimit.RequestsPerMinute
	}

	if override.RateLimit.IsBurstIsSet {
		base.RateLimit.Burst = override.RateLimit.Burst
	}

	return base
}

func applyEnvOverrides(cfg Config) Config {
	cfg.Server.Host = envOrDefault("SERVER_HOST", cfg.Server.Host)
	cfg.Server.Port = envOrDefault("SERVER_PORT", cfg.Server.Port)
	cfg.Log.Level = envOrDefault("LOG_LEVEL", cfg.Log.Level)

	cfg.DB.Host = envOrDefault("DB_HOST", cfg.DB.Host)
	cfg.DB.Port = envOrDefault("DB_PORT", cfg.DB.Port)
	cfg.DB.Name = envOrDefault("DB_NAME", cfg.DB.Name)
	cfg.DB.User = envOrDefault("DB_USER", cfg.DB.User)
	cfg.DB.Password = envOrDefault("DB_PASSWORD", cfg.DB.Password)
	cfg.DB.SSLMode = envOrDefault("DB_SSL_MODE", cfg.DB.SSLMode)

	isRateLimitEnabled, isRateLimitEnabledIsSet := envBool("RATE_LIMIT_ENABLED")
	if isRateLimitEnabledIsSet {
		cfg.RateLimit.IsEnabled = isRateLimitEnabled
		cfg.RateLimit.IsEnabledIsSet = true
	}

	requestsPerMinute, isRequestsPerMinuteIsSet := envInt("RATE_LIMIT_REQUESTS_PER_MINUTE")
	if isRequestsPerMinuteIsSet {
		cfg.RateLimit.RequestsPerMinute = requestsPerMinute
		cfg.RateLimit.IsRequestsPerMinuteIsSet = true
	}

	burst, isBurstIsSet := envInt("RATE_LIMIT_BURST")
	if isBurstIsSet {
		cfg.RateLimit.Burst = burst
		cfg.RateLimit.IsBurstIsSet = true
	}

	return cfg
}

func envOrDefault(key string, defaultValue string) string {
	value, isOk := os.LookupEnv(key)
	if !isOk {
		return defaultValue
	}

	if value == "" {
		return defaultValue
	}

	return value
}

func envBool(key string) (bool, bool) {
	raw, isOk := os.LookupEnv(key)
	if !isOk || raw == "" {
		return false, false
	}

	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return false, false
	}

	return parsed, true
}

func envInt(key string) (int, bool) {
	raw, isOk := os.LookupEnv(key)
	if !isOk || raw == "" {
		return 0, false
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}

	return parsed, true
}

