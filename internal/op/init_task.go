package op

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"tgbot/internal/db"
	"tgbot/internal/model"
	"tgbot/internal/notify"
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

// PushMenuOnce 推送菜单一次
func PushMenuOnce(botData model.Tgbot, menu model.TgbotPushMenu) {
	// 检查配置是否完整
	if botData.MenuSendID == 0 {
		fmt.Printf("[MenuPush] Bot %d: No target chat ID configured\n", botData.ID)
		return
	}

	if botData.Token == "" {
		fmt.Printf("[MenuPush] Bot %d: No token configured\n", botData.ID)
		return
	}

	// 解析菜单数据
	var menuData struct {
		Message  string `json:"message"`
		Keyboard [][]struct {
			Text string `json:"text"`
			URL  string `json:"url"`
		} `json:"keyboard"`
	}

	if err := json.Unmarshal([]byte(menu.Params), &menuData); err != nil {
		fmt.Printf("[MenuPush] Bot %d: Failed to parse menu params: %v\n", botData.ID, err)
		return
	}

	// 构建代理字符串
	proxy := ""
	if botData.ProxyScheme != "" && botData.ProxyValue != "" {
		proxy = botData.ProxyScheme + "://" + botData.ProxyValue
	}

	// 使用 notify 包创建通知实例
	notification, err := notify.NewNotification(botData.Token, botData.MenuSendID, proxy, true)
	if err != nil {
		fmt.Printf("[MenuPush] Bot %d: Failed to create notification: %v\n", botData.ID, err)
		return
	}

	if !notification.Enabled {
		fmt.Printf("[MenuPush] Bot %d: Notification is not enabled\n", botData.ID)
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

	// 发送消息到指定的聊天/群
	fmt.Printf("[MenuPush] Bot %d, Chat %d: Sending menu...\n", botData.ID, botData.MenuSendID)

	msg := tgbotapi.NewMessage(botData.MenuSendID, menuData.Message)
	msg.ReplyMarkup = markup

	_, err = notification.TelegramBot().Send(msg)
	if err != nil {
		errStr := err.Error()
		fmt.Printf("[MenuPush] Bot %d, Chat %d: Failed to send menu: %v\n", botData.ID, botData.MenuSendID, err)

		// 添加常见错误提示
		if strings.Contains(errStr, "chat not found") {
			fmt.Printf("[MenuPush] HINT: Please ensure the bot is added to the group/channel and has permission to send messages\n")
		} else if strings.Contains(errStr, "bot was kicked") {
			fmt.Printf("[MenuPush] HINT: The bot has been kicked from the chat\n")
		} else if strings.Contains(errStr, "not a member") {
			fmt.Printf("[MenuPush] HINT: The bot is not a member of the chat\n")
		}
	} else {
		fmt.Printf("[MenuPush] Bot %d, Chat %d: Menu sent successfully\n", botData.ID, botData.MenuSendID)
	}
}

// InitMenuPushTasks 初始化所有菜单推送任务（一次性推送）
func InitMenuPushTasks() {
	fmt.Println("[MenuPush] Initializing menu push tasks")

	// 获取所有 bot
	botList, err := db.GetTgbotAll()
	if err != nil {
		fmt.Printf("[MenuPush] Failed to get bot list: %v\n", err)
		return
	}

	// 为每个配置了菜单推送的 bot 执行一次推送
	for _, botData := range botList {
		if botData.MenuRelatedID > 0 && botData.MenuSendID != 0 {
			// 获取关联的菜单
			menu, err := db.GetTgbotPushMenuByID(botData.MenuRelatedID)
			if err != nil {
				fmt.Printf("[MenuPush] Failed to get menu %d for bot %d: %v\n", botData.MenuRelatedID, botData.ID, err)
				continue
			}

			if !menu.Status {
				fmt.Printf("[MenuPush] Menu %d is not active, skipping bot %d\n", menu.ID, botData.ID)
				continue
			}

			fmt.Printf("[MenuPush] Pushing menu %d for bot %d to chat %d\n", menu.ID, botData.ID, botData.MenuSendID)

			// 执行一次推送
			PushMenuOnce(botData, *menu)
		}
	}
}
