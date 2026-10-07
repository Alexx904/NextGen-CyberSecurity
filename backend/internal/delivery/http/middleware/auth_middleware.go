package middleware

import (
	"errors"
	"net/http"
	"strings"

	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/pkg/auth"
)

// Authenticate intercetta la richiesta HTTP, estrae e valida il token Bearer JWT
func Authenticate(tokenManager auth.TokenManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Header Authorization mancante")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || strings.TrimSpace(parts[1]) == "" {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Formato Authorization non valido, atteso 'Bearer <token>'")
				return
			}

			tokenStr := strings.TrimSpace(parts[1])
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

