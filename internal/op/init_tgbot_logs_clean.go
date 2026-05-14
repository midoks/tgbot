package op

import (
	"fmt"
	"time"

	"tgbot/internal/db"
)

// InitTgbotLogsCleanTask 初始化 Telegram 日志清理定时任务
func InitTgbotLogsCleanTask() {
	// 立即执行一次清理
	cleanTgbotLogs()

	// 设置定时任务：每天凌晨 3 点执行
	go func() {
		for {
			// 计算距离下一个凌晨3点的时间
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
			duration := next.Sub(now)

			// 等待到下一个执行时间
			time.Sleep(duration)

			// 执行清理
			cleanTgbotLogs()
		}
	}()
}

// cleanTgbotLogs 清理30天之前的日志
func cleanTgbotLogs() {
	before := time.Now().AddDate(0, 0, -30)
	fmt.Printf("Cleaning tgbot logs before %s...\n", before.Format("2006-01-02 15:04:05"))

	if err := db.DeleteTgbotLogsBeforeDate(before); err != nil {
		fmt.Printf("Failed to clean tgbot logs: %v\n", err)
	} else {
		fmt.Println("Tgbot logs cleaned successfully")
	}
}
