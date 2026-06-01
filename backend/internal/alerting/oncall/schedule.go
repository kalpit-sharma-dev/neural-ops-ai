package oncall

import (
	"encoding/json"
	"os"
	"time"
)

// Person represents an on-call engineer.
type Person struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	Start time.Time
	End   time.Time
}

type scheduleEntry struct {
	Team     string `json:"team"`
	Timezone string `json:"timezone"`
	Rotation []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"rotation"`
}

// Schedule resolves current on-call from JSON config.
type Schedule struct {
	entries []scheduleEntry
}

// LoadSchedule reads on-call rotations from JSON file.
func LoadSchedule(path string) (*Schedule, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return &Schedule{}, nil
	}
	var entries []scheduleEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, err
	}
	return &Schedule{entries: entries}, nil
}

// Current returns the on-call person for a team at a given time.
func (s *Schedule) Current(team string, at time.Time) *Person {
	for _, entry := range s.entries {
		if entry.Team != team {
			continue
		}
		for _, slot := range entry.Rotation {
			start, err1 := time.Parse(time.RFC3339, slot.Start)
			end, err2 := time.Parse(time.RFC3339, slot.End)
			if err1 != nil || err2 != nil {
				continue
			}
			if !at.Before(start) && at.Before(end) {
				return &Person{Name: slot.Name, Email: slot.Email, Phone: slot.Phone, Start: start, End: end}
			}
		}
	}
	return nil
}
