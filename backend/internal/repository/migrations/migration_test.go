package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationFilesExistAndNotEmpty(t *testing.T) {
	files := []string{
		"000001_init_schema.up.sql",
		"000001_init_schema.down.sql",
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			// Fallback path check if run from parent directories
			altPath := filepath.Join("..", "..", "internal", "repository", "migrations", file)
			content, err = os.ReadFile(altPath)
			if err != nil {
				t.Fatalf("Impossibile leggere il file di migrazione %s: %v", file, err)
			}
		}

		if len(strings.TrimSpace(string(content))) == 0 {
			t.Errorf("Il file di migrazione %s non può essere vuoto", file)
		}
	}
}

func TestUpMigrationContainsRequiredTables(t *testing.T) {
	content, err := os.ReadFile("000001_init_schema.up.sql")
	if err != nil {
		t.Fatalf("Impossibile leggere 000001_init_schema.up.sql: %v", err)
	}
	sqlText := string(content)

	requiredTables := []string{
		"tenants",
		"users",
		"target_groups",
		"targets",
		"scan_profiles",
		"scans",
		"schedules",
		"cve_entries",
		"findings",
		"report_templates",
	}

	for _, table := range requiredTables {
		expectedCreate := "CREATE TABLE IF NOT EXISTS " + table
		if !strings.Contains(sqlText, expectedCreate) {
			t.Errorf("La migrazione UP non contiene la definizione della tabella %s", table)
		}
	}

	// Verifica presenza indici chiave e GIN
	requiredIndexes := []string{
		"idx_users_tenant_id",
		"idx_targets_tenant_id",
		"idx_scans_tenant_status",
		"idx_findings_scan_id",
		"idx_findings_tenant_severity",
		"idx_tenants_brand_colors",
		"idx_cve_entries_cpe_gin",
	}

	for _, idx := range requiredIndexes {
		if !strings.Contains(sqlText, idx) {
			t.Errorf("La migrazione UP non definisce l'indice %s", idx)
		}
	}
}

func TestDownMigrationContainsRequiredDrops(t *testing.T) {
	content, err := os.ReadFile("000001_init_schema.down.sql")
	if err != nil {
		t.Fatalf("Impossibile leggere 000001_init_schema.down.sql: %v", err)
	}
	sqlText := string(content)

	requiredTables := []string{
		"report_templates",
		"findings",
		"cve_entries",
		"schedules",
		"scans",
		"scan_profiles",
		"targets",
		"target_groups",
		"users",
		"tenants",
	}

	for _, table := range requiredTables {
		expectedDrop := "DROP TABLE IF EXISTS " + table
		if !strings.Contains(sqlText, expectedDrop) {
			t.Errorf("La migrazione DOWN non contiene il DROP della tabella %s", table)
		}
	}
}
