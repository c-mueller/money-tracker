package repository

import (
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"icekalt.dev/money-tracker/ent"
	"icekalt.dev/money-tracker/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

func NewClient(cfg config.DatabaseConfig) (*ent.Client, error) {
	var drv dialect.Driver

	switch cfg.Driver {
	case "sqlite", "sqlite3":
		db, err := sql.Open("sqlite", cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("opening sqlite: %w", err)
		}
		drv = entsql.OpenDB(dialect.SQLite, db)
	case "postgres", "postgresql":
		db, err := sql.Open("pgx", cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("opening postgres: %w", err)
		}
		drv = entsql.OpenDB(dialect.Postgres, db)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	return ent.NewClient(ent.Driver(drv)), nil
}
