package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmptyPassword    = errors.New("la password non può essere vuota")
	ErrPasswordMismatch = errors.New("password non corretta")
	ErrInvalidHash      = errors.New("hash della password non valido")
)

// PasswordHasher definisce il contratto per l'hashing e la verifica delle credenziali
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

// BcryptHasher implementa PasswordHasher utilizzando l'algoritmo Bcrypt
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher crea una nuova istanza con il work factor specificato (min: bcrypt.MinCost, default: 12)
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = 12
	}
	return &BcryptHasher{cost: cost}
}

// Hash genera l'hash Bcrypt della password fornita
func (h *BcryptHasher) Hash(password string) (string, error) {
	if len(password) == 0 {
		return "", ErrEmptyPassword
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("generazione hash fallita: %w", err)
	}

	return string(bytes), nil
}

// Compare verifica se la password in chiaro corrisponde all'hash salvato
func (h *BcryptHasher) Compare(hashedPassword, password string) error {
	if len(password) == 0 {
		return ErrEmptyPassword
	}
	if len(hashedPassword) == 0 {
		return ErrInvalidHash
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrPasswordMismatch
		}
		return ErrInvalidHash
	}

	return nil
}

