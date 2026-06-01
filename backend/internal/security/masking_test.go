package security

import "testing"

func TestMaskPAN(t *testing.T) {
	input := "Payment failed for card 4111111111111111"
	got := MaskPAN(input)
	want := "411111*******1111"
	if got != "Payment failed for card "+want {
		t.Fatalf("MaskPAN() = %q, want masked PAN", got)
	}
}

func TestMaskEmail(t *testing.T) {
	input := "contact john.doe@gmail.com for support"
	got := MaskEmail(input)
	if got != "contact j***@gmail.com for support" {
		t.Fatalf("MaskEmail() = %q", got)
	}
}

func TestMaskPhone(t *testing.T) {
	input := "call +919876543210"
	got := MaskPhone(input)
	if got == input {
		t.Fatalf("expected phone to be masked, got %q", got)
	}
}

func TestMaskMessage(t *testing.T) {
	input := "user jane@test.com paid with 4111111111111111 from +919876543210"
	got := MaskMessage(input)
	if got == input {
		t.Fatal("expected MaskMessage to alter input containing PII")
	}
}
