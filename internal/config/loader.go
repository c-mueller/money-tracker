package config

import (
	"strings"

	"github.com/spf13/viper"
)

func Load(configFile string) (Config, error) {
	cfg := Defaults()

	v := viper.New()
	v.SetEnvPrefix("MONEY_TRACKER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set defaults so viper knows the keys for env binding
	v.SetDefault("server.port", cfg.Server.Port)
	v.SetDefault("server.host", cfg.Server.Host)
	v.SetDefault("database.driver", cfg.Database.Driver)
	v.SetDefault("database.dsn", cfg.Database.DSN)
	v.SetDefault("auth.oidc.issuer", cfg.Auth.OIDC.Issuer)
	v.SetDefault("auth.oidc.client_id", cfg.Auth.OIDC.ClientID)
	v.SetDefault("auth.oidc.client_secret", cfg.Auth.OIDC.ClientSecret)
	v.SetDefault("auth.oidc.redirect_url", cfg.Auth.OIDC.RedirectURL)
	v.SetDefault("auth.session.secret", cfg.Auth.Session.Secret)
	v.SetDefault("auth.session.max_age", cfg.Auth.Session.MaxAge)
	v.SetDefault("logging.level", cfg.Logging.Level)

	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return cfg, err
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
