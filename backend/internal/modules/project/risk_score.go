package project

// RiskScore computes the risk score as probability × impact and derives the
// severity class from the 5×5 risk matrix. Scores are clamped to the valid
// range so business logic stays consistent even if callers pass out-of-range
// values (e.g. partial form data).
//
// Matrix (SRS 6.3, extended to 5 levels):
//
//	score 1-4   → LOW
//	score 5-9   → MEDIUM
//	score 10-15 → HIGH
//	score 16-25 → CRITICAL
func RiskScore(probability, impact int) int {
	p := clampRiskLevel(probability)
	i := clampRiskLevel(impact)
	return p * i
}

// RiskSeverity derives the severity class from probability and impact.
func RiskSeverity(probability, impact int) string {
	switch score := RiskScore(probability, impact); {
	case score >= 16:
		return "CRITICAL"
	case score >= 10:
		return "HIGH"
	case score >= 5:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// clampRiskLevel keeps a 1-5 risk level within bounds.
func clampRiskLevel(v int) int {
	if v < 1 {
		return 1
	}
	if v > 5 {
		return 5
	}
	return v
}
