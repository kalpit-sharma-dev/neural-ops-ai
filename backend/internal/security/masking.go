package security

import (
	"regexp"
	"strings"
)

var (
	panPattern    = regexp.MustCompile(`\b(?:\d[ -]?){12,19}\b`)
	emailPattern  = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	phonePattern  = regexp.MustCompile(`(?:\+?\d{1,3}[-.\s]?)?(?:\(?\d{2,4}\)?[-.\s]?)?\d{3,4}[-.\s]?\d{4,6}`)
	aadhaarPattern = regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`)
	accountPattern = regexp.MustCompile(`\b\d{9,18}\b`)
)

// MaskPAN masks payment card numbers: first 6 + ******* + last 4.
func MaskPAN(input string) string {
	return panPattern.ReplaceAllStringFunc(input, func(match string) string {
		digits := digitsOnly(match)
		if len(digits) < 12 || len(digits) > 19 {
			return match
		}
		return digits[:6] + "*******" + digits[len(digits)-4:]
	})
}

// MaskEmail masks email addresses: j***@gmail.com.
func MaskEmail(input string) string {
	return emailPattern.ReplaceAllStringFunc(input, func(match string) string {
		parts := strings.SplitN(match, "@", 2)
		if len(parts) != 2 || parts[0] == "" {
			return match
		}
		local := parts[0]
		maskedLocal := string(local[0]) + "***"
		return maskedLocal + "@" + parts[1]
	})
}

// MaskPhone masks phone numbers: +91-98****1234.
func MaskPhone(input string) string {
	return phonePattern.ReplaceAllStringFunc(input, func(match string) string {
		digits := digitsOnly(match)
		if len(digits) < 10 {
			return match
		}
		prefix := match
		if strings.HasPrefix(strings.TrimSpace(match), "+") {
			prefix = "+" + digits[:min(2, len(digits)-4)]
		} else if len(digits) >= 12 {
			prefix = "+" + digits[:2]
		} else {
			prefix = digits[:2]
		}
		suffix := digits[len(digits)-4:]
		return prefix + "-** ****" + suffix
	})
}

// MaskAadhaar masks Aadhaar numbers: XXXX-XXXX-1234.
func MaskAadhaar(input string) string {
	return aadhaarPattern.ReplaceAllStringFunc(input, func(match string) string {
		digits := digitsOnly(match)
		if len(digits) != 12 {
			return match
		}
		return "XXXX-XXXX-" + digits[8:]
	})
}

// MaskAccountNumber masks long numeric account identifiers.
func MaskAccountNumber(input string) string {
	return accountPattern.ReplaceAllStringFunc(input, func(match string) string {
		digits := digitsOnly(match)
		if len(digits) < 9 || len(digits) > 18 {
			return match
		}
		if len(digits) <= 4 {
			return match
		}
		return strings.Repeat("*", len(digits)-4) + digits[len(digits)-4:]
	})
}

// MaskMessage applies all PII masking rules to a log message.
func MaskMessage(message string) string {
	if message == "" {
		return message
	}
	out := MaskPAN(message)
	out = MaskAadhaar(out)
	out = MaskAccountNumber(out)
	out = MaskEmail(out)
	out = MaskPhone(out)
	return out
}

func digitsOnly(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
