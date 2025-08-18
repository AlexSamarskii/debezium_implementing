package migration

import (
	"fmt"
	"log"

	"github.com/AlexSamarskii/debezium_implementing/config"
	"github.com/golang-migrate/migrate/v4"
)

type Migrator struct {
	cfg        *config.Config
	migrations string
}

func NewMigrator(cfg *config.Config, migrations string) *Migrator {
	return &Migrator{
		cfg:        cfg,
		migrations: migrations,
	}
}

func (m *Migrator) dsn() string {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		m.cfg.Database.User,
		m.cfg.Database.Password,
		m.cfg.Database.Host,
		m.cfg.Database.Port,
		m.cfg.Database.Name,
	)
	return dsn
}

func (m *Migrator) Apply() error {
	migrator, err := migrate.New(m.migrations, m.dsn())
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No migration changes")
			return nil
		}
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Migrations applied")
	return nil
}
