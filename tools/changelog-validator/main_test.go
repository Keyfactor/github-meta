package main

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestValidateChangelog(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid changelog with all types",
			input: `changelog:
  v1.2.0:
    - type: summary
      description: "This is a test release"
    - type: feature
      description: "Added new feature"
    - type: bugfix
      description: "Fixed a bug"
    - type: breaking
      description: "Breaking change"
      scope: "api"
      issue: 123
  v1.1.0:
    - type: summary
      description: "Previous release"
    - type: enhancement
      description: "Improved performance"`,
			wantError: false,
		},
		{
			name: "empty changelog",
			input: `changelog: {}`,
			wantError: true,
			errorMsg: "Changelog is empty",
		},
		{
			name: "invalid version format",
			input: `changelog:
  invalid-version:
    - type: feature
      description: "Test"`,
			wantError: true,
			errorMsg: "Invalid version format",
		},
		{
			name: "invalid change type",
			input: `changelog:
  v1.0.0:
    - type: invalid-type
      description: "Test"`,
			wantError: true,
			errorMsg: "invalid change type",
		},
		{
			name: "missing description",
			input: `changelog:
  v1.0.0:
    - type: feature`,
			wantError: true,
			errorMsg: "description is required",
		},
		{
			name: "missing summary",
			input: `changelog:
  v1.0.0:
    - type: feature
      description: "Added feature"`,
			wantError: true,
			errorMsg: "missing summary entry",
		},
		{
			name: "description too long",
			input: `changelog:
  v1.0.0:
    - type: summary
      description: "Short summary"
    - type: feature
      description: "` + strings.Repeat("x", 501) + `"`,
			wantError: true,
			errorMsg: "description too long",
		},
		{
			name: "valid semantic versions",
			input: `changelog:
  v1.2.3:
    - type: summary
      description: "Version with v prefix"
    - type: feature
      description: "Test feature"
  2.0.0:
    - type: summary  
      description: "Version without v prefix"
    - type: breaking
      description: "Breaking change"
  v1.0.0-alpha.1:
    - type: summary
      description: "Prerelease version"
    - type: feature
      description: "Alpha feature"`,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tempFile, err := ioutil.TempFile("", "test-changelog-*.yml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			// Write test data
			if _, err := tempFile.WriteString(tt.input); err != nil {
				t.Fatalf("Failed to write test data: %v", err)
			}
			tempFile.Close()

			// Parse and validate
			data, err := ioutil.ReadFile(tempFile.Name())
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			var changelog Changelog
			if err := yaml.Unmarshal(data, &changelog); err != nil {
				if tt.wantError {
					return // Expected YAML parse error
				}
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			errors := validateChangelog(&changelog)

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

func TestValidChangeTypes(t *testing.T) {
	validTypes := []string{
		"summary", "feature", "enhancement", "bugfix", "fix",
		"breaking", "security", "performance", "documentation",
		"internal", "deprecated", "removed",
	}

	for _, changeType := range validTypes {
		if !validChangeTypes[changeType] {
			t.Errorf("Change type '%s' should be valid but is not in validChangeTypes map", changeType)
		}
	}

	// Test invalid types
	invalidTypes := []string{"invalid", "unknown", "test"}
	for _, changeType := range invalidTypes {
		if validChangeTypes[changeType] {
			t.Errorf("Change type '%s' should be invalid but is marked as valid", changeType)
		}
	}
}