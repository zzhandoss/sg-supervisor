package manifest

import (
	"encoding/json"
	"errors"
)

type File struct {
	ProductVersion    string          `json:"productVersion"`
	CoreVersion       string          `json:"coreVersion"`
	SupervisorVersion string          `json:"supervisorVersion"`
	Runtime           Runtime         `json:"runtime"`
	Adapters          []AdapterBundle `json:"adapters,omitempty"`
	Compatibility     Compatibility   `json:"compatibility"`
}

type Runtime struct {
	NodeVersion string `json:"nodeVersion"`
}

type AdapterBundle struct {
	Key      string `json:"key"`
	Version  string `json:"version"`
	Required bool   `json:"required"`
}

type Compatibility struct {
	CoreAPI    int `json:"coreApi"`
	AdapterAPI int `json:"adapterApi"`
}

func Validate(file File) error {
	if file.ProductVersion == "" {
		return errors.New("productVersion is required")
	}
	if file.CoreVersion == "" {
		return errors.New("coreVersion is required")
	}
	if file.SupervisorVersion == "" {
		return errors.New("supervisorVersion is required")
	}
	if file.Runtime.NodeVersion == "" {
		return errors.New("runtime.nodeVersion is required")
	}
	if file.Compatibility.CoreAPI <= 0 {
		return errors.New("compatibility.coreApi must be positive")
	}
	if file.Compatibility.AdapterAPI <= 0 {
		return errors.New("compatibility.adapterApi must be positive")
	}
	for _, adapter := range file.Adapters {
		if adapter.Key == "" || adapter.Version == "" {
			return errors.New("adapter key and version are required")
		}
	}
	return nil
}

func ValidateJSON(data []byte) error {
	var file File
	if err := json.Unmarshal(data, &file); err != nil {
		return err
	}
	return Validate(file)
}
