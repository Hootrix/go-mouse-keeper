package main

import (
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"github.com/go-vgo/robotgo"
	"github.com/skratchdot/open-golang/open"
)

var (
	VERSION      string = "0.1.6"
	URL          string = "https://www.hhtjim.com"
	RuningStatus string = "..." //●
	PauseStatus  string = "   " //○
	Logger       *log.Logger
)

// 全局配置和状态
type Config struct {
	idleTimeout time.Duration // 鼠标静止超时时间
	offsetRange int           // 鼠标移动范围
	isPaused    bool          // 是否暂停
	mu          sync.Mutex
}

func (c *Config) setPaused(paused bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isPaused = paused
}

func (c *Config) getPaused() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isPaused
}

func (c *Config) setIdleTimeout(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.idleTimeout = duration
}

var (
	nRrand *rand.Rand
	config = &Config{
		idleTimeout: 5 * time.Second, // 休息超时时间。超时后keep mouse moving
		offsetRange: 100,             // 移动范围扩大，更真实
		isPaused:    true,            // 默认启动时状态
	}
	mouseKeeper *MouseKeeper
)

type MouseKeeper struct {
	logger       *log.Logger
	pauseMenu    *systray.MenuItem
	lastX        int
	lastY        int
	lastMoveTime time.Time
	timeoutMenus map[time.Duration]*systray.MenuItem
}

func init() {
	seed := time.Now().UnixNano()
	src := rand.NewSource(seed)
	nRrand = rand.New(src)
}

