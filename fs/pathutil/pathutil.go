// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package pathutil provides cross-platform path utilities for OSV-SCALIBR.
package pathutil

import (
	"path/filepath"
	"runtime"
	"strings"
)

// NormalizePath normalizes a path for cross-platform compatibility.
// It handles Windows drive letters and converts backslashes to forward slashes
// for virtual filesystems while preserving the original path for real filesystems.
func NormalizePath(path string, isVirtual bool) string {
	if path == "" {
		return path
	}
	
	// For virtual filesystems (containers, etc.), always use forward slashes
	if isVirtual {
		return filepath.ToSlash(path)
	}
	
	// For real filesystems, use the OS-appropriate separator
	return filepath.Clean(path)
}

// ToVirtualPath converts a path to virtual filesystem format (forward slashes).
// This is used when storing paths in inventory that should be platform-independent.
func ToVirtualPath(path string) string {
	return filepath.ToSlash(path)
}

// FromVirtualPath converts a virtual path to the current OS format.
// This is used when converting stored paths back to OS-specific format.
func FromVirtualPath(path string) string {
	if runtime.GOOS == "windows" {
		return filepath.FromSlash(path)
	}
	return path
}

// JoinVirtual joins path elements using forward slashes, regardless of OS.
// This ensures consistent path handling in virtual filesystems.
func JoinVirtual(elem ...string) string {
	if len(elem) == 0 {
		return ""
	}
	
	// Convert all elements to use forward slashes
	for i, e := range elem {
		elem[i] = filepath.ToSlash(e)
	}
	
	// Join with forward slashes
	result := strings.Join(elem, "/")
	
	// Clean up any double slashes
	for strings.Contains(result, "//") {
		result = strings.ReplaceAll(result, "//", "/")
	}
	
	return result
}

// IsAbsolute checks if a path is absolute, handling both Unix and Windows formats.
func IsAbsolute(path string) bool {
	return filepath.IsAbs(path)
}

// StripDriveLetter removes the Windows drive letter from a path if present.
// This is useful for creating relative paths in container contexts.
func StripDriveLetter(path string) string {
	if runtime.GOOS != "windows" {
		return path
	}
	
	// Check for Windows drive letter (C:, D:, etc.)
	if len(path) >= 2 && path[1] == ':' {
		// Remove drive letter and colon
		path = path[2:]
		// Remove leading slash if present
		if len(path) > 0 && (path[0] == '\\' || path[0] == '/') {
			path = path[1:]
		}
	}
	
	return path
}

// SplitPath splits a path into directory and filename components,
// handling both Unix and Windows separators.
func SplitPath(path string) (dir, file string) {
	// Normalize separators first
	path = filepath.ToSlash(path)
	
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		return "", path
	}
	
	return path[:lastSlash], path[lastSlash+1:]
}

// RelativeTo returns the relative path from base to target.
// Both paths should be in the same format (virtual or OS-specific).
func RelativeTo(base, target string) (string, error) {
	return filepath.Rel(base, target)
}

// ContainsPath checks if child is contained within parent directory.
// This is useful for security checks to prevent path traversal.
func ContainsPath(parent, child string) bool {
	// Clean both paths
	parent = filepath.Clean(parent)
	child = filepath.Clean(child)
	
	// Get relative path
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	
	// Check if relative path goes up directories
	return !strings.HasPrefix(rel, "..") && rel != ".."
}

// ValidatePathSafety checks if a path is safe to use (no path traversal).
func ValidatePathSafety(path string) bool {
	// Clean the path
	cleaned := filepath.Clean(path)
	
	// Check for path traversal attempts
	if strings.Contains(cleaned, "..") {
		return false
	}
	
	// Check for absolute paths that might escape sandbox
	if filepath.IsAbs(cleaned) {
		return false
	}
	
	return true
}

// EnsureTrailingSlash ensures a directory path ends with a slash.
// This is useful for consistent directory handling.
func EnsureTrailingSlash(path string, isVirtual bool) string {
	if path == "" {
		return path
	}
	
	separator := "/"
	if !isVirtual && runtime.GOOS == "windows" {
		separator = "\\"
	}
	
	if !strings.HasSuffix(path, separator) {
		path += separator
	}
	
	return path
}

// RemoveTrailingSlash removes trailing slashes from a path.
func RemoveTrailingSlash(path string) string {
	if path == "" || path == "/" || path == "\\" {
		return path
	}
	
	return strings.TrimRight(path, "/\\")
}

// MapContainerPath maps a Windows host path to container path format.
// This is essential for Windows Docker containers where paths need conversion.
func MapContainerPath(hostPath string) string {
	if hostPath == "" {
		return hostPath
	}
	
	// Convert to forward slashes first
	path := filepath.ToSlash(hostPath)
	
	// Handle Windows drive letters in containers (C:/path -> /c/path)
	if len(path) >= 2 && path[1] == ':' {
		drive := strings.ToLower(string(path[0]))
		path = "/" + drive + path[2:]
	}
	
	return path
}

// NormalizeRegistryPath handles Windows registry paths consistently.
// This is crucial for Windows-based extractors that scan registry.
func NormalizeRegistryPath(regPath string) string {
	if regPath == "" {
		return regPath
	}
	
	// Ensure consistent registry path format (backslashes)
	regPath = strings.ReplaceAll(regPath, "/", "\\")
	
	// Handle common registry root abbreviations
	replacements := map[string]string{
		"HKLM\\": "HKEY_LOCAL_MACHINE\\",
		"HKCU\\": "HKEY_CURRENT_USER\\",
		"HKCR\\": "HKEY_CLASSES_ROOT\\",
		"HKU\\":  "HKEY_USERS\\",
		"HKCC\\": "HKEY_CURRENT_CONFIG\\",
	}
	
	for abbrev, full := range replacements {
		if strings.HasPrefix(regPath, abbrev) {
			regPath = full + regPath[len(abbrev):]
			break
		}
	}
	
	return regPath
}

