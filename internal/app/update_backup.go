package app

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"sg-supervisor/internal/config"
)

type updateBackupCommand struct {
	Name string
	Dir  string
	Exec string
	Args []string
	Env  map[string]string
}

func (a *App) preUpdateBackup(ctx context.Context) error {
	for _, command := range a.preUpdateBackupCommands() {
		if err := ctx.Err(); err != nil {
			return err
		}
		log.Printf("pre-update backup started: %s", command.Name)
		output, err := a.runUpdateBackupCommand(ctx, command)
		if err != nil {
			log.Printf("pre-update backup failed: %s: %v", command.Name, err)
			return err
		}
		if output != "" {
			log.Printf("pre-update backup output [%s]: %s", command.Name, output)
		}
		log.Printf("pre-update backup finished: %s", command.Name)
	}
	return nil
}

func (a *App) preUpdateBackupCommands() []updateBackupCommand {
	nodePath := nodeExecutablePath(a.root)
	coreRoot := filepath.Join(a.layout.InstallDir, "core")
	adapterRoot := filepath.Join(a.layout.InstallDir, "adapters", "dahua-terminal-adapter")

	commands := make([]updateBackupCommand, 0, 2)
	coreScript := filepath.Join(coreRoot, "packages", "ops", "dist", "cli.js")
	if pathExists(nodePath) && pathExists(coreScript) {
		commands = append(commands, updateBackupCommand{
			Name: "school-gate",
			Dir:  coreRoot,
			Exec: nodePath,
			Args: []string{
				filepath.Join("packages", "ops", "dist", "cli.js"),
				"create",
				"--kind", "pre-update",
				"--root-dir", filepath.Join("..", ".."),
				"--env-path", ".env",
				"--include-logs",
			},
			Env: bootstrapCommandEnv(a.root),
		})
	}

	adapterScript := filepath.Join(adapterRoot, "dist", "src", "ops", "backup", "backup-cli.js")
	if pathExists(nodePath) && pathExists(adapterScript) {
		commands = append(commands, updateBackupCommand{
			Name: "dahua-terminal-adapter",
			Dir:  a.root,
			Exec: nodePath,
			Args: []string{
				adapterScript,
				"create",
				"--mode", "pre-update",
				"--root", ".",
				"--backups-dir", filepath.Join(a.layout.BackupsDir, "dahua-terminal-adapter"),
				"--license-dir", a.layout.LicensesDir,
				"--include-logs", "true",
			},
			Env: bootstrapCommandEnv(a.root),
		})
	}

	return commands
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (a *App) runUpdateBackupCommand(ctx context.Context, command updateBackupCommand) (string, error) {
	if command.Name != "dahua-terminal-adapter" {
		return runBootstrapCommand(ctx, command.Dir, command.Env, command.Exec, command.Args...)
	}
	files := map[string]string{
		filepath.Join(a.root, ".env"):         config.AdapterEnvFile(a.layout),
		filepath.Join(a.root, "package.json"): filepath.Join(a.layout.InstallDir, "adapters", "dahua-terminal-adapter", "package.json"),
	}
	return withTemporaryRootFiles(files, func() (string, error) {
		return runBootstrapCommand(ctx, command.Dir, command.Env, command.Exec, command.Args...)
	})
}

func withTemporaryRootFiles(files map[string]string, fn func() (string, error)) (string, error) {
	backups := make(map[string]string, len(files))
	for targetPath, sourcePath := range files {
		backupPath := targetPath + ".bak"
		if pathExists(targetPath) {
			if err := os.Rename(targetPath, backupPath); err != nil {
				rollbackTemporaryRootFiles(files, backups)
				return "", err
			}
			backups[targetPath] = backupPath
		}
		if err := copyFile(sourcePath, targetPath); err != nil {
			rollbackTemporaryRootFiles(files, backups)
			return "", err
		}
	}
	defer rollbackTemporaryRootFiles(files, backups)
	return fn()
}

func rollbackTemporaryRootFiles(files map[string]string, backups map[string]string) {
	for targetPath := range files {
		_ = os.Remove(targetPath)
	}
	for targetPath, backupPath := range backups {
		_ = os.Rename(backupPath, targetPath)
	}
}
