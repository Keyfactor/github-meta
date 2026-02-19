package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type BuildConfig struct {
	Build   BuildSettings   `yaml:"build"`
	Signing SigningSettings `yaml:"signing,omitempty"`
	Release ReleaseSettings `yaml:"release,omitempty"`
	Testing TestingSettings `yaml:"testing,omitempty"`
}

type BuildSettings struct {
	Language         string   `yaml:"language"`
	Solution         string   `yaml:"solution,omitempty"`
	TargetFrameworks []string `yaml:"target_frameworks,omitempty"`
	Dockerfile       string   `yaml:"dockerfile,omitempty"`
	GoModule         string   `yaml:"go_module,omitempty"`
	BuildCommand     string   `yaml:"build_command,omitempty"`
	TestProjects     []string `yaml:"test_projects,omitempty"`
}

type SigningSettings struct {
	Enabled        bool `yaml:"enabled,omitempty"`
	Certificate    string `yaml:"certificate,omitempty"`
	SignAssemblies bool `yaml:"sign_assemblies,omitempty"`
	SignPackages   bool `yaml:"sign_packages,omitempty"`
	SignContainers bool `yaml:"sign_containers,omitempty"`
}

type ReleaseSettings struct {
	ChangelogRequired       bool   `yaml:"changelog_required,omitempty"`
	AutoRelease            bool   `yaml:"auto_release,omitempty"`
	NugetPublish           bool   `yaml:"nuget_publish,omitempty"`
	ContainerRegistry      string `yaml:"container_registry,omitempty"`
	ReleaseNotesFromChangelog bool `yaml:"release_notes_from_changelog,omitempty"`
}

type TestingSettings struct {
	UnitTests          bool    `yaml:"unit_tests,omitempty"`
	IntegrationTests   bool    `yaml:"integration_tests,omitempty"`
	TestCommand        string  `yaml:"test_command,omitempty"`
	CodeCoverage       bool    `yaml:"code_coverage,omitempty"`
	CoverageThreshold  float64 `yaml:"coverage_threshold,omitempty"`
}

type IntegrationManifest struct {
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	IntegrationType string      `json:"integration_type"`
	SupportLevel    string      `json:"support_level"`
	Status          string      `json:"status"`
	LinkGithub      bool        `json:"link_github"`
	UpdateCatalog   bool        `json:"update_catalog"`
	About           interface{} `json:"about,omitempty"`
}

var validLanguages = map[string]bool{
	"dotnet":     true,
	"golang":     true,
	"kubernetes": true,
	"generic":    true,
}

var validIntegrationTypes = map[string]bool{
	"orchestrator":            true,
	"windows-orchestrator":    true,
	"orchestrator-registration": true,
	"iot-orchestrator":        true,
	"registration-handler":    true,
	"ca-gateway":              true,
	"anyca-plugin":            true,
	"approval-handler":        true,
	"metadata":                true,
	"alert-handler":           true,
	"api-client":              true,
	"pam":                     true,
	"terraform":               true,
	"ejbca":                   true,
	"signserver":              true,
}

var validSupportLevels = map[string]bool{
	"community":     true,
	"kf-community":  true,
	"kf-supported":  true,
}

var validStatuses = map[string]bool{
	"prototype":  true,
	"pilot":      true,
	"production": true,
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: manifest-validator <directory>")
	}

	directory := os.Args[1]
	
	// Check for required files
	integrationManifestPath := filepath.Join(directory, "integration-manifest.json")
	buildConfigPath := filepath.Join(directory, ".keyfactor-build.yml")
	
	errors := []string{}
	
	// Validate integration manifest (always required)
	if _, err := os.Stat(integrationManifestPath); os.IsNotExist(err) {
		errors = append(errors, "integration-manifest.json is required but not found")
	} else {
		manifestErrors := validateIntegrationManifest(integrationManifestPath)
		errors = append(errors, manifestErrors...)
	}
	
	// Validate build config if present
	var hasValidBuildConfig bool
	if _, err := os.Stat(buildConfigPath); err == nil {
		buildConfigErrors := validateBuildConfig(buildConfigPath)
		if len(buildConfigErrors) == 0 {
			hasValidBuildConfig = true
		}
		errors = append(errors, buildConfigErrors...)
	}
	
	// Determine validation mode
	mode := "none"
	if len(errors) == 0 && hasValidBuildConfig {
		mode = "full"
	} else if len(errors) == 0 || (len(errors) > 0 && fileExists(integrationManifestPath)) {
		mode = "preview"
	}
	
	// Output results
	if len(errors) > 0 {
		fmt.Printf("❌ Validation failed with %d errors:\n\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		fmt.Printf("\nValidation Mode: %s\n", mode)
		os.Exit(1)
	}

	fmt.Printf("✅ Validation passed!\n")
	fmt.Printf("Validation Mode: %s\n", mode)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func validateIntegrationManifest(filename string) []string {
	var errors []string
	
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Error reading integration-manifest.json: %v", err))
		return errors
	}
	
	var manifest IntegrationManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		errors = append(errors, fmt.Sprintf("Error parsing integration-manifest.json: %v", err))
		return errors
	}
	
	// Validate required fields
	if manifest.Name == "" {
		errors = append(errors, "integration-manifest.json: 'name' field is required")
	}
	
	if manifest.Description == "" {
		errors = append(errors, "integration-manifest.json: 'description' field is required")
	}
	
	if !validIntegrationTypes[manifest.IntegrationType] {
		errors = append(errors, fmt.Sprintf("integration-manifest.json: invalid integration_type '%s'", manifest.IntegrationType))
	}
	
	if !validSupportLevels[manifest.SupportLevel] {
		errors = append(errors, fmt.Sprintf("integration-manifest.json: invalid support_level '%s'", manifest.SupportLevel))
	}
	
	if !validStatuses[manifest.Status] {
		errors = append(errors, fmt.Sprintf("integration-manifest.json: invalid status '%s'", manifest.Status))
	}
	
	return errors
}

