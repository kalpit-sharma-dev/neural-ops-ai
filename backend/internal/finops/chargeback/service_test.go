package chargeback

import "testing"

func TestGenerateStatement(t *testing.T) {
	s := NewService()
	st := s.GenerateStatement(nil, "payments", "chargeback", "monthly")
	if st.Mode != "chargeback" {
		t.Fatalf("expected chargeback mode")
	}
	if st.ExportURL == "" {
		t.Fatal("export URL required")
	}
	csv := s.ExportCSV(st)
	if csv == "" {
		t.Fatal("csv export required")
	}
}
