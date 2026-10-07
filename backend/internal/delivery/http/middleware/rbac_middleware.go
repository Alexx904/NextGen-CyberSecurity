package middleware

import (
	"net/http"

	"nextgen-cybersec/internal/delivery/http/response"
	"nextgen-cybersec/internal/domain"
)

// RequireRole verifica che il ruolo dell'utente appartenga alla lista dei ruoli consentiti
func RequireRole(allowedRoles ...domain.UserRole) func(next http.Handler) http.Handler {
	allowedMap := make(map[domain.UserRole]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		allowedMap[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := GetUserRole(r.Context())
			if !ok {
				response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Ruolo utente non rilevato nel contesto di autenticazione")
				return
			}

			// Se il ruolo non è nella lista ammessa
			if !allowedMap[userRole] {
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "Permessi insufficienti per accedere alla risorsa richiesta")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireSuperAdmin consente l'accesso unicamente al ruolo SuperAdmin
func RequireSuperAdmin() func(next http.Handler) http.Handler {
	return RequireRole(domain.RoleSuperAdmin)
}

// RequireAdmin consente l'accesso agli amministratori (SuperAdmin o TenantAdmin)
func RequireAdmin() func(next http.Handler) http.Handler {
	return RequireRole(domain.RoleSuperAdmin, domain.RoleTenantAdmin)
}

// RequireSecOps consente l'accesso a chi può eseguire operazioni operative di sicurezza e scansione
func RequireSecOps() func(next http.Handler) http.Handler {
	return RequireRole(domain.RoleSuperAdmin, domain.RoleTenantAdmin, domain.RoleSecOps)
}

