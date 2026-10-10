package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrInvalidStreamTicket = errors.New("stream ticket non valido o già consumato")
	ErrExpiredStreamTicket = errors.New("stream ticket scaduto")
)

// StreamTicket rappresenta un ticket monouso a brevissima scadenza per connessioni real-time (SSE e WebSocket)
type StreamTicket struct {
	Ticket    string      `json:"ticket"`
	Claims    *UserClaims `json:"-"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// StreamTicketStore definisce l'interfaccia per emissione e consumo monouso dei ticket
type StreamTicketStore interface {
	IssueTicket(claims *UserClaims, ttl time.Duration) (*StreamTicket, error)
	ConsumeTicket(ticketStr string) (*UserClaims, error)
}

// MemoryStreamTicketStore implementa StreamTicketStore in memoria con mutex e auto-pulizia
type MemoryStreamTicketStore struct {
	mu      sync.Mutex
	tickets map[string]*StreamTicket
}

func NewMemoryStreamTicketStore() *MemoryStreamTicketStore {
	store := &MemoryStreamTicketStore{
		tickets: make(map[string]*StreamTicket),
	}
	// Pulizia in background ogni 60 secondi
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		for range ticker.C {
			store.cleanupExpired()
		}
	}()
	return store
}

func (s *MemoryStreamTicketStore) IssueTicket(claims *UserClaims, ttl time.Duration) (*StreamTicket, error) {
	if claims == nil {
		return nil, errors.New("claims non validi per emissione stream ticket")
	}
	if ttl <= 0 {
		return nil, errors.New("il TTL deve essere maggiore di zero")
	}

	rawBytes := make([]byte, 24)
	if _, err := rand.Read(rawBytes); err != nil {
		return nil, fmt.Errorf("generazione casuale ticket fallita: %w", err)
	}
	tokenStr := "st_" + hex.EncodeToString(rawBytes)

	// Copia difensiva dei claims
	cpClaims := *claims

	ticket := &StreamTicket{
		Ticket:    tokenStr,
		Claims:    &cpClaims,
		ExpiresAt: time.Now().Add(ttl),
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickets[tokenStr] = ticket

	return ticket, nil
}

func (s *MemoryStreamTicketStore) ConsumeTicket(ticketStr string) (*UserClaims, error) {
	if ticketStr == "" {
		return nil, ErrInvalidStreamTicket
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ticket, exists := s.tickets[ticketStr]
	if !exists {
		return nil, ErrInvalidStreamTicket
	}

	// Consumo monouso immediato: cancella il ticket dallo store
	delete(s.tickets, ticketStr)

	if time.Now().After(ticket.ExpiresAt) {
		return nil, ErrExpiredStreamTicket
	}

	return ticket.Claims, nil
}

func (s *MemoryStreamTicketStore) cleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for k, v := range s.tickets {
		if now.After(v.ExpiresAt) {
			delete(s.tickets, k)
		}
	}
}
