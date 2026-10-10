package middleware

import (
	"errors"
	"net/http"
	"strings"

	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/pkg/auth"
)

// Authenticate intercetta la richiesta HTTP, estrae e valida il token Bearer JWT o uno StreamTicket monouso.
// Supporta in via prioritaria '?ticket=<ticket>' per connessioni real-time (SSE e WebSocket),
// l'header standard 'Authorization: Bearer <token>' per API REST, e '?token=<token>' come fallback legacy.
func Authenticate(tokenManager auth.TokenManager, ticketStore ...auth.StreamTicketStore) func(next http.Handler) http.Handler {
	var ts auth.StreamTicketStore
	if len(ticketStore) > 0 {
		ts = ticketStore[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Controllo primario stream-ticket monouso via query param '?ticket=<ticket>'
			if queryTicket := r.URL.Query().Get("ticket"); strings.TrimSpace(queryTicket) != "" && ts != nil {
				claims, err := ts.ConsumeTicket(strings.TrimSpace(queryTicket))
				if err != nil {
					response.Error(w, http.StatusUnauthorized, "INVALID_STREAM_TICKET", "Stream ticket non valido, scaduto o già utilizzato")
					return
				}
				ctx := WithUserClaims(r.Context(), claims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// 2. Controllo header Authorization: Bearer <token>
			var tokenStr string
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || strings.TrimSpace(parts[1]) == "" {
					response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Formato Authorization non valido, atteso 'Bearer <token>'")
					return
				}
				tokenStr = strings.TrimSpace(parts[1])
			} else if queryToken := r.URL.Query().Get("token"); strings.TrimSpace(queryToken) != "" {
				tokenStr = strings.TrimSpace(queryToken)
			} else {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Header Authorization mancante")
				return
			}

			claims, err := tokenManager.ValidateToken(tokenStr)
			if err != nil {
				if errors.Is(err, auth.ErrExpiredToken) {
					response.Error(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "Token di accesso scaduto")
					return
				}
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token non valido o firma non verificabile")
				return
			}

			// Consenti solo token di tipo 'access' sugli endpoint API protetti
			if claims.TokenType != auth.TokenTypeAccess {
				response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN_TYPE", "Tipo di token non valido per l'accesso alle risorse API")
				return
			}

			// Inietta i claim e i riferimenti di identità nel contesto
			ctx := WithUserClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

