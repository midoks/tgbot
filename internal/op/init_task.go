package op

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"tgbot/internal/db"
	"tgbot/internal/tgtask"
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
}

type DomainEntry struct {
	Remark string
	URL    string
}

// ParseDomainEntries 解析域名数据
func ParseDomainEntries(text string) ([]DomainEntry, error) {
	var entries []DomainEntry
	var currentRemark string

	lines := strings.Split(text, "\n")
	urlPattern := regexp.MustCompile(`^https?://`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.Contains(line, "=========================") {
			currentRemark = ""
			continue
		}

		// 检查这一行是否是 URL（以 http:// 或 https:// 开头）
		cleanLine := strings.Trim(line, "`")
		if urlPattern.MatchString(cleanLine) {
			if currentRemark != "" {
				entries = append(entries, DomainEntry{
					Remark: currentRemark,
					URL:    cleanLine,
				})
				currentRemark = ""
			}
		} else if strings.Contains(line, ":") || strings.Contains(line, "：") {
			// 这一行可能是备注行（支持英文和中文冒号）
			colonIndex := strings.Index(line, ":")
			if colonIndex == -1 {
				colonIndex = strings.Index(line, "：")
			}
			if colonIndex != -1 {
				remark := strings.TrimSpace(line[:colonIndex])
				rest := strings.TrimSpace(line[colonIndex+1:])

				// 移除反引号
				rest = strings.Trim(rest, "`")

				if urlPattern.MatchString(rest) {
					// 备注和 URL 在同一行
					entries = append(entries, DomainEntry{
						Remark: remark,
						URL:    rest,
					})
					currentRemark = ""
				} else {
					// 这是备注行，URL 在下一行
					currentRemark = remark
				}
			}
		}
	}

	return entries, nil
}

func CreateMonitorsFromText(text string, gid int64) (successCount, failCount int, err error) {
	entries, err := ParseDomainEntries(text)

	if err != nil {
		return 0, 0, err
	}

	if len(entries) == 0 {
		return 0, 0, nil
	}

	successCount = len(entries)
	return successCount, failCount, nil
}

func CreateMonitorsFromTextAppend(text string, gid int64) (successCount, failCount int, err error) {
	entries, err := ParseDomainEntries(text)

	if err != nil {
		return 0, 0, err
	}

	if len(entries) == 0 {
		return 0, 0, nil
	}

	successCount = len(entries)
	return successCount, failCount, nil
}
