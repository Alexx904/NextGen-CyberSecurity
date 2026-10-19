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

	// Latest source includes append-only assessment lifecycle and evidence.
	if maxVer != 14 {
		t.Errorf("Attesa versione massima 14, ottenuta %d", maxVer)
	}
}
