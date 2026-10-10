package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/auth"
)

func TestStreamTicketStore_IssueAndConsume(t *testing.T) {
	store := auth.NewMemoryStreamTicketStore()

	claims := &auth.UserClaims{
		UserID:    uuid.New(),
		TenantID:  uuid.New(),
		Role:      domain.RoleSecOps,
		Email:     "secops@nextgen.local",
		TokenType: auth.TokenTypeAccess,
	}

	// 1. Emissione ticket
	ticket, err := store.IssueTicket(claims, 5*time.Second)
	if err != nil {
		t.Fatalf("IssueTicket ha restituito errore inatteso: %v", err)
	}

	if ticket == nil || ticket.Ticket == "" {
		t.Fatal("Ticket generato nullo o stringa vuota")
	}

	// 2. Primo consumo: successo
	consumedClaims, err := store.ConsumeTicket(ticket.Ticket)
	if err != nil {
		t.Fatalf("ConsumeTicket ha restituito errore inatteso: %v", err)
	}

	if consumedClaims.UserID != claims.UserID || consumedClaims.TenantID != claims.TenantID {
		t.Errorf("Claims consumati non corrispondono: %+v vs %+v", consumedClaims, claims)
	}

	// 3. Secondo consumo (riutilizzo del medesimo ticket): DEVE fallire (monouso)
	_, errSecond := store.ConsumeTicket(ticket.Ticket)
	if errSecond == nil {
		t.Fatal("Atteso errore al secondo consumo di ticket monouso, ottenuto nil")
	}
}

func TestStreamTicketStore_ExpiredTicket(t *testing.T) {
	store := auth.NewMemoryStreamTicketStore()

	claims := &auth.UserClaims{
		UserID:    uuid.New(),
		TenantID:  uuid.New(),
		Role:      domain.RoleSecOps,
		Email:     "secops@nextgen.local",
		TokenType: auth.TokenTypeAccess,
	}

	// TTL molto breve per verificare scadenza
	ticket, err := store.IssueTicket(claims, 20*time.Millisecond)
	if err != nil {
		t.Fatalf("IssueTicket fallito: %v", err)
	}

	time.Sleep(35 * time.Millisecond)

	_, errConsume := store.ConsumeTicket(ticket.Ticket)
	if errConsume == nil {
		t.Fatal("Atteso errore per ticket scaduto, ottenuto nil")
	}
}
