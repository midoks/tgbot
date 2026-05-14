package op

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"tgbot/internal/db"
	"tgbot/internal/tgtask"
)

// MessageInfo 消息信息结构体
type MessageInfo struct {
	MsgType      string
	MsgContent   string
	ChatName     string
	ChatType     string
	FromUserName string
}

// parseMessageInfo 解析消息信息
func parseMessageInfo(update tgbotapi.Update) MessageInfo {
	msgType := "unknown"
	msgContent := ""
	if update.Message.Text != "" {
		msgType = "text"
		msgContent = update.Message.Text
	} else if len(update.Message.Photo) > 0 {
		msgType = "photo"
		msgContent = fmt.Sprintf("%d photos", len(update.Message.Photo))
	} else if update.Message.Document != nil {
		msgType = "document"
		msgContent = update.Message.Document.FileName
	} else if update.Message.Voice != nil {
		msgType = "voice"
		msgContent = fmt.Sprintf("duration: %ds", update.Message.Voice.Duration)
	} else if update.Message.Video != nil {
		msgType = "video"
		msgContent = fmt.Sprintf("duration: %ds", update.Message.Video.Duration)
	} else if update.Message.Audio != nil {
		msgType = "audio"
		msgContent = update.Message.Audio.Title
	} else if update.Message.Sticker != nil {
		msgType = "sticker"
		msgContent = update.Message.Sticker.Emoji
	} else if update.Message.Location != nil {
		msgType = "location"
		msgContent = fmt.Sprintf("lat: %f, lng: %f", update.Message.Location.Latitude, update.Message.Location.Longitude)
	} else if update.Message.Contact != nil {
		msgType = "contact"
		msgContent = update.Message.Contact.PhoneNumber
	}

	chatName := ""
	if update.Message.Chat.Title != "" {
		chatName = update.Message.Chat.Title
	} else if update.Message.Chat.UserName != "" {
		chatName = "@" + update.Message.Chat.UserName
	} else if update.Message.Chat.FirstName != "" {
		chatName = update.Message.Chat.FirstName
		if update.Message.Chat.LastName != "" {
			chatName += " " + update.Message.Chat.LastName
		}
	} else {
		chatName = fmt.Sprintf("ChatID:%d", update.Message.Chat.ID)
	}

	return MessageInfo{
		MsgType:      msgType,
		MsgContent:   msgContent,
		ChatName:     chatName,
		ChatType:     update.Message.Chat.Type,
		FromUserName: update.Message.From.UserName,
	}
}

// logAndPrintMessage 记录并打印消息
func logAndPrintMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	msgInfo := parseMessageInfo(update)

	// 判断消息是否会被删除（包含@符号）
	op := "0" // 默认正常
	if strings.Contains(msgInfo.MsgContent, "@") {
		op = "1" // 标记为删除
	}

	// 记录消息到日志（完整字段）
	go db.AddTgbotLog(
		bot.Self.ID,
		update.Message.Chat.ID,
		update.Message.From.ID,
		msgInfo.ChatName,
		msgInfo.ChatType,
		msgInfo.FromUserName,
		msgInfo.MsgType,
		msgInfo.MsgContent,
		op,
		"info",
	)

	// 打印消息信息
	fmt.Printf("Received message - Chat: %s, Type: %s, Content: %s, From: %s, Op: %s\n",
		msgInfo.ChatName, msgInfo.MsgType, msgInfo.MsgContent, msgInfo.FromUserName, op)

	// 如果消息内容包含@符号，3秒后删除消息
	if op == "1" {
		go func() {
			time.Sleep(3 * time.Second)
			deleteMsg := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
			_, _ = bot.Send(deleteMsg)
			fmt.Printf("Deleted message %d from chat %d (contained @)\n",
				update.Message.MessageID, update.Message.Chat.ID)
		}()
	}
}

// 未选择策略
func TelegramMessageHandlerStrategyNone(relateMonitorGroupID int64) tgtask.MessageHandler {
	return func(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
		// 根据消息内容做不同处理
		switch update.Message.Text {
		case "/status":
			return HandleStatusCommand(update, bot)
		case "/start":
			fallthrough
		case "/?":
			fallthrough
		case "/help":
			return HandleHelpCommand(update, bot, false)
		default:
			// 记录并打印消息
			logAndPrintMessage(update, bot)

			// msg := tgbotapi.NewMessage(update.Message.Chat.ID, "未选择策略。")
			// _, err := bot.Send(msg)
			return nil
		}
	}
}
