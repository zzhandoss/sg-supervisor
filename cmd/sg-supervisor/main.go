package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"sg-supervisor/internal/app"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("expected command: init-layout | status | generate-activation-request | import-license | import-package-manifest | import-package-bundle | apply-package | bootstrap-install | set-setup-field | install-package | repair | uninstall | assemble-package | build-distribution | build-release | build-release-set | render-service-host | serve")
	}

	switch args[0] {
	case "init-layout":
		return runInitLayout(ctx, args[1:])
	case "status":
		return runStatus(args[1:])
	case "generate-activation-request":
		return runActivationRequest(ctx, args[1:])
	case "import-license":
		return runImportLicense(ctx, args[1:])
	case "import-package-manifest":
		return runImportPackageManifest(ctx, args[1:])
	case "import-package-bundle":
		return runImportPackageBundle(ctx, args[1:])
	case "apply-package":
		return runApplyPackage(ctx, args[1:])
	case "bootstrap-install":
		return runBootstrapInstall(ctx, args[1:])
	case "set-setup-field":
		return runSetSetupField(ctx, args[1:])
	case "install-package":
		return runInstallPackage(ctx, args[1:])
	case "repair":
		return runRepair(ctx, args[1:])
	case "uninstall":
		return runUninstall(ctx, args[1:])
	case "assemble-package":
		return runAssemblePackage(ctx, args[1:])
	case "build-distribution":
		return runBuildDistribution(ctx, args[1:])
	case "build-release":
		return runBuildRelease(ctx, args[1:])
	case "build-release-set":
		return runBuildReleaseSet(ctx, args[1:])
	case "render-service-host":
		return runRenderServiceHost(ctx, args[1:])
	case "serve":
		return runServe(ctx, args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runInitLayout(ctx context.Context, args []string) error {
	supervisor, err := newApp(args)
	if err != nil {
		return err
	}
	return supervisor.EnsureBootstrap(ctx)
}

func runStatus(args []string) error {
	supervisor, err := newApp(args)
	if err != nil {
		return err
	}
	status, err := supervisor.Status(context.Background())
	if err != nil {
		return err
	}
	fmt.Println(status.PrettyString())
	return nil
}

func runActivationRequest(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("generate-activation-request", flag.ContinueOnError)
	root := fs.String("root", ".", "supervisor root")
	customer := fs.String("customer", "", "customer label for request metadata")
	output := fs.String("output", "", "output path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	supervisor, err := app.New(*root)
	if err != nil {
		return err
	}
	path, err := supervisor.GenerateActivationRequest(ctx, *customer, *output)
	if err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}

func runImportLicense(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("import-license", flag.ContinueOnError)
	root := fs.String("root", ".", "supervisor root")
	licensePath := fs.String("license", "", "license file path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *licensePath == "" {
		return errors.New("license path is required")
	}

	supervisor, err := app.New(*root)
	if err != nil {
		return err
	}
	return supervisor.ImportLicense(ctx, *licensePath)
}

func runServe(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	root := fs.String("root", ".", "supervisor root")
	listen := fs.String("listen", "0.0.0.0:8787", "listen address")
	if err := fs.Parse(args); err != nil {
		return err
	}

	supervisor, err := app.New(*root)
	if err != nil {
		return err
	}
	if err := supervisor.EnsureBootstrap(ctx); err != nil {
		return err
	}

	serverCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	return supervisor.Serve(serverCtx, *listen)
}

func runImportPackageManifest(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("import-package-manifest", flag.ContinueOnError)
	root := fs.String("root", ".", "supervisor root")
	manifestPath := fs.String("manifest", "", "manifest file path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *manifestPath == "" {
		return errors.New("manifest path is required")
	}

	supervisor, err := app.New(*root)
	if err != nil {
		return err
	}
	record, err := supervisor.ImportPackageManifest(ctx, *manifestPath)
	if err != nil {
		return err
	}
	fmt.Println(record.StoredPath)
	return nil
}

func runImportPackageBundle(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("import-package-bundle", flag.ContinueOnError)
	root := fs.String("root", ".", "supervisor root")
	bundlePath := fs.String("bundle", "", "bundle zip path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *bundlePath == "" {
		return errors.New("bundle path is required")
	}

	supervisor, err := app.New(*root)
	if err != nil {
		return err
	}
	record, err := supervisor.ImportPackageBundle(ctx, *bundlePath)
	if err != nil {
		return err
	}
	fmt.Println(record.PackageID)
	return nil
}

func runApplyPackage(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("apply-package", flag.ContinueOnError)
	root := fs.String("root", ".", "supervisor root")
	packageID := fs.String("package-id", "", "imported package id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *packageID == "" {
		return errors.New("package-id is required")
	}

	supervisor, err := app.New(*root)
	if err != nil {
		return err
	}
	record, err := supervisor.ApplyPackage(ctx, *packageID)
	if err != nil {
		return err
	}
	fmt.Println(record.PackageID)
	return nil
}

func newApp(args []string) (*app.App, error) {
	fs := flag.NewFlagSet("command", flag.ContinueOnError)
	root := fs.String("root", ".", "supervisor root")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return app.New(filepath.Clean(*root))
}
