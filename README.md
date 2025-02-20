# GoMouseKeeper

![screenshot](screenshot.png)

GoMouseKeeper / GoMouseBusyer keeps your mouse busy. It's particularly useful in scenarios where you need to maintain screen activity, such as preventing automatic screen locks or maintaining online status.

## Features

- **Random Mouse Movement**: Simulates random mouse movements after a specified idle time
- **User Intervention Detection**: Automatically pauses when manual mouse movement is detected
- **Flexible Timeout Settings**: Multiple timeout options available (5 seconds, 1 minute, 5 minutes, 10 minutes, 30 minutes, 60 minutes)
- **System Tray Integration**: Easy program control through system tray icon

## Installation

### Option 1: Quick Installation (Recommended)

```bash
# Linux/macOS users
curl -sfL https://raw.githubusercontent.com/Hootrix/go-mouse-keeper/main/install.sh | sh

# Windows users (PowerShell)
irm https://raw.githubusercontent.com/Hootrix/go-mouse-keeper/main/install.ps1 | iex
```

### Option 2: Manual Installation

从 [GitHub Releases](https://github.com/Hootrix/go-mouse-keeper/releases/latest) 页面下载适合您系统的二进制文件。

### Option 3: Using Go (Requires Go installed)

```bash
go install github.com/Hootrix/go-mouse-keeper/cmd/mouse-keeper@latest
```

### Option 4: Build from Source

```bash
git clone https://github.com/Hootrix/go-mouse-keeper.git
cd go-mouse-keeper
go install ./cmd/mouse-keeper
```

## Usage

```bash
$ mouse-keeper
```

### Command Line Options

```bash
Usage:
  mouse-keeper [command]

Available Commands:
  enable      Start MouseKeeper when system starts
  disable     Do not start MouseKeeper when system starts
  help        Help about any command

Flags:
  -h, --help   help for mouse-keeper
```

### Auto-start Configuration

To configure MouseKeeper to start automatically with your system:

```bash
# Enable auto-start
sudo mouse-keeper enable

# Disable auto-start
sudo mouse-keeper disable
```

### System Tray Usage

![cmd-screenshot](cmd-screenshot.png)

1. After running the program, you'll see an icon in your system tray
2. Click the icon to see the following options:
   - Resume/Pause: Start/Stop mouse movement
   - Check Timeout Settings: Set mouse idle time
   - Quit: Exit the program
3. The program starts in paused state by default, click "Resume" to start
4. The program automatically pauses when manual mouse movement is detected

## Status Icon Guide

- `...` (●): Program is running
- `   ` (○): Program is paused

## Contributing

Issues and Pull Requests are welcome!
