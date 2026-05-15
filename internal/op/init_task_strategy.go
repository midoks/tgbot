package op

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	// "tgbot/internal/db"
)

// HandleStatusCommand 处理/status命令
func HandleStatusCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "正常运行!")
	_, err := bot.Send(msg)
	return err
}

// HandleHelpCommand 处理/help命令
func HandleHelpCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI, includeImportFormat bool) error {
	helpText := `可用命令:
/start - 开始使用
/status - 检查运行状态
/help - 显示此帮助信息`

	if includeImportFormat {
		helpText += ``
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
	_, err := bot.Send(msg)
	return err
}
