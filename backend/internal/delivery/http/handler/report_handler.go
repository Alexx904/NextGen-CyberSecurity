package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

// ReportHandler gestisce gli endpoint di generazione ed esportazione dei report
type ReportHandler struct {
	reportService service.ReportService
}

// NewReportHandler crea una nuova istanza di ReportHandler
func NewReportHandler(reportService service.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
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

	pdfBytes, filename, err := h.reportService.GenerateScanPDFReport(r.Context(), tenantID, scanID)
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


