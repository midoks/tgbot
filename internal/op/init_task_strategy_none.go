package op

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"tgbot/internal/db"
	"tgbot/internal/model"
	"tgbot/internal/tgtask"
)

var (
	banwordsCache     []model.TgbotBanWord
	banwordsCacheTime time.Time
	banwordsCacheMu   sync.RWMutex
	signadCache       map[int64]bool
	signadCacheTime   time.Time
	signadCacheMu     sync.RWMutex
	cacheExpireTime   = 5 * time.Minute
)

func ClearBanwordsCache() {
	banwordsCacheMu.Lock()
	defer banwordsCacheMu.Unlock()
	banwordsCache = nil
	banwordsCacheTime = time.Time{}
}

func ClearSignadCache() {
	signadCacheMu.Lock()
	defer signadCacheMu.Unlock()
	signadCache = nil
	signadCacheTime = time.Time{}
}

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
		if update.Message.Sticker.Emoji != "" {
			msgContent = update.Message.Sticker.Emoji
		} else {
			msgContent = "[Sticker]"
		}
	} else if update.Message.Dice != nil {
		msgType = "dice"
		msgContent = fmt.Sprintf("%s: %d", update.Message.Dice.Emoji, update.Message.Dice.Value)
	} else if update.Message.Location != nil {
		msgType = "location"
		msgContent = fmt.Sprintf("lat: %f, lng: %f", update.Message.Location.Latitude, update.Message.Location.Longitude)
	} else if update.Message.Contact != nil {
		msgType = "contact"
		msgContent = update.Message.Contact.PhoneNumber
	} else {
		msgContent = "[Message]"
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

	fromUserName := ""
	if update.Message.From != nil {
		if update.Message.From.UserName != "" {
			fromUserName = "@" + update.Message.From.UserName
		} else if update.Message.From.FirstName != "" {
			fromUserName = update.Message.From.FirstName
			if update.Message.From.LastName != "" {
				fromUserName += " " + update.Message.From.LastName
			}
		} else {
			fromUserName = fmt.Sprintf("UserID:%d", update.Message.From.ID)
		}
	} else {
		fromUserName = "Unknown"
	}

	return MessageInfo{
		MsgType:      msgType,
		MsgContent:   msgContent,
		ChatName:     chatName,
		ChatType:     update.Message.Chat.Type,
		FromUserName: fromUserName,
	}
}

// checkContentForBanwords 检查内容是否包含违禁词
// 返回是否包含违禁词，以及匹配的违禁词
func checkContentForBanwords(content string) (bool, string) {
	banwords, err := getActiveBanwordsWithCache()
	if err != nil {
		fmt.Printf("Failed to get active banwords: %v\n", err)
		return false, ""
	}

	for _, banword := range banwords {
		words := strings.Split(banword.Word, ",")
		for _, word := range words {
			word = strings.TrimSpace(word)
			if word != "" && strings.Contains(content, word) {
				return true, word
			}
		}
	}

	return false, ""
}

// getActiveBanwordsWithCache 获取违禁词列表（带缓存）
func getActiveBanwordsWithCache() ([]model.TgbotBanWord, error) {
	banwordsCacheMu.RLock()
	if banwordsCache != nil && time.Since(banwordsCacheTime) < cacheExpireTime {
		result := make([]model.TgbotBanWord, len(banwordsCache))
		copy(result, banwordsCache)
		banwordsCacheMu.RUnlock()
		return result, nil
	}
	banwordsCacheMu.RUnlock()

	banwordsCacheMu.Lock()
	defer banwordsCacheMu.Unlock()

	if banwordsCache != nil && time.Since(banwordsCacheTime) < cacheExpireTime {
		result := make([]model.TgbotBanWord, len(banwordsCache))
		copy(result, banwordsCache)
		return result, nil
	}

	list, err := db.GetActiveTgbotBanwords()
	if err != nil {
		return nil, err
	}

	banwordsCache = make([]model.TgbotBanWord, len(list))
	copy(banwordsCache, list)
	banwordsCacheTime = time.Now()

	result := make([]model.TgbotBanWord, len(list))
	copy(result, list)
	return result, nil
}

// isSignadUser 检查用户是否是推广用户（带缓存）
func isSignadUser(userID int64) bool {
	signadCacheMu.RLock()
	if signadCache != nil && time.Since(signadCacheTime) < cacheExpireTime {
		isSignad := signadCache[userID]
		signadCacheMu.RUnlock()
		return isSignad
	}
	signadCacheMu.RUnlock()

	signadCacheMu.Lock()
	defer signadCacheMu.Unlock()

	if signadCache != nil && time.Since(signadCacheTime) < cacheExpireTime {
		return signadCache[userID]
	}

	list, err := db.GetActiveTgbotSignad()
	if err != nil {
		fmt.Printf("Failed to get active signads: %v\n", err)
		return false
	}

	signadCache = make(map[int64]bool)
	for _, signad := range list {
		signadCache[signad.UserID] = true
	}
	signadCacheTime = time.Now()

	return signadCache[userID]
}

// logAndPrintMessage 记录并打印消息
func logAndPrintMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	msgInfo := parseMessageInfo(update)

	// 判断消息是否会被删除（包含违禁词或@符号）
	op := 0 // 默认正常
	matchedWord := ""
	hasBanword, word := checkContentForBanwords(msgInfo.MsgContent)
	if hasBanword {
		op = 1
		matchedWord = word
	}

	// 检查是否是推广用户
	isSignad := isSignadUser(update.Message.From.ID)
	if isSignad {
		op = 2
	}

	// 发联系信息删除
	if msgInfo.MsgType == "contact" {
		op = 1
	}

	// 记录消息到日志
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
	fmt.Printf("Received message - Chat: %s, Type: %s, Content: %s, From: %s, Op: %d\n",
		msgInfo.ChatName, msgInfo.MsgType, msgInfo.MsgContent, msgInfo.FromUserName, op)

	// 如果消息需要删除，违禁词3秒后删除，推广用户30秒后删除
	if op == 1 {
		go func() {
			time.Sleep(2 * time.Second)
			deleteMsg := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
			_, err := bot.Request(deleteMsg)
			if err != nil {
				errStr := err.Error()
				// 忽略消息已被删除的错误（可能已被用户删除或超过48小时）
				if strings.Contains(errStr, "Bad Request: message can't be deleted") {
					fmt.Printf("Message %d already deleted or can't be deleted (matched: %q)\n",
						update.Message.MessageID, matchedWord)
				} else {
					fmt.Printf("Failed to delete message %d from chat %d (matched: %q): %v\n",
						update.Message.MessageID, update.Message.Chat.ID, matchedWord, err)
				}
			} else {
				fmt.Printf("Deleted message %d from chat %d (matched: %q)\n",
					update.Message.MessageID, update.Message.Chat.ID, matchedWord)
			}
		}()
	} else if op == 2 {
		go func() {
			time.Sleep(6 * time.Second)
			deleteMsg := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
			_, err := bot.Request(deleteMsg)
			if err != nil {
				errStr := err.Error()
				// 忽略消息已被删除的错误（可能已被用户删除或超过48小时）
				if strings.Contains(errStr, "Bad Request: message can't be deleted") {
					fmt.Printf("Message %d already deleted or can't be deleted (signad user)\n",
						update.Message.MessageID)
				} else {
					fmt.Printf("Failed to delete message %d from chat %d (signad user): %v\n",
						update.Message.MessageID, update.Message.Chat.ID, err)
				}
			} else {
				fmt.Printf("Deleted message %d from chat %d (signad user)\n",
					update.Message.MessageID, update.Message.Chat.ID)
			}
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
