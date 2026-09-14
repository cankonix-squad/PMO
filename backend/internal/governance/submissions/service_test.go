package governance

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// FSM transition table
// ---------------------------------------------------------------------------

func TestTransitionAllowed(t *testing.T) {
	// Valid transitions per the FSM spec:
	//   DRAFT -> SUBMITTED
	//   SUBMITTED -> IN_REVIEW
	//   IN_REVIEW -> APPROVED | REJECTED
	//   APPROVED -> LOCKED
	//   DRAFT/SUBMITTED -> CANCELLED
	//   REJECTED -> DRAFT
	valid := [][2]string{
		{StatusDraft, StatusSubmitted},
		{StatusSubmitted, StatusInReview},
		{StatusInReview, StatusApproved},
		{StatusInReview, StatusRejected},
		{StatusApproved, StatusLocked},
		{StatusDraft, StatusCancelled},
		{StatusSubmitted, StatusCancelled},
		{StatusRejected, StatusDraft},
	}
	for _, tc := range valid {
		if !transitionAllowed(tc[0], tc[1]) {
			t.Errorf("expected transition %s -> %s to be allowed", tc[0], tc[1])
		}
	}

	// Invalid transitions must be rejected
	invalid := [][2]string{
		{StatusDraft, StatusInReview},     // skip submit
		{StatusDraft, StatusApproved},     // skip submit+review
		{StatusDraft, StatusLocked},       // skip everything
		{StatusSubmitted, StatusApproved}, // skip review
		{StatusSubmitted, StatusRejected}, // skip review
		{StatusInReview, StatusSubmitted}, // backwards
		{StatusInReview, StatusLocked},    // skip approve
		{StatusApproved, StatusInReview},  // backwards
		{StatusApproved, StatusCancelled}, // approved cannot cancel
		{StatusLocked, StatusApproved},    // locked is terminal
		{StatusLocked, StatusDraft},       // locked is terminal
		{StatusCancelled, StatusDraft},    // cancelled is terminal
		{"", StatusDraft},                 // empty from
		{StatusDraft, ""},                 // empty to
	}
	for _, tc := range invalid {
		if transitionAllowed(tc[0], tc[1]) {
			t.Errorf("expected transition %s -> %s to be REJECTED", tc[0], tc[1])
		}
	}
}

// ---------------------------------------------------------------------------
// Validation helpers
// ---------------------------------------------------------------------------

func TestValidItemAction(t *testing.T) {
	valid := []string{ItemActionCreate, ItemActionUpdate, ItemActionDelete, ItemActionUpsert, ItemActionValidateOnly}
	for _, a := range valid {
		if !validItemAction(a) {
			t.Errorf("expected action %s to be valid", a)
		}
	}
	invalid := []string{"", "CREATE_NEW", "delete-now", "APPROVE"}
	for _, a := range invalid {
		if validItemAction(a) {
			t.Errorf("expected action %q to be invalid", a)
		}
	}
}

func TestNormalizeEntityType(t *testing.T) {
	valid := map[string]string{
		"project":            "project",
		"projects":           "project",
		"vendor":             "vendor",
		"vendors":            "vendor",
		"government_mapping": "government_mapping",
	}
	for in, want := range valid {
		got, ok := normalizeEntityType(in)
		if !ok {
			t.Errorf("normalizeEntityType(%q) expected ok", in)
		}
		if got != want {
			t.Errorf("normalizeEntityType(%q) = %q, want %q", in, got, want)
		}
	}

	invalid := []string{"", "UNKNOWN", "projects_typo", "gis_data", "GovernmentMapping", "project "}
	for _, in := range invalid {
		if _, ok := normalizeEntityType(in); ok {
			t.Errorf("normalizeEntityType(%q) expected NOT ok", in)
		}
	}
}

func TestIsUniqueViolation(t *testing.T) {
	if isUniqueViolation(nil) {
		t.Error("isUniqueViolation(nil) = true, want false")
	}
	if isUniqueViolation(errors.New("some error")) {
		t.Error("isUniqueViolation(non-pg error) = true, want false")
	}
}

func TestActionRequiresEntityID(t *testing.T) {
	// UPDATE/DELETE/UPSERT mutate an existing entity → entity_id required.
	requires := []string{ItemActionUpdate, ItemActionDelete, ItemActionUpsert}
	for _, a := range requires {
		if !actionRequiresEntityID(a) {
			t.Errorf("actionRequiresEntityID(%q) = false, want true", a)
		}
	}
	// CREATE (new entity) and VALIDATE_ONLY (payload-only) may omit entity_id.
	optional := []string{ItemActionCreate, ItemActionValidateOnly}
	for _, a := range optional {
		if actionRequiresEntityID(a) {
			t.Errorf("actionRequiresEntityID(%q) = true, want false", a)
		}
	}
}

func TestMonthHelpers(t *testing.T) {
	// nil month -> 0 (full-year lock)
	if got := monthPtrOrZero(nil); got != 0 {
		t.Errorf("monthPtrOrZero(nil) = %d, want 0", got)
	}
	// valid month deref
	m := 7
	if got := monthPtrOrZero(&m); got != 7 {
		t.Errorf("monthPtrOrZero(&7) = %d, want 7", got)
	}
	// out of range -> 0
	m2 := 13
	if got := monthPtrOrZero(&m2); got != 0 {
		t.Errorf("monthPtrOrZero(&13) = %d, want 0", got)
	}
	m3 := 0
	if got := monthPtrOrZero(&m3); got != 0 {
		t.Errorf("monthPtrOrZero(&0) = %d, want 0", got)
	}
}
