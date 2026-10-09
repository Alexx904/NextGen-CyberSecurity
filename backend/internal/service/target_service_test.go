package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/validator"
)

// mockTargetRepo implementa domain.TargetRepository in memoria
type mockTargetRepo struct {
	targets map[string]*domain.Target
}

func newMockTargetRepo() *mockTargetRepo {
	return &mockTargetRepo{
		targets: make(map[string]*domain.Target),
	}
}

func (m *mockTargetRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Target, error) {
	key := tenantID.String() + ":" + id.String()
	t, ok := m.targets[key]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return t, nil
}

func (m *mockTargetRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.TargetFilter) ([]*domain.Target, int, error) {
	var list []*domain.Target
	for _, t := range m.targets {
		if t.TenantID != tenantID {
			continue
		}
		if filter.GroupID != nil && (t.GroupID == nil || *t.GroupID != *filter.GroupID) {
			continue
		}
		if filter.Type != nil && t.Type != *filter.Type {
			continue
		}
		if filter.IsBlacklisted != nil && t.IsBlacklisted != *filter.IsBlacklisted {
			continue
		}
		if filter.Search != "" {
			searchLower := strings.ToLower(filter.Search)
			valMatch := strings.Contains(strings.ToLower(t.Value), searchLower)
			descMatch := t.Description != nil && strings.Contains(strings.ToLower(*t.Description), searchLower)
			if !valMatch && !descMatch {
				continue
			}
		}
		list = append(list, t)
	}
	total := len(list)
	return list, total, nil
}

func (m *mockTargetRepo) Create(ctx context.Context, target *domain.Target) error {
	key := target.TenantID.String() + ":" + target.ID.String()
	m.targets[key] = target
	return nil
}

func (m *mockTargetRepo) Update(ctx context.Context, target *domain.Target) error {
	key := target.TenantID.String() + ":" + target.ID.String()
	if _, ok := m.targets[key]; !ok {
		return domain.ErrNotFound
	}
	m.targets[key] = target
	return nil
}

func (m *mockTargetRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	key := tenantID.String() + ":" + id.String()
	if _, ok := m.targets[key]; !ok {
		return domain.ErrNotFound
	}
	delete(m.targets, key)
	return nil
}

func (m *mockTargetRepo) IsBlacklisted(ctx context.Context, tenantID uuid.UUID, value string) (bool, error) {
	var blacklisted []string
	for _, t := range m.targets {
		if t.TenantID == tenantID && t.IsBlacklisted {
			blacklisted = append(blacklisted, t.Value)
		}
	}
	if len(blacklisted) == 0 {
		return false, nil
	}
	return validator.IsTargetBlacklisted(value, blacklisted)
}

// mockGroupRepo implementa domain.TargetGroupRepository in memoria
type mockGroupRepo struct {
	groups map[string]*domain.TargetGroup
}

func newMockGroupRepo() *mockGroupRepo {
	return &mockGroupRepo{
		groups: make(map[string]*domain.TargetGroup),
	}
}

func (m *mockGroupRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.TargetGroup, error) {
	key := tenantID.String() + ":" + id.String()
	g, ok := m.groups[key]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return g, nil
}

func (m *mockGroupRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.TargetGroup, error) {
	var list []*domain.TargetGroup
	for _, g := range m.groups {
		if g.TenantID == tenantID {
			list = append(list, g)
		}
	}
	return list, nil
}

func (m *mockGroupRepo) Create(ctx context.Context, group *domain.TargetGroup) error {
	key := group.TenantID.String() + ":" + group.ID.String()
	m.groups[key] = group
	return nil
}

func (m *mockGroupRepo) Update(ctx context.Context, group *domain.TargetGroup) error {
	key := group.TenantID.String() + ":" + group.ID.String()
	if _, ok := m.groups[key]; !ok {
		return domain.ErrNotFound
	}
	m.groups[key] = group
	return nil
}

func (m *mockGroupRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	key := tenantID.String() + ":" + id.String()
	if _, ok := m.groups[key]; !ok {
		return domain.ErrNotFound
	}
	delete(m.groups, key)
	return nil
}

func (m *mockGroupRepo) GetTargetCountsByGroup(ctx context.Context, tenantID uuid.UUID) (map[uuid.UUID]int, error) {
	return make(map[uuid.UUID]int), nil
}

