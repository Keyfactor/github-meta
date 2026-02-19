package main

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestValidateIntegrationManifest(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid integration manifest",
			input: `{
				"name": "Test Integration",
				"description": "A test integration",
				"integration_type": "orchestrator",
				"support_level": "kf-supported",
				"status": "production",
				"link_github": true,
				"update_catalog": true
			}`,
			wantError: false,
		},
		{
			name: "missing required fields",
			input: `{
				"name": "Test Integration"
			}`,
			wantError: true,
			errorMsg: "description",
		},
		{
			name: "invalid integration type",
			input: `{
				"name": "Test Integration",
				"description": "A test integration",
				"integration_type": "invalid-type",
				"support_level": "kf-supported",
				"status": "production",
				"link_github": true,
				"update_catalog": true
			}`,
			wantError: true,
			errorMsg: "invalid integration_type",
		},
		{
			name: "invalid support level",
			input: `{
				"name": "Test Integration",
				"description": "A test integration",
				"integration_type": "orchestrator",
				"support_level": "invalid-level",
				"status": "production",
				"link_github": true,
				"update_catalog": true
			}`,
			wantError: true,
			errorMsg: "invalid support_level",
		},
		{
			name: "invalid status",
			input: `{
				"name": "Test Integration",
				"description": "A test integration",
				"integration_type": "orchestrator",
				"support_level": "kf-supported",
				"status": "invalid-status",
				"link_github": true,
				"update_catalog": true
			}`,
			wantError: true,
			errorMsg: "invalid status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tempFile, err := ioutil.TempFile("", "test-manifest-*.json")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			// Write test data
			if _, err := tempFile.WriteString(tt.input); err != nil {
				t.Fatalf("Failed to write test data: %v", err)
			}
			tempFile.Close()

			// Validate
			errors := validateIntegrationManifest(tempFile.Name())

			if tt.wantError {
				if len(errors) == 0 {
					t.Errorf("Expected validation errors but got none")
					return
				}
				
				if tt.errorMsg != "" {
					found := false
					for _, err := range errors {
						if strings.Contains(err, tt.errorMsg) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error message containing '%s', got errors: %v", tt.errorMsg, errors)
					}
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("Expected no validation errors but got: %v", errors)
				}
			}
		})
	}
}

func TestValidateBuildConfig(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid dotnet config",
			input: `build:
  language: dotnet
  solution: "Test.sln"
  target_frameworks:
    - "net6.0"
    - "net8.0"
signing:
  enabled: true
  certificate: "keyfactor-code-signing"`,
			wantError: false,
		},
		{
			name: "valid golang config",
			input: `build:
  language: golang
  go_module: "github.com/test/project"
signing:
  enabled: true`,
			wantError: false,
		},
		{
			name: "valid kubernetes config",
			input: `build:
  language: kubernetes
  dockerfile: "Dockerfile"
signing:
  sign_containers: true`,
			wantError: false,
		},
		{
			name: "valid generic config",
			input: `build:
  language: generic
  build_command: "make build"`,
			wantError: false,
		},
		{
			name: "missing language",
			input: `build:
  solution: "Test.sln"`,
			wantError: true,
			errorMsg: "build.language is required",
		},
		{
			name: "invalid language",
			input: `build:
  language: python`,
			wantError: true,
			errorMsg: "invalid language 'python'",
		},
		{
			name: "invalid dotnet framework",
			input: `build:
  language: dotnet
  target_frameworks:
    - "net5.0"`,
			wantError: true,
			errorMsg: "invalid target framework",
		},
		{
			name: "invalid certificate",
			input: `build:
  language: dotnet
signing:
  certificate: "unknown-cert"`,
			wantError: true,
			errorMsg: "unknown certificate",
		},
		{
			name: "invalid coverage threshold",
			input: `build:
  language: dotnet
testing:
  coverage_threshold: 150`,
			wantError: true,
			errorMsg: "coverage_threshold must be between 0 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tempFile, err := ioutil.TempFile("", "test-build-config-*.yml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			// Write test data
			if _, err := tempFile.WriteString(tt.input); err != nil {
				t.Fatalf("Failed to write test data: %v", err)
			}
			tempFile.Close()

			// Validate
			errors := validateBuildConfig(tempFile.Name())

			if tt.wantError {
				if len(errors) == 0 {
					t.Errorf("Expected validation errors but got none")
					return
				}
				
				if tt.errorMsg != "" {
					found := false
					for _, err := range errors {
						if strings.Contains(err, tt.errorMsg) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error message containing '%s', got errors: %v", tt.errorMsg, errors)
					}
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("Expected no validation errors but got: %v", errors)
				}
			}
		})
	}
}

func TestIsValidDotNetFramework(t *testing.T) {
	tests := []struct {
		framework string
		valid     bool
	}{
		{"net6.0", true},
		{"net8.0", true},
		{"net10.0", true},
		{"netstandard2.0", true},
		{"netstandard2.1", true},
		{"netframework4.8", true},
		{"net5.0", false}, // Out of support
		{"net4.0", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.framework, func(t *testing.T) {
			result := isValidDotNetFramework(tt.framework)
			if result != tt.valid {
				t.Errorf("isValidDotNetFramework(%s) = %v, want %v", tt.framework, result, tt.valid)
			}
		})
	}
}

func TestValidLanguages(t *testing.T) {
	expectedLanguages := []string{"dotnet", "golang", "kubernetes", "generic"}
	
	for _, lang := range expectedLanguages {
		if !validLanguages[lang] {
			t.Errorf("Language '%s' should be valid but is not in validLanguages map", lang)
		}
	}
	
	// Test that we have exactly the expected number of languages
	if len(validLanguages) != len(expectedLanguages) {
		t.Errorf("Expected %d valid languages, but found %d", len(expectedLanguages), len(validLanguages))
	}
}

func TestValidIntegrationTypes(t *testing.T) {
	// Test some key integration types
	keyTypes := []string{
		"orchestrator", "windows-orchestrator", "ca-gateway", 
		"pam", "terraform", "api-client",
	}
	
	for _, integrationType := range keyTypes {
		if !validIntegrationTypes[integrationType] {
			t.Errorf("Integration type '%s' should be valid but is not in validIntegrationTypes map", integrationType)
		}
	}
}