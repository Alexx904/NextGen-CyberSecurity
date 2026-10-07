package postgres

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"nextgen-cybersec/internal/repository/migrations"
)

// RunMigrations esegue le migrazioni pending sfruttando le definizioni DDL embeddate in migrations.FS
func RunMigrations(dsn string) error {
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("impossibile inizializzare il driver source iofs: %w", err)
	}

	// Il driver golang-migrate per pgx v5 si registra con lo schema "pgx5"
	migratorDSN := dsn
	if strings.HasPrefix(migratorDSN, "postgres://") {
		migratorDSN = "pgx5://" + strings.TrimPrefix(migratorDSN, "postgres://")
	} else if strings.HasPrefix(migratorDSN, "postgresql://") {
		migratorDSN = "pgx5://" + strings.TrimPrefix(migratorDSN, "postgresql://")
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, migratorDSN)
	if err != nil {
		return fmt.Errorf("impossibile istanziare il migrator con DSN specificato: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			slog.Debug("Chiusura sorgente migrazioni", "info", srcErr)
		}
		if dbErr != nil {
			slog.Debug("Chiusura db driver migrazioni", "info", dbErr)
		}
	}()

	slog.Info("Applicazione migrazioni database in corso...")
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("Nessuna nuova migrazione: lo schema del database è già aggiornato")
			return nil
		}
		return fmt.Errorf("errore durante l'esecuzione delle migrazioni DDL UP: %w", err)
	}

	slog.Info("Migrazioni DDL applicate con successo")
	return nil
}

