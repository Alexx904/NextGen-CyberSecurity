package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/report"
)

type ComplianceHandler struct{ service *service.ComplianceService }

func NewComplianceHandler(s *service.ComplianceService) *ComplianceHandler {
	return &ComplianceHandler{s}
}
func complianceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.Error(w, 404, "NOT_FOUND", "Risorsa fuori scope o non trovata")
	case errors.Is(err, domain.ErrConflict):
		response.Error(w, 409, "CONFLICT", "Versione non aggiornata o osservazione già associata: ricarica i dati")
	case errors.Is(err, domain.ErrValidation):
		response.Error(w, 400, "VALIDATION", "Dati, cronologia o transizione non validi")
	default:
		response.Error(w, 500, "INTERNAL", "Operazione compliance non riuscita")
	}
}
func decodeCompliance(w http.ResponseWriter, r *http.Request, dest any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 65536)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		response.Error(w, 400, "VALIDATION", "Richiesta non valida")
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		response.Error(w, 400, "VALIDATION", "Una sola richiesta JSON è consentita")
		return false
	}
	return true
}
func complianceIdentity(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	tenant, ok := middleware.GetTenantID(r.Context())
	actor, actorOK := middleware.GetUserID(r.Context())
	if !ok || !actorOK || tenant == uuid.Nil || actor == uuid.Nil {
		response.Error(w, 401, "UNAUTHORIZED", "Sessione richiesta")
		return uuid.Nil, uuid.Nil, false
	}
	return tenant, actor, true
}
func (h *ComplianceHandler) Snapshot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	tenant, _, ok := complianceIdentity(w, r)
	if !ok {
		return
	}
	scope := domain.ComplianceScope{}
	company, err := uuid.Parse(r.URL.Query().Get("company_id"))
	if err != nil {
		complianceError(w, domain.ErrValidation)
		return
	}
	scope.CompanyID = company
	if value := r.URL.Query().Get("department_id"); value != "" {
		scope.DepartmentID, err = uuid.Parse(value)
		if err != nil {
			complianceError(w, domain.ErrValidation)
			return
		}
	}
	s, err := h.service.Snapshot(r.Context(), tenant, scope)
	if err != nil {
		complianceError(w, err)
		return
	}
	response.JSON(w, 200, s)
}
func (h *ComplianceHandler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	tenant, _, ok := complianceIdentity(w, r)
	if !ok {
		return
	}
	var input struct {
		GroupID uuid.UUID `json:"group_id"`
		Name    string    `json:"name"`
	}
	if !decodeCompliance(w, r, &input) {
		return
	}
	asset, err := h.service.CreateAsset(r.Context(), tenant, domain.AssessmentAsset{GroupID: input.GroupID, Name: input.Name})
	if err != nil {
		complianceError(w, err)
		return
	}
	response.JSON(w, 201, asset)
}
func (h *ComplianceHandler) Observe(w http.ResponseWriter, r *http.Request) {
	tenant, actor, ok := complianceIdentity(w, r)
	if !ok {
		return
	}
	var input struct {
		AssetID   uuid.UUID `json:"asset_id"`
		FindingID uuid.UUID `json:"finding_id"`
	}
	if !decodeCompliance(w, r, &input) {
		return
	}
	id, err := h.service.Observe(r.Context(), tenant, actor, input.AssetID, input.FindingID)
	if err != nil {
		complianceError(w, err)
		return
	}
	response.JSON(w, 200, map[string]any{"issue_id": id})
}
func (h *ComplianceHandler) Command(w http.ResponseWriter, r *http.Request) {
	tenant, actor, ok := complianceIdentity(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		complianceError(w, domain.ErrValidation)
		return
	}
	var input struct {
		Action          string                    `json:"action"`
		ExpectedVersion int                       `json:"expected_version"`
		Reason          string                    `json:"reason"`
		Owner           uuid.UUID                 `json:"owner"`
		PolicyID        string                    `json:"policy_id"`
		PolicyVersion   string                    `json:"policy_version"`
		SLAHours        int                       `json:"sla_hours"`
		Verification    *domain.IssueVerification `json:"verification"`
	}
	if !decodeCompliance(w, r, &input) {
		return
	}
	command := domain.IssueCommand{Action: input.Action, ExpectedVersion: input.ExpectedVersion, Reason: input.Reason, Owner: input.Owner, PolicyID: input.PolicyID, PolicyVersion: input.PolicyVersion, SLAHours: input.SLAHours, Verification: input.Verification, Actor: actor}
	if err = h.service.Command(r.Context(), tenant, id, command); err != nil {
		complianceError(w, err)
		return
	}
	response.JSON(w, 200, map[string]bool{"updated": true})
}
func (h *ComplianceHandler) Evidence(w http.ResponseWriter, r *http.Request) {
	tenant, actor, ok := complianceIdentity(w, r)
	if !ok {
		return
	}
	var input struct {
		GroupID       uuid.UUID `json:"group_id"`
		ControlID     string    `json:"control_id"`
		Status        string    `json:"status"`
		Reason        string    `json:"reason"`
		Reference     string    `json:"reference"`
		SourceVersion string    `json:"source_version"`
	}
	if !decodeCompliance(w, r, &input) {
		return
	}
	e := domain.ControlEvidence{GroupID: input.GroupID, ControlID: input.ControlID, Status: input.Status, Reason: input.Reason, Reference: input.Reference, SourceVersion: input.SourceVersion, Actor: actor}
	if err := h.service.Evidence(r.Context(), tenant, e); err != nil {
		complianceError(w, err)
		return
	}
	response.JSON(w, 201, map[string]bool{"recorded": true})
}
func (h *ComplianceHandler) Audit(w http.ResponseWriter, r *http.Request) {
	tenant, _, ok := complianceIdentity(w, r)
	if !ok {
		return
	}
	var input struct {
		Scope    domain.ComplianceScope `json:"scope"`
		Metadata report.AuditMetadata   `json:"metadata"`
	}
	if !decodeCompliance(w, r, &input) {
		return
	}
	pdf, err := h.service.Audit(r.Context(), tenant, input.Scope, input.Metadata)
	if err != nil {
		complianceError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="assessment_audit.pdf"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdf)
}
