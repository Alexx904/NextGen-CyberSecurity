package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"nextgen-cybersec/internal/delivery/http/middleware"
	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
	"nextgen-cybersec/pkg/auth"
)

// LoginRequest definisce il payload per l'autenticazione
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest definisce il payload per il rinnovo dei token
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// UserResponse modella il profilo pubblico dell'utente (senza hash password)
type UserResponse struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginResponse raggruppa le informazioni di profilo e la coppia di token
type LoginResponse struct {
	User   UserResponse    `json:"user"`
	Tokens *auth.TokenPair `json:"tokens"`
}

// AuthHandler gestisce gli endpoint di autenticazione e identità
type AuthHandler struct {
	authService service.AuthService
	ticketStore auth.StreamTicketStore
}

// NewAuthHandler crea una nuova istanza del controller Auth
func NewAuthHandler(authService service.AuthService, ticketStore ...auth.StreamTicketStore) *AuthHandler {
	var ts auth.StreamTicketStore
	if len(ticketStore) > 0 {
		ts = ticketStore[0]
	}
	return &AuthHandler{
		authService: authService,
		ticketStore: ts,
	}
}

// IssueStreamTicket emette un ticket monouso di breve durata per stream SSE e WebSocket POST /api/v1/auth/stream-ticket
func (h *AuthHandler) IssueStreamTicket(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok || claims == nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autenticazione richiesta")
		return
	}

	if h.ticketStore == nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gestore stream ticket non configurato")
		return
	}

	ticket, err := h.ticketStore.IssueTicket(claims, 30*time.Second)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "TICKET_GENERATION_FAILED", "Impossibile generare stream ticket")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"ticket":     ticket.Ticket,
		"expires_in": 30,
	})
}

// Login elabora le credenziali inviate via POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Payload JSON della richiesta non valido")
		return
	}

	out, err := h.authService.Login(r.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", "Email e password sono campi obbligatori")
			return
		}
		if errors.Is(err, domain.ErrUnauthorized) {
			response.Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Credenziali non valide o account disattivato")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Errore interno durante il login")
		return
	}

	resp := LoginResponse{
		User:   toUserResponse(out.User),
		Tokens: out.Tokens,
	}

	response.JSON(w, http.StatusOK, resp)
}

// Me restituisce il profilo dell'utente autenticato GET /api/v1/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, okT := middleware.GetTenantID(ctx)
	userID, okU := middleware.GetUserID(ctx)

	if !okT || !okU {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Contesto di autenticazione non valido")
		return
	}

	user, err := h.authService.GetCurrentUser(ctx, tenantID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", "Profilo utente non trovato")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Impossibile recuperare il profilo utente")
		return
	}

	response.JSON(w, http.StatusOK, toUserResponse(user))
}

// RefreshToken rinnova la coppia di token JWT POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Campo 'refresh_token' mancante o non valido")
		return
	}

	tokens, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "Refresh token non valido, scaduto o revocato")
		return
	}

	response.JSON(w, http.StatusOK, tokens)
}

func toUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID.String(),
		TenantID:  u.TenantID.String(),
		Email:     u.Email,
		FullName:  u.FullName,
		Role:      string(u.Role),
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

