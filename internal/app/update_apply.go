package app

import (
	"context"
	"fmt"

	"sg-supervisor/internal/updates"
)

func (a *App) ApplyPackage(ctx context.Context, packageID string) (updates.ActiveRecord, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return updates.ActiveRecord{}, err
	}

	licenseStatus, err := a.license.Status(ctx)
	if err != nil {
		return updates.ActiveRecord{}, err
	}
	previousActive, err := a.updates.Active(ctx)
	if err != nil {
		return updates.ActiveRecord{}, err
	}

	runningServices := a.runtime.RunningServiceNames()
	if err := a.stopServicesForUpdate(ctx, runningServices); err != nil {
		return updates.ActiveRecord{}, err
	}

	active, err := a.updates.Apply(ctx, packageID)
	if err != nil {
		return updates.ActiveRecord{}, err
	}
	active.StoppedServices = append(active.StoppedServices, runningServices...)

	restarted, restartErr := a.restartServices(ctx, runningServices, licenseStatus.Valid)
	active.RestartedServices = restarted
	if restartErr == nil {
		if err := a.updates.SaveActive(active); err != nil {
			return updates.ActiveRecord{}, err
		}
		_ = a.saveUpdateOperation("succeeded", "not-needed", active.PackageID, active.PackageID, "")
		return active, nil
	}

	if active.BackupPath == "" {
		_ = a.updates.SaveActive(active)
		_ = a.saveUpdateOperation("failed", "not-available", active.PackageID, active.PackageID, restartErr.Error())
		return active, fmt.Errorf("package applied but failed to restart services: %w", restartErr)
	}
	if err := a.updates.Rollback(ctx, active.BackupPath); err != nil {
		_ = a.updates.SaveActive(active)
		_ = a.saveUpdateOperation("rollback-failed", "failed", active.PackageID, active.PackageID, restartErr.Error()+"; rollback: "+err.Error())
		return active, fmt.Errorf("package applied, failed to restart services, and rollback failed: %w", err)
	}

	rolledBack, restoreErr := a.restartServices(ctx, runningServices, licenseStatus.Valid)
	if previousActive.PackageID == "" {
		if err := a.updates.ClearActive(); err != nil {
			return updates.ActiveRecord{}, err
		}
	} else {
		previousActive.StoppedServices = append([]string(nil), runningServices...)
		previousActive.RestartedServices = rolledBack
		if err := a.updates.SaveActive(previousActive); err != nil {
			return updates.ActiveRecord{}, err
		}
	}
	if restoreErr != nil {
		_ = a.saveUpdateOperation("rolled-back", "succeeded", packageID, previousActive.PackageID, restartErr.Error()+"; restore: "+restoreErr.Error())
		return previousActive, fmt.Errorf("package restart failed, rollback succeeded, but service restore failed: %w", restoreErr)
	}
	_ = a.saveUpdateOperation("rolled-back", "succeeded", packageID, previousActive.PackageID, restartErr.Error())
	return previousActive, fmt.Errorf("package restart failed and supervisor rolled back to the previous package: %w", restartErr)
}

func (a *App) stopServicesForUpdate(ctx context.Context, names []string) error {
	if len(names) == 0 {
		return nil
	}
	if err := a.runtime.StopMany(names); err != nil {
		return err
	}
	waitCtx, cancel := context.WithTimeout(ctx, updateStopTimeout)
	defer cancel()
	return a.runtime.WaitForStopped(waitCtx, names)
}

func (a *App) restartServices(ctx context.Context, names []string, licenseValid bool) ([]string, error) {
	restarted := make([]string, 0, len(names))
	for _, name := range names {
		if err := a.runtime.Start(ctx, name, licenseValid); err != nil {
			return restarted, fmt.Errorf("%s: %w", name, err)
		}
		restarted = append(restarted, name)
	}
	return restarted, nil
}

func (a *App) saveUpdateOperation(outcome, rollbackOutcome, packageID, activePackageID, message string) error {
	return a.updates.SaveOperation(updates.OperationStatus{
		Action:          "apply-package",
		PackageID:       packageID,
		Outcome:         outcome,
		Message:         message,
		RollbackOutcome: rollbackOutcome,
		ActivePackageID: activePackageID,
	})
}
