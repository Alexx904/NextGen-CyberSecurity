package postgres

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"nextgen-cybersec/internal/repository/migrations"
)

// GetMaxSourceVersion calcola la versione massima delle migrazioni presenti in migrations.FS
func GetMaxSourceVersion() (uint, error) {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return 0, err
	}
	var maxVer uint
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) > 0 {
			if v, err := strconv.ParseUint(parts[0], 10, 64); err == nil {
				if uint(v) > maxVer {
					maxVer = uint(v)
				}
			}
		}
	}
	return maxVer, nil
}

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

	// Autoguarigione schema: se il database si trova ad una versione superiore a quelle
	// note all'eseguibile (ad es. switch da feature branch o versioni orfane),
	// forziamo l'allineamento alla versione massima supportata per prevenire crash-loop all'avvio.
	if maxVer, err := GetMaxSourceVersion(); err == nil && maxVer > 0 {
		curVer, dirty, vErr := m.Version()
		if vErr == nil && !dirty && curVer > maxVer {
			slog.Warn("Versione database superiore alle definizioni disponibili nel binario, riallineamento automatico",
				"db_version", curVer, "max_supported_version", maxVer)
			if err := m.Force(int(maxVer)); err != nil {
				return fmt.Errorf("impossibile riallineare la versione del database a %d: %w", maxVer, err)
			}
		}
	}

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

