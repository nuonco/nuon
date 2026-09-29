package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var (
	workspaceFolders   []protocol.WorkspaceFolder
	workspaceFoldersMu sync.RWMutex
)

func SetWorkspaceFolders(folders []protocol.WorkspaceFolder) {
	workspaceFoldersMu.Lock()
	defer workspaceFoldersMu.Unlock()
	workspaceFolders = folders
}

func GetWorkspaceFolders() []protocol.WorkspaceFolder {
	workspaceFoldersMu.RLock()
	defer workspaceFoldersMu.RUnlock()
	return workspaceFolders
}

func ScanWorkspaceForDiagnostics(ctx *glsp.Context) {
	folders := GetWorkspaceFolders()
	if len(folders) == 0 {
		return
	}

	log.Infof("🔍 Starting workspace scan for TOML files...")

	for _, folder := range folders {
		folderPath := uriToPath(string(folder.URI))
		if folderPath == "" {
			log.Warningf("⚠️  Could not convert URI to path: %s", folder.URI)
			continue
		}

		log.Infof("📁 Scanning workspace folder: %s", folderPath)

		tomlFiles, err := findTOMLFiles(folderPath)
		if err != nil {
			log.Errorf("❌ Error scanning folder %s: %v", folderPath, err)
			continue
		}

		log.Infof("✅ Found %d TOML file(s) in %s", len(tomlFiles), folderPath)

		for _, filePath := range tomlFiles {
			uri := pathToURI(filePath)

			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Warningf("⚠️  Could not read file %s: %v", filePath, err)
				continue
			}

			PublishDiagnostics(ctx, uri, string(content))
		}
	}

	log.Infof("✅ Workspace scan complete")
}

func findTOMLFiles(rootPath string) ([]string, error) {
	var tomlFiles []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(strings.ToLower(info.Name()), ".toml") {
			tomlFiles = append(tomlFiles, path)
		}

		return nil
	})

	return tomlFiles, err
}

func uriToPath(uri string) string {
	if strings.HasPrefix(uri, "file://") {
		uri = uri[7:]
	}

	uri = strings.ReplaceAll(uri, "%20", " ")

	return uri
}

func pathToURI(path string) protocol.DocumentUri {
	path = strings.ReplaceAll(path, " ", "%20")

	if !strings.HasPrefix(path, "file://") {
		path = "file://" + path
	}

	return protocol.DocumentUri(path)
}
