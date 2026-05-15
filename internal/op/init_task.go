package op

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"tgbot/internal/db"
	"tgbot/internal/model"
	"tgbot/internal/notify"
	"tgbot/internal/tgtask"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var manager = tgtask.GetManager()

// MenuPushTask 菜单推送任务
type MenuPushTask struct {
	botID     int64
	chatID    int64
	menu      model.TgbotPushMenu
	freq      int64 // 推送频率（秒）
	stopChan  chan struct{}
	isRunning bool
	mutex     sync.Mutex
	notify    *notify.Notification
}

var menuPushTasks sync.Map // 存储所有菜单推送任务

// StartMenuPushTask 启动菜单推送任务
func StartMenuPushTask(botData model.Tgbot, menu model.TgbotPushMenu, freq int64) *MenuPushTask {
	// 构建代理字符串
	proxy := ""
	if botData.ProxyScheme != "" && botData.ProxyValue != "" {
		proxy = botData.ProxyScheme + "://" + botData.ProxyValue
	}

	// 创建通知实例
	notification, err := notify.NewNotification(botData.Token, botData.MenuSendID, proxy, true)
	if err != nil {
		fmt.Printf("[MenuPush] Failed to create notification for bot %d: %v\n", botData.ID, err)
		return nil
	}

	task := &MenuPushTask{
		botID:     botData.ID,
		chatID:    botData.MenuSendID,
		menu:      menu,
		freq:      freq,
		stopChan:  make(chan struct{}),
		isRunning: true,
		notify:    notification,
	}

	// 保存到全局任务 map
	menuPushTasks.Store(botData.ID, task)

	go task.run()
	return task
}

// run 运行推送任务
func (t *MenuPushTask) run() {
	defer func() {
		t.mutex.Lock()
		t.isRunning = false
		t.mutex.Unlock()
		menuPushTasks.Delete(t.botID)
	}()

	// 立即执行一次推送
	t.pushMenu()

	// 如果频率为 0 或负数，只执行一次
	if t.freq <= 0 {
		fmt.Printf("[MenuPush] Bot %d: One-time push completed, freq=%d\n", t.botID, t.freq)
		return
	}

	// 创建定时器
	ticker := time.NewTicker(time.Duration(t.freq) * time.Second)
	defer ticker.Stop()

	// fmt.Printf("[MenuPush] Bot %d: Started periodic push, freq=%d seconds\n", t.botID, t.freq)
	for {
		select {
		case <-ticker.C:
			t.pushMenu()
		case <-t.stopChan:
			fmt.Printf("[MenuPush] Bot %d: Task stopped\n", t.botID)
			return
		}
	}
}

// pushMenu 推送菜单
func (t *MenuPushTask) pushMenu() {
	// fmt.Printf("[MenuPush] Bot %d, Chat %d: Starting pushMenu\n", t.botID, t.chatID)

	// 检查 notify 实例
	if t.notify == nil {
		fmt.Printf("[MenuPush] Bot %d: notify instance is nil\n", t.botID)
		return
	}

	if !t.notify.Enabled {
		fmt.Printf("[MenuPush] Bot %d: notify is not enabled\n", t.botID)
		return
	}

	// 检查聊天 ID
	if t.chatID == 0 {
		fmt.Printf("[MenuPush] Bot %d: chatID is 0\n", t.botID)
		return
	}

	// 解析菜单数据
	// fmt.Printf("[MenuPush] Bot %d: Parsing menu params: %s\n", t.botID, t.menu.Params)
	var menuData struct {
		Message  string `json:"message"`
		Keyboard [][]struct {
			Text string `json:"text"`
			URL  string `json:"url"`
		} `json:"keyboard"`
	}

	if err := json.Unmarshal([]byte(t.menu.Params), &menuData); err != nil {
		fmt.Printf("[MenuPush] Bot %d: Failed to parse menu params: %v\n", t.botID, err)
		return
	}

	// fmt.Printf("[MenuPush] Bot %d: Menu message length: %d\n", t.botID, len(menuData.Message))
	// fmt.Printf("[MenuPush] Bot %d: Keyboard rows: %d\n", t.botID, len(menuData.Keyboard))

	// 构建内联键盘
	var inlineKeyboard [][]tgbotapi.InlineKeyboardButton
	for _, row := range menuData.Keyboard {
		var keyboardRow []tgbotapi.InlineKeyboardButton
		for _, btn := range row {
			urlStr := btn.URL
			keyboardRow = append(keyboardRow, tgbotapi.InlineKeyboardButton{
				Text: btn.Text,
				URL:  &urlStr,
			})
		}
		inlineKeyboard = append(inlineKeyboard, keyboardRow)
	}

	// fmt.Printf("[MenuPush] Bot %d, Chat %d: Sending menu...\n", t.botID, t.chatID)
	// 使用 notify 包的 SendWithKeyboardReturningID 方法发送并获取消息 ID
	messageID, err := t.notify.SendWithKeyboardReturningID(context.Background(), menuData.Message, inlineKeyboard)
	if err == nil {
		// fmt.Printf("[MenuPush] Bot %d, Chat %d: Menu sent successfully, messageID=%d\n", t.botID, t.chatID, messageID)
		t.deleteMessageAfter(t.chatID, messageID, t.freq-1)
	}
}

