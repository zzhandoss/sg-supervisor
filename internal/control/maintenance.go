package control

import (
	"context"
	"net/http"

	"sg-supervisor/internal/maintenance"
)

type InstallReport struct {
	Completed       bool     `json:"completed"`
	PackageID       string   `json:"packageId"`
	ActivePackageID string   `json:"activePackageId"`
	ServiceName     string   `json:"serviceName"`
	WrittenFiles    []string `json:"writtenFiles"`
	InstallHints    []string `json:"installHints"`
	Issues          []Issue  `json:"issues,omitempty"`
}

type RepairReport struct {
	Completed           bool     `json:"completed"`
	EnsuredPaths        []string `json:"ensuredPaths"`
	ServiceArtifacts    []string `json:"serviceArtifacts"`
	ActivePackageID     string   `json:"activePackageId,omitempty"`
	NeedsPackageInstall bool     `json:"needsPackageInstall"`
	Issues              []Issue  `json:"issues,omitempty"`
}

type Issue struct {
	Step     string `json:"step"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type UninstallReport struct {
	Mode            string   `json:"mode"`
	Completed       bool     `json:"completed"`
	RemovedPaths    []string `json:"removedPaths"`
	KeptPaths       []string `json:"keptPaths"`
	StoppedServices []string `json:"stoppedServices,omitempty"`
	UninstallHints  []string `json:"uninstallHints,omitempty"`
	Issues          []Issue  `json:"issues,omitempty"`
}

func (s *Server) handleInstallPackage(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PackageID  string `json:"packageId"`
		BinaryPath string `json:"binaryPath"`
	}
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	report, err := s.deps.InstallPackage(r.Context(), request.PackageID, request.BinaryPath)
	if err != nil {
		if len(report.Issues) > 0 || report.ActivePackageID != "" || len(report.WrittenFiles) > 0 || report.Completed {
			writeErrorWithData(w, http.StatusConflict, err, report)
			return
		}
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": report})
}

func (s *Server) handleRepair(w http.ResponseWriter, r *http.Request) {
	var request struct {
		BinaryPath string `json:"binaryPath"`
	}
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	report, err := s.deps.Repair(r.Context(), request.BinaryPath)
	if err != nil {
		if len(report.Issues) > 0 || len(report.EnsuredPaths) > 0 || len(report.ServiceArtifacts) > 0 || report.Completed {
			writeErrorWithData(w, http.StatusConflict, err, report)
			return
		}
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": report})
}

func (s *Server) handleUninstall(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Mode string `json:"mode"`
	}
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	report, err := s.deps.Uninstall(r.Context(), request.Mode)
	if err != nil {
		if len(report.Issues) > 0 || len(report.RemovedPaths) > 0 || report.Completed {
			writeErrorWithData(w, http.StatusConflict, err, report)
			return
		}
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": report})
}

type Installer func(context.Context, string, string) (InstallReport, error)
type Repairer func(context.Context, string) (RepairReport, error)
type Uninstaller func(context.Context, string) (UninstallReport, error)

func mapIssues(issues []maintenance.Issue) []Issue {
	result := make([]Issue, 0, len(issues))
	for _, issue := range issues {
		result = append(result, Issue{
			Step:     issue.Step,
			Severity: issue.Severity,
			Message:  issue.Message,
		})
	}
	return result
}
