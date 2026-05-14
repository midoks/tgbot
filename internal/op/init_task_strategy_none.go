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
			fmt.Println("update", update)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "未选择策略。")
			_, err := bot.Send(msg)
			return err
		}
	}
}
