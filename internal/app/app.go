package app

import (
	"context"
	"path/filepath"

	"sg-supervisor/internal/config"
	"sg-supervisor/internal/control"
	"sg-supervisor/internal/license"
	"sg-supervisor/internal/manifest"
	"sg-supervisor/internal/runtime"
	"sg-supervisor/internal/servicehost"
	"sg-supervisor/internal/setup"
	"sg-supervisor/internal/updates"
)

type App struct {
	root    string
	layout  config.Layout
	cfg     config.SupervisorConfig
	license *license.Store
	product *config.ProductStore
	runtime *runtime.Manager
	setup   *setup.Store
	updates *updates.Store
	runner  servicehost.Runner
}

func New(root string) (*App, error) {
	layout := config.NewLayout(root)
	cfg, err := config.LoadOrCreate(layout.ConfigFile)
	if err != nil {
		return nil, err
	}
	product := config.NewProductStore(layout)
	if err := product.Ensure(); err != nil {
		return nil, err
	}
	productCfg, err := product.Load()
	if err != nil {
		return nil, err
	}
	catalog, err := config.LoadServiceCatalog(layout)
	if err != nil {
		return nil, err
	}
	catalog = config.ApplyProductConfig(layout, catalog, productCfg)
	return &App{
		root:    root,
		layout:  layout,
		cfg:     cfg,
		license: license.NewStore(layout, cfg),
		product: product,
		runtime: runtime.NewManager(catalog),
		setup:   setup.NewStore(layout),
		updates: updates.NewStore(layout, cfg),
		runner:  servicehost.ExecRunner{},
	}, nil
}

func (a *App) EnsureBootstrap(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := config.EnsureLayout(a.layout); err != nil {
		return err
	}
	if err := config.EnsureServiceCatalog(a.layout); err != nil {
		return err
	}
	if err := a.product.Ensure(); err != nil {
		return err
	}
	if err := a.setup.Ensure(ctx); err != nil {
		return err
	}
	return a.syncRuntimeConfig()
}

func (a *App) Status(ctx context.Context) (control.StatusResponse, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return control.StatusResponse{}, err
	}

	licenseStatus, err := a.license.Status(ctx)
	if err != nil {
		return control.StatusResponse{}, err
	}
	setupStatus, err := a.SetupStatus(ctx)
	if err != nil {
		return control.StatusResponse{}, err
	}
	importedPackages, err := a.updates.List(ctx)
	if err != nil {
		return control.StatusResponse{}, err
	}
	lastUpdate, err := a.updates.Operation(ctx)
	if err != nil {
		return control.StatusResponse{}, err
	}
	activePackage, err := a.updates.Active(ctx)
	if err != nil {
		return control.StatusResponse{}, err
	}
	serviceStatuses := a.runtime.StatusesWithHealth(ctx)
	productConfig, err := a.ProductConfigStatus(ctx)
	if err != nil {
		return control.StatusResponse{}, err
	}

	return control.StatusResponse{
		ProductName:   a.cfg.ProductName,
		Root:          a.root,
		ListenAddr:    a.cfg.ListenAddress,
		SetupRequired: !setupStatus.Complete,
		Setup:         setupStatus,
		Directories: control.DirectoryStatus{
			Install:  a.layout.InstallDir,
			Config:   a.layout.ConfigDir,
			Data:     a.layout.DataDir,
			Logs:     a.layout.LogsDir,
			Licenses: a.layout.LicensesDir,
			Backups:  a.layout.BackupsDir,
			Runtime:  a.layout.RuntimeDir,
			Updates:  a.layout.UpdatesDir,
		},
		License:          licenseStatus,
		LastUpdate:       mapLastUpdate(lastUpdate),
		ManagedServices:  managedServiceNames(serviceStatuses),
		Services:         serviceStatuses,
		ImportedPackages: mapPackageRecords(importedPackages),
		ActivePackage:    mapActivePackage(activePackage),
		ProductConfig:    productConfig,
	}, nil
}

func (a *App) GenerateActivationRequest(ctx context.Context, customer, outputPath string) (string, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return "", err
	}
	if outputPath == "" {
		outputPath = filepath.Join(a.layout.LicensesDir, "activation-request.json")
	}
	request, err := license.BuildActivationRequest(customer)
	if err != nil {
		return "", err
	}
	if err := license.WriteActivationRequest(outputPath, request); err != nil {
		return "", err
	}
	return outputPath, nil
}

func (a *App) ImportLicense(ctx context.Context, path string) error {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return err
	}
	return a.license.Import(ctx, path)
}

func (a *App) StartService(ctx context.Context, name string) error {
	status, err := a.license.Status(ctx)
	if err != nil {
		return err
	}
	return a.runtime.Start(ctx, name, status.Valid)
}

func (a *App) StopService(name string) error {
	return a.runtime.Stop(name)
}

func (a *App) RestartService(ctx context.Context, name string) error {
	status, err := a.license.Status(ctx)
	if err != nil {
		return err
	}
	return a.runtime.Restart(ctx, name, status.Valid)
}