func (m *mockGroupRepo) GetSeverityCountsByGroup(ctx context.Context, tenantID uuid.UUID) (map[uuid.UUID]map[domain.SeverityLevel]int, error) {
	return make(map[uuid.UUID]map[domain.SeverityLevel]int), nil
}

func TestTargetService_CreateTarget_TypesInference(t *testing.T) {
	targetRepo := newMockTargetRepo()
	groupRepo := newMockGroupRepo()
	svc := service.NewTargetService(targetRepo, groupRepo)

	ctx := context.Background()
	tenantID := uuid.New()

	testCases := []struct {
		inputVal     string
		expectedType domain.TargetType
	}{
		{"192.168.1.10", domain.TargetTypeIP},
		{"::1", domain.TargetTypeIP},
		{"10.0.0.0/24", domain.TargetTypeCIDR},
		{"192.168.1.1-50", domain.TargetTypeRange},
		{"scanme.nmap.org", domain.TargetTypeFQDN},
	}

	for _, tc := range testCases {
		target, err := svc.CreateTarget(ctx, service.CreateTargetInput{
			TenantID: tenantID,
			Value:    tc.inputVal,
		})
		if err != nil {
			t.Fatalf("CreateTarget fallito per '%s': %v", tc.inputVal, err)
		}
		if target.Type != tc.expectedType {
			t.Errorf("Per input '%s': atteso tipo %v, ottenuto %v", tc.inputVal, tc.expectedType, target.Type)
		}
	}
}

func TestTargetService_CreateTarget_ValidationErrors(t *testing.T) {
	targetRepo := newMockTargetRepo()
	groupRepo := newMockGroupRepo()
	svc := service.NewTargetService(targetRepo, groupRepo)

	ctx := context.Background()
	tenantID := uuid.New()

	malicious := []string{
		"-sS",
		"--script=all",
		"192.168.1.1; ls",
		"",
		"invalid-single-word",
	}

	for _, m := range malicious {
		_, err := svc.CreateTarget(ctx, service.CreateTargetInput{
			TenantID: tenantID,
			Value:    m,
		})
		if !errors.Is(err, domain.ErrTargetInvalid) {
			t.Errorf("Per input '%s': atteso ErrTargetInvalid, ottenuto %v", m, err)
		}
	}
}

func TestTargetService_CreateTarget_BlacklistProtection(t *testing.T) {
	targetRepo := newMockTargetRepo()
	groupRepo := newMockGroupRepo()
	svc := service.NewTargetService(targetRepo, groupRepo)

	ctx := context.Background()
	tenantID := uuid.New()

	// 1. Inserisci una regola in blacklist per la subnet 10.0.0.0/8
	blTarget, err := svc.CreateTarget(ctx, service.CreateTargetInput{
		TenantID:      tenantID,
		Value:         "10.0.0.0/8",
		IsBlacklisted: true,
	})
	if err != nil {
		t.Fatalf("Creazione regola blacklist fallita: %v", err)
	}
	if !blTarget.IsBlacklisted {
		t.Fatal("Atteso target contrassegnato come blacklisted")
	}

	// 2. Prova a inserire un bersaglio operativo che cade nella subnet
	_, err = svc.CreateTarget(ctx, service.CreateTargetInput{
		TenantID:      tenantID,
		Value:         "10.5.1.20",
		IsBlacklisted: false,
	})
	if !errors.Is(err, domain.ErrTargetBlacklisted) {
		t.Fatalf("Atteso ErrTargetBlacklisted, ottenuto: %v", err)
	}

	// 3. Bersaglio esterno alla blacklist consentito
	_, err = svc.CreateTarget(ctx, service.CreateTargetInput{
		TenantID:      tenantID,
		Value:         "192.168.1.1",
		IsBlacklisted: false,
	})
	if err != nil {
		t.Fatalf("Creazione target consentito fallita: %v", err)
	}
}

func TestTargetService_MultiTenantIsolation(t *testing.T) {
	targetRepo := newMockTargetRepo()
	groupRepo := newMockGroupRepo()
	svc := service.NewTargetService(targetRepo, groupRepo)

	ctx := context.Background()
	tenantA := uuid.New()
	tenantB := uuid.New()

	targetA, err := svc.CreateTarget(ctx, service.CreateTargetInput{
		TenantID: tenantA,
		Value:    "192.168.10.1",
	})
	if err != nil {
		t.Fatalf("Creazione target tenant A fallita: %v", err)
	}

	// Tenant B prova a recuperare il target di Tenant A
	_, err = svc.GetTarget(ctx, tenantB, targetA.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Atteso ErrNotFound per accesso cross-tenant, ottenuto: %v", err)
	}

	// Tenant B prova a cancellare il target di Tenant A
	err = svc.DeleteTarget(ctx, tenantB, targetA.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Atteso ErrNotFound per cancellazione cross-tenant, ottenuto: %v", err)
	}
}

