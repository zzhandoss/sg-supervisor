package manifest

import "testing"

func TestValidateJSON(t *testing.T) {
	valid := []byte(`{
  "productVersion":"1.2.0",
  "coreVersion":"1.2.0",
  "supervisorVersion":"0.1.0",
  "runtime":{"nodeVersion":"20.x"},
  "adapters":[{"key":"dahua-terminal-adapter","version":"0.2.0","required":true}],
  "compatibility":{"coreApi":1,"adapterApi":1}
}`)

	if err := ValidateJSON(valid); err != nil {
		t.Fatalf("expected valid manifest, got %v", err)
	}
}

func TestValidateRejectsMissingVersions(t *testing.T) {
	invalid := File{
		Runtime:       Runtime{NodeVersion: "20.x"},
		Compatibility: Compatibility{CoreAPI: 1, AdapterAPI: 1},
	}

	if err := Validate(invalid); err == nil {
		t.Fatalf("expected validation error")
	}
}
