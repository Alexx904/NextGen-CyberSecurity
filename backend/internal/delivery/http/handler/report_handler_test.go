package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"nextgen-cybersec/internal/delivery/http/handler"
	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/auth"
)

type mockReportService struct {
	listReportsFunc              func(ctx context.Context, tenantID uuid.UUID, filter domain.ReportFilter) ([]*domain.ReportItem, int, error)
	generateScanPDFReportFunc    func(ctx context.Context, tenantID, scanID uuid.UUID, template domain.ReportTemplate) ([]byte, string, error)
	generateScanCSVReportFunc    func(ctx context.Context, tenantID, scanID uuid.UUID) ([]byte, string, error)
	generateCompanyPDFReportFunc func(ctx context.Context, tenantID, groupID uuid.UUID, template domain.ReportTemplate) ([]byte, string, error)
	generateCompanyCSVReportFunc func(ctx context.Context, tenantID, groupID uuid.UUID) ([]byte, string, error)
}

func (m *mockReportService) ListReports(ctx context.Context, tenantID uuid.UUID, filter domain.ReportFilter) ([]*domain.ReportItem, int, error) {
	if m.listReportsFunc != nil {
		return m.listReportsFunc(ctx, tenantID, filter)
	}
	return nil, 0, nil
}

func (m *mockReportService) GenerateScanPDFReport(ctx context.Context, tenantID, scanID uuid.UUID, template domain.ReportTemplate) ([]byte, string, error) {
	if m.generateScanPDFReportFunc != nil {
		return m.generateScanPDFReportFunc(ctx, tenantID, scanID, template)
	}
	return nil, "", domain.ErrNotFound
}

func (m *mockReportService) GenerateScanCSVReport(ctx context.Context, tenantID, scanID uuid.UUID) ([]byte, string, error) {
	if m.generateScanCSVReportFunc != nil {
		return m.generateScanCSVReportFunc(ctx, tenantID, scanID)
	}
	return nil, "", domain.ErrNotFound
}

func (m *mockReportService) GenerateCompanyPDFReport(ctx context.Context, tenantID, groupID uuid.UUID, template domain.ReportTemplate) ([]byte, string, error) {
	if m.generateCompanyPDFReportFunc != nil {
		return m.generateCompanyPDFReportFunc(ctx, tenantID, groupID, template)
	}
	return nil, "", domain.ErrNotFound
}

func (m *mockReportService) GenerateCompanyCSVReport(ctx context.Context, tenantID, groupID uuid.UUID) ([]byte, string, error) {
	if m.generateCompanyCSVReportFunc != nil {
		return m.generateCompanyCSVReportFunc(ctx, tenantID, groupID)
	}
	return nil, "", domain.ErrNotFound
}

