package finops

import (
	"fmt"
	"regexp"
)

var safeFilter = regexp.MustCompile(`^[a-zA-Z0-9_*.\-]+$`)

// ValidateFilter rejects injection-prone filter values (REQ §8, §10).
func ValidateFilter(value, field string) error {
	if value == "" || value == "all" {
		return nil
	}
	if len(value) > 128 {
		return fmt.Errorf("%s too long", field)
	}
	if !safeFilter.MatchString(value) {
		return fmt.Errorf("invalid %s", field)
	}
	return nil
}
