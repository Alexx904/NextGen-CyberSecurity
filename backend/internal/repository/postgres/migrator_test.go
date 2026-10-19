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

	// Latest source includes immutable scan authorization evidence.
	if maxVer != 13 {
		t.Errorf("Attesa versione massima 13, ottenuta %d", maxVer)
	}
}