func setupReportRouter(h *handler.ReportHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/reports", h.List)
	r.Get("/reports/{id}/download", h.DownloadReport)
	r.Get("/reports/company/{groupId}/download", h.DownloadCompanyReport)
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

func TestReportHandler_List_Success(t *testing.T) {
	tenantID := uuid.New()
	now := time.Now().UTC()
	groupName := "External Perimeter"
	groupID := uuid.New()

	mockItems := []*domain.ReportItem{
		{
			ScanID:       uuid.New(),
			TenantID:     tenantID,
			TargetID:     uuid.New(),
			TargetValue:  "203.0.113.1",
			TargetType:   domain.TargetTypeIP,
			GroupID:      &groupID,
			GroupName:    &groupName,
			ProfileID:    uuid.New(),
			ProfileName:  "Standard Service Scan",
			Status:       domain.ScanStatusCompleted,
			TriggerType:  domain.TriggerManual,
			StartedAt:    &now,
			CompletedAt:  &now,
			MaxSeverity:  domain.SeverityCritical,
			PostureScore: 40,
			PostureGrade: "D",
			CreatedAt:    now,
		},
	}

	mockSvc := &mockReportService{
		listReportsFunc: func(ctx context.Context, tID uuid.UUID, filter domain.ReportFilter) ([]*domain.ReportItem, int, error) {
			if tID != tenantID {
				t.Fatalf("tenantID mismatch: atteso %s, ottenuto %s", tenantID, tID)
			}
			return mockItems, 1, nil
		},
	}

	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/reports?limit=10&offset=0", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200, ottenuto %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Items  []domain.ReportItem `json:"items"`
			Total  int                 `json:"total"`
			Limit  int                 `json:"limit"`
			Offset int                 `json:"offset"`
		} `json:"data"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decodifica risposta fallita: %v", err)
	}

	if !resp.Success {
		t.Errorf("atteso success true")
	}
	if resp.Data.Total != 1 || len(resp.Data.Items) != 1 {
		t.Errorf("dati report restituiti errati: total=%d, items=%d", resp.Data.Total, len(resp.Data.Items))
	}
	if resp.Data.Items[0].PostureScore != 40 || resp.Data.Items[0].PostureGrade != "D" {
		t.Errorf("posture score errato: %+v", resp.Data.Items[0])
	}
}

func TestReportHandler_List_WithFilters(t *testing.T) {
	tenantID := uuid.New()
	targetID := uuid.New()
	groupID := uuid.New()

	filterReceived := false
	mockSvc := &mockReportService{
		listReportsFunc: func(ctx context.Context, tID uuid.UUID, filter domain.ReportFilter) ([]*domain.ReportItem, int, error) {
			filterReceived = true
			if filter.TargetID == nil || *filter.TargetID != targetID {
				t.Errorf("TargetID errato nel filtro: %v", filter.TargetID)
			}
			if filter.GroupID == nil || *filter.GroupID != groupID {
				t.Errorf("GroupID errato nel filtro: %v", filter.GroupID)
			}
			if filter.Severity == nil || *filter.Severity != domain.SeverityHigh {
				t.Errorf("Severity errata nel filtro: %v", filter.Severity)
			}
			if filter.TargetType == nil || *filter.TargetType != domain.TargetTypeCIDR {
				t.Errorf("TargetType errato nel filtro: %v", filter.TargetType)
			}
			if filter.Status == nil || *filter.Status != domain.ScanStatusCompleted {
				t.Errorf("Status errato nel filtro: %v", filter.Status)
			}
			if filter.DateFrom == nil || filter.DateTo == nil {
				t.Errorf("DateFrom o DateTo mancanti nel filtro")
			}
			if filter.Search != "test-subnet" {
				t.Errorf("Search errato: %s", filter.Search)
			}
			return []*domain.ReportItem{}, 0, nil
		},
	}

	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	url := fmt.Sprintf("/reports?target_id=%s&group_id=%s&severity=HIGH&target_type=CIDR&status=COMPLETED&date_from=2026-01-01&date_to=2026-12-31&search=test-subnet", targetID, groupID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200, ottenuto %d", rec.Code)
	}
	if !filterReceived {
		t.Errorf("il filtro non è stato inoltrato al service")
	}
}

func TestReportHandler_List_Unauthorized(t *testing.T) {
	mockSvc := &mockReportService{}
	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/reports", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("atteso 401 Unauthorized, ottenuto %d", rec.Code)
	}
}

func TestReportHandler_List_InvalidParams(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockReportService{}
	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	// Target ID non valido
	req1 := httptest.NewRequest(http.MethodGet, "/reports?target_id=invalid-uuid", nil)
	req1 = withReportUserClaims(req1, tenantID)
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusBadRequest {
		t.Errorf("atteso 400 per target_id non valido, ottenuto %d", rec1.Code)
	}

	// Data non valida
	req2 := httptest.NewRequest(http.MethodGet, "/reports?date_from=not-a-date", nil)
	req2 = withReportUserClaims(req2, tenantID)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("atteso 400 per data non valida, ottenuto %d", rec2.Code)
	}
}

func TestReportHandler_DownloadReport_PDF_Executive(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()
	dummyPDF := []byte("%PDF-1.4 mock executive report")
	dummyFilename := "report_executive_scan.pdf"

	mockSvc := &mockReportService{
		generateScanPDFReportFunc: func(ctx context.Context, tID, sID uuid.UUID, tmpl domain.ReportTemplate) ([]byte, string, error) {
			if tID != tenantID || sID != scanID {
				t.Fatalf("parametri non corrispondenti: tID=%v, sID=%v", tID, sID)
			}
			if tmpl != domain.ReportTemplateExecutive {
				t.Fatalf("atteso template %s, ottenuto %s", domain.ReportTemplateExecutive, tmpl)
			}
			return dummyPDF, dummyFilename, nil
		},
	}

	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/reports/"+scanID.String()+"/download?type=pdf&template=executive", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200, ottenuto %d", rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/pdf" {
		t.Errorf("Content-Type errato: %s", rec.Header().Get("Content-Type"))
	}
	if !strings.Contains(rec.Header().Get("Content-Disposition"), dummyFilename) {
		t.Errorf("Content-Disposition non contiene filename: %s", rec.Header().Get("Content-Disposition"))
	}
}

func TestReportHandler_DownloadReport_PDF_Technical(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()
	dummyPDF := []byte("%PDF-1.4 mock technical report")
	dummyFilename := "report_technical_scan.pdf"

	mockSvc := &mockReportService{
		generateScanPDFReportFunc: func(ctx context.Context, tID, sID uuid.UUID, tmpl domain.ReportTemplate) ([]byte, string, error) {
			if tmpl != domain.ReportTemplateTechnical {
				t.Fatalf("atteso template %s, ottenuto %s", domain.ReportTemplateTechnical, tmpl)
			}
			return dummyPDF, dummyFilename, nil
		},
	}

	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/reports/"+scanID.String()+"/download?type=pdf&template=technical", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200, ottenuto %d", rec.Code)
	}
}

func TestReportHandler_DownloadReport_CSV(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()
	dummyCSV := []byte("finding_id,host_ip,port\n1,192.168.1.1,80\n")
	dummyFilename := "findings_scan.csv"

	mockSvc := &mockReportService{
		generateScanCSVReportFunc: func(ctx context.Context, tID, sID uuid.UUID) ([]byte, string, error) {
			return dummyCSV, dummyFilename, nil
		},
	}

	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/reports/"+scanID.String()+"/download?type=csv", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200, ottenuto %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/csv") {
		t.Errorf("atteso text/csv, ottenuto %s", rec.Header().Get("Content-Type"))
	}
}

func TestReportHandler_DownloadReport_NotFound(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()

	mockSvc := &mockReportService{
		generateScanPDFReportFunc: func(ctx context.Context, tID, sID uuid.UUID, tmpl domain.ReportTemplate) ([]byte, string, error) {
			return nil, "", domain.ErrNotFound
		},
	}

	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/reports/"+scanID.String()+"/download", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("atteso 404 Not Found, ottenuto %d", rec.Code)
	}
}

func TestReportHandler_DownloadReport_InvalidID(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockReportService{}
	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/reports/invalid-uuid/download", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("atteso 400 Bad Request, ottenuto %d", rec.Code)
	}
}

func TestReportHandler_DownloadPDF_Success(t *testing.T) {
	tenantID := uuid.New()
	scanID := uuid.New()
	dummyPDF := []byte("%PDF-1.4 mock content stream")
	dummyFilename := "report_scan_test.pdf"

	mockSvc := &mockReportService{
		generateScanPDFReportFunc: func(ctx context.Context, tID, sID uuid.UUID, tmpl domain.ReportTemplate) ([]byte, string, error) {
			if tID != tenantID || sID != scanID {
				t.Fatalf("parametri non corrispondenti: tID=%v, sID=%v", tID, sID)
			}
			return dummyPDF, dummyFilename, nil
		},
	}

	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/scans/"+scanID.String()+"/reports/pdf?template=technical", nil)
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
		generateScanPDFReportFunc: func(ctx context.Context, tID, sID uuid.UUID, tmpl domain.ReportTemplate) ([]byte, string, error) {
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
		generateScanPDFReportFunc: func(ctx context.Context, tID, sID uuid.UUID, tmpl domain.ReportTemplate) ([]byte, string, error) {
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

func TestReportHandler_DownloadCompanyReport_PDF_Success(t *testing.T) {
	tenantID := uuid.New()
	groupID := uuid.New()
	dummyPDF := []byte("%PDF-1.4 Company Report")
	expectedFilename := "report_azienda_executive_acme_corp_20261010.pdf"

	mockSvc := &mockReportService{
		generateCompanyPDFReportFunc: func(ctx context.Context, tID, gID uuid.UUID, tmpl domain.ReportTemplate) ([]byte, string, error) {
			if tID != tenantID || gID != groupID {
				return nil, "", domain.ErrNotFound
			}
			return dummyPDF, expectedFilename, nil
		},
	}

	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/reports/company/"+groupID.String()+"/download?type=pdf&template=executive", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200 OK, ottenuto %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("atteso Content-Type application/pdf, ottenuto %s", ct)
	}
	if !strings.Contains(rec.Header().Get("Content-Disposition"), expectedFilename) {
		t.Errorf("atteso header filename contenente %s", expectedFilename)
	}
	if rec.Body.String() != string(dummyPDF) {
		t.Errorf("contenuto PDF non corrispondente")
	}
}

func TestReportHandler_DownloadCompanyReport_CSV_Success(t *testing.T) {
	tenantID := uuid.New()
	groupID := uuid.New()
	dummyCSV := []byte("Host IP,Port,Protocol,Service,Severity\n192.168.1.10,80,tcp,http,MEDIUM\n")
	expectedFilename := "findings_azienda_acme_corp_20261010.csv"

	mockSvc := &mockReportService{
		generateCompanyCSVReportFunc: func(ctx context.Context, tID, gID uuid.UUID) ([]byte, string, error) {
			if tID != tenantID || gID != groupID {
				return nil, "", domain.ErrNotFound
			}
			return dummyCSV, expectedFilename, nil
		},
	}

	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/reports/company/"+groupID.String()+"/download?type=csv", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("atteso status 200 OK, ottenuto %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/csv") {
		t.Errorf("atteso Content-Type text/csv, ottenuto %s", ct)
	}
	if !strings.Contains(rec.Header().Get("Content-Disposition"), expectedFilename) {
		t.Errorf("atteso header filename contenente %s", expectedFilename)
	}
}

func TestReportHandler_DownloadCompanyReport_InvalidGroupID(t *testing.T) {
	tenantID := uuid.New()
	mockSvc := &mockReportService{}
	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/reports/company/invalid-uuid/download", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("atteso status 400 Bad Request, ottenuto %d", rec.Code)
	}
}

func TestReportHandler_DownloadCompanyReport_NotFound(t *testing.T) {
	tenantID := uuid.New()
	groupID := uuid.New()

	mockSvc := &mockReportService{
		generateCompanyPDFReportFunc: func(ctx context.Context, tID, gID uuid.UUID, tmpl domain.ReportTemplate) ([]byte, string, error) {
			return nil, "", domain.ErrNotFound
		},
	}

	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/reports/company/"+groupID.String()+"/download", nil)
	req = withReportUserClaims(req, tenantID)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("atteso status 404 Not Found, ottenuto %d", rec.Code)
	}
}

func TestReportHandler_DownloadCompanyReport_Unauthorized(t *testing.T) {
	groupID := uuid.New()
	mockSvc := &mockReportService{}
	h := handler.NewReportHandler(mockSvc)
	router := setupReportRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/reports/company/"+groupID.String()+"/download", nil)
	// Nessun claim di tenant iniettato nel contesto
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("atteso status 401 Unauthorized, ottenuto %d", rec.Code)
	}
}

