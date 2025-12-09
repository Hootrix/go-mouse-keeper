package main

import (
	_ "embed"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/go-vgo/robotgo"
	"github.com/kardianos/service"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
)

//go:embed icon-run.png
var iconRunData []byte

//go:embed icon-pause.png
var iconPauseData []byte

var (
	VERSION      string = "0.2.0"
	URL          string = "https://www.hhtjim.com"
	RuningStatus string = "" //● ...
	PauseStatus  string = "" //○
	Logger       *slog.Logger
	logHandler   *slog.TextHandler
	done         = make(chan struct{}) // 用于通知所有 goroutine 退出
)

// 全局配置和状态
type Config struct {
	idleTimeout time.Duration // 鼠标静止超时时间
	offsetRange int           // 鼠标移动范围
	isPaused    bool          // 是否暂停
	mu          sync.RWMutex
}

func (c *Config) setPaused(paused bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isPaused = paused
}

func (c *Config) getPaused() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isPaused
}

func (c *Config) getIdleTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.idleTimeout
}

func (c *Config) setIdleTimeout(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.idleTimeout = duration
}

var (
	config = &Config{
		idleTimeout: 5 * time.Second, // 休息超时时间。超时后keep mouse moving
		offsetRange: 100,             // 移动范围扩大，更真实
		isPaused:    true,            // 默认启动时状态
	}
	mouseKeeper *MouseKeeper
)

type MouseKeeper struct {
	logger        *slog.Logger
	pauseMenu     *systray.MenuItem
	lastX         int
	lastY         int
	lastMoveTime  time.Time
	timeoutMenus  map[time.Duration]*systray.MenuItem
	logLevelMenus map[slog.Level]*systray.MenuItem
	mu            sync.RWMutex // 保护 lastX, lastY, lastMoveTime
	isMoving      bool         // 系统正在移动鼠标
}

func (mk *MouseKeeper) updateMenuState(isPaused bool) {
	if mk.pauseMenu == nil {
		return
	}
	if isPaused {
		mk.pauseMenu.SetTitle("Resume")
		// systray.SetTitle(PauseStatus)
		systray.SetIcon(iconPauseData)
		mk.logger.Info("System paused. Click 'Resume' to start mouse movement")
	} else {
		mk.pauseMenu.SetTitle("Pause")
		// systray.SetTitle(RuningStatus)
		systray.SetIcon(iconRunData)
		mk.logger.Info("System resumed", "idle_timeout", config.idleTimeout)
	}
}

func onReady() {
	systray.SetIcon(iconRunData)
	systray.SetTitle(PauseStatus)
	systray.SetTooltip("MouseKeeper")

	// Initialize timeout menu items map
	mouseKeeper.timeoutMenus = make(map[time.Duration]*systray.MenuItem)
	mouseKeeper.logLevelMenus = make(map[slog.Level]*systray.MenuItem)

	// Add menu items
	mouseKeeper.pauseMenu = systray.AddMenuItem("Resume", "Resume/Pause mouse movement") // Default show Resume

	// Add timeout settings submenu
	timeoutSubmenu := systray.AddMenuItem("Idle Timeout", "Set idle timeout")

	// Create timeout menu items with initial state
	timeouts := []struct {
		duration time.Duration
		label    string
	}{
		{5 * time.Second, "5 Second"},
		{1 * time.Minute, "1 Minute"},
		{5 * time.Minute, "5 Minutes"},
		{10 * time.Minute, "10 Minutes"},
		{30 * time.Minute, "30 Minutes"},
		{60 * time.Minute, "60 Minutes"},
	}

	for _, t := range timeouts {
		item := timeoutSubmenu.AddSubMenuItem(t.label, "Set timeout to "+t.label)
		mouseKeeper.timeoutMenus[t.duration] = item
		// Set initial check state
		if config.idleTimeout == t.duration {
			item.Check()
		}
	}

	// 子菜单
	systray.AddSeparator()
	logLevelSubmenu := systray.AddMenuItem("Log Level", "Set log level")
	levels := []struct {
		level slog.Level
		name  string
	}{
		// {slog.LevelDebug, "Debug"},
		{slog.LevelInfo, "Info"},
		{slog.LevelWarn, "Warn"},
		{slog.LevelError, "Error"},
	}

	// Create menu items for each log level
	for _, l := range levels {
		menuItem := logLevelSubmenu.AddSubMenuItem(l.name, "Set log level to "+l.name)
		mouseKeeper.logLevelMenus[l.level] = menuItem
		if l.level == slog.LevelInfo {
			menuItem.Check() // Default level is Info
		}

		// Set up click handler
		go func(level slog.Level, item *systray.MenuItem) {
			for {
				select {
				case <-item.ClickedCh:
					// Uncheck all items
					for _, mi := range mouseKeeper.logLevelMenus {
						mi.Uncheck()
					}

					// Check the selected item
					item.Check()

					// Log before changing the level
					Logger.Warn("Changing log level", "to", level)

					// Update log level
					logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
						Level: level,
					})
					Logger = slog.New(logHandler)
					mouseKeeper.logger = Logger
				case <-done:
					return // 退出 goroutine
				}
			}
		}(l.level, menuItem)
	}

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the application")
	mVersion := systray.AddMenuItem("Version "+VERSION, "")
	mVersion.Disable()
	about := systray.AddMenuItem("About", URL)

	go func() {
		for range mouseKeeper.pauseMenu.ClickedCh {
			isPaused := config.getPaused()
			newState := !isPaused // 新状态: isPaused=true -> newState=false(运行)

			if newState {
				// newState=true 表示切换到暂停，不需要特殊处理
				config.setPaused(newState)
				mouseKeeper.updateMenuState(newState)
			} else {
				// newState=false 表示切换到运行（Resume）
				// 先添加延迟，避免检测到点击菜单的鼠标移动
				time.Sleep(time.Second)

				// 重置最后位置和时间
				mouseKeeper.lastX, mouseKeeper.lastY = robotgo.Location()
				mouseKeeper.lastMoveTime = time.Now()

				// 最后再更新状态
				config.setPaused(newState)
				mouseKeeper.updateMenuState(newState)
			}
		}
	}()

	// Handle timeout settings
	for _, t := range timeouts {
		duration := t.duration // Create a new variable to avoid closure problems
		menuItem := mouseKeeper.timeoutMenus[duration]

		go func() {
			for {
				select {
				case <-menuItem.ClickedCh:
					// Uncheck all items first
					for _, item := range mouseKeeper.timeoutMenus {
						item.Uncheck()
					}
					// Check the selected item
					menuItem.Check()
					config.setIdleTimeout(duration)
					Logger.Warn("Idle timeout set", "duration", duration)
				case <-done:
					return // 退出 goroutine
				}
			}
		}()
	}

	go func() {
		for range mQuit.ClickedCh {
			systray.Quit()
			return
		}
	}()

	go func() {
		for range about.ClickedCh {
			open.Run(URL)
		}
	}()
}

