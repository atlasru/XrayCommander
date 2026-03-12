# XrayCommander Architecture

## Overview

XrayCommander is a Terminal User Interface (TUI) application for managing Xray VLESS proxy connections.

## Architecture Components

### 1. Entry Point (`cmd/xraycommander/`)
- **main.go**: Application entry point

### 2. Configuration Management (`internal/config/`)
- **manager.go**: Profile CRUD, file operations, Xray detection

### 3. TUI Components (`internal/tui/`)
- **model.go**: Main application model, state management
- **views.go**: UI rendering functions
- **keys.go**: Keyboard input handlers

### 4. Xray Service (`internal/xray/`)
- **service.go**: Process management, logs, monitoring

### 5. Utilities (`internal/utils/`)
- **helpers.go**: UUID generation, VLESS link parsing

### 6. Data Models (`pkg/models/`)
- **profile.go**: Core data structures

## Data Flow

User Input → TUI → Model Update → View Render
                ↓
        Config Manager
                ↓
        Xray Service
