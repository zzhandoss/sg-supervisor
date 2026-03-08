package distribution

import (
	"path/filepath"
	"strings"
)

func shouldSkipWiXPath(relativePath string, isDir bool) bool {
	cleanPath := filepath.ToSlash(filepath.Clean(relativePath))
	if cleanPath == "." {
		return false
	}
	segments := strings.Split(cleanPath, "/")
	for _, segment := range segments[:len(segments)-1] {
		if isWiXExcludedDir(segment) {
			return true
		}
	}
	name := segments[len(segments)-1]
	if isDir {
		return isWiXExcludedDir(name)
	}
	return isWiXExcludedFile(name)
}

func isWiXExcludedDir(name string) bool {
	switch strings.ToLower(name) {
	case ".github", ".vscode", ".idea", "__tests__", "test", "tests", "testing", "docs", "doc", "example", "examples", "coverage", "benchmark", "benchmarks":
		return true
	default:
		return false
	}
}

func isWiXExcludedFile(name string) bool {
	lowerName := strings.ToLower(name)
	switch lowerName {
	case ".gitattributes", ".gitignore", ".npmignore", ".editorconfig", ".eslintrc", ".eslintignore", ".prettierignore", ".prettierrc", ".prettierrc.json", ".prettierrc.js", "eslint.config.js":
		return true
	}
	if strings.HasPrefix(lowerName, "tsconfig") && strings.HasSuffix(lowerName, ".json") {
		return true
	}
	if strings.HasSuffix(lowerName, ".d.ts") || strings.HasSuffix(lowerName, ".d.mts") || strings.HasSuffix(lowerName, ".d.cts") {
		return true
	}
	switch filepath.Ext(lowerName) {
	case ".map", ".md", ".mkd", ".ts", ".tsx", ".mts", ".cts":
		return true
	default:
		return false
	}
}
