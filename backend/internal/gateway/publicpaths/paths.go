package publicpaths

import "strings"

// IsPublic reports whether a path should bypass authentication and RBAC.
func IsPublic(path string) bool {
	switch path {
	case "/health", "/ready", "/live", "/metrics", "/api/v1/info":
		return true
	}
	if strings.HasPrefix(path, "/swagger") {
		return true
	}
	if strings.HasPrefix(path, "/api/v1/auth/") {
		switch path {
		case "/api/v1/auth/config",
			"/api/v1/auth/oidc/start",
			"/api/v1/auth/oidc/callback",
			"/api/v1/auth/oidc/exchange",
			"/api/v1/auth/refresh",
			"/api/v1/auth/dev/login",
			"/api/v1/auth/logout",
			"/api/v1/auth/saml/metadata",
			"/api/v1/auth/saml/login",
			"/api/v1/auth/saml/acs":
			return true
		}
	}
	return false
}
