package postgres_test

import (
	"testing"

	"nextgen-cybersec/internal/repository/postgres"
)

func TestGetMaxSourceVersion(t *testing.T) {
	maxVer, err := postgres.GetMaxSourceVersion()
	if err != nil {
		t.Fatalf("GetMaxSourceVersion ha restituito un errore imprevisto: %v", err)
	}

	// Al momento ci sono 7 migrazioni (000001 .. 000007)
	if maxVer != 7 {
		t.Errorf("Attesa versione massima 7, ottenuta %d", maxVer)
	}
}

