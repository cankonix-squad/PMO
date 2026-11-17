package vendorportal

import (
	"strings"
	"testing"
)

func TestNormalizeDecision(t *testing.T) {
	if d, _, err := normalizeDecision(ReviewRequest{Decision: "verified"}); err != nil || d != StatusVerified {
		t.Fatalf("verified: %v %v", d, err)
	}
	if _, _, err := normalizeDecision(ReviewRequest{Decision: "REJECTED"}); err == nil {
		t.Fatal("reject without reason must fail")
	}
	if d, n, err := normalizeDecision(ReviewRequest{Decision: "REJECTED", Notes: " foto buram "}); err != nil || d != StatusRejected || n != "foto buram" {
		t.Fatalf("reject: %v %q %v", d, n, err)
	}
	if _, _, err := normalizeDecision(ReviewRequest{Decision: "APPROVED"}); err == nil {
		t.Fatal("unknown decision must fail")
	}
}

func TestTemporaryPassword(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		p, err := temporaryPassword()
		if err != nil {
			t.Fatal(err)
		}
		if len(p) != 16 {
			t.Fatalf("length %d", len(p))
		}
		if !strings.ContainsAny(p, "ABCDEFGHJKLMNPQRSTUVWXYZ") || !strings.ContainsAny(p, "abcdefghijkmnpqrstuvwxyz") ||
			!strings.ContainsAny(p, "23456789") || !strings.ContainsAny(p, "!@#$%*?") {
			t.Fatalf("missing character class: %s", p)
		}
		if seen[p] {
			t.Fatal("duplicate password")
		}
		seen[p] = true
	}
}
