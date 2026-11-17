package vendorportal

import (
	"testing"

	"github.com/google/uuid"
)

func pct(v float64) *float64 { return &v }

func TestProjectPhysicalPct(t *testing.T) {
	id := uuid.New
	cases := []struct {
		name string
		rows []contractProgress
		want float64
	}{
		{"no contracts", nil, 0},
		{"single", []contractProgress{{id(), 100, pct(42.5)}}, 42.5},
		{"weighted by value", []contractProgress{{id(), 300, pct(50)}, {id(), 100, pct(10)}}, 40},
		{"missing report counts zero", []contractProgress{{id(), 100, pct(80)}, {id(), 100, nil}}, 40},
		{"all zero value weighs equally", []contractProgress{{id(), 0, pct(30)}, {id(), 0, pct(60)}}, 45},
		{"rounded to 2 decimals", []contractProgress{{id(), 1, pct(10)}, {id(), 2, pct(20)}}, 16.67},
	}
	for _, c := range cases {
		if got := projectPhysicalPct(c.rows); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

func TestPeriodLabel(t *testing.T) {
	if got := periodLabel(2026, 3); got != "Maret 2026" {
		t.Fatalf("got %q", got)
	}
}