// ResolveWindowsServicePath resolves Windows service executable paths.
// Services often have quoted paths with arguments that need parsing.
func ResolveWindowsServicePath(servicePath string) string {
	if servicePath == "" {
		return servicePath
	}
	
	// Handle quoted paths with arguments
	if strings.HasPrefix(servicePath, "\"") {
		endQuote := strings.Index(servicePath[1:], "\"")
		if endQuote != -1 {
			return servicePath[1 : endQuote+1]
		}
	}
	
	// Handle unquoted paths (split on first space)
	parts := strings.Fields(servicePath)
	if len(parts) > 0 {
		return parts[0]
	}
	
	return servicePath
}

// ExpandWindowsPath expands Windows environment variables in paths.
// Common in Windows configurations and registry entries.
func ExpandWindowsPath(path string) string {
	if runtime.GOOS != "windows" {
		return path
	}
	
	// Common Windows environment variable expansions
	expansions := map[string]string{
		"%SystemRoot%":    "C:\\Windows",
		"%ProgramFiles%":  "C:\\Program Files",
		"%ProgramFiles(x86)%": "C:\\Program Files (x86)",
		"%USERPROFILE%":   "C:\\Users\\Default",
		"%APPDATA%":       "C:\\Users\\Default\\AppData\\Roaming",
		"%LOCALAPPDATA%":  "C:\\Users\\Default\\AppData\\Local",
		"%TEMP%":          "C:\\Windows\\Temp",
		"%WINDIR%":        "C:\\Windows",
	}
	
	for envVar, expansion := range expansions {
		path = strings.ReplaceAll(path, envVar, expansion)
	}
	
	return path
}

// IsWindowsReservedName checks if a filename is a Windows reserved name.
// Important for cross-platform compatibility and security.
func IsWindowsReservedName(name string) bool {
	if name == "" {
		return false
	}
	
	// Remove extension for checking
	baseName := strings.ToUpper(name)
	if dotIndex := strings.LastIndex(baseName, "."); dotIndex != -1 {
		baseName = baseName[:dotIndex]
	}
	
	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}
	
	for _, reserved := range reservedNames {
		if baseName == reserved {
			return true
		}
	}
	
	return false
}

// MapDockerVolume handles Docker Desktop Windows path mapping.
// This addresses common issues with Windows Docker volume mounting.
func MapDockerVolume(hostPath, containerPath string) (string, error) {
	if hostPath == "" {
		return containerPath, nil
	}
	
	// Handle Docker Desktop Windows path mapping
	if runtime.GOOS == "windows" {
		// Convert Windows paths to Docker-compatible format
		dockerPath := MapContainerPath(hostPath)
		
		// Handle WSL2 path conversion if needed
		if strings.HasPrefix(dockerPath, "/c/") {
			// Docker Desktop maps C: to /c/ in WSL2
			return dockerPath, nil
		}
	}
	
	return hostPath, nil
}

// ResolveSymlinks safely resolves symlinks with depth limit.
// Prevents infinite loops from circular symlinks.
func ResolveSymlinks(path string, maxDepth int) (string, error) {
	if maxDepth <= 0 {
		return path, nil
	}
	
	resolved := path
	for i := 0; i < maxDepth; i++ {
		info, err := filepath.EvalSymlinks(resolved)
		if err != nil {
			// If we can't resolve, return what we have
			return resolved, nil
		}
		
		if info == resolved {
			// No more symlinks to resolve
			break
		}
		
		resolved = info
	}
	
	return resolved, nil
}

// DetectProjectType analyzes files to determine project type.
// Useful for multi-language project detection.
func DetectProjectType(files []string) []string {
	projectTypes := make(map[string]bool)
	
	for _, file := range files {
		base := filepath.Base(file)
		
		switch base {
		case "package.json":
			projectTypes["nodejs"] = true
		case "pom.xml":
			projectTypes["maven"] = true
		case "build.gradle", "build.gradle.kts":
			projectTypes["gradle"] = true
		case "Cargo.toml":
			projectTypes["rust"] = true
		case "go.mod":
			projectTypes["golang"] = true
		case "requirements.txt", "setup.py", "pyproject.toml":
			projectTypes["python"] = true
		case "composer.json":
			projectTypes["php"] = true
		case "Gemfile":
			projectTypes["ruby"] = true
		}
	}
	
	var types []string
	for projectType := range projectTypes {
		types = append(types, projectType)
	}
	
	return types
}

// IsMonorepo detects if the project structure indicates a monorepo.
// Helps with complex project structure analysis.
func IsMonorepo(files []string) bool {
	indicators := []string{
		"lerna.json",
		"nx.json", 
		"rush.json",
		"pnpm-workspace.yaml",
		"workspace.json",
		".gitmodules",
	}
	
	for _, file := range files {
		base := filepath.Base(file)
		for _, indicator := range indicators {
			if base == indicator {
				return true
			}
		}
	}
	
	// Check for multiple package.json files (common monorepo pattern)
	packageJsonCount := 0
	for _, file := range files {
		if filepath.Base(file) == "package.json" {
			packageJsonCount++
		}
	}
	
	return packageJsonCount > 1
}