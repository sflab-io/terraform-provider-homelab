# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a custom OpenTofu provider that implements standardized naming conventions for homelab infrastructure. The provider is built using the Terraform Plugin Framework and currently provides a single data source (`homelab_naming`) that generates consistent resource names following the pattern `<app>-<env>`, with special handling for production environments (`prod`/`production`) which omit the environment suffix.

**Key Architecture Points:**
- Provider address: `registry.terraform.io/sflab-io/homelab`
- Built using HashiCorp's terraform-plugin-framework (v1.16.1)
- Currently data-source only (no managed resources)
- No provider-level configuration required
- Local development uses dev_overrides in `.tofurc`

## Development Commands

### Available Mise Tasks

This project uses mise for task automation. List all available tasks:

```bash
mise tasks
# Output:
# provider:build                       Build and verify the tofu homeLab provider
# provider:install                     Install the Terraform HomeLab provider
# provider:examples:naming:plan        Plan tofu homeLab provider naming data sources examples
```

### Building and Installing

```bash
# Format, vet, and tidy dependencies (quick validation)
mise run provider:build

# Build and install the provider using GoReleaser (local snapshot build)
mise run provider:install [version]  # Default version: 0.2.0

# Or manually:
go fmt .
go vet .
go mod tidy
# For GoReleaser build:
goreleaser build --snapshot --single-target --clean
cp dist/terraform-provider-homelab_*/terraform-provider-homelab "$(go env GOBIN)/terraform-provider-homelab"
```

**Important Notes:**
- The install task uses GoReleaser in snapshot mode for consistency with release builds
- Adds ~2-3s compared to `go install`, but ensures identical build flags and provides meaningful version strings (e.g., `0.2.0-next+20250128.abc123`)
- **WARNING**: `mise run provider:install` removes existing installations:
  - Removes `$(go env GOBIN)/terraform-provider-homelab`
  - Removes entire `~/.local/share/opentofu/plugins/` directory
- For fastest iteration during debugging, you can use `go install .` directly (builds with version="dev")

### Testing

```bash
# Test the provider with example configuration
mise run provider:examples:naming:plan

# Or manually:
cd examples/data-sources/naming
TF_CLI_CONFIG_FILE=.tofurc tofu plan
```

**Note:** This provider uses OpenTofu (`tofu`) for testing, not Terraform. The example directory contains a `.tofurc` with dev_overrides pointing to the local Go bin directory. When using dev_overrides, `tofu init` is not needed and may produce errors.

### Tools and Dependencies