// deleteMessageAfter 延迟删除消息
func (t *MenuPushTask) deleteMessageAfter(chatID int64, messageID int, delaySeconds int64) {
	go func() {
		select {
		case <-time.After(time.Duration(delaySeconds) * time.Second):
			// fmt.Printf("[MenuPush] Bot %d: Deleting message %d after %d seconds\n", t.botID, messageID, delaySeconds)

			deleteConfig := tgbotapi.NewDeleteMessage(chatID, messageID)
			_, err := t.notify.TelegramBot().Request(deleteConfig)
			if err != nil {
				// fmt.Printf("[MenuPush] Bot %d: Failed to delete message %d: %v\n", t.botID, messageID, err)
			} else {
				// fmt.Printf("[MenuPush] Bot %d: Message %d deleted successfully\n", t.botID, messageID)
			}
		case <-t.stopChan:
			// fmt.Printf("[MenuPush] Bot %d: Stopping, cancel message deletion for message %d\n", t.botID, messageID)
			return
		}
	}()
}

// Stop 停止推送任务
func (t *MenuPushTask) Stop() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.isRunning {
		close(t.stopChan)
		t.isRunning = false
	}
}

// IsRunning 检查任务是否正在运行
func (t *MenuPushTask) IsRunning() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.isRunning
}

// StopAllMenuPushTasks 停止所有菜单推送任务
func StopAllMenuPushTasks() {
	// fmt.Println("[MenuPush] Stopping all menu push tasks")
	menuPushTasks.Range(func(key, value interface{}) bool {
		if task, ok := value.(*MenuPushTask); ok {
			task.Stop()
		}
		return true
	})
}

// GetMenuPushTask 获取指定 bot 的推送任务
func GetMenuPushTask(botID int64) *MenuPushTask {
	if value, ok := menuPushTasks.Load(botID); ok {
		return value.(*MenuPushTask)
	}
	return nil
}

// RestartMenuPushTask 重启指定 bot 的推送任务
func RestartMenuPushTask(botData model.Tgbot, menu model.TgbotPushMenu, freq int64) {
	// 先停止现有任务
	if existingTask := GetMenuPushTask(botData.ID); existingTask != nil {
		existingTask.Stop()
		time.Sleep(1 * time.Second) // 等待任务完全停止
	}

	// 启动新任务
	if freq > 0 || botData.MenuSendID != 0 {
		StartMenuPushTask(botData, menu, freq)
	}
}

