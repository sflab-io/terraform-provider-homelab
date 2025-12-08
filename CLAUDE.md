# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a custom Terraform/OpenTofu provider that implements standardized naming conventions for homelab infrastructure. The provider is built using the Terraform Plugin Framework and currently provides a single data source (`homelab_naming`) that generates consistent resource names following the pattern `<app>-<env>`.

**Key Architecture Points:**
- Provider address: `registry.terraform.io/abes140377/homelab`
- Built using HashiCorp's terraform-plugin-framework (v1.16.1)
- Currently data-source only (no managed resources)
- No provider-level configuration required
- Local development uses dev_overrides in `.terraformrc`

## Development Commands

### Building and Installing

```bash
# Format, vet, and tidy dependencies
mise run provider:build

# Install the provider to Go bin directory
mise run provider:install

# Or manually:
go fmt .
go vet .
go mod tidy
go install .
```

### Testing

```bash
# Test the provider with example configuration
mise run provider:examples:naming:plan

# Or manually:
cd examples/data-sources/naming
TF_CLI_CONFIG_FILE=.tofurc tofu plan
```

**Note:** This provider uses OpenTofu (`tofu`) for testing, not Terraform. The example directory contains a `.tofurc` with dev_overrides pointing to the local Go bin directory. When using dev_overrides, `tofu init` is not needed and may produce errors.

### Go Version

This project requires Go 1.24.2, managed via mise in this project. The mise configuration provides task automation for building, installing, and testing the provider.

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
│       ├── install                      # Install provider to both Go bin AND OpenTofu plugins dir
│       └── examples/naming/plan         # Test with naming example (tofu plan only, no init)
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
│                                        # - Read logic: concatenates app-env with hyphen
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

**IMPORTANT:** The naming pattern in `naming_data_source.go:76` concatenates as `app-env`:

```go
name := fmt.Sprintf("%s-%s", data.App.ValueString(), data.Env.ValueString())
```

So `env="dev"` and `app="web"` produces `"web-dev"`. The documentation has been updated to reflect this pattern correctly.

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

1. **Install**: `mise run provider:install` puts binary in Go bin directory (`~/.local/share/mise/installs/go/1.24.2/bin/terraform-provider-homelab`)
2. **Configure**: `.tofurc` in example directories points to that bin path via `dev_overrides`
3. **Use**: `TF_CLI_CONFIG_FILE=.tofurc` tells OpenTofu to use dev overrides
4. **No init needed**: Skip `tofu init` when using dev_overrides (not necessary and may error)

**Advantages:**
- Changes available immediately after `go install`
- No version management needed
- Fastest iteration cycle

**Note:** Dev overrides warnings are expected and normal during local development.

### Approach 2: filesystem_mirror

The `mise run provider:install` task also copies the provider to the OpenTofu plugins directory structure:
- Target: `~/.local/share/opentofu/plugins/registry.terraform.io/abes140377/homelab/0.1.0/darwin_arm64/`
- Binary name: `terraform-provider-homelab_v0.1.0`

This approach mimics a registry installation and requires:
- Proper version directory structure
- Running `tofu init` to discover the provider
- Configuring `filesystem_mirror` in `.tofurc`

**Use this approach when:**
- Testing provider versioning behavior
- Simulating registry-like installation
- Sharing the provider locally without dev_overrides

## Limitations

- No automated acceptance tests (manual testing only via examples)
- No managed resources (data sources only)
- No provider functions
- Simple string concatenation with no validation
- Not published to Terraform Registry (local development only)
