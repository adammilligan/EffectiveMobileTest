package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
	Log    LogConfig
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

	return Config{
		Server: yc.Server,
		Log:    yc.Log,
		DB:     yc.DB,
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

