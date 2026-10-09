package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

// ReportHandler gestisce gli endpoint di consultazione ed esportazione per l'Hub Centralizzato dei Report
type ReportHandler struct {
	reportService service.ReportService
}

// NewReportHandler crea una nuova istanza di ReportHandler
func NewReportHandler(reportService service.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

// List gestisce GET /api/v1/reports per interrogare lo storico aggregato con filtri e paginazione
func (h *ReportHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Identificativo tenant mancante o non valido")
		return
	}

	filter := domain.ReportFilter{
		Limit:  20,
		Offset: 0,
	}

	if targetParam := r.URL.Query().Get("target_id"); targetParam != "" {
		if id, err := uuid.Parse(targetParam); err == nil {
			filter.TargetID = &id
		} else {
			response.Error(w, http.StatusBadRequest, "INVALID_PARAM", "target_id non valido")
			return
		}
	}

	if groupParam := r.URL.Query().Get("group_id"); groupParam != "" {
		if id, err := uuid.Parse(groupParam); err == nil {
			filter.GroupID = &id
		} else {
			response.Error(w, http.StatusBadRequest, "INVALID_PARAM", "group_id non valido")
			return
		}
	}

	sevParam := r.URL.Query().Get("severity")
	if sevParam == "" {
		sevParam = r.URL.Query().Get("max_severity")
	}
	if sevParam != "" {
		sev := domain.SeverityLevel(strings.ToUpper(sevParam))
		filter.Severity = &sev
	}

	if typeParam := r.URL.Query().Get("type"); typeParam != "" {
		if typeParam != "pdf" && typeParam != "csv" { // Evita conflitto con query export format se riutilizzato
			tType := domain.TargetType(strings.ToUpper(typeParam))
			if tType.IsValid() {
				filter.TargetType = &tType
			}
		}
	}
	if targetTypeParam := r.URL.Query().Get("target_type"); targetTypeParam != "" {
		tType := domain.TargetType(strings.ToUpper(targetTypeParam))
		if tType.IsValid() {
			filter.TargetType = &tType
		}
	}

	if statusParam := r.URL.Query().Get("status"); statusParam != "" {
		st := domain.ScanStatus(strings.ToUpper(statusParam))
		if st.IsValid() {
			filter.Status = &st
		}
	}

	dateFromStr := r.URL.Query().Get("date_from")
	if dateFromStr == "" {
		dateFromStr = r.URL.Query().Get("from")
	}
	if dateFromStr != "" {
		dFrom, err := parseDate(dateFromStr, false)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "INVALID_PARAM", "Formato date_from non valido")
			return
		}
		filter.DateFrom = dFrom
	}

	dateToStr := r.URL.Query().Get("date_to")
	if dateToStr == "" {
		dateToStr = r.URL.Query().Get("to")
	}
	if dateToStr != "" {
		dTo, err := parseDate(dateToStr, true)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "INVALID_PARAM", "Formato date_to non valido")
			return
		}
		filter.DateTo = dTo
	}

	filter.Search = strings.TrimSpace(r.URL.Query().Get("search"))

	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			filter.Limit = val
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			filter.Offset = val
		}
	}

	items, total, err := h.reportService.ListReports(r.Context(), tenantID, filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante il recupero dello storico report")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"items":  items,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// DownloadReport gestisce GET /api/v1/reports/{id}/download con supporto unificato a type=pdf|csv e template=executive|technical
func (h *ReportHandler) DownloadReport(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Identificativo tenant mancante o non valido")
		return
	}

	idParam := chi.URLParam(r, "id")
	scanID, err := uuid.Parse(idParam)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID scansione non valido")
		return
	}

	formatType := strings.ToLower(r.URL.Query().Get("type"))
	if formatType == "csv" {
		csvBytes, filename, err := h.reportService.GenerateScanCSVReport(r.Context(), tenantID, scanID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				response.Error(w, http.StatusNotFound, "NOT_FOUND", "Scansione non trovata")
				return
			}
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante la generazione del report CSV")
			return
		}

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(csvBytes)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(csvBytes)
		return
	}

	// Default: Formato PDF
	templateParam := domain.ReportTemplate(strings.ToLower(r.URL.Query().Get("template")))
	if !templateParam.IsValid() {
		templateParam = domain.ReportTemplateExecutive
	}

	pdfBytes, filename, err := h.reportService.GenerateScanPDFReport(r.Context(), tenantID, scanID, templateParam)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Scansione o risorse correlate non trovate")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante la generazione del report PDF")
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}

// DownloadPDF gestisce GET /api/v1/scans/{id}/reports/pdf per il download del report PDF White-Label
func (h *ReportHandler) DownloadPDF(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Identificativo tenant mancante o non valido")
		return
	}

	idParam := chi.URLParam(r, "id")
	scanID, err := uuid.Parse(idParam)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID scansione non valido")
		return
	}

	templateParam := domain.ReportTemplate(strings.ToLower(r.URL.Query().Get("template")))
	if !templateParam.IsValid() {
		templateParam = domain.ReportTemplateExecutive
	}

	pdfBytes, filename, err := h.reportService.GenerateScanPDFReport(r.Context(), tenantID, scanID, templateParam)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Scansione o risorse correlate non trovate")
			return
		}
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", "Parametri di richiesta report non validi")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante la generazione del report PDF")
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}

// DownloadCSV gestisce GET /api/v1/scans/{id}/reports/csv per il download dei finding in formato CSV (RFC 4180)
func (h *ReportHandler) DownloadCSV(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Identificativo tenant mancante o non valido")
		return
	}

	idParam := chi.URLParam(r, "id")
	scanID, err := uuid.Parse(idParam)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "ID scansione non valido")
		return
	}

	csvBytes, filename, err := h.reportService.GenerateScanCSVReport(r.Context(), tenantID, scanID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Scansione o risorse correlate non trovate")
			return
		}
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", "Parametri di richiesta report non validi")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore durante la generazione del report CSV")
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(csvBytes)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvBytes)
}

func parseDate(dateStr string, isEndOfDay bool) (*time.Time, error) {
	clean := strings.TrimSpace(dateStr)
	if clean == "" {
		return nil, nil
	}

	if t, err := time.Parse(time.RFC3339, clean); err == nil {
		return &t, nil
	}

	if t, err := time.Parse("2006-01-02", clean); err == nil {
		if isEndOfDay {
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
		return &t, nil
	}

	return nil, fmt.Errorf("formato data non supportato: %s", dateStr)
}
