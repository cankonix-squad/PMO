package project

import "testing"

func TestNormalizeAgreementForm(t *testing.T) {
	if f, err := normalizeAgreementForm(""); f != nil || err != nil {
		t.Error("empty must be nil")
	}
	if f, err := normalizeAgreementForm(" spk "); err != nil || *f != "SPK" {
		t.Error("spk must normalize")
	}
	if _, err := normalizeAgreementForm("MOU"); err != ErrInvalidAgreementForm {
		t.Error("MOU must be rejected")
	}
}
