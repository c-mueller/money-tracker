package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Database.Driver != "sqlite" {
		t.Errorf("expected driver sqlite, got %s", cfg.Database.Driver)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("expected level info, got %s", cfg.Logging.Level)
	}
}

func TestENVOverride(t *testing.T) {
	t.Setenv("MONEY_TRACKER_SERVER_PORT", "9090")
	t.Setenv("MONEY_TRACKER_DATABASE_DRIVER", "postgres")
	t.Setenv("MONEY_TRACKER_LOGGING_LEVEL", "debug")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Database.Driver != "postgres" {
		t.Errorf("expected driver postgres, got %s", cfg.Database.Driver)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("expected level debug, got %s", cfg.Logging.Level)
	}
}

func TestFileOverride(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := []byte("server:\n  port: 3000\nlogging:\n  level: warn\n")
	if err := os.WriteFile(cfgPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != 3000 {
		t.Errorf("expected port 3000, got %d", cfg.Server.Port)
	}
	if cfg.Logging.Level != "warn" {
		t.Errorf("expected level warn, got %s", cfg.Logging.Level)
	}
	// Unset values should keep defaults
	if cfg.Database.Driver != "sqlite" {
		t.Errorf("expected driver sqlite, got %s", cfg.Database.Driver)
	}
}
