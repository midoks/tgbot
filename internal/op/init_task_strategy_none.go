package op

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"tgbot/internal/tgtask"
)

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
			// 打印消息类型和内容
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
			// 获取聊天名称
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
			fmt.Printf("Received message - Chat: %s, Type: %s, Content: %s, From: %s\n", chatName, msgType, msgContent, update.Message.From.UserName)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "未选择策略。")
			_, err := bot.Send(msg)
			return err
		}
	}
}
