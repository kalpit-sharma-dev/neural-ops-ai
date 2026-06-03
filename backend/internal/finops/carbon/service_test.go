package carbon

import "testing"

func TestFootprint(t *testing.T) {
	s := NewService()
	fp := s.Footprint(nil, "all", "team")
	if fp.Methodology == "" {
		t.Fatal("methodology required")
	}
	if fp.SCI.RequestsPerMonth == 0 {
		t.Fatal("SCI required")
	}
}

func TestRecommendations(t *testing.T) {
	recs := NewService().Recommendations()
	if len(recs) == 0 {
		t.Fatal("expected carbon recommendations")
	}
}
