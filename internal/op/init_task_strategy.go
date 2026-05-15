package op

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
/admin_addr - 获取后台管理地址
/help - 显示此帮助信息`

	if includeImportFormat {
		helpText += ``
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
	_, err := bot.Send(msg)
	return err
}

// HandleAdminAddrCommand 处理/admin_addr命令
func HandleAdminAddrCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	ip := getLocalIP()
	addr := fmt.Sprintf("后台管理地址: http://%s:9393/tgbot", ip)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, addr)
	_, err := bot.Send(msg)
	return err
}
