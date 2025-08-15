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

package pathutil

import (
	"runtime"
	"sort"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		isVirtual bool
		expected  string
	}{
		{
			name:      "virtual_path_with_backslashes",
			path:      "app\\src\\main.go",
			isVirtual: true,
			expected:  "app/src/main.go",
		},
		{
			name:      "virtual_path_already_normalized",
			path:      "app/src/main.go",
			isVirtual: true,
			expected:  "app/src/main.go",
		},
		{
			name:      "real_path_unix",
			path:      "app/src/main.go",
			isVirtual: false,
			expected:  "app/src/main.go",
		},
		{
			name:      "empty_path",
			path:      "",
			isVirtual: true,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePath(tt.path, tt.isVirtual)
			if got != tt.expected {
				t.Errorf("NormalizePath(%q, %v) = %q, want %q", tt.path, tt.isVirtual, got, tt.expected)
			}
		})
	}
}

func TestToVirtualPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "windows_path",
			path:     "C:\\Users\\test\\file.txt",
			expected: "C:/Users/test/file.txt",
		},
		{
			name:     "unix_path",
			path:     "/home/test/file.txt",
			expected: "/home/test/file.txt",
		},
		{
			name:     "mixed_separators",
			path:     "app\\src/main.go",
			expected: "app/src/main.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToVirtualPath(tt.path)
			if got != tt.expected {
				t.Errorf("ToVirtualPath(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestJoinVirtual(t *testing.T) {
	tests := []struct {
		name     string
		elements []string
		expected string
	}{
		{
			name:     "simple_join",
			elements: []string{"app", "src", "main.go"},
			expected: "app/src/main.go",
		},
		{
			name:     "with_backslashes",
			elements: []string{"app\\src", "test", "file.txt"},
			expected: "app/src/test/file.txt",
		},
		{
			name:     "empty_elements",
			elements: []string{},
			expected: "",
		},
		{
			name:     "single_element",
			elements: []string{"file.txt"},
			expected: "file.txt",
		},
		{
			name:     "with_double_slashes",
			elements: []string{"app//src", "test"},
			expected: "app/src/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JoinVirtual(tt.elements...)
			if got != tt.expected {
				t.Errorf("JoinVirtual(%v) = %q, want %q", tt.elements, got, tt.expected)
			}
		})
	}
}

func TestStripDriveLetter(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "windows_drive_with_backslash",
			path:     "C:\\Users\\test",
			expected: "Users\\test",
		},
		{
			name:     "windows_drive_with_slash",
			path:     "C:/Users/test",
			expected: "Users/test",
		},
		{
			name:     "unix_absolute_path",
			path:     "/home/test",
			expected: "/home/test", // Should be unchanged on non-Windows
		},
		{
			name:     "relative_path",
			path:     "app/src/main.go",
			expected: "app/src/main.go",
		},
		{
			name:     "drive_only",
			path:     "C:",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripDriveLetter(tt.path)
			
			// On non-Windows systems, paths should be unchanged unless they have drive letters
			if runtime.GOOS != "windows" && !containsDriveLetter(tt.path) {
				if got != tt.path {
					t.Errorf("StripDriveLetter(%q) = %q, want %q (unchanged on non-Windows)", tt.path, got, tt.path)
				}
				return
			}
			
			if got != tt.expected {
				t.Errorf("StripDriveLetter(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectedDir  string
		expectedFile string
	}{
		{
			name:         "simple_path",
			path:         "app/src/main.go",
			expectedDir:  "app/src",
			expectedFile: "main.go",
		},
		{
			name:         "windows_path",
			path:         "app\\src\\main.go",
			expectedDir:  "app/src",
			expectedFile: "main.go",
		},
		{
			name:         "file_only",
			path:         "main.go",
			expectedDir:  "",
			expectedFile: "main.go",
		},
		{
			name:         "root_file",
			path:         "/main.go",
			expectedDir:  "",
			expectedFile: "main.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDir, gotFile := SplitPath(tt.path)
			if gotDir != tt.expectedDir || gotFile != tt.expectedFile {
				t.Errorf("SplitPath(%q) = (%q, %q), want (%q, %q)", 
					tt.path, gotDir, gotFile, tt.expectedDir, tt.expectedFile)
			}
		})
	}
}

func TestContainsPath(t *testing.T) {
	tests := []struct {
		name     string
		parent   string
		child    string
		expected bool
	}{
		{
			name:     "child_in_parent",
			parent:   "/app",
			child:    "/app/src/main.go",
			expected: true,
		},
		{
			name:     "child_equals_parent",
			parent:   "/app",
			child:    "/app",
			expected: true,
		},
		{
			name:     "child_outside_parent",
			parent:   "/app",
			child:    "/other/file.txt",
			expected: false,
		},
		{
			name:     "path_traversal_attempt",
			parent:   "/app",
			child:    "/app/../etc/passwd",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsPath(tt.parent, tt.child)
			if got != tt.expected {
				t.Errorf("ContainsPath(%q, %q) = %v, want %v", tt.parent, tt.child, got, tt.expected)
			}
		})
	}
}

