package security

// Standard audit actions for authentication events.
const (
	AuditActionLoginSuccess  = "AUTH_LOGIN_SUCCESS"
	AuditActionLoginFailure  = "AUTH_LOGIN_FAILURE"
	AuditActionLogout        = "AUTH_LOGOUT"
	AuditActionTokenRefresh  = "AUTH_TOKEN_REFRESH"
	AuditActionOIDCExchange  = "AUTH_OIDC_EXCHANGE"
)