func onExit() {
	close(done) // 通知所有 goroutine 退出
	os.Exit(0)
}

// 模拟真实的鼠标移动
func (mk *MouseKeeper) simulateRealisticMouseMovement() {
	// 标记开始移动
	mk.mu.Lock()
	mk.isMoving = true
	startX, startY := mk.lastX, mk.lastY
	mk.mu.Unlock()

	// 确保移动结束后重置标志
	defer func() {
		mk.mu.Lock()
		mk.lastX, mk.lastY = robotgo.Location()
		mk.lastMoveTime = time.Now()
		mk.isMoving = false
		mk.mu.Unlock()
	}()

	// 生成随机目标位置，使用配置的 offsetRange
	offset := config.offsetRange
	targetX := startX + rand.Intn(offset*2) - offset
	targetY := startY + rand.Intn(offset*2) - offset

	// 确保目标位置在屏幕范围内
	width, height := robotgo.GetScreenSize()
	targetX = max(0, min(width-1, targetX))
	targetY = max(0, min(height-1, targetY))

	mk.logger.Info("Starting mouse movement", "target_x", targetX, "target_y", targetY)

	robotgo.MoveSmooth(targetX, targetY, 1.0, 1.0)

	mk.logger.Info("Mouse movement completed", "current_x", targetX, "current_y", targetY)
}

// 检查用户活动，返回 true 表示检测到用户活动
func (mk *MouseKeeper) checkUserActivity() bool {
	mk.mu.RLock()
	// 系统正在移动鼠标时，不检测用户活动
	if mk.isMoving {
		mk.mu.RUnlock()
		return false
	}
	lastX, lastY := mk.lastX, mk.lastY
	mk.mu.RUnlock()

	currentX, currentY := robotgo.Location()
	deltaX := abs(currentX - lastX)
	deltaY := abs(currentY - lastY)

	// 检测到鼠标位置变化
	if deltaX > 5 || deltaY > 5 {
		mk.mu.Lock()
		mk.lastX, mk.lastY = currentX, currentY
		mk.lastMoveTime = time.Now() // 更新时间，重置空闲计时
		mk.mu.Unlock()
		return true
	}
	return false
}

