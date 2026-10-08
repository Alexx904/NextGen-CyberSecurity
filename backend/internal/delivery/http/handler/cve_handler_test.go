package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nextgen-cybersec/internal/delivery/http/handler"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

type mockCVERepository struct {
	getByIDFunc        func(ctx context.Context, cveID string) (*domain.CVEEntry, error)
	searchByCPEURIFunc func(ctx context.Context, cpeURI string) ([]*domain.CVEEntry, error)
	searchByCPEFunc    func(ctx context.Context, cpePrefix string, version string) ([]*domain.CVEEntry, error)
	searchFunc         func(ctx context.Context, filter domain.CVEFilter) ([]*domain.CVEEntry, int, error)
	upsertBatchFunc    func(ctx context.Context, entries []*domain.CVEEntry) error
	countFunc          func(ctx context.Context) (int64, error)
	getStatsFunc       func(ctx context.Context) (*domain.CVEStats, error)
}

func (m *mockCVERepository) GetByID(ctx context.Context, cveID string) (*domain.CVEEntry, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, cveID)
	}
	return nil, nil
}
func (m *mockCVERepository) SearchByCPEURI(ctx context.Context, cpeURI string) ([]*domain.CVEEntry, error) {
	if m.searchByCPEURIFunc != nil {
		return m.searchByCPEURIFunc(ctx, cpeURI)
	}
	return nil, nil
}
func (m *mockCVERepository) SearchByCPE(ctx context.Context, cpePrefix string, version string) ([]*domain.CVEEntry, error) {
	if m.searchByCPEFunc != nil {
		return m.searchByCPEFunc(ctx, cpePrefix, version)
	}
	return nil, nil
}
func (m *mockCVERepository) Search(ctx context.Context, filter domain.CVEFilter) ([]*domain.CVEEntry, int, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, filter)
	}
	return nil, 0, nil
}
func (m *mockCVERepository) UpsertBatch(ctx context.Context, entries []*domain.CVEEntry) error {
	if m.upsertBatchFunc != nil {
		return m.upsertBatchFunc(ctx, entries)
	}
	return nil
}
func (m *mockCVERepository) Count(ctx context.Context) (int64, error) {
	if m.countFunc != nil {
		return m.countFunc(ctx)
	}
	return 0, nil
}
func (m *mockCVERepository) GetStats(ctx context.Context) (*domain.CVEStats, error) {
	if m.getStatsFunc != nil {
		return m.getStatsFunc(ctx)
	}
	return &domain.CVEStats{}, nil
}

func TestCVEHandler_GetStats_Success(t *testing.T) {
	now := time.Now().UTC()
	expectedStats := &domain.CVEStats{
		TotalCount: 42,
		SeverityCounts: map[domain.SeverityLevel]int64{
			domain.SeverityCritical: 10,
			domain.SeverityHigh:     20,
			domain.SeverityMedium:   12,
		},
		LastUpdatedAt: &now,
	}

	mockRepo := &mockCVERepository{
		getStatsFunc: func(ctx context.Context) (*domain.CVEStats, error) {
			return expectedStats, nil
		},
	}

	h := handler.NewCVEHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cve/stats", nil)
	w := httptest.NewRecorder()

	h.GetStats(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Atteso status 200, ottenuto %d", resp.StatusCode)
	}

	var jsonResp struct {
		Success bool            `json:"success"`
		Data    domain.CVEStats `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		t.Fatalf("Errore unmarshal JSON response: %v", err)
	}

	if !jsonResp.Success {
		t.Errorf("Atteso success true, ottenuto false")
	}
	if jsonResp.Data.TotalCount != 42 {
		t.Errorf("Atteso TotalCount 42, ottenuto %d", jsonResp.Data.TotalCount)
	}
}

func TestCVEHandler_GetStats_RepositoryError(t *testing.T) {
	mockRepo := &mockCVERepository{
		getStatsFunc: func(ctx context.Context) (*domain.CVEStats, error) {
			return nil, errors.New("database connection refused")
		},
	}

	h := handler.NewCVEHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cve/stats", nil)
	w := httptest.NewRecorder()

	h.GetStats(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Atteso status 500, ottenuto %d", resp.StatusCode)
	}
}

func TestCVEHandler_Sync_Success(t *testing.T) {
	mockRepo := &mockCVERepository{
		upsertBatchFunc: func(ctx context.Context, entries []*domain.CVEEntry) error {
			return nil
		},
	}
	syncService := service.NewCVESyncService(mockRepo, 100)

	h := handler.NewCVEHandler(mockRepo, syncService)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cve/sync", nil)
	w := httptest.NewRecorder()

	h.Sync(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Atteso status 200, ottenuto %d", resp.StatusCode)
	}

	var jsonResp struct {
		Success bool              `json:"success"`
		Data    service.SyncStats `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		t.Fatalf("Errore unmarshal JSON response: %v", err)
	}

	if !jsonResp.Success {
		t.Errorf("Atteso success true")
	}
	if jsonResp.Data.TotalProcessed == 0 {
		t.Errorf("Atteso TotalProcessed > 0 dal feed di default")
	}
}

func TestCVEHandler_Sync_NilService(t *testing.T) {
	mockRepo := &mockCVERepository{}
	h := handler.NewCVEHandler(mockRepo, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cve/sync", nil)
	w := httptest.NewRecorder()

	h.Sync(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("Atteso status 503 per servizio non configurato, ottenuto %d", resp.StatusCode)
	}
}

