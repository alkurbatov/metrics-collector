package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	// Migrate tools.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(url security.DatabaseURL) error {
	var (
		retries  uint = 20
		migrator *migrate.Migrate
		err      error
	)

	for retries > 0 {
		migrator, err = migrate.New("file://migrations", string(url))
		if err == nil {
			break
		}

		retries--

		log.Info().Msg("Trying to connect to " + url.String())
		time.Sleep(time.Second)
	}

	if err != nil {
		return fmt.Errorf("DB migration failed: %w", err)
	}

	err = migrator.Up()
	defer migrator.Close()

	if err == nil {
		log.Info().Msg("Applying migrations: success")
		return nil
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Info().Msg("Applying migrations: no change")
		return nil
	}

	return fmt.Errorf("DB migration failed: %w", err)
}
