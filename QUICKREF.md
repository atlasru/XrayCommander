# XrayCommander Quick Reference

## Build & Run

```bash
# Quick build
./build.sh

# Manual build
go build -o xraycommander ./cmd/xraycommander

# Run
./xraycommander
```

## Project Structure

```
xraycommander/
├── cmd/xraycommander/     # Entry point
├── internal/
│   ├── config/            # Config management
│   ├── tui/               # UI components
│   ├── xray/              # Xray service
│   └── utils/             # Helpers
├── pkg/models/            # Data models
├── go.mod                 # Dependencies
├── Makefile               # Build tasks
└── build.sh               # Build script
```

## Key Components

| Component | File | Purpose |
|-----------|------|---------|
| Main | `cmd/xraycommander/main.go` | Entry point |
| Config | `internal/config/manager.go` | Profile storage |
| Model | `internal/tui/model.go` | App state |
| Views | `internal/tui/views.go` | UI rendering |
| Keys | `internal/tui/keys.go` | Input handling |
| Service | `internal/xray/service.go` | Xray process |
| Models | `pkg/models/profile.go` | Data structures |

## Dependencies

- `bubbletea` - TUI framework
- `bubbles` - UI components
- `lipgloss` - Styling
- `go-toml` - Config parsing
- `google/uuid` - UUID generation

## VLESS Profile Fields

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| name | Yes | - | Profile name |
| address | Yes | - | Server address |
| port | Yes | - | Server port |
| uuid | Yes | - | VLESS UUID |
| encryption | No | none | Encryption method |
| flow | No | - | Flow control |
| network | No | tcp | Transport protocol |
| security | No | - | Security type |
| sni | No | - | Server name indication |
| fp | No | - | TLS fingerprint |

## Keyboard Shortcuts

### Global
- `Ctrl+Q` - Quit
- `Esc` - Back/Cancel

### Main
- `Ctrl+O` - Connect
- `Ctrl+X` - Disconnect
- `Ctrl+R` - Restart
- `Ctrl+L` - Logs
- `P` - Profiles

### Profiles
- `Enter` - Connect
- `N` - New
- `E` - Edit
- `D` - Delete
- `T` - Test
