package op

import (
	"fmt"
	"math/rand"
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

	// 检查向其它(非本群的机器人发送消息)，导致无法删除恢复消息解决。

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
		switch update.Message.Text {
		case "/status":
			return HandleStatusCommand(update, bot)
		case "/admin_addr":
			return HandleAdminAddrCommand(update, bot)
		case "/start":
			fallthrough
		case "/?":
			fallthrough
		case "/help":
			return HandleHelpCommand(update, bot, false)
		default:
			// 检查是否是验证答案
			if handleVerificationResponse(update, bot) {
				return nil
			}
			logAndPrintMessage(update, bot)
			return nil
		}
	}
}

type verificationChallenge struct {
	Num1       int
	Num2       int
	Answer     int
	Options    []int // 4个选项
	ExpireTime time.Time
}

var (
	verificationCache      = make(map[int64]verificationChallenge)
	verificationCacheMu    sync.RWMutex
	verificationExpire     = 5 * time.Minute
	maxVerificationRetries = 3
)

func generateVerification() (int, int, int, []int) {
	rand.Seed(time.Now().UnixNano())
	num1 := rand.Intn(40) + 10
	num2 := rand.Intn(40) + 10
	answer := num1 + num2

	// 生成3个错误选项和1个正确选项
	options := make([]int, 4)
	options[0] = answer

	// 生成3个不同的错误答案
	for i := 1; i < 4; {
		wrongAnswer := answer + rand.Intn(30) - 15 // 在正确答案附近随机
		if wrongAnswer != answer && wrongAnswer > 0 {
			// 检查是否已存在
			exists := false
			for j := 0; j < i; j++ {
				if options[j] == wrongAnswer {
					exists = true
					break
				}
			}
			if !exists {
				options[i] = wrongAnswer
				i++
			}
		}
	}

	// 打乱选项顺序
	for i := range options {
		j := rand.Intn(i + 1)
		options[i], options[j] = options[j], options[i]
	}

	return num1, num2, answer, options
}

func clearVerification(userID int64) {
	verificationCacheMu.Lock()
	defer verificationCacheMu.Unlock()
	delete(verificationCache, userID)
}

// TelegramCallbackHandlerStrategyNone 处理内联按钮回调
func TelegramCallbackHandlerStrategyNone() tgtask.CallbackHandler {
	return func(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
		callback := update.CallbackQuery
		if callback == nil {
			return nil
		}

		userID := callback.From.ID
		chatID := callback.Message.Chat.ID
		data := callback.Data

		fmt.Printf("收到回调 - UserID: %d, ChatID: %d, Data: %s\n", userID, chatID, data)

		// 检查是否是验证回调
		if strings.HasPrefix(data, "verify_") {
			var answer int
			if _, err := fmt.Sscanf(data, "verify_%d", &answer); err != nil {
				fmt.Printf("解析回调数据失败: %v\n", err)
				return nil
			}

			verificationCacheMu.RLock()
			challenge, exists := verificationCache[userID]
			verificationCacheMu.RUnlock()

			if !exists {
				fmt.Printf("用户 %d 没有待验证的挑战\n", userID)
				return nil
			}

			if time.Now().After(challenge.ExpireTime) {
				clearVerification(userID)
				// 验证超时 - 用户保持禁言状态
				msg := tgbotapi.NewMessage(chatID, "验证超时！您已被禁言，请联系管理员。")
				_, _ = bot.Send(msg)
				fmt.Printf("User %d verification timeout (remains restricted)\n", userID)
				return nil
			}

			var responseMsg string
			if answer == challenge.Answer {
				// 验证成功 - 解除禁言
				clearVerification(userID)

				// 解除禁言，恢复所有权限
				unrestrictPermissions := tgbotapi.ChatPermissions{
					CanSendMessages:       true,
					CanSendMediaMessages:  true,
					CanSendPolls:          true,
					CanSendOtherMessages:  true,
					CanAddWebPagePreviews: true,
					CanChangeInfo:         false,
					CanInviteUsers:        true,
					CanPinMessages:        false,
				}
				unrestrictConfig := tgbotapi.RestrictChatMemberConfig{
					ChatMemberConfig: tgbotapi.ChatMemberConfig{
						ChatID: chatID,
						UserID: userID,
					},
					Permissions: &unrestrictPermissions,
				}
				_, err := bot.Request(unrestrictConfig)
				if err != nil {
					fmt.Printf("Failed to unrestrict user %d: %v\n", userID, err)
				} else {
					fmt.Printf("User %d unrestricted (verified)\n", userID)
				}

				responseMsg = "验证成功！欢迎加入群聊！"
				fmt.Printf("用户 %d 验证成功\n", userID)
			} else {
				// 验证失败 - 保持禁言状态
				responseMsg = fmt.Sprintf("答案错误！请重新选择：\n%d + %d = ?", challenge.Num1, challenge.Num2)
				fmt.Printf("用户 %d 验证失败，选择了 %d，正确答案是 %d\n", userID, answer, challenge.Answer)
			}

			// 发送回复消息
			msg := tgbotapi.NewMessage(chatID, responseMsg)
			_, err := bot.Send(msg)
			if err != nil {
				fmt.Printf("发送消息失败: %v\n", err)
			}

			// 回复回调（必须调用，否则按钮会一直显示加载状态）
			callbackConfig := tgbotapi.NewCallback(callback.ID, "")
			_, _ = bot.Request(callbackConfig)
		}

		return nil
	}
}

