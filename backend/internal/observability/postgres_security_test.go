package observability

import "testing"

func TestExtractCVE(t *testing.T) {
	got := extractCVE("scanner found cve-2024-12345 in dependency")
	if got != "CVE-2024-12345" {
		t.Fatalf("unexpected cve extraction: %q", got)
	}
	if extractCVE("no identifier here") != "" {
		t.Fatal("expected empty cve when no token exists")
	}
}

func TestExtractIP(t *testing.T) {
	got := extractIP("blocked request from 10.1.2.3 path=/login")
	if got != "10.1.2.3" {
		t.Fatalf("unexpected ip extraction: %q", got)
	}
	if extractIP("no ip present") != "" {
		t.Fatal("expected empty ip when none exists")
	}
}

func TestInferAttackType(t *testing.T) {
	cases := []struct {
		msg  string
		cls  string
		want string
	}{
		{"possible sql injection payload", "", "SQL injection"},
		{"reflected xss attempt", "", "Cross-site scripting"},
		{"../../etc/passwd path traversal", "", "Path traversal"},
		{"outbound callback indicates ssrf", "", "SSRF attempt"},
		{"rce probe detected", "", "Remote code execution attempt"},
		{"odd behavior", "threat", "Suspicious request"},
	}
	for _, tc := range cases {
		if got := inferAttackType(tc.msg, tc.cls); got != tc.want {
			t.Fatalf("inferAttackType(%q,%q) = %q, want %q", tc.msg, tc.cls, got, tc.want)
		}
	}
}

func TestTruncateForUI(t *testing.T) {
	if got := truncateForUI("abcdef", 10); got != "abcdef" {
		t.Fatalf("expected unchanged string, got %q", got)
	}
	if got := truncateForUI("abcdefghijklmnopqrstuvwxyz", 5); got != "abcde…" {
		t.Fatalf("unexpected truncation output: %q", got)
	}
}
