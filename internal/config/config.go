package config

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Language string         `mapstructure:"language"`
	MCP      MCPConfig      `mapstructure:"mcp"`
}

type MCPConfig struct {
	URL   string `mapstructure:"url"`
	Token string `mapstructure:"token"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

type AuthConfig struct {
	OIDC    OIDCConfig    `mapstructure:"oidc"`
	Session SessionConfig `mapstructure:"session"`
}

type OIDCConfig struct {
	Issuer       string `mapstructure:"issuer"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

type SessionConfig struct {
	Secret string `mapstructure:"secret"`
	MaxAge int    `mapstructure:"max_age"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

func Defaults() Config {
	return Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
			DSN:    "money-tracker.db?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)",
		},
		Auth: AuthConfig{
			Session: SessionConfig{
				MaxAge: 86400,
			},
		},
		Logging: LoggingConfig{
			Level: "info",
		},
		Language: "de",
	}
}
