# spinctl

A terminal UI configuration tool for [Spinnaker](https://spinnaker.io/), replacing [Halyard](https://github.com/spinnaker/halyard).

spinctl provides an interactive TUI that works over SSH for managing Spinnaker configuration and deploying services via Debian packages. It imports existing Halyard configurations and migrates them to a cleaner format.

## Features

- **Terminal UI** -- interactive configuration editor built with [Bubbletea](https://github.com/charmbracelet/bubbletea), works over SSH
- **Halyard import** -- migrate existing `~/.hal/config` to spinctl format with automatic backup
- **Drill-down YAML editor** -- navigate and edit deeply nested Spinnaker configuration
- **Debian deployment** -- install and manage Spinnaker service packages via apt
- **Version management** -- pin Spinnaker release versions or override individual services via BOM
- **Config validation** -- structural, type, and cross-field validation before deployment
- **File locking** -- prevents concurrent access from multiple spinctl instances
- **Signal handling** -- graceful interruption during deploy with resume support

## Installation

### From source

Requires Go 1.26.1 or later.

```bash
git clone https://github.com/ashleykleynhans/spinctl.git
cd spinctl
make build
```

The binary will be at `bin/spinctl`.

### Usage

```bash
# Launch the interactive TUI
spinctl

# Import existing Halyard configuration
spinctl import --hal-dir ~/.hal

# Deploy all enabled services
spinctl deploy

# Deploy specific services only (warns if dependencies are missing)
spinctl deploy --services gate,orca

# Preview what would be deployed without executing
spinctl deploy --dry-run

# Show current configuration as YAML
spinctl show
```

## Configuration

spinctl stores its configuration at `~/.spinctl/config.yaml`. The format is a simplified version of the Halyard config:

```yaml
schema_version: 1
version: "1.35.0"
apt_repository: "https://us-apt.pkg.dev/projects/spinnaker-community"

services:
  gate:
    enabled: true
    host: localhost
    port: 8084
    settings:
      server:
        ssl:
          enabled: false
  clouddriver:
    enabled: true
    host: localhost
    port: 7002
  orca:
    enabled: true
    host: localhost
    port: 8083

providers:
  kubernetes:
    enabled: true
    accounts:
      - name: prod
        context: prod-context

security:
  authn:
    enabled: false
  authz:
    enabled: false

features:
  artifacts: true
  pipelineTemplates: false
```

### Service overrides

Pin individual service versions instead of using the BOM defaults:

```yaml
service_overrides:
  clouddriver: "5.82.1"
  gate: "6.99.0"
```

## Spinnaker Services

spinctl manages these Spinnaker microservices:

| Service | Default Port | Purpose |
|---------|-------------|---------|
| clouddriver | 7002 | Cloud provider integrations |
| orca | 8083 | Pipeline orchestration |
| gate | 8084 | API gateway |
| front50 | 8080 | Metadata persistence |
| echo | 8089 | Event routing / CRON |
| igor | 8088 | CI/SCM integrations |
| fiat | 7003 | Authorization |
| rosco | 8087 | Image bakery |
| kayenta | 8090 | Canary analysis |
| deck | 9000 | Web UI |

### Deployment order

Services are deployed in dependency order:

1. front50
2. fiat (depends on front50)
3. clouddriver
4. orca (depends on clouddriver)
5. echo (depends on orca)
6. igor, rosco, kayenta (independent tier)
7. gate (depends on all backend services)
8. deck (depends on gate)

## TUI Navigation

| Key | Action |
|-----|--------|
| Up/Down or j/k | Navigate items |
| Enter | Drill into / edit field |
| Esc or Backspace | Go back one level |
| a | Add item (in lists) |
| d | Delete item (in lists) |
| s | Save config |
| q | Quit (prompts to save if unsaved) |
| ? | Help overlay |

## Halyard Migration

spinctl can import your existing Halyard configuration:

```bash
spinctl import --hal-dir ~/.hal
```

This will:

1. Read `~/.hal/config`
2. Back up the `.hal` directory to `~/.hal.backup.<timestamp>`
3. Map known fields (providers, security, features, services) to spinctl format
4. Preserve unmapped fields in a `custom` catch-all section
5. Write the new config to `~/.spinctl/config.yaml`

If your Halyard config has multiple deployment configurations, spinctl will import the `default` deployment. Multi-deployment support is planned.

## File Permissions

- `~/.spinctl/config.yaml` is written with mode `0600` (may contain credentials)
- Files under `/opt/spinnaker/config/` are written with mode `0640`, owned by `root:spinnaker`

## Development

### Prerequisites

- Go 1.26.1+
- Make

### Build and test

```bash
make build        # Build binary to bin/spinctl
make test         # Run tests with race detector
make coverage     # Generate coverage report (fails below 90%)
make lint         # Run golangci-lint
make clean        # Remove build artifacts
```

### Project structure

```
spinctl/
├── cmd/spinctl/          # CLI entry point (cobra commands)
├── internal/
│   ├── config/           # Config model, Load/Save, validation, locking
│   ├── halimport/        # Halyard config parser and migration
│   ├── deploy/           # Debian deployer, BOM resolution, deploy runner
│   ├── tui/              # Bubbletea TUI (app, pages, components)
│   └── model/            # Shared types (ServiceName, deployment order)
├── Makefile
├── go.mod
└── LICENSE               # Apache 2.0
```

### Architecture

spinctl is a single Go binary with no daemon process. This is a deliberate departure from Halyard's client/daemon architecture, which was a common source of friction.

Key design decisions:

- **`yaml.Node`** for service settings preserves YAML types, comments, and ordering during round-trips
- **Executor interface** for system commands (apt-get, systemctl) enables testing without running real commands
- **File locking** via `flock` prevents concurrent access
- **Schema versioning** (`schema_version` field) enables future config format migrations

## Roadmap

- [ ] Docker deployment target
- [ ] Kubernetes (Helm) deployment target
- [ ] Remote/multi-server deployment
- [ ] Halyard profile file import (`~/.hal/<deployment>/profiles/`)

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.
