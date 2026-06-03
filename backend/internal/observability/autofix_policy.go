package observability

import (
	"os"
	"strings"
)

// AutoFixPolicy governs autonomous remediation (AI-02, bank safety).
type AutoFixPolicy struct {
	Enabled         bool
	RequireApproval bool
	AllowedRoles    map[string]bool
}

// LoadAutoFixPolicy reads AUTOFIX_* environment variables.
func LoadAutoFixPolicy() AutoFixPolicy {
	roles := map[string]bool{}
	for _, r := range strings.Split(os.Getenv("AUTOFIX_ALLOWED_ROLES"), ",") {
		r = strings.TrimSpace(strings.ToUpper(r))
		if r != "" {
			roles[r] = true
		}
	}
	if len(roles) == 0 {
		roles["ADMIN"] = true
		roles["SRE"] = true
	}
	return AutoFixPolicy{
		Enabled:         envTruthy(os.Getenv("AUTOFIX_ENABLED")),
		RequireApproval: !envFalsy(os.Getenv("AUTOFIX_REQUIRE_APPROVAL")),
		AllowedRoles:    roles,
	}
}

func (p AutoFixPolicy) RoleMayExecute(role string) bool {
	if len(p.AllowedRoles) == 0 {
		return true
	}
	return p.AllowedRoles[strings.ToUpper(strings.TrimSpace(role))]
}

func envTruthy(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "1" || v == "true" || v == "yes"
}

func envFalsy(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "0" || v == "false" || v == "no"
}
