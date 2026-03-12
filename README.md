# XrayCommander

A modern Terminal User Interface (TUI) application for managing Xray VLESS proxy connections. Built with Go and Bubble Tea.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-green)

## Features

- **Profile Management**: Create, edit, delete VLESS connection profiles
- **Real-time Monitoring**: View connection status, upload/download speeds, traffic statistics
- **Log Viewer**: Monitor Xray logs in real-time with color-coded output
- **Connection Testing**: Test server latency and availability
- **Security**: Secure storage of configurations with proper file permissions (0600)
- **Cross-platform**: Supports Linux, macOS, and Windows

## Installation

### Prerequisites

- Go 1.21 or higher
- Xray-core binary (will be prompted to install if not found)

### From Source

```bash
# Clone the repository
git clone https://github.com/atlasru/xraycommander.git
cd xraycommander

# Build the application
go build -o xraycommander ./cmd/xraycommander

# Or use the build script
./build.sh
```

### Install to System (Optional)

```bash
# Linux/macOS
sudo mv xraycommander /usr/local/bin/

# Windows (run as Administrator)
move xraycommander.exe C:\Windows\System32\
```

## Usage

```bash
# Run the application
./xraycommander
```

### Keyboard Shortcuts

#### Global
- `Ctrl+Q` / `q` - Quit application
- `?` - Show help

#### Main Screen
- `Ctrl+O` - Start connection (select profile)
- `Ctrl+X` - Stop connection
- `Ctrl+R` - Restart service
- `Ctrl+L` - View logs
- `P` - Manage profiles

#### Profile Management
- `↑/k` - Move up
- `↓/j` - Move down
- `Enter` - Connect to selected profile
- `N` - Create new profile
- `E` - Edit selected profile
- `D` - Delete selected profile
- `T` - Test connection latency
- `Esc` - Back to main

#### Profile Form
- `Tab` - Next field
- `Shift+Tab` - Previous field
- `Ctrl+S` - Save profile
- `Esc` - Cancel

## Configuration

Configuration files are stored in:
- **Linux/macOS**: `~/.config/xraycommander/`
- **Windows**: `%USERPROFILE%\.config\xraycommander\`

### Files
- `config.toml` - Application settings
- `profiles/` - VLESS profile configurations (JSON format)
- `xray.log` - Xray process logs

### Profile Format

```json
{
  "id": "uuid-string",
  "name": "My Server",
  "address": "example.com",
  "port": 443,
  "uuid": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "encryption": "none",
  "flow": "xtls-rprx-vision",
  "network": "tcp",
  "security": "tls",
  "sni": "example.com",
  "fp": "chrome"
}
```

## Architecture

```
xraycommander/
├── cmd/xraycommander/    # Application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── tui/              # Terminal UI components
│   ├── xray/             # Xray process management
│   └── utils/            # Utility functions
├── pkg/
│   └── models/           # Data models
└── go.mod
```

## Troubleshooting

### Missing go.sum

If you get errors about missing `go.sum` entries:

```bash
go mod tidy
```

### Xray Not Found

The application will prompt you to install Xray-core automatically, or you can install it manually from [XTLS/Xray-core](https://github.com/XTLS/Xray-core).

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Xray-core](https://github.com/XTLS/Xray-core) - Proxy core
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling library
