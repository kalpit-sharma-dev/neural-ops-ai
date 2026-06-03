package finops

import "testing"

func TestAllowedScope(t *testing.T) {
	if !AllowedScope(nil, "all") {
		t.Fatal("all scope should be allowed")
	}
	if !AllowedScope([]string{"payments"}, "payments") {
		t.Fatal("matching scope should be allowed")
	}
	if AllowedScope([]string{"platform"}, "payments") {
		t.Fatal("mismatched scope should be denied")
	}
	if !AllowedScope([]string{"*"}, "payments") {
		t.Fatal("wildcard should allow any scope")
	}
}
