package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

func rootFromArgs(args []string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == "--root" && i+1 < len(args) {
			return args[i+1]
		}
		if len(args[i]) > len("--root=") && args[i][:7] == "--root=" {
			return args[i][7:]
		}
	}
	return "."
}

func setupLogging(root string) (func(), error) {
	logDir := filepath.Join(root, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filepath.Join(logDir, "sg-supervisor.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(io.MultiWriter(os.Stderr, file))
	return func() {
		_ = file.Close()
	}, nil
}
