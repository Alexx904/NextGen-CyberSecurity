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

	// Latest source includes persisted discovery snapshots.
	if maxVer != 11 {
		t.Errorf("Attesa versione massima 11, ottenuta %d", maxVer)
	}
}
