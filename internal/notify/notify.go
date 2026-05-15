package notify

import (
	"context"
	"fmt"
	"strings"
	"time"

	"tgbot/internal/tgtask"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Notification 通知服务
type Notification struct {
	telegramBot *tgbotapi.BotAPI
	ChatID      int64
	Enabled     bool
	Token       string
}

// NewNotification 创建通知实例
func NewNotification(token string, chatID int64, proxy string, enabled bool) (*Notification, error) {
	if !enabled {
		return &Notification{Enabled: false}, nil
	}

	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// 使用 tgtask 包的 CreateBotWithProxy 函数创建 bot 实例
	bot, err := tgtask.CreateBotWithProxy(token, proxy)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	return &Notification{
		telegramBot: bot,
		ChatID:      chatID,
		Enabled:     true,
		Token:       token,
	}, nil
}

// Send 发送标准通知消息
func (n *Notification) Send(ctx context.Context, title, content string) error {
	if !n.Enabled {
		return nil
	}

	message := fmt.Sprintf("%s:\n%s", title, content)
	return n.SendText(ctx, message)
}

// SendText 发送纯文本消息
func (n *Notification) SendText(ctx context.Context, text string) error {
	if !n.Enabled {
		return nil
	}

	if text == "" {
		return fmt.Errorf("message text cannot be empty")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		msg := tgbotapi.NewMessage(n.ChatID, text)
		_, err := n.telegramBot.Send(msg)
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}
}

// SendHTML 发送 HTML 格式消息
func (n *Notification) SendHTML(ctx context.Context, html string) error {
	if !n.Enabled {
		return nil
	}

	if html == "" {
		return fmt.Errorf("HTML content cannot be empty")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		msg := tgbotapi.NewMessage(n.ChatID, html)
		msg.ParseMode = "HTML"
		_, err := n.telegramBot.Send(msg)
		if err != nil {
			return fmt.Errorf("failed to send HTML message: %w", err)
		}
		return nil
	}
}

// SendMarkdown 发送 Markdown 格式消息
func (n *Notification) SendMarkdown(ctx context.Context, markdown string) error {
	if !n.Enabled {
		return nil
	}

	if markdown == "" {
		return fmt.Errorf("Markdown content cannot be empty")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		msg := tgbotapi.NewMessage(n.ChatID, markdown)
		msg.ParseMode = "MarkdownV2"
		_, err := n.telegramBot.Send(msg)
		if err != nil {
			return fmt.Errorf("failed to send Markdown message: %w", err)
		}
		return nil
	}
}

// SendWithKeyboard 发送带内联键盘的消息
func (n *Notification) SendWithKeyboard(ctx context.Context, text string, keyboard [][]tgbotapi.InlineKeyboardButton) error {
	if !n.Enabled {
		return nil
	}

	if text == "" {
		return fmt.Errorf("message text cannot be empty")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		msg := tgbotapi.NewMessage(n.ChatID, text)

		// 只有当键盘不为空时才设置
		if len(keyboard) > 0 {
			markup := tgbotapi.NewInlineKeyboardMarkup(keyboard...)
			msg.ReplyMarkup = markup
		}

		_, err := n.telegramBot.Send(msg)
		if err != nil {
			return fmt.Errorf("failed to send message with keyboard: %w", err)
		}
		return nil
	}
}

// SendPhoto 发送照片
func (n *Notification) SendPhoto(ctx context.Context, photoURL string, caption string) error {
	if !n.Enabled {
		return nil
	}

	if photoURL == "" {
		return fmt.Errorf("photo URL cannot be empty")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		photo := tgbotapi.NewPhoto(n.ChatID, tgbotapi.FileURL(photoURL))
		if caption != "" {
			photo.Caption = caption
		}

		_, err := n.telegramBot.Send(photo)
		if err != nil {
			return fmt.Errorf("failed to send photo: %w", err)
		}
		return nil
	}
}

// SendWithRetry 发送消息并支持重试
func (n *Notification) SendWithRetry(ctx context.Context, text string, maxRetries int) error {
	if !n.Enabled {
		return nil
	}

	var err error
	for i := 0; i < maxRetries; i++ {
		err = n.SendText(ctx, text)
		if err == nil {
			return nil
		}

		// 检查是否是可重试的错误
		if !IsRetryableError(err) {
			return err
		}

		// 指数退避重试
		delay := time.Duration(i+1) * 2 * time.Second
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			continue
		}
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}

// IsRetryableError 判断错误是否可重试
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"500 Internal Server Error",
		"502 Bad Gateway",
		"503 Service Unavailable",
		"504 Gateway Timeout",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(retryable)) {
			return true
		}
	}

	return false
}

// IsChatNotFoundError 判断是否是聊天不存在错误
func IsChatNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "chat not found")
}

// IsBotKickedError 判断是否是 bot 被踢出错误
func IsBotKickedError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "bot was kicked")
}

// IsNotMemberError 判断是否是 bot 不是成员错误
func IsNotMemberError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "not a member")
}

// GetErrorHint 获取错误提示信息
func GetErrorHint(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()
	if strings.Contains(errStr, "chat not found") {
		return "Please ensure the bot is added to the group/channel and has permission to send messages"
	} else if strings.Contains(errStr, "bot was kicked") {
		return "The bot has been kicked from the chat"
	} else if strings.Contains(errStr, "not a member") {
		return "The bot is not a member of the chat"
	} else if strings.Contains(errStr, "permission denied") {
		return "The bot does not have permission to send messages in this chat"
	}

	return ""
}

// TelegramBot 返回底层的 Telegram Bot 实例
func (n *Notification) TelegramBot() *tgbotapi.BotAPI {
	return n.telegramBot
}

// SetChatID 设置目标聊天 ID
func (n *Notification) SetChatID(chatID int64) {
	n.ChatID = chatID
}

// SetEnabled 设置启用状态
func (n *Notification) SetEnabled(enabled bool) {
	n.Enabled = enabled
}
