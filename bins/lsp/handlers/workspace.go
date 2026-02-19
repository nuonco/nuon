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

// SetWorkspaceFolders stores the workspace folders from initialization
func SetWorkspaceFolders(folders []protocol.WorkspaceFolder) {
	workspaceFoldersMu.Lock()
	defer workspaceFoldersMu.Unlock()
	workspaceFolders = folders
}

// GetWorkspaceFolders returns the current workspace folders
func GetWorkspaceFolders() []protocol.WorkspaceFolder {
	workspaceFoldersMu.RLock()
	defer workspaceFoldersMu.RUnlock()
	return workspaceFolders
}

// ScanWorkspaceForDiagnostics scans all workspace folders for TOML files and publishes diagnostics
func ScanWorkspaceForDiagnostics(ctx *glsp.Context) {
	folders := GetWorkspaceFolders()
	if len(folders) == 0 {
		return
	}

	log.Infof("🔍 Starting workspace scan for TOML files...")

	for _, folder := range folders {
		// Convert URI to file path
		folderPath := uriToPath(string(folder.URI))
		if folderPath == "" {
			log.Warningf("⚠️  Could not convert URI to path: %s", folder.URI)
			continue
		}

		log.Infof("📁 Scanning workspace folder: %s", folderPath)

		// Find all .toml files
		tomlFiles, err := findTOMLFiles(folderPath)
		if err != nil {
			log.Errorf("❌ Error scanning folder %s: %v", folderPath, err)
			continue
		}

		log.Infof("✅ Found %d TOML file(s) in %s", len(tomlFiles), folderPath)

		// Publish diagnostics for each file
		for _, filePath := range tomlFiles {
			uri := pathToURI(filePath)

			// Read file content
			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Warningf("⚠️  Could not read file %s: %v", filePath, err)
				continue
			}

			// Publish diagnostics for this file
			PublishDiagnostics(ctx, uri, string(content))
		}
	}

	log.Infof("✅ Workspace scan complete")
}

// findTOMLFiles recursively finds all .toml files in a directory
func findTOMLFiles(rootPath string) ([]string, error) {
	var tomlFiles []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and common ignore patterns
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file has .toml extension
		if strings.HasSuffix(strings.ToLower(info.Name()), ".toml") {
			tomlFiles = append(tomlFiles, path)
		}

		return nil
	})

	return tomlFiles, err
}

// uriToPath converts a file:// URI to a local file path
func uriToPath(uri string) string {
	// Remove file:// prefix
	if strings.HasPrefix(uri, "file://") {
		uri = uri[7:]
	}

	// URL decode (basic implementation for common cases)
	uri = strings.ReplaceAll(uri, "%20", " ")

	return uri
}

// pathToURI converts a local file path to a file:// URI
func pathToURI(path string) protocol.DocumentUri {
	// Basic implementation - replace spaces with %20
	path = strings.ReplaceAll(path, " ", "%20")

	// Ensure file:// prefix
	if !strings.HasPrefix(path, "file://") {
		path = "file://" + path
	}

	return protocol.DocumentUri(path)
}
