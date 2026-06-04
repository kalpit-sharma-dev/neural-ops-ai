package observability

import (
	"regexp"
	"strings"
)

// demoFiringAlertServices mirrors typical POC firing alerts for silence preview in demo mode.
var demoFiringAlertServices = []string{
	"payment-service",
	"payment-service",
	"api-gateway",
	"auth-service",
	"ledger-service",
}

// MatchServicePattern returns true when service matches a glob pattern (* wildcard).
func MatchServicePattern(pattern, service string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" || pattern == "*" {
		return true
	}
	escaped := regexp.QuoteMeta(pattern)
	escaped = strings.ReplaceAll(escaped, "\\*", ".*")
	re, err := regexp.Compile("(?i)^" + escaped + "$")
	if err != nil {
		return strings.EqualFold(service, pattern)
	}
	return re.MatchString(service)
}

// CountSilencePreviewMatches counts demo firing alerts that would match a silence pattern.
func CountSilencePreviewMatches(pattern string) int {
	n := 0
	for _, svc := range demoFiringAlertServices {
		if MatchServicePattern(pattern, svc) {
			n++
		}
	}
	return n
}
