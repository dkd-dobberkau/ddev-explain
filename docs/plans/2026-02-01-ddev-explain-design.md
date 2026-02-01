# Design: ddev-explain

Ein Go CLI Tool das DDEV-Projekte analysiert und übersichtlich zusammenfasst, mit besonderem Fokus auf Entwicklungsverzeichnisse.

## CLI Interface

### Aufruf

```bash
# Im oder unterhalb eines DDEV-Projekts
ddev-explain

# Alle Projekte auf dem System
ddev-explain --all

# Ausgabeformat wählen
ddev-explain --format=json
ddev-explain --format=markdown

# Als DDEV Custom Command (nach Installation)
ddev explain
```

### Flags

| Flag | Kurz | Beschreibung |
|------|------|--------------|
| `--format` | `-f` | Ausgabeformat: `text` (Standard), `json`, `markdown` |
| `--all` | `-a` | Alle bekannten DDEV-Projekte anzeigen |
| `--dev-paths` | | Nur Entwicklungsverzeichnisse anzeigen (Quick-View) |
| `--verbose` | `-v` | Zusätzliche Details (Hooks, Umgebungsvariablen) |
| `--install-command` | | DDEV Custom Command installieren |
| `--help` | `-h` | Hilfe |

## Projektstruktur

```
ddev-explain/
├── cmd/
│   └── root.go              # CLI mit cobra
├── internal/
│   ├── ddev/
│   │   ├── config.go        # DDEV Config Parser
│   │   ├── services.go      # Service-Erkennung
│   │   └── commands.go      # Custom Commands Parser
│   ├── composer/
│   │   └── composer.go      # Composer.json Parser
│   ├── detector/
│   │   ├── detector.go      # Dev-Verzeichnis Erkennung
│   │   ├── pathrepo.go      # Composer Path Repositories
│   │   ├── symlinks.go      # Symlink-Analyse
│   │   ├── conventions.go   # Konventionelle Verzeichnisse
│   │   └── mounts.go        # Docker Mounts
│   ├── finder/
│   │   └── finder.go        # Projekt-Suche (aufwärts, global)
│   └── output/
│       ├── text.go          # Text Formatter (mit Farben)
│       ├── json.go          # JSON Formatter
│       └── markdown.go      # Markdown Formatter
├── main.go
└── go.mod
```

## Analysierte Informationen

### Basis-Informationen

Aus `.ddev/config.yaml`:
- Projektname, Typ (typo3, php, drupal, etc.)
- PHP-Version, Webserver (nginx-fpm/apache-fpm)
- Datenbank (mariadb/mysql/postgres + Version)
- URLs (HTTP/HTTPS, Router-URL)
- Node.js Version (falls konfiguriert)

### Services

Aus `.ddev/config.yaml` und `.ddev/docker-compose.*.yaml`:
- Zusätzliche Services: Solr, Redis, Elasticsearch, Mailhog, etc.
- Ports und Verbindungsdetails
- Service-spezifische Konfiguration

### Entwicklungsverzeichnisse (Hauptfokus)

Beispiel-Ausgabe:
```
📁 Development Paths
├── Composer Path Repositories
│   └── packages/* → ./packages (symlinked)
├── Local Packages Found
│   ├── packages/my-sitepackage/
│   ├── packages/my-extension/
│   └── packages/ext-solr/ (→ ../solr-project)
├── Symlinks in vendor/
│   └── vendor/myvendor/pkg → ../../packages/pkg
└── Additional Mounts
    └── ../shared-lib → /var/www/shared (ro)
```

### Custom Commands & Hooks

- Projekt-Commands aus `.ddev/commands/`
- Hooks (post-start, pre-commit, etc.)
- Provider-Konfiguration (falls vorhanden)

## Erkennung der Entwicklungsverzeichnisse

### 1. Composer Path Repositories

Parst `composer.json` → `repositories[]` und sucht nach `type: "path"` Einträgen:

```json
{
  "repositories": [
    {"type": "path", "url": "./packages/*"},
    {"type": "path", "url": "../shared-extensions/my-ext"}
  ]
}
```

Glob-Patterns werden aufgelöst, relative Pfade zu absoluten konvertiert.

### 2. Symlink-Analyse

- Scannt `vendor/` rekursiv nach Symlinks
- Filtert: Ziel außerhalb `vendor/` = Entwicklungspfad
- Ignoriert: Symlinks innerhalb `vendor/` (normale Composer-Links)

### 3. Konventionelle Verzeichnisse

Prüft bekannte Patterns:
- `packages/`, `local/`, `local-packages/`
- `typo3conf/ext/` (bei TYPO3 < 12)
- Verzeichnisse mit `composer.json` die `"type": "library|typo3-cms-*"` haben

### 4. Docker Mounts

- Parst `.ddev/config.yaml` → `additional_hostnames`, `additional_fqdns`
- Parst `.ddev/docker-compose.*.yaml` → `volumes` mit Host-Pfaden
- Filtert: Nur Pfade außerhalb des Projektverzeichnisses

### Datenstruktur

```go
type DevPath struct {
    Path        string   // Absoluter Pfad
    Type        string   // "composer-path", "symlink", "mount", "convention"
    Source      string   // Wo erkannt (composer.json, docker-compose, etc.)
    MountTarget string   // Falls Mount: Ziel im Container
    Packages    []string // Gefundene Packages in diesem Pfad
}
```

## Installation

### Standalone Binary

```bash
# Via Go
go install github.com/[user]/ddev-explain@latest

# Oder Binary Download (für Releases)
curl -L https://github.com/[user]/ddev-explain/releases/latest/download/ddev-explain-$(uname -s)-$(uname -m) \
  -o /usr/local/bin/ddev-explain
chmod +x /usr/local/bin/ddev-explain
```

### DDEV Custom Command

Das Tool installiert sich selbst als DDEV Command:

```bash
ddev-explain --install-command
# → Erstellt ~/.ddev/commands/host/explain
```

Der Command ist ein Wrapper-Script:

```bash
#!/bin/bash
## Description: Summarize DDEV project configuration
## Usage: explain [flags]
## Example: ddev explain --format=json

ddev-explain "$@"
```

## Projektfindung

### Standard: Aufwärts suchen

Startet im aktuellen Verzeichnis und sucht nach oben bis ein `.ddev/config.yaml` gefunden wird (analog zu git).

### Mit --all: Globale Liste

1. `~/.ddev/global_config.yaml` → `project_list` (falls vorhanden)
2. Fallback: Scannt bekannte Verzeichnisse (`~/Projects`, `~/Sites`, etc.)
3. Optional: Cache-File mit zuletzt gefundenen Projekten

## Abhängigkeiten

**Build-Zeit:**
- Go 1.21+

**Runtime:**
- Keine (single binary)

**Go Libraries:**
- `github.com/spf13/cobra` - CLI Framework
- `gopkg.in/yaml.v3` - YAML Parsing
- `github.com/fatih/color` - Terminal-Farben
