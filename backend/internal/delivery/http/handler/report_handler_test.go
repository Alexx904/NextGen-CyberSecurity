package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"nextgen-cybersec/internal/delivery/http/handler"
	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/auth"
)

type mockReportService struct {
	generateScanPDFReportFunc func(ctx context.Context, tenantID, scanID uuid.UUID) ([]byte, string, error)
	generateScanCSVReportFunc func(ctx context.Context, tenantID, scanID uuid.UUID) ([]byte, string, error)
}

func (m *mockReportService) GenerateScanPDFReport(ctx context.Context, tenantID, scanID uuid.UUID) ([]byte, string, error) {
	if m.generateScanPDFReportFunc != nil {
		return m.generateScanPDFReportFunc(ctx, tenantID, scanID)
	}
	return nil, "", domain.ErrNotFound
}

func (m *mockReportService) GenerateScanCSVReport(ctx context.Context, tenantID, scanID uuid.UUID) ([]byte, string, error) {
	if m.generateScanCSVReportFunc != nil {
		return m.generateScanCSVReportFunc(ctx, tenantID, scanID)
	}
	return nil, "", domain.ErrNotFound
}

func setupReportRouter(h *handler.ReportHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/scans/{id}/reports/pdf", h.DownloadPDF)
	r.Get("/scans/{id}/reports/csv", h.DownloadCSV)
	return r
}

func withReportUserClaims(r *http.Request, tenantID uuid.UUID) *http.Request {
	claims := &auth.UserClaims{
		UserID:    uuid.New(),
		TenantID:  tenantID,
		Role:      domain.RoleSecOps,
		Email:     "secops@test.local",
		TokenType: auth.TokenTypeAccess,
	}
	return r.WithContext(middleware.WithUserClaims(r.Context(), claims))
}

func TestReportHandler_DownloadPDF_Success(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()
	dummyPDF := []byte("%PDF-1.4 mock content stream")
	dummyFilename := "report_scan_test.pdf"

	mockSvc := &mockReportService{
		generateScanPDFReportFunc: func(ctx context.Context, tID, sID uuid.UUID) ([]byte, string, error) {
			if tID != tenantID || sID != scanID {
				t.Fatalf("parametri non corrispondenti: tID=%v, sID=%v", tID, sID)
			}
			return dummyPDF, dummyFilename, nil
		},
	}

	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+scanID.String()+"/reports/pdf", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200 OK, ottenuto %d", rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/pdf" {
		t.Errorf("atteso Content-Type application/pdf, ottenuto %s", rec.Header().Get("Content-Type"))
	}

	disposition := rec.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, dummyFilename) {
		t.Errorf("Content-Disposition non contiene il filename corretto: %s", disposition)
	}

	if rec.Body.String() != string(dummyPDF) {
		t.Errorf("body restituito non corrisponde al contenuto PDF")
	}
}

func TestReportHandler_DownloadPDF_Unauthorized(t *testing.T) {
	mockSvc := &mockReportService{}
	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+uuid.NewString()+"/reports/pdf", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("atteso status 401 Unauthorized, ottenuto %d", rec.Code)
	}
}

func TestReportHandler_DownloadPDF_InvalidScanID(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockReportService{}
	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/not-a-valid-uuid/reports/pdf", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("atteso status 400 Bad Request, ottenuto %d", rec.Code)
	}
}

func TestReportHandler_DownloadPDF_NotFound(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	mockSvc := &mockReportService{
		generateScanPDFReportFunc: func(ctx context.Context, tID, sID uuid.UUID) ([]byte, string, error) {
			return nil, "", domain.ErrNotFound
		},
	}

	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+scanID.String()+"/reports/pdf", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("atteso status 404 Not Found, ottenuto %d", rec.Code)
	}
}

func TestReportHandler_DownloadPDF_InternalError(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	mockSvc := &mockReportService{
		generateScanPDFReportFunc: func(ctx context.Context, tID, sID uuid.UUID) ([]byte, string, error) {
			return nil, "", errors.New("generazione fallita")
		},
	}

	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+scanID.String()+"/reports/pdf", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("atteso status 500 Internal Server Error, ottenuto %d", rec.Code)
	}
}

func TestReportHandler_DownloadCSV_Success(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()
	dummyCSV := []byte("finding_id,scan_id,host_ip\n1,2,192.168.1.1\n")
	dummyFilename := "findings_scan_test.csv"

	mockSvc := &mockReportService{
		generateScanCSVReportFunc: func(ctx context.Context, tID, sID uuid.UUID) ([]byte, string, error) {
			if tID != tenantID || sID != scanID {
				t.Fatalf("parametri non corrispondenti: tID=%v, sID=%v", tID, sID)
			}
			return dummyCSV, dummyFilename, nil
		},
	}

	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+scanID.String()+"/reports/csv", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200 OK, ottenuto %d", rec.Code)
	}

	if !strings.Contains(rec.Header().Get("Content-Type"), "text/csv") {
		t.Errorf("atteso Content-Type text/csv, ottenuto %s", rec.Header().Get("Content-Type"))
	}

	disposition := rec.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, dummyFilename) {
		t.Errorf("Content-Disposition non contiene il filename corretto: %s", disposition)
	}

	if rec.Body.String() != string(dummyCSV) {
		t.Errorf("body restituito non corrisponde al contenuto CSV")
	}
}

func TestReportHandler_DownloadCSV_Unauthorized(t *testing.T) {
	mockSvc := &mockReportService{}
	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+uuid.NewString()+"/reports/csv", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("atteso status 401 Unauthorized, ottenuto %d", rec.Code)
	}
}

func TestReportHandler_DownloadCSV_InvalidScanID(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockReportService{}
	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/invalid-uuid/reports/csv", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("atteso status 400 Bad Request, ottenuto %d", rec.Code)
	}
}

func TestReportHandler_DownloadCSV_NotFound(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	mockSvc := &mockReportService{
		generateScanCSVReportFunc: func(ctx context.Context, tID, sID uuid.UUID) ([]byte, string, error) {
			return nil, "", domain.ErrNotFound
		},
	}

	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+scanID.String()+"/reports/csv", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("atteso status 404 Not Found, ottenuto %d", rec.Code)
	}
}


