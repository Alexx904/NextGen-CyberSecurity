package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
)

// RequireTenant garantisce che ogni richiesta elaborata sia associata a un tenant valido (NFR-SEC-03)
func RequireTenant() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			tenantID, ok := GetTenantID(ctx)
			if !ok || tenantID == uuid.Nil {
				response.Error(w, http.StatusUnauthorized, "TENANT_REQUIRED", "Tenant non identificato nel contesto di sicurezza")
				return
			}

			// Se è presente l'header X-Tenant-ID (es. SuperAdmin che gestisce specifici tenant)
			if overrideHeader := r.Header.Get("X-Tenant-ID"); overrideHeader != "" {
				role, _ := GetUserRole(ctx)
				overrideUUID, err := uuid.Parse(overrideHeader)
				if err != nil {
					response.Error(w, http.StatusBadRequest, "INVALID_TENANT_ID", "Formato UUID non valido nell'header X-Tenant-ID")
					return
				}

				// Solo SuperAdmin può accedere ad altri tenant tramite header
				if role != domain.RoleSuperAdmin && overrideUUID != tenantID {
					response.Error(w, http.StatusForbidden, "FORBIDDEN", "Non sei autorizzato a operare su un tenant differente dal tuo")
					return
				}

				if role == domain.RoleSuperAdmin {
					ctx = context.WithValue(ctx, tenantIDKey, overrideUUID)
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