func (a *App) ImportPackageManifest(ctx context.Context, path string) (updates.Record, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return updates.Record{}, err
	}
	return a.updates.ImportManifest(ctx, path)
}

func (a *App) ImportPackageBundle(ctx context.Context, path string) (updates.Record, error) {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return updates.Record{}, err
	}
	return a.updates.ImportBundle(ctx, path)
}

func (a *App) Serve(ctx context.Context, listen string) error {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return err
	}
	if listen != "" {
		a.cfg.ListenAddress = listen
	}
	server := control.NewServer(a.cfg.ListenAddress, control.HandlerDependencies{
		Status: func(ctx context.Context) (control.StatusResponse, error) {
			return a.Status(ctx)
		},
		GenerateActivationRequest: func(ctx context.Context, customer, outputPath string) (string, error) {
			return a.GenerateActivationRequest(ctx, customer, outputPath)
		},
		ImportLicense: func(ctx context.Context, path string) error {
			return a.ImportLicense(ctx, path)
		},
		StartService: func(ctx context.Context, name string) error {
			return a.StartService(ctx, name)
		},
		StopService: func(name string) error {
			return a.StopService(name)
		},
		RestartService: func(ctx context.Context, name string) error {
			return a.RestartService(ctx, name)
		},
		ImportPackageManifest: func(ctx context.Context, path string) (control.PackageRecord, error) {
			record, err := a.ImportPackageManifest(ctx, path)
			if err != nil {
				return control.PackageRecord{}, err
			}
			return mapPackageRecord(record), nil
		},
		ImportPackageBundle: func(ctx context.Context, path string) (control.PackageRecord, error) {
			record, err := a.ImportPackageBundle(ctx, path)
			if err != nil {
				return control.PackageRecord{}, err
			}
			return mapPackageRecord(record), nil
		},
		ApplyLocalBundle: func(ctx context.Context, path string) (control.ActivePackageRecord, error) {
			record, err := a.ApplyLocalBundle(ctx, path)
			if err != nil {
				return control.ActivePackageRecord{}, err
			}
			return mapActivePackage(record), nil
		},
		ApplyPackage: func(ctx context.Context, packageID string) (control.ActivePackageRecord, error) {
			record, err := a.ApplyPackage(ctx, packageID)
			if err != nil {
				return control.ActivePackageRecord{}, err
			}
			return mapActivePackage(record), nil
		},
		UpdateSetupField:           a.UpdateSetupField,
		UpdateProductConfig:        a.UpdateProductConfig,
		InstallPackage:             a.InstallPackage,
		Repair:                     a.Repair,
		Uninstall:                  a.Uninstall,
		RenderServiceHostArtifacts: a.RenderServiceHostArtifacts,
		ValidateManifest: func(data []byte) error {
			return manifest.ValidateJSON(data)
		},
	})
	return server.Run(ctx)
}

func managedServiceNames(statuses []runtime.ServiceStatus) []string {
	names := make([]string, 0, len(statuses))
	for _, status := range statuses {
		names = append(names, status.Name)
	}
	return names
}

func mapPackageRecords(records []updates.Record) []control.PackageRecord {
	result := make([]control.PackageRecord, 0, len(records))
	for _, record := range records {
		result = append(result, mapPackageRecord(record))
	}
	return result
}

func mapPackageRecord(record updates.Record) control.PackageRecord {
	return control.PackageRecord{
		PackageID:         record.PackageID,
		SourceType:        record.SourceType,
		SourcePath:        record.SourcePath,
		ArchivePath:       record.ArchivePath,
		StoredPath:        record.StoredPath,
		StageDir:          record.StageDir,
		ImportedAt:        record.ImportedAt,
		ProductVersion:    record.Manifest.ProductVersion,
		CoreVersion:       record.Manifest.CoreVersion,
		SupervisorVersion: record.Manifest.SupervisorVersion,
		Adapters:          updatesAdapterLabels(record),
	}
}

func mapActivePackage(record updates.ActiveRecord) control.ActivePackageRecord {
	return control.ActivePackageRecord{
		PackageID:         record.PackageID,
		AppliedAt:         record.AppliedAt,
		ProductVersion:    record.ProductVersion,
		CoreVersion:       record.CoreVersion,
		SupervisorVersion: record.SupervisorVersion,
		ManifestPath:      record.ManifestPath,
		BackupPath:        record.BackupPath,
		Adapters:          record.Adapters,
		StoppedServices:   record.StoppedServices,
		RestartedServices: record.RestartedServices,
	}
}

func mapLastUpdate(status updates.OperationStatus) control.UpdateOperationStatus {
	return control.UpdateOperationStatus{
		Action:          status.Action,
		PackageID:       status.PackageID,
		Outcome:         status.Outcome,
		Message:         status.Message,
		RollbackOutcome: status.RollbackOutcome,
		ActivePackageID: status.ActivePackageID,
		RecordedAt:      status.RecordedAt,
	}
}

func updatesAdapterLabels(record updates.Record) []string {
	adapters := make([]string, 0, len(record.Manifest.Adapters))
	for _, adapter := range record.Manifest.Adapters {
		adapters = append(adapters, adapter.Key+"@"+adapter.Version)
	}
	return adapters
}
