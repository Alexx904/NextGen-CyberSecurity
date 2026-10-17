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
		"000002_cve_indexes.up.sql",
		"000002_cve_indexes.down.sql",
		"000003_drop_report_templates.up.sql",
		"000003_drop_report_templates.down.sql",
		"000004_tree_target_groups.up.sql",
		"000004_tree_target_groups.down.sql",
		"000005_ssh_servers_firewall.up.sql",
		"000005_ssh_servers_firewall.down.sql",
		"000008_ssh_client_labels.up.sql",
		"000008_ssh_client_labels.down.sql",
		"000009_company_department_depth.up.sql",
		"000009_company_department_depth.down.sql",
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

func TestSSHServersFirewallMigration(t *testing.T) {
	contentUp, err := os.ReadFile("000005_ssh_servers_firewall.up.sql")
	if err != nil {
		altPath := filepath.Join("..", "..", "internal", "repository", "migrations", "000005_ssh_servers_firewall.up.sql")
		contentUp, err = os.ReadFile(altPath)
		if err != nil {
			t.Fatalf("Impossibile leggere 000005_ssh_servers_firewall.up.sql: %v", err)
		}
	}
	sqlText := string(contentUp)

	expectedKeywords := []string{
		"CREATE TABLE IF NOT EXISTS ssh_servers",
		"CREATE TABLE IF NOT EXISTS firewall_backups",
		"idx_ssh_servers_tenant_id",
		"idx_ssh_servers_status",
		"idx_firewall_backups_tenant_id",
		"idx_firewall_backups_server_id",
		"idx_firewall_backups_created_at",
		"encrypted_secret",
		"auth_type",
		"raw_iptables_rules",
	}

	for _, kw := range expectedKeywords {
		if !strings.Contains(sqlText, kw) {
			t.Errorf("La migrazione 000005 UP non contiene la specifica attesa %q", kw)
		}
	}

	contentDown, err := os.ReadFile("000005_ssh_servers_firewall.down.sql")
	if err != nil {
		altPath := filepath.Join("..", "..", "internal", "repository", "migrations", "000005_ssh_servers_firewall.down.sql")
		contentDown, err = os.ReadFile(altPath)
		if err != nil {
			t.Fatalf("Impossibile leggere 000005_ssh_servers_firewall.down.sql: %v", err)
		}
	}
	downSql := string(contentDown)
	if !strings.Contains(downSql, "DROP TABLE IF EXISTS firewall_backups") {
		t.Errorf("La migrazione 000005 DOWN non contiene il drop di firewall_backups")
	}
	if !strings.Contains(downSql, "DROP TABLE IF EXISTS ssh_servers") {
		t.Errorf("La migrazione 000005 DOWN non contiene il drop di ssh_servers")
	}
}

func TestTreeTargetGroupsMigration(t *testing.T) {
	content, err := os.ReadFile("000004_tree_target_groups.up.sql")
	if err != nil {
		altPath := filepath.Join("..", "..", "internal", "repository", "migrations", "000004_tree_target_groups.up.sql")
		content, err = os.ReadFile(altPath)
		if err != nil {
			t.Fatalf("Impossibile leggere 000004_tree_target_groups.up.sql: %v", err)
		}
	}
	sqlText := string(content)

	if !strings.Contains(sqlText, "parent_id") {
		t.Errorf("La migrazione 000004 UP non contiene la colonna parent_id")
	}
	if !strings.Contains(sqlText, "color_tag") {
		t.Errorf("La migrazione 000004 UP non contiene la colonna color_tag")
	}
	if !strings.Contains(sqlText, "idx_target_groups_parent_id") {
		t.Errorf("La migrazione 000004 UP non contiene l'indice idx_target_groups_parent_id")
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
