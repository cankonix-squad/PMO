package government

import "os"

// ---------------------------------------------------------------------------
// Config — reads non-secret connector configuration from environment variables
// ---------------------------------------------------------------------------
// IMPORTANT: This file MUST NOT read, log, or return API keys, passwords,
// tokens, or any other secrets. Only return metadata safe for API responses.
// ---------------------------------------------------------------------------

// envOrDefault returns the env var value or a fallback.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envBool returns true if the env var is set to "true" or "1".
func envBool(key string) bool {
	v := os.Getenv(key)
	return v == "true" || v == "1"
}

// connectorState derives the ConnectorState constant based on config.
// NOT_CONFIGURED → connector disabled or no base URL set
// SANDBOX_SAMPLE → enabled but sandbox/mock flag active
// ACTIVE         → fully configured and live
func connectorState(enabled bool, baseURL string, sandbox bool) string {
	if !enabled || baseURL == "" {
		return ConnectorStateNotConfigured
	}
	if sandbox {
		return ConnectorStateSandboxSample
	}
	return ConnectorStateActive
}

// LoadConfig reads government connector configuration from environment
// variables and returns a map of ConnectorConfig keyed by connector key.
//
// Environment variables (all optional, defaults to not-configured/sandbox):
//
//	GOV_PROJECT_REGISTRY_ENABLED      = true|false
//	GOV_PROJECT_REGISTRY_BASE_URL     = https://...
//	GOV_PROJECT_REGISTRY_SANDBOX      = true|false
//
//	GOV_BUDGET_REF_ENABLED            = true|false
//	GOV_BUDGET_REF_BASE_URL           = https://...
//	GOV_BUDGET_REF_SANDBOX            = true|false
//
//	GOV_LOCATION_REF_ENABLED          = true|false
//	GOV_LOCATION_REF_BASE_URL         = https://...
//	GOV_LOCATION_REF_SANDBOX          = true|false
//
//	GOV_VENDOR_REF_ENABLED            = true|false
//	GOV_VENDOR_REF_BASE_URL           = https://...
//	GOV_VENDOR_REF_SANDBOX            = true|false
func LoadConfig() map[string]ConnectorConfig {
	configs := []struct {
		key        string
		enabledEnv string
		baseURLEnv string
		sandboxEnv string
	}{
		{
			key:        ConnectorProjectRegistry,
			enabledEnv: "GOV_PROJECT_REGISTRY_ENABLED",
			baseURLEnv: "GOV_PROJECT_REGISTRY_BASE_URL",
			sandboxEnv: "GOV_PROJECT_REGISTRY_SANDBOX",
		},
		{
			key:        ConnectorBudgetReference,
			enabledEnv: "GOV_BUDGET_REF_ENABLED",
			baseURLEnv: "GOV_BUDGET_REF_BASE_URL",
			sandboxEnv: "GOV_BUDGET_REF_SANDBOX",
		},
		{
			key:        ConnectorLocationReference,
			enabledEnv: "GOV_LOCATION_REF_ENABLED",
			baseURLEnv: "GOV_LOCATION_REF_BASE_URL",
			sandboxEnv: "GOV_LOCATION_REF_SANDBOX",
		},
		{
			key:        ConnectorVendorReference,
			enabledEnv: "GOV_VENDOR_REF_ENABLED",
			baseURLEnv: "GOV_VENDOR_REF_BASE_URL",
			sandboxEnv: "GOV_VENDOR_REF_SANDBOX",
		},
	}

	result := make(map[string]ConnectorConfig, len(configs))
	for _, c := range configs {
		enabled := envBool(c.enabledEnv)
		baseURL := envOrDefault(c.baseURLEnv, "")
		sandbox := envBool(c.sandboxEnv)
		state := connectorState(enabled, baseURL, sandbox)

		result[c.key] = ConnectorConfig{
			ConnectorKey: c.key,
			Enabled:      enabled,
			BaseURL:      baseURL, // safe: no secrets; URL only
			State:        state,
			SandboxMode:  sandbox,
		}
	}
	return result
}
