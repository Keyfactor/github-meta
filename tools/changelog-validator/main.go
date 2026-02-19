package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
	"github.com/Masterminds/semver/v3"
)

type Changelog struct {
	Changelog map[string][]ChangeEntry `yaml:"changelog"`
}

type ChangeEntry struct {
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Scope       string `yaml:"scope,omitempty"`
	Issue       int    `yaml:"issue,omitempty"`
	PR          int    `yaml:"pr,omitempty"`
}

var validChangeTypes = map[string]bool{
	"summary":       true,
	"feature":       true,
	"enhancement":   true,
	"bugfix":        true,
	"fix":           true,
	"breaking":      true,
	"security":      true,
	"performance":   true,
	"documentation": true,
	"internal":      true,
	"deprecated":    true,
	"removed":       true,
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: changelog-validator <changelog.yml>")
	}

	filename := os.Args[1]
	
	// Read the changelog file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading file %s: %v", filename, err)
	}

	// Parse the YAML
	var changelog Changelog
	if err := yaml.Unmarshal(data, &changelog); err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}

	// Validate the changelog
	errors := validateChangelog(&changelog)
	
	if len(errors) > 0 {
		fmt.Printf("❌ Changelog validation failed with %d errors:\n\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		os.Exit(1)
	}

	fmt.Println("✅ Changelog validation passed!")
}

func validateChangelog(changelog *Changelog) []string {
	var errors []string

	if len(changelog.Changelog) == 0 {
		errors = append(errors, "Changelog is empty")
		return errors
	}

	// Version format regex
	versionRegex := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?$`)

	for version, changes := range changelog.Changelog {
		// Validate version format
		if !versionRegex.MatchString(version) {
			errors = append(errors, fmt.Sprintf("Invalid version format: %s (should be semantic version like v1.2.3 or 1.2.3)", version))
			continue
		}

		// Try to parse as semantic version (remove 'v' prefix if present)
		cleanVersion := version
		if cleanVersion[0] == 'v' {
			cleanVersion = cleanVersion[1:]
		}
		
		if _, err := semver.NewVersion(cleanVersion); err != nil {
			errors = append(errors, fmt.Sprintf("Invalid semantic version: %s (%v)", version, err))
		}

		// Validate changes
		if len(changes) == 0 {
			errors = append(errors, fmt.Sprintf("Version %s has no changelog entries", version))
			continue
		}

		for i, change := range changes {
			// Validate change type
			if !validChangeTypes[change.Type] {
				errors = append(errors, fmt.Sprintf("Version %s, entry %d: invalid change type '%s'", version, i+1, change.Type))
			}

			// Validate description
			if change.Description == "" {
				errors = append(errors, fmt.Sprintf("Version %s, entry %d: description is required", version, i+1))
			}

			// Validate description length
			if len(change.Description) > 500 {
				errors = append(errors, fmt.Sprintf("Version %s, entry %d: description too long (%d chars, max 500)", version, i+1, len(change.Description)))
			}
		}

		// Check for required summary in releases
		hasSummary := false
		for _, change := range changes {
			if change.Type == "summary" {
				hasSummary = true
				break
			}
		}

		if !hasSummary {
			errors = append(errors, fmt.Sprintf("Version %s: missing summary entry (recommended for releases)", version))
		}
	}

	return errors
}