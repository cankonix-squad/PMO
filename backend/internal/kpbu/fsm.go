package kpbu

// preContractStages can be put ON_HOLD or CANCELLED.
var preContractStages = map[string]bool{
	StageIdentified: true, StagePlanning: true, StagePreparationOBC: true,
	StagePreparationFBC: true, StageTransaction: true,
}

// contractedStages can be TERMINATED (early termination of the agreement).
var contractedStages = map[string]bool{
	StageContractSigned: true, StageFinancialClose: true, StageConstruction: true, StageOperation: true,
}

var nextMainStage = map[string]string{
	StageIdentified:     StagePlanning,
	StagePlanning:       StagePreparationOBC,
	StagePreparationOBC: StagePreparationFBC,
	StagePreparationFBC: StageTransaction,
	StageTransaction:    StageContractSigned,
	StageContractSigned: StageFinancialClose,
	StageFinancialClose: StageConstruction,
	StageConstruction:   StageOperation,
	StageOperation:      StageHandover,
	StageHandover:       StageClosed,
}

// IsKnownStage reports whether s is a valid stage value.
func IsKnownStage(s string) bool {
	if _, ok := nextMainStage[s]; ok {
		return true
	}
	switch s {
	case StageClosed, StageOnHold, StageCancelled, StageTerminated:
		return true
	}
	return false
}

// AllowedTransitions lists the stages reachable from current.
// previous is the stage a project was in before ON_HOLD (resume target).
func AllowedTransitions(current string, previous *string) []string {
	if current == StageOnHold {
		out := []string{}
		if previous != nil && preContractStages[*previous] {
			out = append(out, *previous)
		}
		return append(out, StageCancelled)
	}
	out := []string{}
	if next, ok := nextMainStage[current]; ok {
		out = append(out, next)
	}
	if preContractStages[current] {
		out = append(out, StageOnHold, StageCancelled)
	}
	if contractedStages[current] {
		out = append(out, StageTerminated)
	}
	return out
}

// CanTransition reports whether current → to is allowed.
func CanTransition(current string, previous *string, to string) bool {
	for _, s := range AllowedTransitions(current, previous) {
		if s == to {
			return true
		}
	}
	return false
}

// isMainStage reports whether s is tracked with a stage milestone.
func isMainStage(s string) bool {
	for _, m := range MainStages {
		if m == s {
			return true
		}
	}
	return false
}
