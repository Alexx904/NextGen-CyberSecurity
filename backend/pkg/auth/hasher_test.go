package auth_test

import (
	"errors"
	"testing"

	"nextgen-cybersec/pkg/auth"
)

func TestBcryptHasher_Success(t *testing.T) {
	hasher := auth.NewBcryptHasher(10) // cost 10 for fast tests

	password := "NextGenCyberSec#2026!"
	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Hash ha ritornato un errore inaspettato: %v", err)
	}

	if hash == "" || hash == password {
		t.Fatalf("Hash generato non valido o identico alla password in chiaro: %s", hash)
	}

	// Verifica corretta
	if err := hasher.Compare(hash, password); err != nil {
		t.Fatalf("Compare ha fallito con la password corretta: %v", err)
	}
}

func TestBcryptHasher_WrongPassword(t *testing.T) {
	hasher := auth.NewBcryptHasher(10)

	password := "NextGenCyberSec#2026!"
	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Hash ha ritornato errore: %v", err)
	}

	err = hasher.Compare(hash, "WrongPassword#123")
	if !errors.Is(err, auth.ErrPasswordMismatch) {
		t.Fatalf("Atteso ErrPasswordMismatch, ottenuto: %v", err)
	}
}

func TestBcryptHasher_EmptyInputs(t *testing.T) {
	hasher := auth.NewBcryptHasher(10)

	// Hash con password vuota
	_, err := hasher.Hash("")
	if !errors.Is(err, auth.ErrEmptyPassword) {
		t.Fatalf("Atteso ErrEmptyPassword durante Hash, ottenuto: %v", err)
	}

	// Compare con password vuota
	err = hasher.Compare("$2a$10$validhashplaceholder", "")
	if !errors.Is(err, auth.ErrEmptyPassword) {
		t.Fatalf("Atteso ErrEmptyPassword durante Compare, ottenuto: %v", err)
	}

	// Compare con hash vuoto
	err = hasher.Compare("", "somePassword")
	if !errors.Is(err, auth.ErrInvalidHash) {
		t.Fatalf("Atteso ErrInvalidHash per hash vuoto, ottenuto: %v", err)
	}

	// Compare con hash malformato
	err = hasher.Compare("invalid-non-bcrypt-hash", "somePassword")
	if !errors.Is(err, auth.ErrInvalidHash) {
		t.Fatalf("Atteso ErrInvalidHash per hash malformato, ottenuto: %v", err)
	}
}

func TestBcryptHasher_CostFallback(t *testing.T) {
	// Costo non valido (< bcrypt.MinCost = 4)
	hasherLow := auth.NewBcryptHasher(1)
	hash, err := hasherLow.Hash("testPassword123")
	if err != nil {
		t.Fatalf("Hash con cost fallback ha ritornato errore: %v", err)
	}
	if err := hasherLow.Compare(hash, "testPassword123"); err != nil {
		t.Fatalf("Compare ha fallito: %v", err)
	}

	// Costo non valido (> bcrypt.MaxCost = 31)
	hasherHigh := auth.NewBcryptHasher(50)
	hash2, err := hasherHigh.Hash("testPassword123")
	if err != nil {
		t.Fatalf("Hash con cost fallback ha ritornato errore: %v", err)
	}
	if err := hasherHigh.Compare(hash2, "testPassword123"); err != nil {
		t.Fatalf("Compare ha fallito: %v", err)
	}
}

