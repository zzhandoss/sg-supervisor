package distribution

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

func renderWiXSource(stageDir string) string {
	stageRoot := stageDir
	if absoluteStageDir, err := filepath.Abs(stageDir); err == nil {
		stageRoot = absoluteStageDir
	}
	tree := buildWiXTree(stageRoot)
	supervisorFileID := wiXID("fil_supervisor/sg-supervisor.exe")
	return strings.Join([]string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<Wix xmlns="http://wixtoolset.org/schemas/v4/wxs">`,
		`  <Package Name="School Gate" Manufacturer="School Gate" Version="1.0.0" UpgradeCode="B4E0F2B1-55B5-4C50-9C3A-0F4B3D48E111" Compressed="yes" InstallerVersion="500" Scope="perMachine">`,
		`    <MajorUpgrade DowngradeErrorMessage="A newer version of School Gate is already installed." />`,
		`    <MediaTemplate EmbedCab="yes" />`,
		`    <StandardDirectory Id="ProgramFiles64Folder">`,
		`      <Directory Id="INSTALLFOLDER" Name="School Gate">`,
		renderWiXTree(tree, stageRoot, "        "),
		`      </Directory>`,
		`    </StandardDirectory>`,
		`    <Feature Id="MainFeature" Title="School Gate" Level="1">`,
		renderWiXComponentRefs(tree),
		`    </Feature>`,
		`    <CustomAction Id="BootstrapInstall" FileRef="` + supervisorFileID + `" ExeCommand='bootstrap-install --root "[INSTALLFOLDER]" --bundle-dir "[SourceDir]payload"' Execute="deferred" Impersonate="no" Return="check" />`,
		`    <InstallExecuteSequence>`,
		`      <Custom Action="BootstrapInstall" Before="InstallFinalize">NOT REMOVE</Custom>`,
		`    </InstallExecuteSequence>`,
		`  </Package>`,
		`</Wix>`,
		"",
	}, "\n")
}

type wiXNode struct {
	Name     string
	Files    []string
	Children map[string]*wiXNode
}

func buildWiXTree(stageDir string) *wiXNode {
	root := &wiXNode{Children: map[string]*wiXNode{}}
	_ = filepath.Walk(stageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || path == stageDir {
			return err
		}
		relativePath, relErr := filepath.Rel(stageDir, path)
		if relErr != nil {
			return relErr
		}
		if shouldSkipWiXPath(relativePath, info.IsDir()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			if _, ok := resolveWiXSourcePath(path); !ok {
				return nil
			}
		}
		parts := strings.Split(filepath.ToSlash(relativePath), "/")
		node := root
		for _, part := range parts[:len(parts)-1] {
			if node.Children[part] == nil {
				node.Children[part] = &wiXNode{Name: part, Children: map[string]*wiXNode{}}
			}
			node = node.Children[part]
		}
		if info.IsDir() {
			if node.Children[parts[len(parts)-1]] == nil {
				node.Children[parts[len(parts)-1]] = &wiXNode{Name: parts[len(parts)-1], Children: map[string]*wiXNode{}}
			}
			return nil
		}
		node.Files = append(node.Files, filepath.ToSlash(relativePath))
		return nil
	})
	return root
}

func renderWiXTree(node *wiXNode, stageDir, indent string) string {
	return renderWiXTreeWithPath(node, stageDir, "", indent)
}

func renderWiXTreeWithPath(node *wiXNode, stageDir, basePath, indent string) string {
	lines := make([]string, 0)
	files := append([]string(nil), node.Files...)
	sort.Strings(files)
	for _, relativePath := range files {
		componentID := wiXID("cmp_" + relativePath)
		fileID := wiXID("fil_" + relativePath)
		sourcePath, ok := resolveWiXSourcePath(filepath.Join(stageDir, filepath.FromSlash(relativePath)))
		if !ok {
			continue
		}
		lines = append(lines,
			indent+`<Component Id="`+componentID+`" Guid="`+wiXGUID(relativePath)+`">`,
			indent+`  <File Id="`+fileID+`" Source="`+sourcePath+`" KeyPath="yes" />`,
			indent+`</Component>`,
		)
	}
	names := make([]string, 0, len(node.Children))
	for name := range node.Children {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		child := node.Children[name]
		childPath := name
		if basePath != "" {
			childPath = basePath + "/" + name
		}
		lines = append(lines, indent+`<Directory Id="`+wiXID("dir_"+childPath)+`" Name="`+name+`">`)
		lines = append(lines, renderWiXTreeWithPath(child, stageDir, childPath, indent+"  "))
		lines = append(lines, indent+`</Directory>`)
	}
	return strings.Join(lines, "\n")
}

func renderWiXComponentRefs(node *wiXNode) string {
	lines := make([]string, 0)
	renderWiXRefs(node, &lines)
	return strings.Join(lines, "\n")
}

func renderWiXRefs(node *wiXNode, lines *[]string) {
	files := append([]string(nil), node.Files...)
	sort.Strings(files)
	for _, filePath := range files {
		*lines = append(*lines, `      <ComponentRef Id="`+wiXID("cmp_"+filepath.ToSlash(filePath))+`" />`)
	}
	names := make([]string, 0, len(node.Children))
	for name := range node.Children {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		renderWiXRefs(node.Children[name], lines)
	}
}

func wiXID(value string) string {
	const maxPrefixLen = 24
	var builder strings.Builder
	builder.Grow(maxPrefixLen)
	for _, runeValue := range value {
		if builder.Len() >= maxPrefixLen {
			break
		}
		switch {
		case unicode.IsLetter(runeValue), unicode.IsDigit(runeValue), runeValue == '_', runeValue == '.':
			builder.WriteRune(unicode.ToLower(runeValue))
		default:
			builder.WriteByte('_')
		}
	}
	sum := sha1.Sum([]byte(value))
	return "id_" + builder.String() + "_" + strings.ToLower(hex.EncodeToString(sum[:6]))
}

func wiXGUID(value string) string {
	sum := sha1.Sum([]byte(value))
	hexValue := strings.ToUpper(hex.EncodeToString(sum[:16]))
	return strings.Join([]string{
		hexValue[0:8],
		hexValue[8:12],
		hexValue[12:16],
		hexValue[16:20],
		hexValue[20:32],
	}, "-")
}

func resolveWiXSourcePath(path string) (string, bool) {
	if resolvedPath, err := filepath.EvalSymlinks(path); err == nil {
		path = resolvedPath
	}
	if _, err := os.Stat(path); err != nil {
		return "", false
	}
	return filepath.ToSlash(path), true
}
