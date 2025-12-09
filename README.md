# GoMouseKeeper

![screenshot](screenshot.png)

GoMouseKeeper / GoMouseBusyer keeps your mouse busy. It's particularly useful in scenarios where you need to maintain screen activity, such as preventing automatic screen locks or maintaining online status.

## Features

- **Random Mouse Movement**: Simulates random mouse movements after a specified idle time
- **User Intervention Detection**: Automatically pauses when manual mouse movement is detected
- **Flexible Timeout Settings**: Multiple timeout options available (5 seconds, 1 minute, 5 minutes, 10 minutes, 30 minutes, 60 minutes)
- **System Tray Integration**: Easy program control through system tray icon

## Installation

### Option 1: Download App (Recommended)

Download the latest `.dmg` file from [Releases](https://github.com/Hootrix/go-mouse-keeper/releases/latest):

1. Download `MouseKeeper-x.x.x.dmg`
2. Open the DMG file
3. Drag `MouseKeeper.app` to Applications folder
4. Launch from Applications or Spotlight

> **Note**: On first launch, you may need to right-click and select "Open" to bypass Gatekeeper.

### ⚠️ Accessibility Permission Required

MouseKeeper needs **Accessibility permission** to control the mouse:

1. Open **System Settings** → **Privacy & Security** → **Accessibility**
2. Click the **+** button and add `MouseKeeper.app` (from Applications folder)
3. Make sure the checkbox is enabled
4. Restart MouseKeeper if it was already running

### Option 2: Go Install

If you have Go installed:

```bash
go install github.com/Hootrix/go-mouse-keeper/cmd/mouse-keeper@latest
```

Then run `mouse-keeper` in your terminal.

### Option 3: Build from Source

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


## Contributing

Issues and Pull Requests are welcome!
