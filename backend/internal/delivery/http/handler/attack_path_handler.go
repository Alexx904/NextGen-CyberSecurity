package handler

import (
	"net/http"
	"strconv"

	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

// AttackPathHandler gestisce gli endpoint REST per la consultazione del grafo delle minacce e dei chokepoint.
type AttackPathHandler struct {
	attackPathService service.AttackPathService
}

// NewAttackPathHandler crea una nuova istanza di AttackPathHandler.
func NewAttackPathHandler(attackPathService service.AttackPathService) *AttackPathHandler {
	return &AttackPathHandler{
		attackPathService: attackPathService,
	}
}

// GetGraph gestisce GET /api/v1/attack-paths/graph restituendo la topologia completa del grafo e i percorsi di penetrazione.
func (h *AttackPathHandler) GetGraph(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	minCVSS := 0.0
	if minCVSSStr := r.URL.Query().Get("min_cvss"); minCVSSStr != "" {
		if val, err := strconv.ParseFloat(minCVSSStr, 64); err == nil && val >= 0.0 {
			minCVSS = val
		}
	}

	graph, err := h.attackPathService.GenerateAttackGraph(r.Context(), tenantID, minCVSS)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Impossibile generare il grafo di attacco")
		return
	}

	response.JSON(w, http.StatusOK, graph)
}

// GetChokepoints gestisce GET /api/v1/attack-paths/chokepoints restituendo la lista prioritizzata dei punti critici di remediation.
func (h *AttackPathHandler) GetChokepoints(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Tenant non identificato nel contesto")
		return
	}

	chokepoints, err := h.attackPathService.GetChokepoints(r.Context(), tenantID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Impossibile recuperare i chokepoint")
		return
	}

	if chokepoints == nil {
		chokepoints = make([]domain.Chokepoint, 0)
	}

	response.JSON(w, http.StatusOK, chokepoints)
}