func validateBuildConfig(filename string) []string {
	var errors []string
	
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Error reading .keyfactor-build.yml: %v", err))
		return errors
	}
	
	var config BuildConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		errors = append(errors, fmt.Sprintf("Error parsing .keyfactor-build.yml: %v", err))
		return errors
	}
	
	// Validate required build settings
	if config.Build.Language == "" {
		errors = append(errors, ".keyfactor-build.yml: build.language is required")
		return errors
	}
	
	if !validLanguages[config.Build.Language] {
		errors = append(errors, fmt.Sprintf(".keyfactor-build.yml: invalid language '%s' (must be: dotnet, golang, kubernetes, generic)", config.Build.Language))
		return errors
	}
	
	// Language-specific validation
	switch config.Build.Language {
	case "dotnet":
		errors = append(errors, validateDotNetConfig(&config.Build)...)
	case "golang":
		errors = append(errors, validateGoConfig(&config.Build)...)
	case "kubernetes":
		errors = append(errors, validateKubernetesConfig(&config.Build)...)
	case "generic":
		errors = append(errors, validateGenericConfig(&config.Build)...)
	}
	
	// Validate signing settings if present
	if config.Signing.Certificate != "" && config.Signing.Certificate != "keyfactor-code-signing" {
		errors = append(errors, fmt.Sprintf(".keyfactor-build.yml: unknown certificate '%s'", config.Signing.Certificate))
	}
	
	// Validate testing settings
	if config.Testing.CoverageThreshold < 0 || config.Testing.CoverageThreshold > 100 {
		errors = append(errors, ".keyfactor-build.yml: testing.coverage_threshold must be between 0 and 100")
	}
	
	return errors
}

func validateDotNetConfig(build *BuildSettings) []string {
	var errors []string
	
	// Validate target frameworks if specified
	if len(build.TargetFrameworks) > 0 {
		for _, framework := range build.TargetFrameworks {
			if !isValidDotNetFramework(framework) {
				errors = append(errors, fmt.Sprintf(".keyfactor-build.yml: invalid target framework '%s'", framework))
			}
		}
	}
	
	return errors
}

func validateGoConfig(build *BuildSettings) []string {
	var errors []string
	
	// Go projects should specify module name
	if build.GoModule == "" {
		errors = append(errors, ".keyfactor-build.yml: build.go_module is recommended for Go projects")
	}
	
	return errors
}

func validateKubernetesConfig(build *BuildSettings) []string {
	var errors []string
	
	// Kubernetes projects should have a Dockerfile
	if build.Dockerfile == "" {
		errors = append(errors, ".keyfactor-build.yml: build.dockerfile is recommended for kubernetes projects")
	}
	
	return errors
}

func validateGenericConfig(build *BuildSettings) []string {
	var errors []string
	
	// Generic projects should specify build command
	if build.BuildCommand == "" {
		errors = append(errors, ".keyfactor-build.yml: build.build_command is recommended for generic projects")
	}
	
	return errors
}

func isValidDotNetFramework(framework string) bool {
	validFrameworks := map[string]bool{
		"net6.0":           true,
		"net7.0":           true,
		"net8.0":           true,
		"net9.0":           true,
		"net10.0":          true,
		"net11.0":          true,
		"net12.0":          true,
		"netstandard2.0":   true,
		"netstandard2.1":   true,
		"netframework4.5":  true,
		"netframework4.6":  true,
		"netframework4.7":  true,
		"netframework4.8":  true,
	}
	
	return validFrameworks[framework]
}