// 辅助函数：计算绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (mk *MouseKeeper) start() {
	// 用户活动检测 goroutine
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// 检查用户活动
				if mk.checkUserActivity() {
					// 用户活动中，切换到暂停状态
					if !config.getPaused() {
						config.setPaused(true)
						mk.updateMenuState(true)
						mk.logger.Info("User activity detected, pausing")
					}
				}
			case <-done:
				return
			}
		}
	}()

	// 空闲检测 + 自动移动 goroutine
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// 检查空闲超时（只在暂停状态下检查）
				if config.getPaused() {
					mk.mu.RLock()
					lastMoveTime := mk.lastMoveTime
					mk.mu.RUnlock()

					idleTime := time.Since(lastMoveTime)
					idleTimeout := config.getIdleTimeout()

					if idleTime >= idleTimeout {
						// 空闲超时，开始自动移动
						config.setPaused(false)
						mk.updateMenuState(false)
						mk.logger.Info("Idle timeout reached, starting simulation", "idle_time", idleTime, "timeout", idleTimeout)
					}
				}

				// 运行状态下，持续移动鼠标
				if !config.getPaused() {
					mk.simulateRealisticMouseMovement()
					// 移动后等待随机 1-5 秒
					time.Sleep(time.Duration(1+rand.Intn(5)) * time.Second)
				}
			case <-done:
				return
			}
		}
	}()
}

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	Logger.Info(time.Now().Format("2006-01-02 03:04:05 PM") + " Service started")
	go p.run()
	return nil
}
func (p *program) run() {
	mouseKeeper = &MouseKeeper{
		logger:        Logger,
		timeoutMenus:  make(map[time.Duration]*systray.MenuItem),
		logLevelMenus: make(map[slog.Level]*systray.MenuItem),
		lastX:         0,
		lastY:         0,
		lastMoveTime:  time.Now(),
	}

	Logger.Info("MouseKeeper started (Paused). Click 'Resume' to start")

	// 初始化鼠标位置
	mouseKeeper.lastX, mouseKeeper.lastY = robotgo.Location()
	mouseKeeper.lastMoveTime = time.Now()
	mouseKeeper.start()
}
func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

func main() {
	// Initialize logger with default level (Info)
	logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Logger = slog.New(logHandler)

	svcConfig := &service.Config{
		Name:        "com.hhtjim.mousekeeper", // 使用反域名格式
		DisplayName: "Mouse Keeper Service",
		Description: "Keeps your mouse moving to prevent system sleep or away status",
		Option: map[string]interface{}{
			"RunAtLoad":   true,  // 用户登录时立即启动
			"KeepAlive":   false, // false 停止后禁止运行
			"UserService": true,  // 安装为用户级服务 ~/Library/LaunchAgents/
			// "SessionCreate": true,
		},
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		panic(err)
	}

	// Setup command line tool
	var rootCmd = &cobra.Command{
		Use:   "mouse-keeper",
		Short: "MouseKeeper - Keep your mouse moving",
		Long: `MouseKeeper is a system tray app.

It moves your mouse sometimes to:
- Prevent screen saver
- Keep your status as "online"

You can control it from system tray icon.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Setup signal handling for Ctrl+C
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			go func() {
				<-sigChan
				Logger.Info("Received interrupt signal, shutting down...")
				systray.Quit()
			}()

			// Run service in background
			go func() {
				err = s.Run()
				if err != nil {
					Logger.Error("Service error:", "error", err)
					os.Exit(1)
				}
			}()

			// Run system tray in main thread
			systray.Run(onReady, onExit)
		},
	}

	// Add 'enable' command
	// launchctl load ~/Library/LaunchAgents/com.hhtjim.mousekeeper.plist
	// test run： launchctl start com.hhtjim.mousekeeper
	// test stop： launchctl stop com.hhtjim.mousekeeper
	var enableCmd = &cobra.Command{
		Use:   "enable",
		Short: "Start MouseKeeper when system starts(Root Permissions required)",
		Long:  "Register MouseKeeper as a system service. It will start when system boots.",
		Run: func(cmd *cobra.Command, args []string) {
			err = s.Install()
			if err != nil {
				panic(err)
			} else {
				Logger.Info("Auto-start enabled. MouseKeeper will start with system.")
			}
			os.Exit(0)
		},
	}

	// Add 'disable' command
	// launchctl unload ~/Library/LaunchAgents/com.hhtjim.mousekeeper.plist
	var disableCmd = &cobra.Command{
		Use:   "disable",
		Short: "Do not start MouseKeeper when system starts(Root Permissions required)",
		Long:  "Remove MouseKeeper from system services. It will not start when system boots.",
		Run: func(cmd *cobra.Command, args []string) {
			err = s.Uninstall()
			if err != nil {
				panic(err)
			} else {
				Logger.Info("Auto-start disabled. MouseKeeper will not start with system.")
			}
			os.Exit(0)
		},
	}

	// Disable completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Add commands to root
	rootCmd.AddCommand(enableCmd)
	rootCmd.AddCommand(disableCmd)

	// Run command line tool
	rootCmd.Execute()
}