func (mk *MouseKeeper) updateMenuState(isPaused bool) {
	if mk.pauseMenu == nil {
		return
	}
	if isPaused {
		mk.pauseMenu.SetTitle("Resume")
		systray.SetTitle(PauseStatus)
		mk.logger.Printf("System paused. Click 'Resume' to start mouse movement")
	} else {
		mk.pauseMenu.SetTitle("Pause")
		systray.SetTitle(RuningStatus)
		mk.logger.Printf("System resumed. Will start moving mouse after %v of inactivity", config.idleTimeout)
	}
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle(PauseStatus)
	systray.SetTooltip("MouseKeeper")

	// Initialize timeout menu items map
	mouseKeeper.timeoutMenus = make(map[time.Duration]*systray.MenuItem)

	// Add menu items
	mouseKeeper.pauseMenu = systray.AddMenuItem("Resume", "Resume/Pause mouse movement") // Default show Resume

	// Add timeout settings submenu
	mTimeout := systray.AddMenuItem("Check Timeout Settings", "Check mouse rest time and start simulation") // 检查鼠标休憩时间后开始模拟

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
		item := mTimeout.AddSubMenuItem(t.label, "Set timeout to "+t.label)
		mouseKeeper.timeoutMenus[t.duration] = item
		// Set initial check state
		if config.idleTimeout == t.duration {
			item.Check()
		}
	}

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the application")
	mVersion := systray.AddMenuItem("Version "+VERSION, "")
	mVersion.Disable()
	about := systray.AddMenuItem("About", URL)

	go func() {
		for range mouseKeeper.pauseMenu.ClickedCh {
			isPaused := config.getPaused()
			config.setPaused(!isPaused)
			mouseKeeper.updateMenuState(!isPaused)

			if !isPaused {
				// 添加一个短暂的延迟，避免检测到点击菜单的鼠标移动
				time.Sleep(time.Second)

				// 重置最后移动时间，避免立即开始移动
				mouseKeeper.lastMoveTime = time.Now()

				mouseKeeper.lastX, mouseKeeper.lastY = robotgo.Location()
			}
		}
	}()

	// Handle timeout settings
	for _, t := range timeouts {
		duration := t.duration // Create a new variable to avoid closure problems
		menuItem := mouseKeeper.timeoutMenus[duration]

		go func() {
			for range menuItem.ClickedCh {
				// Uncheck all items first
				for _, item := range mouseKeeper.timeoutMenus {
					item.Uncheck()
				}
				// Check the selected item
				menuItem.Check()
				config.setIdleTimeout(duration)
				Logger.Printf("Idle timeout set to %v", duration)
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
	// 清理工作
}

// 模拟真实的鼠标移动
func (mk *MouseKeeper) simulateRealisticMouseMovement(startX, startY int) {
	// 在移动前先检查一次用户活动
	if mk.checkUserActivity() {
		return
	}

	// 生成随机目标位置
	targetX := startX + nRrand.Intn(500) - 250
	targetY := startY + nRrand.Intn(500) - 250

	// 确保目标位置在屏幕范围内
	width, height := robotgo.GetScreenSize()
	targetX = max(0, min(width-1, targetX))
	targetY = max(0, min(height-1, targetY))

	mk.logger.Printf("Starting mouse movement to position (%d, %d)", targetX, targetY)

	// 记录这是系统移动
	mk.lastMoveTime = time.Now()
	robotgo.MoveSmooth(targetX, targetY, 1.0, 1.0)

	// 更新最后位置
	mk.lastX, mk.lastY = robotgo.Location()
	mk.logger.Printf("Mouse movement completed")
}

// 检查用户活动
func (mk *MouseKeeper) checkUserActivity() bool {
	currentX, currentY := robotgo.Location()
	deltaX := abs(currentX - mk.lastX)
	deltaY := abs(currentY - mk.lastY)

	// 检查是否是系统移动造成的位置变化
	if time.Since(mk.lastMoveTime) < time.Second {
		return false
	}

	if deltaX > 5 || deltaY > 5 {
		mk.logger.Printf("User activity detected (moved %d,%d pixels). System paused.", deltaX, deltaY)
		isPaused := config.getPaused()
		if !isPaused {
			config.setPaused(true)
			mk.updateMenuState(true)
		}
		mk.lastX, mk.lastY = currentX, currentY
		mk.lastMoveTime = time.Now()
		return true
	}
	return false
}

func (mk *MouseKeeper) start() {
	// 检查用户活动的goroutine
	go func() {
		for {
			time.Sleep(100 * time.Millisecond) // 更频繁地检查用户活动

			if config.getPaused() {
				continue
			}

			// 检查用户活动
			if mk.checkUserActivity() {
				continue
			}
		}
	}()

	// 自动移动鼠标的goroutine
	go func() {
		for {

			time.Sleep(time.Second) // 每秒检查一次

			// 检查是否超过空闲时间
			if time.Since(mk.lastMoveTime) >= config.idleTimeout {
				config.setPaused(false) // 恢复运行
				mk.updateMenuState(false)
				mk.logger.Printf("No mouse movement detected for %v, starting simulation", config.idleTimeout)
			}

			if config.getPaused() {
				continue
			}

			mouseKeeper.updateMenuState(config.isPaused) //确保初始化菜单状态正确

			mk.simulateRealisticMouseMovement(mk.lastX, mk.lastY)
			time.Sleep(time.Duration(1+nRrand.Intn(5)) * time.Second) // 随机等待1-5秒再次移动
			// time.Sleep(time.Duration(2) * time.Second) //DEBUG
		}
	}()

}

// 辅助函数：取最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 辅助函数：取最大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 辅助函数：计算绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func main() {
	// Initialize logger
	// Logger = log.New(os.Stdout, "INFO: ", log.LstdFlags|log.Lshortfile)
	Logger = log.New(os.Stdout, "INFO: ", log.LstdFlags)

	// Initialize random seed
	nRrand = rand.New(rand.NewSource(time.Now().UnixNano()))

	mouseKeeper = &MouseKeeper{
		logger:       Logger,
		timeoutMenus: make(map[time.Duration]*systray.MenuItem),
		lastX:        0,
		lastY:        0,
		lastMoveTime: time.Now(),
	}

	Logger.Printf("MouseKeeper started (Paused). Click 'Resume' to start")

	// 初始化鼠标位置
	mouseKeeper.lastX, mouseKeeper.lastY = robotgo.Location()
	mouseKeeper.lastMoveTime = time.Now()

	mouseKeeper.start()
	systray.Run(onReady, onExit)
}
