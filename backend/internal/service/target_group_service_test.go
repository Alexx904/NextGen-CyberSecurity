package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

// mockTargetGroupRepo implementa in-memory domain.TargetGroupRepository per i test di TargetGroupService
func TestCompanyHierarchyRejectsNestedDepartments(t *testing.T) {
	repo := newMockTargetGroupRepo()
	svc := service.NewTargetGroupService(repo)
	tenant := uuid.New()
	ctx := context.Background()
	company, err := svc.CreateGroup(ctx, service.CreateGroupInput{TenantID: tenant, Name: "Company"})
	if err != nil {
		t.Fatal(err)
	}
	department, err := svc.CreateGroup(ctx, service.CreateGroupInput{TenantID: tenant, ParentID: &company.ID, Name: "IT"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.CreateGroup(ctx, service.CreateGroupInput{TenantID: tenant, ParentID: &department.ID, Name: "Nested"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("nested department accepted: %v", err)
	}
	other, err := svc.CreateGroup(ctx, service.CreateGroupInput{TenantID: tenant, Name: "Other"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.UpdateGroup(ctx, service.UpdateGroupInput{TenantID: tenant, ID: other.ID, ParentID: &department.ID, Name: "Nested"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("nested update accepted: %v", err)
	}
	if _, err = svc.UpdateGroup(ctx, service.UpdateGroupInput{TenantID: tenant, ID: company.ID, ParentID: &other.ID, Name: "Company"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("company with children reparented: %v", err)
	}
}

type mockTargetGroupRepo struct {
	groups         map[string]*domain.TargetGroup
	targetCounts   map[string]map[uuid.UUID]int
	severityCounts map[string]map[uuid.UUID]map[domain.SeverityLevel]int
}

func newMockTargetGroupRepo() *mockTargetGroupRepo {
	return &mockTargetGroupRepo{
		groups:         make(map[string]*domain.TargetGroup),
		targetCounts:   make(map[string]map[uuid.UUID]int),
		severityCounts: make(map[string]map[uuid.UUID]map[domain.SeverityLevel]int),
	}
}

func (m *mockTargetGroupRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.TargetGroup, error) {
	key := tenantID.String() + ":" + id.String()
	g, ok := m.groups[key]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return g, nil
}

func (m *mockTargetGroupRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.TargetGroup, error) {
	var list []*domain.TargetGroup
	for _, g := range m.groups {
		if g.TenantID == tenantID {
			list = append(list, g)
		}
	}
	return list, nil
}

func (m *mockTargetGroupRepo) Create(ctx context.Context, group *domain.TargetGroup) error {
	key := group.TenantID.String() + ":" + group.ID.String()
	m.groups[key] = group
	return nil
}

func (m *mockTargetGroupRepo) Update(ctx context.Context, group *domain.TargetGroup) error {
	key := group.TenantID.String() + ":" + group.ID.String()
	if _, ok := m.groups[key]; !ok {
		return domain.ErrNotFound
	}
	m.groups[key] = group
	return nil
}

func (m *mockTargetGroupRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	key := tenantID.String() + ":" + id.String()
	if _, ok := m.groups[key]; !ok {
		return domain.ErrNotFound
	}
	delete(m.groups, key)
	return nil
}

func (m *mockTargetGroupRepo) GetTargetCountsByGroup(ctx context.Context, tenantID uuid.UUID) (map[uuid.UUID]int, error) {
	counts, ok := m.targetCounts[tenantID.String()]
	if !ok {
		return make(map[uuid.UUID]int), nil
	}
	res := make(map[uuid.UUID]int, len(counts))
	for k, v := range counts {
		res[k] = v
	}
	return res, nil
}

func (m *mockTargetGroupRepo) GetSeverityCountsByGroup(ctx context.Context, tenantID uuid.UUID) (map[uuid.UUID]map[domain.SeverityLevel]int, error) {
	counts, ok := m.severityCounts[tenantID.String()]
	if !ok {
		return make(map[uuid.UUID]map[domain.SeverityLevel]int), nil
	}
	res := make(map[uuid.UUID]map[domain.SeverityLevel]int, len(counts))
	for gid, sevs := range counts {
		res[gid] = make(map[domain.SeverityLevel]int, len(sevs))
		for k, v := range sevs {
			res[gid][k] = v
		}
	}
	return res, nil
}

func TestTargetGroupService_CreateGroup(t *testing.T) {
	ctx := context.Background()
	repo := newMockTargetGroupRepo()
	svc := service.NewTargetGroupService(repo)

	tenantID := uuid.New()
	colorBlue := "#2563eb"
	descRoot := "Cartella Radice"

	// 1. Creazione gruppo root con successo
	root, err := svc.CreateGroup(ctx, service.CreateGroupInput{
		TenantID:    tenantID,
		Name:        "Root Infrastructure",
		Description: &descRoot,
		ColorTag:    &colorBlue,
	})
	if err != nil {
		t.Fatalf("Creazione root group inattesa fallita: %v", err)
	}
	if root.ParentID != nil {
		t.Errorf("Atteso ParentID nil per root, ottenuto %v", root.ParentID)
	}
	if root.ColorTag == nil || *root.ColorTag != colorBlue {
		t.Errorf("Atteso color_tag %s, ottenuto %v", colorBlue, root.ColorTag)
	}

	// 2. Creazione sotto-gruppo valido
	child, err := svc.CreateGroup(ctx, service.CreateGroupInput{
		TenantID: tenantID,
		ParentID: &root.ID,
		Name:     "Subfolder DMZ",
	})
	if err != nil {
		t.Fatalf("Creazione child group inattesa fallita: %v", err)
	}
	if child.ParentID == nil || *child.ParentID != root.ID {
		t.Errorf("Atteso ParentID %v, ottenuto %v", root.ID, child.ParentID)
	}

	// 3. Fallimento: genitore inesistente
	invalidParentID := uuid.New()
	_, err = svc.CreateGroup(ctx, service.CreateGroupInput{
		TenantID: tenantID,
		ParentID: &invalidParentID,
		Name:     "Invalid Parent Folder",
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Atteso ErrNotFound per parent inesistente, ottenuto: %v", err)
	}

	// 4. Fallimento: validazione nome vuoto
	_, err = svc.CreateGroup(ctx, service.CreateGroupInput{
		TenantID: tenantID,
		Name:     "   ",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Atteso ErrValidation per nome vuoto, ottenuto: %v", err)
	}

	// 5. Fallimento: TenantID nullo
	_, err = svc.CreateGroup(ctx, service.CreateGroupInput{
		TenantID: uuid.Nil,
		Name:     "No Tenant",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Atteso ErrValidation per TenantID nullo, ottenuto: %v", err)
	}
}

func TestTargetGroupService_UpdateGroup_CycleDetection(t *testing.T) {
	ctx := context.Background()
	repo := newMockTargetGroupRepo()
	svc := service.NewTargetGroupService(repo)

	tenantID := uuid.New()

	// Creazione catena: NodeA -> NodeB -> NodeC
	nodeA, _ := svc.CreateGroup(ctx, service.CreateGroupInput{
		TenantID: tenantID,
		Name:     "Node A",
	})
	nodeB, _ := svc.CreateGroup(ctx, service.CreateGroupInput{
		TenantID: tenantID,
		ParentID: &nodeA.ID,
		Name:     "Node B",
	})
	// Seed legacy depth directly: new API writes must reject this shape.
	nodeC := &domain.TargetGroup{ID: uuid.New(), TenantID: tenantID, ParentID: &nodeB.ID, Name: "Node C"}
	if err := repo.Create(ctx, nodeC); err != nil {
		t.Fatal(err)
	}

	// 1. Errore: un nodo non può essere padre di sé stesso
	_, err := svc.UpdateGroup(ctx, service.UpdateGroupInput{
		TenantID: tenantID,
		ID:       nodeA.ID,
		ParentID: &nodeA.ID,
		Name:     "Node A Self",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Atteso ErrValidation su auto-referenza genitore, ottenuto: %v", err)
	}

	// 2. Errore: impostare come genitore di NodeA il discendente NodeC crea un ciclo
	_, err = svc.UpdateGroup(ctx, service.UpdateGroupInput{
		TenantID: tenantID,
		ID:       nodeA.ID,
		ParentID: &nodeC.ID,
		Name:     "Node A Cycle",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Atteso ErrValidation su rilevamento ciclo, ottenuto: %v", err)
	}

	// 3. Spostamento valido: spostare NodeC sotto la root (ParentID = nil)
	updatedC, err := svc.UpdateGroup(ctx, service.UpdateGroupInput{
		TenantID: tenantID,
		ID:       nodeC.ID,
		ParentID: nil,
		Name:     "Node C Promoted",
	})
	if err != nil {
		t.Fatalf("Aggiornamento valido fallito: %v", err)
	}
	if updatedC.ParentID != nil {
		t.Errorf("Atteso ParentID nil, ottenuto %v", updatedC.ParentID)
	}
}

func TestTargetGroupService_GetTree_HierarchyAndCounts(t *testing.T) {
	ctx := context.Background()
	repo := newMockTargetGroupRepo()
	svc := service.NewTargetGroupService(repo)

	tenantID := uuid.New()

	// Costruzione gerarchia:
	// Root 1
	//  ├── Child 1.1
	//  │    └── Leaf 1.1.1
	//  └── Child 1.2
	// Root 2 (senza figli)
	root1, _ := svc.CreateGroup(ctx, service.CreateGroupInput{TenantID: tenantID, Name: "Root 1"})
	child11, _ := svc.CreateGroup(ctx, service.CreateGroupInput{TenantID: tenantID, ParentID: &root1.ID, Name: "Child 1.1"})
	leaf111 := &domain.TargetGroup{ID: uuid.New(), TenantID: tenantID, ParentID: &child11.ID, Name: "Leaf 1.1.1"}
	if err := repo.Create(ctx, leaf111); err != nil {
		t.Fatal(err)
	}
	child12, _ := svc.CreateGroup(ctx, service.CreateGroupInput{TenantID: tenantID, ParentID: &root1.ID, Name: "Child 1.2"})
	root2, _ := svc.CreateGroup(ctx, service.CreateGroupInput{TenantID: tenantID, Name: "Root 2"})

	// Impostazione conteggio target diretti:
	// leaf111: 5
	// child11: 2
	// child12: 3
	// root1: 1
	// root2: 4
	repo.targetCounts[tenantID.String()] = map[uuid.UUID]int{
		leaf111.ID: 5,
		child11.ID: 2,
		child12.ID: 3,
		root1.ID:   1,
		root2.ID:   4,
	}

	tree, err := svc.GetTree(ctx, tenantID)
	if err != nil {
		t.Fatalf("GetTree fallito: %v", err)
	}

	if len(tree) != 2 {
		t.Fatalf("Attesi 2 nodi radice, ottenuti %d", len(tree))
	}

	// Ordinamento alfabetico radici: Root 1, Root 2
	r1 := tree[0]
	r2 := tree[1]
	if r1.Name != "Root 1" || r2.Name != "Root 2" {
		t.Fatalf("Ordinamento radici non corretto: [%s, %s]", r1.Name, r2.Name)
	}

	// Verifica Root 1:
	// DirectTargetCount = 1
	// Recursive TargetCount = 1 + (2 + 5) + 3 = 11
	if r1.DirectTargetCount != 1 {
		t.Errorf("Root 1 DirectTargetCount atteso 1, ottenuto %d", r1.DirectTargetCount)
	}
	if r1.TargetCount != 11 {
		t.Errorf("Root 1 TargetCount ricorsivo atteso 11, ottenuto %d", r1.TargetCount)
	}
	if len(r1.Children) != 2 {
		t.Fatalf("Root 1 attesi 2 figli, ottenuti %d", len(r1.Children))
	}

	// Verifica Child 1.1:
	c11 := r1.Children[0] // Ordinamento alfabetico: Child 1.1, Child 1.2
	if c11.Name != "Child 1.1" {
		t.Fatalf("Atteso Child 1.1, ottenuto %s", c11.Name)
	}
	if c11.DirectTargetCount != 2 {
		t.Errorf("Child 1.1 DirectTargetCount atteso 2, ottenuto %d", c11.DirectTargetCount)
	}
	if c11.TargetCount != 7 { // 2 diretti + 5 di leaf111
		t.Errorf("Child 1.1 TargetCount ricorsivo atteso 7, ottenuto %d", c11.TargetCount)
	}

	// Verifica Leaf 1.1.1:
	if len(c11.Children) != 1 {
		t.Fatalf("Child 1.1 atteso 1 figlio, ottenuti %d", len(c11.Children))
	}
	l111 := c11.Children[0]
	if l111.DirectTargetCount != 5 || l111.TargetCount != 5 {
		t.Errorf("Leaf 1.1.1 attesi 5 target, ottenuti direct=%d, total=%d", l111.DirectTargetCount, l111.TargetCount)
	}

	// Verifica Root 2:
	if r2.DirectTargetCount != 4 || r2.TargetCount != 4 {
		t.Errorf("Root 2 attesi 4 target, ottenuti direct=%d, total=%d", r2.DirectTargetCount, r2.TargetCount)
	}
}

func TestTargetGroupService_GetTree_AggregateRisk(t *testing.T) {
	ctx := context.Background()
	repo := newMockTargetGroupRepo()
	svc := service.NewTargetGroupService(repo)

	tenantID := uuid.New()

	// Radice -> Figlio -> Foglia
	root, _ := svc.CreateGroup(ctx, service.CreateGroupInput{TenantID: tenantID, Name: "DC Milan"})
	child, _ := svc.CreateGroup(ctx, service.CreateGroupInput{TenantID: tenantID, ParentID: &root.ID, Name: "DMZ"})
	leaf := &domain.TargetGroup{ID: uuid.New(), TenantID: tenantID, ParentID: &child.ID, Name: "Web Farm"}
	if err := repo.Create(ctx, leaf); err != nil {
		t.Fatal(err)
	}

	// Imposta severità:
	// leaf ha una vulnerabilità CRITICAL
	// child ha solo MEDIUM
	// root non ha finding diretti
	repo.severityCounts[tenantID.String()] = map[uuid.UUID]map[domain.SeverityLevel]int{
		leaf.ID: {
			domain.SeverityCritical: 1,
			domain.SeverityLow:      3,
		},
		child.ID: {
			domain.SeverityMedium: 2,
		},
	}

	tree, err := svc.GetTree(ctx, tenantID)
	if err != nil {
		t.Fatalf("GetTree fallito: %v", err)
	}

	if len(tree) != 1 {
		t.Fatalf("Atteso 1 nodo radice, ottenuti %d", len(tree))
	}

	rootNode := tree[0]
	childNode := rootNode.Children[0]
	leafNode := childNode.Children[0]

	// Verifica propagazione del rischio CRITICAL a tutti i livelli superiori
	if leafNode.AggregateRisk != domain.SeverityCritical {
		t.Errorf("LeafNode atteso CRITICAL, ottenuto %s", leafNode.AggregateRisk)
	}
	if childNode.AggregateRisk != domain.SeverityCritical {
		t.Errorf("ChildNode atteso CRITICAL propagato, ottenuto %s", childNode.AggregateRisk)
	}
	if rootNode.AggregateRisk != domain.SeverityCritical {
		t.Errorf("RootNode atteso CRITICAL propagato, ottenuto %s", rootNode.AggregateRisk)
	}

	// Conteggio complessivo vulnerabilità propagate a Root
	if rootNode.SeverityCounts[domain.SeverityCritical] != 1 {
		t.Errorf("RootNode atteso 1 Critical, ottenuto %d", rootNode.SeverityCounts[domain.SeverityCritical])
	}
	if rootNode.SeverityCounts[domain.SeverityMedium] != 2 {
		t.Errorf("RootNode atteso 2 Medium, ottenuto %d", rootNode.SeverityCounts[domain.SeverityMedium])
	}
	if rootNode.SeverityCounts[domain.SeverityLow] != 3 {
		t.Errorf("RootNode atteso 3 Low, ottenuto %d", rootNode.SeverityCounts[domain.SeverityLow])
	}
}

func TestTargetGroupService_TenantIsolation(t *testing.T) {
	ctx := context.Background()
	repo := newMockTargetGroupRepo()
	svc := service.NewTargetGroupService(repo)

	tenantA := uuid.New()
	tenantB := uuid.New()

	// Tenant A crea un gruppo
	groupA, err := svc.CreateGroup(ctx, service.CreateGroupInput{
		TenantID: tenantA,
		Name:     "Tenant A Group",
	})
	if err != nil {
		t.Fatalf("Creazione gruppo Tenant A fallita: %v", err)
	}

	// Tenant B tenta di usare groupA come genitore -> deve fallire (ErrNotFound)
	_, err = svc.CreateGroup(ctx, service.CreateGroupInput{
		TenantID: tenantB,
		ParentID: &groupA.ID,
		Name:     "Tenant B Rogue Group",
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Atteso ErrNotFound per violazione tenant isolamento su parent, ottenuto: %v", err)
	}

	// Verifica albero di Tenant B: deve essere vuoto
	treeB, err := svc.GetTree(ctx, tenantB)
	if err != nil {
		t.Fatalf("GetTree Tenant B fallito: %v", err)
	}
	if len(treeB) != 0 {
		t.Errorf("Atteso albero vuoto per Tenant B, ottenuti %d nodi", len(treeB))
	}
}