func TestValidatePathSafety(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "safe_relative_path",
			path:     "app/src/main.go",
			expected: true,
		},
		{
			name:     "path_traversal",
			path:     "../../../etc/passwd",
			expected: false,
		},
		{
			name:     "absolute_path",
			path:     "/etc/passwd",
			expected: false,
		},
		{
			name:     "current_directory",
			path:     ".",
			expected: true,
		},
		{
			name:     "hidden_traversal",
			path:     "app/../../../etc/passwd",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePathSafety(tt.path)
			if got != tt.expected {
				t.Errorf("ValidatePathSafety(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestEnsureTrailingSlash(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		isVirtual bool
		expected  string
	}{
		{
			name:      "virtual_without_slash",
			path:      "app/src",
			isVirtual: true,
			expected:  "app/src/",
		},
		{
			name:      "virtual_with_slash",
			path:      "app/src/",
			isVirtual: true,
			expected:  "app/src/",
		},
		{
			name:      "empty_path",
			path:      "",
			isVirtual: true,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnsureTrailingSlash(tt.path, tt.isVirtual)
			if got != tt.expected {
				t.Errorf("EnsureTrailingSlash(%q, %v) = %q, want %q", tt.path, tt.isVirtual, got, tt.expected)
			}
		})
	}
}

func TestRemoveTrailingSlash(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "with_trailing_slash",
			path:     "app/src/",
			expected: "app/src",
		},
		{
			name:     "with_trailing_backslash",
			path:     "app\\src\\",
			expected: "app\\src",
		},
		{
			name:     "without_trailing_slash",
			path:     "app/src",
			expected: "app/src",
		},
		{
			name:     "root_slash",
			path:     "/",
			expected: "/",
		},
		{
			name:     "empty_path",
			path:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveTrailingSlash(tt.path)
			if got != tt.expected {
				t.Errorf("RemoveTrailingSlash(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

// Helper function to check if a path contains a Windows drive letter
func containsDriveLetter(path string) bool {
	return len(path) >= 2 && path[1] == ':'
}f
unc TestMapContainerPath(t *testing.T) {
	tests := []struct {
		name     string
		hostPath string
		expected string
	}{
		{
			name:     "Windows C drive",
			hostPath: "C:\\Users\\test",
			expected: "/c/Users/test",
		},
		{
			name:     "Windows D drive",
			hostPath: "D:/data/files",
			expected: "/d/data/files",
		},
		{
			name:     "Unix path unchanged",
			hostPath: "/home/user/data",
			expected: "/home/user/data",
		},
		{
			name:     "empty path",
			hostPath: "",
			expected: "",
		},
		{
			name:     "relative path",
			hostPath: "relative/path",
			expected: "relative/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapContainerPath(tt.hostPath)
			if result != tt.expected {
				t.Errorf("MapContainerPath(%q) = %q, want %q", tt.hostPath, result, tt.expected)
			}
		})
	}
}

func TestNormalizeRegistryPath(t *testing.T) {
	tests := []struct {
		name     string
		regPath  string
		expected string
	}{
		{
			name:     "HKLM abbreviation",
			regPath:  "HKLM\\Software\\Test",
			expected: "HKEY_LOCAL_MACHINE\\Software\\Test",
		},
		{
			name:     "HKCU abbreviation",
			regPath:  "HKCU\\Software\\Test",
			expected: "HKEY_CURRENT_USER\\Software\\Test",
		},
		{
			name:     "forward slashes converted",
			regPath:  "HKLM/Software/Test",
			expected: "HKEY_LOCAL_MACHINE\\Software\\Test",
		},
		{
			name:     "already full path",
			regPath:  "HKEY_LOCAL_MACHINE\\Software\\Test",
			expected: "HKEY_LOCAL_MACHINE\\Software\\Test",
		},
		{
			name:     "empty path",
			regPath:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeRegistryPath(tt.regPath)
			if result != tt.expected {
				t.Errorf("NormalizeRegistryPath(%q) = %q, want %q", tt.regPath, result, tt.expected)
			}
		})
	}
}

func TestResolveWindowsServicePath(t *testing.T) {
	tests := []struct {
		name        string
		servicePath string
		expected    string
	}{
		{
			name:        "quoted path with arguments",
			servicePath: "\"C:\\Program Files\\Service\\service.exe\" -arg1 -arg2",
			expected:    "C:\\Program Files\\Service\\service.exe",
		},
		{
			name:        "unquoted path with arguments",
			servicePath: "C:\\Windows\\System32\\svchost.exe -k NetworkService",
			expected:    "C:\\Windows\\System32\\svchost.exe",
		},
		{
			name:        "path without arguments",
			servicePath: "C:\\Windows\\System32\\services.exe",
			expected:    "C:\\Windows\\System32\\services.exe",
		},
		{
			name:        "empty path",
			servicePath: "",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveWindowsServicePath(tt.servicePath)
			if result != tt.expected {
				t.Errorf("ResolveWindowsServicePath(%q) = %q, want %q", tt.servicePath, result, tt.expected)
			}
		})
	}
}

func TestExpandWindowsPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "SystemRoot expansion",
			path:     "%SystemRoot%\\System32\\drivers",
			expected: "C:\\Windows\\System32\\drivers",
		},
		{
			name:     "ProgramFiles expansion",
			path:     "%ProgramFiles%\\Common Files",
			expected: "C:\\Program Files\\Common Files",
		},
		{
			name:     "multiple expansions",
			path:     "%SystemRoot%\\%TEMP%\\test",
			expected: "C:\\Windows\\C:\\Windows\\Temp\\test",
		},
		{
			name:     "no expansion needed",
			path:     "C:\\Direct\\Path",
			expected: "C:\\Direct\\Path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandWindowsPath(tt.path)
			if result != tt.expected {
				t.Errorf("ExpandWindowsPath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsWindowsReservedName(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "CON reserved",
			filename: "CON",
			expected: true,
		},
		{
			name:     "CON with extension",
			filename: "CON.txt",
			expected: true,
		},
		{
			name:     "COM1 reserved",
			filename: "COM1",
			expected: true,
		},
		{
			name:     "LPT1 reserved",
			filename: "LPT1.dat",
			expected: true,
		},
		{
			name:     "normal filename",
			filename: "document.txt",
			expected: false,
		},
		{
			name:     "contains reserved but not exact",
			filename: "CONSOLE.exe",
			expected: false,
		},
		{
			name:     "empty name",
			filename: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWindowsReservedName(tt.filename)
			if result != tt.expected {
				t.Errorf("IsWindowsReservedName(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}f
unc TestMapDockerVolume(t *testing.T) {
	tests := []struct {
		name          string
		hostPath      string
		containerPath string
		expected      string
	}{
		{
			name:          "Windows C drive mapping",
			hostPath:      "C:\\Users\\test",
			containerPath: "/app",
			expected:      "/c/Users/test",
		},
		{
			name:          "empty host path",
			hostPath:      "",
			containerPath: "/app",
			expected:      "/app",
		},
		{
			name:          "Unix path unchanged",
			hostPath:      "/home/user",
			containerPath: "/app",
			expected:      "/home/user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MapDockerVolume(tt.hostPath, tt.containerPath)
			if err != nil {
				t.Errorf("MapDockerVolume(%q, %q) error = %v", tt.hostPath, tt.containerPath, err)
				return
			}
			if result != tt.expected {
				t.Errorf("MapDockerVolume(%q, %q) = %q, want %q", tt.hostPath, tt.containerPath, result, tt.expected)
			}
		})
	}
}

func TestDetectProjectType(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected []string
	}{
		{
			name:     "Node.js project",
			files:    []string{"package.json", "src/index.js"},
			expected: []string{"nodejs"},
		},
		{
			name:     "Multi-language project",
			files:    []string{"package.json", "pom.xml", "go.mod"},
			expected: []string{"nodejs", "maven", "golang"},
		},
		{
			name:     "Gradle Kotlin project",
			files:    []string{"build.gradle.kts", "settings.gradle.kts"},
			expected: []string{"gradle"},
		},
		{
			name:     "Python project",
			files:    []string{"requirements.txt", "setup.py", "src/main.py"},
			expected: []string{"python"},
		},
		{
			name:     "no project files",
			files:    []string{"README.md", "src/code.txt"},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectProjectType(tt.files)
			
			// Sort both slices for comparison
			sort.Strings(result)
			sort.Strings(tt.expected)
			
			if len(result) != len(tt.expected) {
				t.Errorf("DetectProjectType(%v) = %v, want %v", tt.files, result, tt.expected)
				return
			}
			
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("DetectProjectType(%v) = %v, want %v", tt.files, result, tt.expected)
					break
				}
			}
		})
	}
}

func TestIsMonorepo(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected bool
	}{
		{
			name:     "Lerna monorepo",
			files:    []string{"lerna.json", "packages/app1/package.json", "packages/app2/package.json"},
			expected: true,
		},
		{
			name:     "Nx monorepo",
			files:    []string{"nx.json", "workspace.json", "apps/app1/package.json"},
			expected: true,
		},
		{
			name:     "Multiple package.json files",
			files:    []string{"package.json", "frontend/package.json", "backend/package.json"},
			expected: true,
		},
		{
			name:     "Single project",
			files:    []string{"package.json", "src/index.js", "README.md"},
			expected: false,
		},
		{
			name:     "No package files",
			files:    []string{"README.md", "src/code.go"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMonorepo(tt.files)
			if result != tt.expected {
				t.Errorf("IsMonorepo(%v) = %v, want %v", tt.files, result, tt.expected)
			}
		})
	}
}