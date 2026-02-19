# CA Gateway Template - Build System Integration Demo

This repository serves as a **demonstration** of the Keyfactor centralized build system integration. It shows how to configure any repository to use the standardized build, test, and release process.

## 🎯 Build System Integration Status

This template repository is **fully integrated** with the centralized build system and demonstrates:

- ✅ **Configuration Validation** - Proper manifest and build config files
- ✅ **Automated Builds** - Multi-framework .NET builds on every PR/push
- ✅ **Release Management** - Automatic versioning, tagging, and GitHub releases
- ✅ **Compliance Checking** - Build system compliance validation and reporting
- ✅ **Preview Mode** - Safe validation builds for testing integration
- ✅ **Artifact Management** - Build artifact collection and signing preparation

## 📋 Integration Files

### Required Configuration Files

#### 1. `integration-manifest.json`
Standard integration catalog metadata - defines the integration type, support level, and compatibility.

#### 2. `.keyfactor-build.yml`
Build system configuration - specifies language, build settings, and release preferences:
```yaml
build:
  language: dotnet  # REQUIRED: Explicit language specification
  solution: "cagateway-template.sln"
  target_frameworks: ["net472"]
  configuration: Release

release:
  changelog_required: true
  auto_release: true
  
signing:
  enabled: true
```

#### 3. `changelog.yml` 
Structured changelog for automated release notes:
```yaml
changelog:
  v1.0.0:
    - type: summary
      description: "Release description"
    - type: feature
      description: "New feature added"
```

### GitHub Actions Workflow

#### `.github/workflows/build.yml`
Demonstrates the complete build pipeline:

1. **Validation** - Checks configuration files and compliance
2. **Build** - Compiles the .NET Framework project with multiple targets
3. **Testing** - Runs tests (if present) and collects coverage
4. **Compliance** - Validates build system usage and generates fix suggestions
5. **Release** - Creates GitHub releases with artifacts and changelogs
6. **Preview** - Demonstrates safe validation for testing integration

## 🔧 How This Demonstrates Build System Usage

### Multi-Job Workflow
- **validate**: Shows configuration validation before building
- **build**: Demonstrates .NET Framework builds on Windows runners
- **compliance-check**: Shows automated compliance reporting
- **release**: Shows automated release creation on main branch
- **preview-build**: Demonstrates preview mode for testing

### Build System Features Showcased

#### Configuration Validation
```yaml
- uses: Keyfactor/github-meta/.github/actions/validate-manifest@main
  with:
    fail-on-missing: true
    require-build-config: true
```

#### Multi-Framework .NET Builds
```yaml
- uses: Keyfactor/github-meta/.github/actions/build-dotnet@main
  with:
    solution-file: "cagateway-template.sln"
    target-frameworks: '["net472"]'
    configuration: "Release"
    enable-signing: true
```

#### Automated Release Management
```yaml
- uses: Keyfactor/github-meta/.github/actions/release-manager@main
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    artifacts-path: "artifacts"
    changelog-required: true
    create-release: true
```

#### Compliance Checking
```yaml
- uses: Keyfactor/github-meta/.github/actions/sync-checker@main
  with:
    enforce-compliance: false
    generate-fixes: true
```

## 🚀 Testing the Build System

### Pull Request Testing
1. Create a branch with changes
2. Open a pull request
3. Observe:
   - Configuration validation
   - Build execution on Windows
   - Compliance checking with generated report
   - Preview build demonstration

### Release Testing  
1. Merge changes to `main` branch
2. Observe:
   - Automatic version detection from changelog
   - Tag creation
   - GitHub release with artifacts
   - Build artifacts uploaded

### Compliance Testing
1. Modify or remove configuration files
2. Create PR to see compliance issues detected
3. Check generated `compliance-fixes.md` for specific guidance

## 📊 Build Outputs

### Artifacts Created
- **Build Artifacts**: Compiled .NET Framework assemblies
- **Compliance Report**: Detailed compliance status and fix suggestions
- **Release Assets**: Signed artifacts attached to GitHub releases

### GitHub Actions Summary
Each build provides detailed summaries showing:
- Configuration validation results
- Build success/failure for each target framework
- Compliance status and recommendations
- Preview mode analysis and guidance

## 🎓 Learning from This Template

### For Repository Owners
This template shows exactly what files and configuration are needed to integrate with the build system.

### For Build System Development
This serves as a test case for validating build system functionality across different project types.

### For Migration Planning
Demonstrates the gradual migration path from existing build processes to centralized automation.

## 🔗 Related Documentation

- [Build System Integration Guide](../docs/build-system-guide.md)
- [Integration Manifest Schema](../integration-manifest-schema.json)
- [Build Configuration Examples](../templates/)

---

This template repository demonstrates **production-ready integration** with the Keyfactor centralized build system, showcasing automated builds, releases, and compliance validation.
