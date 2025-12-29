# OpenTofu Provider: Homelab

A custom OpenTofu provider for standardized resource naming in homelab infrastructure environments.

## Overview

The Homelab provider enables consistent naming conventions across all infrastructure resources (LXC containers, VMs, resource pools, DNS records) through a simple, reusable datasource. This proof-of-concept implementation demonstrates the viability of custom providers for centralized naming logic.

## Features

- **Standardized Naming**: Generate consistent names following the pattern `<app>-<env>` (with special handling for production environments)
- **Type Safety**: OpenTofu validates inputs at plan time
- **Zero Configuration**: No provider-level configuration required
- **Extensible**: Foundation for future datasources, resources, and functions

## Requirements

### For Users
- [OpenTofu](https://opentofu.org/) >= 1.9.0

### For Development
This project uses [mise](https://mise.jdx.dev/) for tool management and task automation.

**Required Tools (automatically managed by mise):**
- [Go](https://golang.org/doc/install) 1.24.2
- [OpenTofu](https://opentofu.org/) 1.9.0
- [GoReleaser](https://goreleaser.com/) 2.13.1
- [golangci-lint](https://golangci-lint.run/) v2.7.2

**Setup:**
```bash
# Install mise (if not already installed)
# See: https://mise.jdx.dev/getting-started.html

# Clone and enter the repository
cd terraform-provider-homelab

# Mise will automatically install all required tools
# (triggered by mise.toml on directory entry)
```

## Installation

### Local Development Setup

1. **Build and install the provider:**

```bash
# Using mise tasks (recommended - builds with GoReleaser)
mise run provider:install

# Or quick install with Go (faster, but version="dev")
go install .
```

**Installation Details:**
- Mise task installs to: `~/.local/share/mise/installs/go/1.24.2/bin/terraform-provider-homelab`
- Also creates symlink in: `~/.local/share/opentofu/plugins/registry.terraform.io/sflab-io/homelab/0.2.0/<os_arch>/`
- Version format (GoReleaser): `0.2.0-next+20250129.abc123`
- Version format (go install): `dev`

2. **Configure dev overrides:**

Create or edit `~/.tofurc` with the following content:

```hcl
  provider_installation {
    dev_overrides {
      "registry.terraform.io/sflab-io/homelab" = "/path/to/your/go/bin"
    }

    direct {}
  }
```

   **To find your Go bin directory:**
```bash
  # If using mise (project setup):
  echo ~/.local/share/mise/installs/go/1.24.2/bin

  # Or check GOBIN:
  go env GOBIN

  # If GOBIN is empty, use GOPATH/bin:
  echo "$(go env GOPATH)/bin"
```

3. **Verify installation:**

```bash
# Using mise task (recommended)
mise run provider:examples:naming:plan

# Or manually
cd examples/data-sources/naming
TF_CLI_CONFIG_FILE=.tofurc \
  tofu plan
```

You should see output showing the generated names.

### Installation Approaches

The provider supports two installation approaches:

#### Approach 1: dev_overrides (Recommended)
Uses the `.tofurc` configuration above to override provider resolution.

**Advantages:**
- Instant updates after rebuild (no `tofu init` needed)
- Simpler workflow for active development
- Expected "dev overrides" warnings are normal

**Workflow:**
```bash
mise run provider:install          # Build and install
cd examples/data-sources/naming
TF_CLI_CONFIG_FILE=.tofurc tofu plan  # Use immediately
```

#### Approach 2: filesystem_mirror
Uses the symlinked installation in `~/.local/share/opentofu/plugins/`.

**Advantages:**
- Tests version-based resolution
- Mimics registry installation behavior
- No dev override warnings

**Setup:**
```hcl
# In ~/.tofurc or project .tofurc
provider_installation {
  filesystem_mirror {
    path    = "/Users/yourusername/.local/share/opentofu/plugins"
    include = ["registry.terraform.io/sflab-io/*"]
  }

  direct {
    exclude = ["registry.terraform.io/sflab-io/*"]
  }
}
```

See `examples/.tofurc.example` for a complete configuration example.

## Usage

### Basic Example

```hcl
terraform {
  required_providers {
    homelab = {
      source  = "registry.terraform.io/sflab-io/homelab"
      version = ">= 0.2.0"
    }
  }
}

provider "homelab" {}

# Generate a name for a development web server
data "homelab_naming" "dev_web" {
  app = "web"
  env = "dev"
}

# Use the generated name
output "resource_name" {
  value = data.homelab_naming.dev_web.name  # Output: "web-dev"
}
```

### Data Source: homelab_naming

Generates standardized names based on environment and application identifiers.

#### Arguments

- `app` (String, Required) - The application name (e.g., `web`, `db`, `api`)
- `env` (String, Required) - The environment name (e.g., `dev`, `staging`, `prod`)

#### Attributes

- `name` (String) - The generated name following the pattern `<app>-<env>`, or just `<app>` for `prod`/`production` environments

#### Examples

```hcl
# Production environment - no suffix
data "homelab_naming" "prod_database" {
  app = "db"
  env = "prod"
}

output "database_name" {
  value = data.homelab_naming.prod_database.name  # Output: "db"
}

# Non-production environment - includes suffix
data "homelab_naming" "dev_web" {
  app = "web"
  env = "dev"
}

output "web_server_name" {
  value = data.homelab_naming.dev_web.name  # Output: "web-dev"
}
```

### Integration with Other Resources

The naming datasource can be used to ensure consistent naming across your infrastructure:

```hcl
# Generate a name for production API (output: "api")
data "homelab_naming" "app_server" {
  app = "api"
  env = "prod"
}

# Use it in a Proxmox LXC container
resource "proxmox_virtual_environment_container" "app" {
  hostname = data.homelab_naming.app_server.name  # "api"
  # ... other configuration
}

# Use it in DNS records
resource "dns_a_record_set" "app" {
  zone      = "example.com."
  name      = data.homelab_naming.app_server.name  # "api"
  addresses = [proxmox_virtual_environment_container.app.ipv4_addresses[0]]
}

# For staging environment (output: "api-staging")
data "homelab_naming" "staging_api" {
  app = "api"
  env = "staging"
}
```

## Development

### Available Mise Tasks

This project uses [mise](https://mise.jdx.dev/) for task automation:

```bash
# List all available tasks
mise tasks

# Available tasks:
# provider:build                 - Format, vet, and tidy code (go fmt/vet/mod tidy)
# provider:install [version]     - Build with GoReleaser and install (default: 0.2.0)
# provider:examples:naming:plan  - Test provider with naming example
```

**Common workflows:**
```bash
# Quick validation before commit
mise run provider:build

# Full build and install for testing
mise run provider:install

# Test with example configuration
mise run provider:examples:naming:plan
```

### Building from Source

```bash
# Install dependencies
go mod tidy

# Quick validation (format, vet, tidy)
mise run provider:build

# Build and install with GoReleaser (recommended)
mise run provider:install

# Or manually with GoReleaser (for all platforms)
goreleaser build --snapshot --clean

# Or quick install with Go (version will be "dev")
go install .
```

**Important Notes:**
- `mise run provider:install` uses GoReleaser in snapshot mode for consistency with releases
- Snapshot builds generate version strings like: `0.2.0-next+20250129.abc123`
- **WARNING**: The install task cleans previous installations:
  - Removes `$(go env GOBIN)/terraform-provider-homelab`
  - Removes entire `~/.local/share/opentofu/plugins/` directory
- For fastest iteration, use `go install .` (builds with version="dev")

### Running Tests

```bash
# Format, vet, and verify code
mise run provider:build

# Test with example configuration
mise run provider:examples:naming:plan

# Or manually
cd examples/data-sources/naming
TF_CLI_CONFIG_FILE=.tofurc tofu plan
```

### Project Structure

```
terraform-provider-homelab/
├── main.go                          # Provider server entry point
├── go.mod                           # Go module definition
├── go.sum                           # Go dependencies checksums
├── CLAUDE.md                        # Claude Code guidance
├── README.md                        # This file
├── .goreleaser.yml                  # GoReleaser build configuration
├── mise.toml                        # Mise tool and task configuration
├── .pre-commit-config.yaml          # Pre-commit hooks
├── .mise/
│   └── tasks/
│       └── provider/                # Mise task definitions
│           ├── build                # Format, vet, tidy
│           ├── install              # Install provider with GoReleaser
│           └── examples/naming/plan # Test with examples
├── internal/
│   └── provider/
│       ├── provider.go              # Provider implementation
│       └── naming_data_source.go    # Naming datasource implementation
└── examples/
    ├── .tofurc.example              # Example OpenTofu configuration
    └── data-sources/naming/
        ├── main.tf                  # Usage example
        └── .tofurc                  # Dev overrides config (active)
```

## Limitations

This is a proof-of-concept implementation with the following limitations:

- **No Registry Publication**: Provider must be installed locally; not available in OpenTofu Registry
- **Simple Logic**: Concatenates `app` and `env` with a hyphen (pattern: `<app>-<env>`), with special handling for `prod`/`production` environments (no suffix); no advanced validation or transformations
- **No Resources**: Only implements a datasource; no managed resources
- **No Functions**: Provider functions not implemented
- **Manual Testing**: No automated test framework

## Future Enhancements

Potential improvements for future iterations:

1. **Enhanced Validation**: Add validation for allowed environment names and character restrictions
2. **Additional Attributes**: Support resource type, location, project identifiers
3. **Resources**: Implement managed resources for naming conventions storage
4. **Provider Functions**: Add utility functions for name manipulation
5. **Registry Publication**: Package and publish to OpenTofu Registry

## Troubleshooting

### Provider Not Found

**Issue**: OpenTofu cannot find the provider.

**Solution**:
1. Verify the provider is installed: `ls -la $(go env GOBIN)/terraform-provider-homelab`
2. Check your `.tofurc` configuration points to the correct directory
3. Ensure the provider address matches: `registry.terraform.io/sflab-io/homelab`

### Dev Override Warnings

**Issue**: Seeing warnings about "Provider development overrides are in effect".

**Solution**: This is expected behavior when using dev overrides. The warnings remind you that you're using a local build instead of a registry version. You can skip `tofu init` when using dev overrides.

### Build Failures

**Issue**: `go install` or `mise run provider:install` fails.

**Solution**:
1. Verify Go version: `go version` (should be 1.24.2)
2. Ensure mise tools are installed: `mise install`
3. Clean and rebuild: `go clean -cache && go mod tidy && mise run provider:install`
4. Check for missing dependencies: `go mod download`
5. Verify GoReleaser is available: `goreleaser --version`

### Mise Tool Issues

**Issue**: Commands fail with "tool not found" errors.

**Solution**:
1. Install mise: See https://mise.jdx.dev/getting-started.html
2. Install project tools: `mise install`
3. Verify installations: `mise list`
4. Check mise configuration: `mise doctor`

## Contributing

This is a proof-of-concept provider developed for the terragrunt-infrastructure-catalog-homelab project. Future enhancements will be considered based on real-world usage patterns and feedback.

## License

This provider is part of the terragrunt-infrastructure-catalog-homelab project and follows the same licensing terms.

## Resources

### Provider Development
- [Terraform Plugin Framework Documentation](https://developer.hashicorp.com/terraform/plugin/framework)
- [HashiCorp Provider Scaffolding Framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework)
- [OpenTofu Documentation](https://opentofu.org/docs/)

### Tools
- [Mise - Dev Tools Manager](https://mise.jdx.dev/)
- [GoReleaser - Release Automation](https://goreleaser.com/)
- [golangci-lint - Go Linter](https://golangci-lint.run/)
