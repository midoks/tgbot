package op

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"tgbot/internal/db"
	"tgbot/internal/model"
	"tgbot/internal/tgtask"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ReloadTelegramTask() {
	manager := tgtask.GetManager()
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
}

func InitTelegramTask() {
	fmt.Println("InitTelegramTask")
	manager := tgtask.GetManager()

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

type DomainEntry struct {
	Remark string
	URL    string
}

// ParseDomainEntries 解析域名数据
func ParseDomainEntries(text string) ([]DomainEntry, error) {
	var entries []DomainEntry
	var currentRemark string

	lines := strings.Split(text, "\n")
	urlPattern := regexp.MustCompile(`^https?://`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.Contains(line, "=========================") {
			currentRemark = ""
			continue
		}

		// 检查这一行是否是 URL（以 http:// 或 https:// 开头）
		cleanLine := strings.Trim(line, "`")
		if urlPattern.MatchString(cleanLine) {
			if currentRemark != "" {
				entries = append(entries, DomainEntry{
					Remark: currentRemark,
					URL:    cleanLine,
				})
				currentRemark = ""
			}
		} else if strings.Contains(line, ":") || strings.Contains(line, "：") {
			// 这一行可能是备注行（支持英文和中文冒号）
			colonIndex := strings.Index(line, ":")
			if colonIndex == -1 {
				colonIndex = strings.Index(line, "：")
			}
			if colonIndex != -1 {
				remark := strings.TrimSpace(line[:colonIndex])
				rest := strings.TrimSpace(line[colonIndex+1:])

				// 移除反引号
				rest = strings.Trim(rest, "`")

				if urlPattern.MatchString(rest) {
					// 备注和 URL 在同一行
					entries = append(entries, DomainEntry{
						Remark: remark,
						URL:    rest,
					})
					currentRemark = ""
				} else {
					// 这是备注行，URL 在下一行
					currentRemark = remark
				}
			}
		}
	}

	return entries, nil
}

func CreateMonitorsFromText(text string, gid int64) (successCount, failCount int, err error) {
	entries, err := ParseDomainEntries(text)

	if err != nil {
		return 0, 0, err
	}

	if len(entries) == 0 {
		return 0, 0, nil
	}

	successCount = len(entries)
	return successCount, failCount, nil
}

func CreateMonitorsFromTextAppend(text string, gid int64) (successCount, failCount int, err error) {
	entries, err := ParseDomainEntries(text)

	if err != nil {
		return 0, 0, err
	}

	if len(entries) == 0 {
		return 0, 0, nil
	}

	successCount = len(entries)
	return successCount, failCount, nil
}

// MenuPushTask 菜单推送任务
type MenuPushTask struct {
	botID     int64
	menu      model.TgbotPushMenu
	freq      int64 // 推送频率（秒）
	stopChan  chan struct{}
	isRunning bool
}

// StartMenuPushTask 启动菜单推送任务
func StartMenuPushTask(botID int64, menu model.TgbotPushMenu, freq int64) *MenuPushTask {
	task := &MenuPushTask{
		botID:     botID,
		menu:      menu,
		freq:      freq,
		stopChan:  make(chan struct{}),
		isRunning: true,
	}

	go task.run()
	return task
}

// run 运行推送任务
func (t *MenuPushTask) run() {
	ticker := time.NewTicker(time.Duration(t.freq) * time.Second)
	defer ticker.Stop()

	// 立即执行一次
	t.pushMenu()

	for {
		select {
		case <-ticker.C:
			t.pushMenu()
		case <-t.stopChan:
			fmt.Printf("Menu push task for bot %d stopped\n", t.botID)
			return
		}
	}
}

// pushMenu 推送菜单
func (t *MenuPushTask) pushMenu() {
	// 解析菜单数据
	var menuData struct {
		Message  string `json:"message"`
		Keyboard [][]struct {
			Text string `json:"text"`
			URL  string `json:"url"`
		} `json:"keyboard"`
	}

	if err := json.Unmarshal([]byte(t.menu.Params), &menuData); err != nil {
		fmt.Printf("Failed to parse menu params for bot %d: %v\n", t.botID, err)
		return
	}

	// 获取 bot 实例
	manager := tgtask.GetManager()
	botInstance := manager.GetBot(t.botID)
	if botInstance == nil || botInstance.BotAPI == nil {
		fmt.Printf("Bot %d not found\n", t.botID)
		return
	}

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

	// 创建内联键盘标记
	markup := tgbotapi.NewInlineKeyboardMarkup(inlineKeyboard...)

	// 这里需要获取要推送的聊天 ID，可以从配置中读取或者从数据库获取
	// 暂时先打印日志
	fmt.Printf("Pushing menu to bot %d: %s\n", t.botID, menuData.Message)

	// 实际推送逻辑需要根据具体需求实现
	// 例如：推送到特定的群组或频道
	chatIDs := []int64{} // 从配置或数据库获取

	for _, chatID := range chatIDs {
		msg := tgbotapi.NewMessage(chatID, menuData.Message)
		msg.ReplyMarkup = markup

		_, err := botInstance.BotAPI.Send(msg)
		if err != nil {
			fmt.Printf("Failed to send menu to chat %d: %v\n", chatID, err)
		} else {
			fmt.Printf("Menu sent to chat %d successfully\n", chatID)
		}
	}
}

// Stop 停止推送任务
func (t *MenuPushTask) Stop() {
	if t.isRunning {
		close(t.stopChan)
		t.isRunning = false
	}
}

// InitMenuPushTasks 初始化所有菜单推送任务
func InitMenuPushTasks() {
	fmt.Println("InitMenuPushTasks")

	// 获取所有激活的 bot
	botList, err := db.GetTgbotAll()
	if err != nil {
		fmt.Printf("Failed to get bot list: %v\n", err)
		return
	}

	// 为每个 bot 启动菜单推送任务
	for _, botData := range botList {
		if botData.MenuRelatedID > 0 && botData.MenuFreq > 0 {
			// 获取关联的菜单
			menu, err := db.GetTgbotPushMenuByID(botData.MenuRelatedID)
			if err != nil {
				fmt.Printf("Failed to get menu %d for bot %d: %v\n", botData.MenuRelatedID, botData.ID, err)
				continue
			}

			if !menu.Status {
				fmt.Printf("Menu %d is not active, skipping\n", menu.ID)
				continue
			}

			fmt.Printf("Starting menu push task for bot %d with menu %d, freq %d seconds\n",
				botData.ID, menu.ID, botData.MenuFreq)

			// 启动推送任务
			StartMenuPushTask(botData.ID, *menu, botData.MenuFreq)
		}
	}
}