This project uses [mise](https://mise.jdx.dev/) for managing tools and task automation:

**Required Tools (managed by mise):**
- Go 1.24.2
- golangci-lint v2.7.2
- goreleaser 2.13.1
- opentofu 1.9.0

**Setup:**
```bash
# Install mise if not already installed
# See: https://mise.jdx.dev/getting-started.html

# Enter project directory (triggers mise hooks)
cd terraform-provider-homelab

# Or manually install tools
mise install
```

The mise configuration (`mise.toml`) automatically:
- Installs required tools on directory entry
- Loads environment variables from `.env` and `.creds.env.yaml`
- Installs pre-commit hooks

## Code Structure

```
├── main.go                              # Provider server entry point
│                                        # - Sets provider address
│                                        # - Configures debug mode support
│                                        # - Calls providerserver.Serve()
│
├── .mise/                               # Mise task automation
│   └── tasks/provider/                  # Provider-specific tasks
│       ├── build                        # Format, vet, tidy (go fmt/vet/mod tidy)
│       ├── install                      # Build with GoReleaser and install to GOBIN AND OpenTofu plugins dir
│       └── examples/naming/plan         # Test with naming example (tofu plan only, no init)
│
├── .goreleaser.yml                      # GoReleaser configuration for builds and releases
│                                        # - Snapshot mode for local dev (--snapshot --single-target)
│                                        # - Generates version strings like: 0.2.0-next+20250128.abc123
│
├── mise.toml                            # Mise configuration
│                                        # - Tool versions (Go, GoReleaser, OpenTofu, golangci-lint)
│                                        # - Environment variables and file loading
│                                        # - Setup hooks (pre-commit installation)
│
├── .pre-commit-config.yaml              # Pre-commit hooks configuration
│
├── .creds.env.yaml                      # Encrypted credentials (gitignored, optional)
│
├── internal/provider/
│   ├── provider.go                      # Provider implementation
│   │                                    # - Defines homelabProvider type
│   │                                    # - Implements Provider interface
│   │                                    # - No provider-level config needed
│   │                                    # - Returns list of data sources
│   │
│   └── naming_data_source.go            # Naming data source
│                                        # - Implements DataSource interface
│                                        # - Schema: env (string), app (string), name (computed string)
│                                        # - Read logic: concatenates app-env with hyphen, except for prod/production
│                                        # - Uses terraform-plugin-framework types
│
└── examples/
    ├── .tofurc.example                  # Example config with both dev_overrides and filesystem_mirror
    └── data-sources/naming/
        ├── main.tf                      # Example usage of homelab_naming
        └── .tofurc                      # Dev overrides config (active, used by plan task)
```

## Key Implementation Patterns

### Data Source Pattern

The naming data source follows the standard terraform-plugin-framework pattern:

1. **Model struct** with `tfsdk` tags mapping to Terraform schema
2. **Metadata()** returns the data source type name (appends to provider name)
3. **Schema()** defines required/computed attributes with descriptions
4. **Read()** implements the data source logic:
   - Read config into model
   - Perform computation/lookup
   - Set computed values
   - Save to state

### Current Naming Logic

**IMPORTANT:** The naming logic in `naming_data_source.go:76-84` implements special handling for production environments:

```go
// If env is "prod" or "production", return only the app name (no suffix)
if data.Env.ValueString() == "prod" || data.Env.ValueString() == "production" {
    data.Name = types.StringValue(data.App.ValueString())
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }
    return
}

// Generate the name by concatenating app and env with a hyphen
name := fmt.Sprintf("%s-%s", data.App.ValueString(), data.Env.ValueString())
data.Name = types.StringValue(name)
```

**Examples:**
- `env="prod"` and `app="web"` produces `"web"`
- `env="production"` and `app="db"` produces `"db"`
- `env="dev"` and `app="web"` produces `"web-dev"`
- `env="staging"` and `app="api"` produces `"api-staging"`

### Adding New Data Sources

To add a new data source:

1. Create a new file in `internal/provider/` (e.g., `foo_data_source.go`)
2. Define the model struct with `tfsdk` tags
3. Implement the `datasource.DataSource` interface:
   - `Metadata()` - set type name
   - `Schema()` - define attributes
   - `Read()` - implement logic
4. Create a `NewFooDataSource()` constructor
5. Add to `provider.go` DataSources() slice
6. Add examples in `examples/data-sources/foo/`

### Testing Workflow

Since there are no automated tests:

1. Build and install: `mise run provider:install`
2. Run plan: `mise run provider:examples:naming:plan`
3. Verify output shows expected generated names
4. Check for errors in provider logs

## Local Development Setup

The provider supports two installation approaches for local development:

### Approach 1: dev_overrides (Recommended for Development)

This is the recommended approach during active development:

1. **Install**: `mise run provider:install` builds with GoReleaser and installs to GOBIN
   - GOBIN location: `~/.local/share/mise/installs/go/1.24.2/bin` (when using mise)
   - Binary name: `terraform-provider-homelab`
   - Version format: `0.2.0-next+20250128.abc123` (snapshot builds include timestamp and git hash)

2. **Configure**: `.tofurc` in example directories points to GOBIN via `dev_overrides`
   ```hcl
   provider_installation {
     dev_overrides {
       "registry.terraform.io/sflab-io/homelab" = "/Users/seba/.local/share/mise/installs/go/1.24.2/bin"
     }
     direct {}
   }
   ```

3. **Use**: `TF_CLI_CONFIG_FILE=.tofurc` tells OpenTofu to use dev overrides
4. **No init needed**: Skip `tofu init` when using dev_overrides (not necessary and may error)

**Advantages:**
- Changes available immediately after rebuild
- Consistent build process with releases (uses GoReleaser)
- Meaningful version strings for debugging
- Fastest iteration cycle

**Note:** Dev overrides warnings are expected and normal during local development.

### Approach 2: filesystem_mirror

The `mise run provider:install` task also symlinks the provider to the OpenTofu plugins directory structure:
- Target: `~/.local/share/opentofu/plugins/registry.terraform.io/sflab-io/homelab/0.2.0/<os_arch>/`
  - Example for macOS ARM64: `~/.local/share/opentofu/plugins/registry.terraform.io/sflab-io/homelab/0.2.0/darwin_arm64/`
- Binary name: `terraform-provider-homelab_v0.2.0`
- Implementation: Creates symlink from GOBIN to plugins directory

This approach mimics a registry installation and requires:
- Proper version directory structure
- Running `tofu init` to discover the provider
- Configuring `filesystem_mirror` in `.tofurc` (see `examples/.tofurc.example`)

**Use this approach when:**
- Testing provider versioning behavior
- Simulating registry-like installation
- Sharing the provider locally without dev_overrides

## Limitations

- No automated acceptance tests (manual testing only via examples)
- No managed resources (data sources only)
- No provider functions
- Simple naming logic with special prod/production handling; no advanced validation or transformations
- Not published to Terraform Registry (local development only)
