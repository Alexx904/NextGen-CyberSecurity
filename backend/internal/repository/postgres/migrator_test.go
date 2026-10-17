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

	// Latest source includes the independent SSH host identity pin.
	if maxVer != 10 {
		t.Errorf("Attesa versione massima 10, ottenuta %d", maxVer)
	}
}