// InitMenuPushTasks 初始化所有菜单推送任务
func InitMenuPushTasks() {
	// fmt.Println("[MenuPush] Initializing menu push tasks")

	// 获取所有 bot
	botList, err := db.GetTgbotAll()
	if err != nil {
		// fmt.Printf("[MenuPush] Failed to get bot list: %v\n", err)
		return
	}

	// 为每个配置了菜单推送的 bot 启动推送任务
	for _, botData := range botList {
		if botData.MenuRelatedID > 0 && botData.MenuSendID != 0 {
			// 获取关联的菜单
			menu, err := db.GetTgbotPushMenuByID(botData.MenuRelatedID)
			if err != nil {
				// fmt.Printf("[MenuPush] Failed to get menu %d for bot %d: %v\n", botData.MenuRelatedID, botData.ID, err)
				continue
			}

			if !menu.Status {
				// fmt.Printf("[MenuPush] Menu %d is not active, skipping bot %d\n", menu.ID, botData.ID)
				continue
			}

			// fmt.Printf("[MenuPush] Starting menu push task for bot %d with menu %d, chat %d, freq %d seconds\n",
			// botData.ID, menu.ID, botData.MenuSendID, botData.MenuFreq)

			// 启动推送任务
			StartMenuPushTask(botData, *menu, botData.MenuFreq)
		}
	}
}

// ReloadMenuPushTasks 重新加载所有菜单推送任务
func ReloadMenuPushTasks() {
	// fmt.Println("[MenuPush] Reloading menu push tasks")

	// 先停止所有现有任务
	StopAllMenuPushTasks()
	time.Sleep(2 * time.Second) // 等待所有任务完全停止

	// 重新初始化
	InitMenuPushTasks()
}

func ReloadTelegramTask() {
	// 停止所有菜单推送任务
	StopAllMenuPushTasks()

	// 获取当前的 Telegram 配置
	telegram_list, err := db.GetTgbotAll()
	if err != nil {
		fmt.Printf("failed to get tgbot data: %v\n", err)
		return
	}

	// 移除所有现有的机器人
	if err := manager.RemoveAllBots(); err != nil {
		fmt.Printf("failed to remove all bots: %v\n", err)
	}

	// 等待足够的时间确保所有 bot 实例完全停止
	time.Sleep(3 * time.Second)
	// fmt.Println("移除所有现有的机器人成功!!")

	// 重新添加和启动机器人
	if len(telegram_list) > 0 {
		for _, data := range telegram_list {
			if !data.ListenEnable {
				continue
			}

			botID := data.ID
			proxy := ""
			if data.ProxyScheme != "" && data.ProxyValue != "" {
				proxy = data.ProxyScheme + "://" + data.ProxyValue
			}

			var handler tgtask.MessageHandler
			handler = TelegramMessageHandlerStrategyNone(0)

			if err := manager.AddBot(botID, data.Token, proxy, 0, handler); err != nil {
				fmt.Printf("failed to add bot: %v\n", err)
				continue
			}

			if err := manager.StartBot(botID); err != nil {
				fmt.Printf("failed to start bot: %v\n", err)
				continue
			}
		}
	}

	// 重新初始化菜单推送任务
	InitMenuPushTasks()
}

func InitTelegramTask() {
	// fmt.Println("InitTelegramTask")

	// 获取所有 bot 配置
	telegram_list, err := db.GetTgbotAll()
	if err != nil {
		fmt.Printf("failed to get tgbot data: %v\n", err)
		return
	}

	if len(telegram_list) == 0 {
		return
	}

	for _, data := range telegram_list {
		if !data.ListenEnable {
			continue
		}

		botID := data.ID
		proxy := ""
		if data.ProxyScheme != "" && data.ProxyValue != "" {
			proxy = data.ProxyScheme + "://" + data.ProxyValue
		}

		var handler tgtask.MessageHandler
		handler = TelegramMessageHandlerStrategyNone(0)

		if err := manager.AddBot(botID, data.Token, proxy, 0, handler); err != nil {
			fmt.Printf("failed to add bot: %v\n", err)
			continue
		}

		if err := manager.StartBot(botID); err != nil {
			fmt.Printf("failed to start bot: %v\n", err)
			continue
		}
	}

	// 初始化菜单推送任务
	InitMenuPushTasks()
}