func TestTargetService_Groups(t *testing.T) {
	targetRepo := newMockTargetRepo()
	groupRepo := newMockGroupRepo()
	svc := service.NewTargetService(targetRepo, groupRepo)

	ctx := context.Background()
	tenantID := uuid.New()
	otherTenantID := uuid.New()

	desc := "Gruppo server DMZ"
	group, err := svc.CreateGroup(ctx, tenantID, "DMZ Servers", &desc)
	if err != nil {
		t.Fatalf("CreateGroup fallito: %v", err)
	}

	// Associazione valida a target
	target, err := svc.CreateTarget(ctx, service.CreateTargetInput{
		TenantID: tenantID,
		GroupID:  &group.ID,
		Value:    "192.168.100.5",
	})
	if err != nil {
		t.Fatalf("CreateTarget con gruppo valido fallito: %v", err)
	}
	if target.GroupID == nil || *target.GroupID != group.ID {
		t.Fatalf("GroupID non associato correttamente")
	}

	// Tentativo di associazione a gruppo di un altro tenant
	otherGroup, _ := svc.CreateGroup(ctx, otherTenantID, "Other Group", nil)
	_, err = svc.CreateTarget(ctx, service.CreateTargetInput{
		TenantID: tenantID,
		GroupID:  &otherGroup.ID,
		Value:    "192.168.100.6",
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Atteso ErrNotFound per associazione a gruppo di altro tenant, ottenuto: %v", err)
	}

	// Elenco gruppi
	groups, err := svc.ListGroups(ctx, tenantID)
	if err != nil {
		t.Fatalf("ListGroups fallito: %v", err)
	}
	if len(groups) != 1 {
		t.Errorf("Atteso 1 gruppo, ottenuti %d", len(groups))
	}

	// Cancellazione gruppo
	if err := svc.DeleteGroup(ctx, tenantID, group.ID); err != nil {
		t.Fatalf("DeleteGroup fallito: %v", err)
	}
}

func TestTargetService_ListAndUpdate(t *testing.T) {
	targetRepo := newMockTargetRepo()
	groupRepo := newMockGroupRepo()
	svc := service.NewTargetService(targetRepo, groupRepo)

	ctx := context.Background()
	tenantID := uuid.New()

	descA := "Server Primario"
	tA, _ := svc.CreateTarget(ctx, service.CreateTargetInput{
		TenantID:    tenantID,
		Value:       "192.168.1.10",
		Description: &descA,
	})
	descB := "Server Secondario"
	svc.CreateTarget(ctx, service.CreateTargetInput{
		TenantID:    tenantID,
		Value:       "192.168.1.20",
		Description: &descB,
	})

	// 1. List con Search
	results, total, err := svc.ListTargets(ctx, tenantID, domain.TargetFilter{
		Search: "Primario",
	})
	if err != nil {
		t.Fatalf("ListTargets fallito: %v", err)
	}
	if total != 1 || len(results) != 1 {
		t.Errorf("Atteso 1 risultato per search Primario, ottenuti total=%d len=%d", total, len(results))
	}

	// 2. Update target
	newDesc := "Server Primario Aggiornato"
	updated, err := svc.UpdateTarget(ctx, service.UpdateTargetInput{
		TenantID:      tenantID,
		ID:            tA.ID,
		Description:   &newDesc,
		IsBlacklisted: true,
	})
	if err != nil {
		t.Fatalf("UpdateTarget fallito: %v", err)
	}
	if updated.Description == nil || *updated.Description != newDesc {
		t.Errorf("Descrizione non aggiornata")
	}
	if !updated.IsBlacklisted {
		t.Errorf("Flag is_blacklisted non aggiornato")
	}

	// 3. Delete target
	if err := svc.DeleteTarget(ctx, tenantID, tA.ID); err != nil {
		t.Fatalf("DeleteTarget fallito: %v", err)
	}

	// Verifica eliminazione
	_, err = svc.GetTarget(ctx, tenantID, tA.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Atteso ErrNotFound dopo cancellazione, ottenuto: %v", err)
	}
}

