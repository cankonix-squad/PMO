package government

// ---------------------------------------------------------------------------
// Registry — hardcoded default connector definitions
// ---------------------------------------------------------------------------
// Each connector maps to one or more government data sources (SIRUP,
// SIMPONI, OM-SPAN, or their mock equivalents when credentials are absent).
// ---------------------------------------------------------------------------

// defaultConnectors is the canonical list of government connectors.
// Keys must match the ConnectorXxx constants in model.go.
var defaultConnectors = []ConnectorDefinition{
	{
		Key:  ConnectorProjectRegistry,
		Name: "Government Project Registry",
		Description: "Ingests project reference data from the government project " +
			"registry (SIRUP or equivalent). Provides procurement plan numbers, " +
			"project codes, and linked budget references for matching against " +
			"PMO projects.",
		DatasetTypes: []string{DatasetProjects},
	},
	{
		Key:  ConnectorBudgetReference,
		Name: "Government Budget Reference",
		Description: "Ingests budget allocation and realisation data from OM-SPAN " +
			"or SIMPONI. Enables comparison between planned budget (DIPA) and " +
			"actual spend reported in the government financial system.",
		DatasetTypes: []string{DatasetBudgetAllocation},
	},
	{
		Key:  ConnectorLocationReference,
		Name: "Government Location Reference",
		Description: "Ingests administrative location hierarchy (province, " +
			"district, sub-district) from the government reference data service. " +
			"Used to enrich project spatial attributes and validate field inspection " +
			"location codes.",
		DatasetTypes: []string{DatasetLocations},
	},
	{
		Key:  ConnectorVendorReference,
		Name: "Government Vendor Reference",
		Description: "Ingests registered vendor (rekanan) data from the government " +
			"procurement system (SIRUP/e-Katalog). Used to cross-reference PMO " +
			"vendor records against official procurement registrations.",
		DatasetTypes: []string{DatasetVendors},
	},
}

// AllowedConnectorKeys is the set of valid connector keys for validation.
var AllowedConnectorKeys = func() map[string]bool {
	m := make(map[string]bool, len(defaultConnectors))
	for _, c := range defaultConnectors {
		m[c.Key] = true
	}
	return m
}()

// AllowedDatasetTypes is the set of valid dataset types for validation.
var AllowedDatasetTypes = map[string]bool{
	DatasetProjects:         true,
	DatasetBudgetAllocation: true,
	DatasetLocations:        true,
	DatasetVendors:          true,
}

// AllowedModes is the set of valid sync modes for validation.
var AllowedModes = map[string]bool{
	ModeSample: true,
	ModeDryRun: true,
	ModeCommit: true,
}

// connectorDatasetTypes maps connector key → allowed dataset types.
var connectorDatasetTypes = func() map[string]map[string]bool {
	m := make(map[string]map[string]bool, len(defaultConnectors))
	for _, c := range defaultConnectors {
		inner := make(map[string]bool, len(c.DatasetTypes))
		for _, dt := range c.DatasetTypes {
			inner[dt] = true
		}
		m[c.Key] = inner
	}
	return m
}()

// ListConnectors returns a copy of all registered connector definitions
// enriched with the current operational state from cfg.
func ListConnectors(cfg map[string]ConnectorConfig) []ConnectorDefinition {
	result := make([]ConnectorDefinition, len(defaultConnectors))
	for i, c := range defaultConnectors {
		copy := c
		if cc, ok := cfg[c.Key]; ok {
			copy.State = cc.State
			if cc.BaseURL != "" {
				copy.BaseURL = cc.BaseURL
			}
		} else {
			copy.State = ConnectorStateNotConfigured
		}
		result[i] = copy
	}
	return result
}

// GetConnector returns a single connector definition by key, or false if not found.
func GetConnector(key string, cfg map[string]ConnectorConfig) (ConnectorDefinition, bool) {
	for _, c := range defaultConnectors {
		if c.Key == key {
			copy := c
			if cc, ok := cfg[key]; ok {
				copy.State = cc.State
				if cc.BaseURL != "" {
					copy.BaseURL = cc.BaseURL
				}
			} else {
				copy.State = ConnectorStateNotConfigured
			}
			return copy, true
		}
	}
	return ConnectorDefinition{}, false
}

// DatasetTypesForConnector returns the allowed dataset types for a connector key.
func DatasetTypesForConnector(key string) (map[string]bool, bool) {
	m, ok := connectorDatasetTypes[key]
	return m, ok
}
