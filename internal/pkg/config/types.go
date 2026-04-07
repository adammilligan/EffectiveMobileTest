package config

// ServerConfig describes HTTP server binding parameters.
type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

// DBConfig describes PostgreSQL connection parameters.
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
}

// LogConfig describes application logging parameters.
type LogConfig struct {
	Level string `yaml:"level"`
}

type yamlConfig struct {
	Server ServerConfig `yaml:"server"`
	Log    LogConfig    `yaml:"log"`
	DB     DBConfig     `yaml:"db"`
}

