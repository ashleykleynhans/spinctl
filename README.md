# spinctl

A terminal UI configuration tool for [Spinnaker](https://spinnaker.io/), replacing [Halyard](https://github.com/spinnaker/halyard).

spinctl provides an interactive TUI that works over SSH for managing Spinnaker configuration and deploying services via Debian packages. It imports existing Halyard configurations and migrates them to a cleaner format.

## Features

- **Terminal UI** -- polished interactive configuration editor built with [Bubbletea](https://github.com/charmbracelet/bubbletea), works over SSH
- **Halyard import** -- migrate existing `~/.hal/config`, service-settings, and profiles to spinctl format with automatic backup
- **Drill-down YAML editor** -- navigate and edit deeply nested Spinnaker configuration with breadcrumb navigation
- **ON/OFF status** -- visual `[ ON]`/`[OFF]` badges for all toggleable settings, spacebar to toggle from list view
- **TUI Deploy** -- build deploy plan, confirm, and execute directly from the TUI with live status updates
- **Debian deployment** -- install and manage Spinnaker service packages via apt in dependency order
- **Version management** -- pin Spinnaker release versions (e.g. `2025.3.2`) or override individual services via BOM
- **Config validation** -- structural, type, and cross-field validation before deployment
- **Save control** -- explicit save with `s`, dirty tracking with revert detection, quit confirmation for unsaved changes
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

### From release

Download the latest release from the [releases page](https://github.com/ashleykleynhans/spinctl/releases).

```bash
# Debian/Ubuntu (arm64)
curl -LO https://github.com/ashleykleynhans/spinctl/releases/latest/download/spinctl_<version>_arm64.deb
sudo dpkg -i spinctl_<version>_arm64.deb

# RPM (x86_64)
curl -LO https://github.com/ashleykleynhans/spinctl/releases/latest/download/spinctl-<version>-1.x86_64.rpm
sudo rpm -i spinctl-<version>-1.x86_64.rpm
```

## Usage

```bash
# Launch the interactive TUI
spinctl

# Import existing Halyard configuration
spinctl import

# Import from a custom path
spinctl --hal-dir /path/to/.hal import

# Deploy all enabled services (CLI mode)
spinctl deploy

# Deploy specific services only
spinctl deploy --services gate,orca

# Preview what would be deployed without executing
spinctl deploy --dry-run

# Show current configuration as YAML
spinctl show

# Use custom config and lock paths
spinctl --config /path/to/config.yaml --lock /path/to/.lock
```

## Configuration

spinctl stores its configuration at `~/.spinctl/config.yaml`. The format is a clean restructuring of the Halyard config:

```yaml
schema_version: 1
version: "2025.3.2"
apt_repository: "https://us-apt.pkg.dev/projects/spinnaker-community"

services:
  gate:
    enabled: true
    host: 0.0.0.0
    port: 8084
    settings:
      server:
        ssl:
          enabled: true
          keyStore: /path/to/keystore.jks
      spring:
        security:
          oauth2:
            client:
              registration:
                github:
                  client-id: your-client-id
                  client-secret: your-secret
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
  aws:
    enabled: true
    accounts:
      - name: my-aws
        accountId: "123456789012"
        regions:
          - name: us-west-2
          - name: us-east-1

security:
  authn:
    enabled: true
  authz:
    enabled: true

features:
  artifacts: true
  chaos: false

artifacts:
  github:
    enabled: true
  s3:
    enabled: true

persistent_storage:
  persistentStoreType: s3
  s3:
    bucket: my-spinnaker-bucket
    region: us-west-2

notifications:
  slack:
    enabled: true
    botName: spinnaker

ci:
  jenkins:
    enabled: true

canary:
  enabled: true
  serviceIntegrations:
    - ...

metric_stores:
  prometheus:
    enabled: false
  datadog:
    enabled: false

timezone: America/Los_Angeles
```

### Service overrides

Pin individual service versions instead of using the BOM defaults:

```yaml
service_overrides:
  clouddriver: "2025.3.2"
  gate: "2025.3.2"
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
| Enter | Drill into / edit field / toggle boolean |
| Space | Toggle ON/OFF (booleans and maps with enabled key) |
| Esc | Go back one level |
| + | Add new item (maps and lists) |
| d | Delete item (with confirmation) |
| s | Save config to disk |
| q | Quit (prompts to save if unsaved changes) |

### Home screen sections

- **Services** -- enable/disable and configure each Spinnaker service
- **Providers** -- cloud provider accounts (AWS, Kubernetes, GCP, etc.)
- **Security** -- authn/authz settings, OAuth2, SSL configuration
- **Features** -- feature flag toggles
- **Artifacts** -- artifact source configuration
- **Persistent Storage** -- storage backend selection and settings
- **Notifications** -- Slack, email, SMS notification channels
- **CI** -- Jenkins, Travis, CodeBuild, etc.
- **Repository** -- Artifactory and other artifact repositories
- **Pub/Sub** -- Google Pub/Sub and other integrations
- **Canary** -- Kayenta canary analysis configuration
- **Webhook** -- webhook trust and configuration
- **Metric Stores** -- Prometheus, Datadog, Stackdriver, etc.
- **Deployment Environment** -- distributed/local deployment settings
- **Version** -- Spinnaker release version
- **Import from Halyard** -- migrate existing halyard configuration
- **Deploy** -- build plan, confirm, and deploy services

## Halyard Migration

spinctl imports your existing Halyard configuration:

```bash
spinctl import
```

Or specify a custom path:

```bash
spinctl --hal-dir /path/to/.hal import
```

The import process:

1. Reads `~/.hal/config` (main halyard configuration)
2. Reads `~/.hal/<deployment>/service-settings/*.yml` (per-service host, port, enabled overrides)
3. Reads `~/.hal/<deployment>/profiles/*-local.yml` (Spring Boot config overrides -- SSL, CORS, OAuth2, etc.)
4. Maps all known sections (providers, security, features, artifacts, CI, canary, notifications, etc.) to dedicated config fields
5. Writes the config to `~/.spinctl/config.yaml`

All services are enabled by default after import (matching Halyard's behavior). Service-specific overrides from `service-settings/` can disable individual services.

You can also edit the halyard path directly in the TUI import page by pressing `e`.

## Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Path to config file | `~/.spinctl/config.yaml` |
| `--lock` | Path to lock file | `~/.spinctl/.lock` |
| `--hal-dir` | Path to .hal directory | `~/.hal` |

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
│   ├── halimport/        # Halyard config parser, profiles, and migration
│   ├── deploy/           # Debian deployer, BOM resolution, deploy runner
│   ├── tui/              # Bubbletea TUI (app, pages, editor, components)
│   │   └── components/   # Reusable form inputs and status bar
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
- **Snapshot-based dirty tracking** -- compares current config against last saved state, so reverting a change clears the modified flag

## Roadmap

- [ ] Docker deployment target
- [ ] Kubernetes (Helm) deployment target
- [ ] Remote/multi-server deployment
- [ ] Multi-deployment configuration support

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.
