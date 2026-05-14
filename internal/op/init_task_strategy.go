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
		helpText += `

批量导入格式:
备注: https://example.com
备注: https://test.com
=========================
备注: https://domain.com

说明:
- 每行格式: 备注: URL
- 使用 ========================= 作为分组分隔符
- URL 必须以 http:// 或 https:// 开头`
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
	_, err := bot.Send(msg)
	return err
}
