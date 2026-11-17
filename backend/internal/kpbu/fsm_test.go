package kpbu

import (
	"testing"
	"time"

	"github.com/harmanto-49/cankora/internal/shared/types"
)

func strp(s string) *string { return &s }

func TestCanTransition(t *testing.T) {
	cases := []struct {
		from string
		prev *string
		to   string
		want bool
	}{
		{StageIdentified, nil, StagePlanning, true},
		{StageIdentified, nil, StageTransaction, false},
		{StageTransaction, nil, StageContractSigned, true},
		{StageTransaction, nil, StageOnHold, true},
		{StageTransaction, nil, StageTerminated, false},
		{StageConstruction, nil, StageTerminated, true},
		{StageConstruction, nil, StageOnHold, false},
		{StageConstruction, nil, StageCancelled, false},
		{StageHandover, nil, StageClosed, true},
		{StageClosed, nil, StagePlanning, false},
		{StageCancelled, nil, StageIdentified, false},
		{StageOnHold, strp(StagePreparationOBC), StagePreparationOBC, true},
		{StageOnHold, strp(StagePreparationOBC), StagePlanning, false},
		{StageOnHold, nil, StageCancelled, true},
	}
	for _, c := range cases {
		if got := CanTransition(c.from, c.prev, c.to); got != c.want {
			t.Errorf("%s -> %s: got %v want %v", c.from, c.to, got, c.want)
		}
	}
}

func TestIsKnownStage(t *testing.T) {
	for _, s := range append(MainStages, StageClosed, StageOnHold, StageCancelled, StageTerminated) {
		if !IsKnownStage(s) {
			t.Errorf("%s should be known", s)
		}
	}
	if IsKnownStage("FOO") {
		t.Error("FOO should be unknown")
	}
}

func TestDelay(t *testing.T) {
	d := func(s string) *types.FlexTime {
		tm, _ := time.Parse("2006-01-02", s)
		return &types.FlexTime{Time: tm}
	}
	now, _ := time.Parse("2006-01-02", "2026-11-17")
	cases := []struct {
		name            string
		planned, actual *types.FlexTime
		want            int
	}{
		{"no plan", nil, nil, 0},
		{"future plan", d("2026-12-01"), nil, 0},
		{"overdue open", d("2026-11-07"), nil, 10},
		{"finished late", d("2026-10-01"), d("2026-10-06"), 5},
		{"finished early", d("2026-10-01"), d("2026-09-20"), 0},
	}
	for _, c := range cases {
		if got := delay(c.planned, c.actual, now); got != c.want {
			t.Errorf("%s: got %d want %d", c.name, got, c.want)
		}
	}
}