// handleVerificationResponse 处理用户的验证答案回复
// 返回 true 表示这是一个验证回复，false 表示不是
func handleVerificationResponse(update tgbotapi.Update, bot *tgbotapi.BotAPI) bool {
	userID := update.Message.From.ID

	verificationCacheMu.RLock()
	challenge, exists := verificationCache[userID]
	verificationCacheMu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(challenge.ExpireTime) {
		clearVerification(userID)
		return false
	}

	// 尝试解析用户回复
	userInput := strings.TrimSpace(strings.ToUpper(update.Message.Text))

	var answer int
	var isValidInput bool

	// 检查是否是选项字母（A、B、C、D）
	if len(userInput) == 1 {
		switch userInput {
		case "A":
			if len(challenge.Options) > 0 {
				answer = challenge.Options[0]
				isValidInput = true
			}
		case "B":
			if len(challenge.Options) > 1 {
				answer = challenge.Options[1]
				isValidInput = true
			}
		case "C":
			if len(challenge.Options) > 2 {
				answer = challenge.Options[2]
				isValidInput = true
			}
		case "D":
			if len(challenge.Options) > 3 {
				answer = challenge.Options[3]
				isValidInput = true
			}
		}
	}

	// 如果不是选项字母，尝试解析数字
	if !isValidInput {
		if _, err := fmt.Sscanf(update.Message.Text, "%d", &answer); err != nil {
			return false
		}
		isValidInput = true
	}

	if !isValidInput {
		return false
	}

	if answer == challenge.Answer {
		// 验证成功
		clearVerification(userID)
		successMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("验证成功！欢迎加入群聊！"))
		_, err := bot.Send(successMsg)
		if err != nil {
			fmt.Printf("Failed to send verification success message: %v\n", err)
		} else {
			fmt.Printf("User %d passed verification\n", userID)
		}
		return true
	}

	// 验证失败
	failMsg := tgbotapi.NewMessage(update.Message.Chat.ID,
		fmt.Sprintf("答案错误！请重新选择：\n%d + %d = ?\n\nA. %d\nB. %d\nC. %d\nD. %d",
			challenge.Num1, challenge.Num2,
			challenge.Options[0], challenge.Options[1],
			challenge.Options[2], challenge.Options[3]))
	_, err := bot.Send(failMsg)
	if err != nil {
		fmt.Printf("Failed to send verification fail message: %v\n", err)
	} else {
		fmt.Printf("User %d failed verification (answered %d, expected %d)\n", userID, answer, challenge.Answer)
	}
	return true
}

