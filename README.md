# ddev-explain

A CLI tool that summarizes DDEV project configurations with focus on development directories.

## Installation

```bash
go install github.com/dkd-dobberkau/ddev-explain@latest
```

## Usage

```bash
# In a DDEV project directory
ddev-explain

# Show all projects (note: not fully implemented yet)
ddev-explain --all

# Different output formats
ddev-explain --format=json
ddev-explain --format=markdown

# Show only development paths
ddev-explain --dev-paths

# Verbose output (includes hooks, commands)
ddev-explain -v
```

## Install as DDEV Command

```bash
ddev-explain --install-command
ddev explain
```

## Features

- Parses DDEV configuration
- Detects development directories:
  - Composer path repositories
  - Symlinks in vendor/
  - Conventional directories (packages/, local/)
  - Docker mounts
- Lists additional services
- Shows custom commands and hooks
- Multiple output formats (text, JSON, Markdown)

## Development

### Build

```bash
go build -o ddev-explain
```

### Test

```bash
go test ./... -v
```

## License

MIT
