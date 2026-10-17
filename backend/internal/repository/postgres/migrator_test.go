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

	// Latest source includes SSH client labels and Company -> Department guard.
	if maxVer != 9 {
		t.Errorf("Attesa versione massima 9, ottenuta %d", maxVer)
	}
}