func TelegramChatMemberHandlerStrategyNone(relateMonitorGroupID int64) tgtask.ChatMemberHandler {
	return func(update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
		var chatID int64
		var userID int64
		var userName string
		var oldStatus string
		var newStatus string

		fmt.Printf("=== ChatMemberHandler 被调用 ===\n")
		fmt.Printf("update.ChatMember != nil: %v\n", update.ChatMember != nil)
		fmt.Printf("update.MyChatMember != nil: %v\n", update.MyChatMember != nil)

		if update.ChatMember != nil {
			// 新成员加入群聊事件
			fmt.Printf("处理 update.ChatMember 事件\n")
			chatID = update.ChatMember.Chat.ID
			userID = update.ChatMember.NewChatMember.User.ID
			if update.ChatMember.NewChatMember.User.UserName != "" {
				userName = update.ChatMember.NewChatMember.User.UserName
			} else if update.ChatMember.NewChatMember.User.FirstName != "" {
				userName = update.ChatMember.NewChatMember.User.FirstName
			} else {
				userName = "Unknown"
			}
			oldStatus = update.ChatMember.OldChatMember.Status
			newStatus = update.ChatMember.NewChatMember.Status
		} else if update.MyChatMember != nil {
			// 机器人自身状态变化事件
			fmt.Printf("处理 update.MyChatMember 事件\n")
			chatID = update.MyChatMember.Chat.ID
			userID = update.MyChatMember.From.ID
			if update.MyChatMember.From.UserName != "" {
				userName = update.MyChatMember.From.UserName
			} else if update.MyChatMember.From.FirstName != "" {
				userName = update.MyChatMember.From.FirstName
			} else {
				userName = "Unknown"
			}
			oldStatus = update.MyChatMember.OldChatMember.Status
			newStatus = update.MyChatMember.NewChatMember.Status
		} else {
			fmt.Printf("ChatMemberHandler: 没有有效的 ChatMember 更新\n")
			return nil
		}

		fmt.Printf("ChatMember update - ChatID: %d, User: %s (ID: %d), OldStatus: %s, NewStatus: %s\n",
			chatID, userName, userID, oldStatus, newStatus)

		// 用户加入群聊（包括重新加入）
		// 当状态从非 member 变为 member 时触发验证
		if newStatus == "member" && oldStatus != "member" {
			fmt.Printf("触发验证逻辑\n")
			// 先禁言用户，只允许阅读消息
			permissions := tgbotapi.ChatPermissions{
				CanSendMessages:       false,
				CanSendMediaMessages:  false,
				CanSendPolls:          false,
				CanSendOtherMessages:  false,
				CanAddWebPagePreviews: false,
				CanChangeInfo:         false,
				CanInviteUsers:        false,
				CanPinMessages:        false,
			}
			restrictConfig := tgbotapi.RestrictChatMemberConfig{
				ChatMemberConfig: tgbotapi.ChatMemberConfig{
					ChatID: chatID,
					UserID: userID,
				},
				Permissions: &permissions,
			}
			_, err := bot.Request(restrictConfig)
			if err != nil {
				fmt.Printf("Failed to restrict user %d: %v\n", userID, err)
			} else {
				fmt.Printf("User %d restricted (only can read messages)\n", userID)
			}

			num1, num2, answer, options := generateVerification()

			verificationCacheMu.Lock()
			verificationCache[userID] = verificationChallenge{
				Num1:       num1,
				Num2:       num2,
				Answer:     answer,
				Options:    options,
				ExpireTime: time.Now().Add(verificationExpire),
			}
			verificationCacheMu.Unlock()

			// 构造带内联键盘的验证消息
			verificationMsg := tgbotapi.NewMessage(chatID,
				fmt.Sprintf("欢迎 %s 加入群聊！\n请在5分钟内完成验证，否则将被禁言：\n\n%d + %d = ?", userName, num1, num2))

			// 创建内联键盘 - 4个按钮一行显示
			var keyboard tgbotapi.InlineKeyboardMarkup
			keyboard.InlineKeyboard = make([][]tgbotapi.InlineKeyboardButton, 1)
			keyboard.InlineKeyboard[0] = []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData("A. "+fmt.Sprintf("%d", options[0]), "verify_"+fmt.Sprintf("%d", options[0])),
				tgbotapi.NewInlineKeyboardButtonData("B. "+fmt.Sprintf("%d", options[1]), "verify_"+fmt.Sprintf("%d", options[1])),
				tgbotapi.NewInlineKeyboardButtonData("C. "+fmt.Sprintf("%d", options[2]), "verify_"+fmt.Sprintf("%d", options[2])),
				tgbotapi.NewInlineKeyboardButtonData("D. "+fmt.Sprintf("%d", options[3]), "verify_"+fmt.Sprintf("%d", options[3])),
			}
			verificationMsg.ReplyMarkup = keyboard
			_, err = bot.Send(verificationMsg)
			if err != nil {
				fmt.Printf("Failed to send verification message to user %d: %v\n", userID, err)
			} else {
				fmt.Printf("Sent verification challenge to user %d: %d + %d = %d\n", userID, num1, num2, answer)
			}
		}

		return nil
	}
}

func HandleVerificationAnswer(userID int64, answer int) bool {
	verificationCacheMu.RLock()
	challenge, exists := verificationCache[userID]
	verificationCacheMu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(challenge.ExpireTime) {
		clearVerification(userID)
		return false
	}

	if answer == challenge.Answer {
		clearVerification(userID)
		return true
	}

	return false
}
