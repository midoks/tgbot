package db

import (
	"fmt"
	"time"

	"tgbot/internal/model"
)

// getTgbotLogTableName 根据日期获取分表名称
func getTgbotLogTableName(t time.Time) string {
	return fmt.Sprintf("tgbot_logs_%s", t.Format("20060102"))
}

// getTgbotLogTableNamesInRange 获取日期范围内所有分表名称
func getTgbotLogTableNamesInRange(start, end time.Time) []string {
	var tables []string
	current := start
	for !current.After(end) {
		tables = append(tables, getTgbotLogTableName(current))
		current = current.AddDate(0, 0, 1)
	}
	return tables
}

// createTgbotLogTable 如果表不存在则创建
func createTgbotLogTable(tableName string) error {
	// 检查表是否存在
	var exists bool
	err := db.Raw("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name=?)", tableName).Scan(&exists).Error
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// 创建表
	createSQL := fmt.Sprintf(`
	CREATE TABLE %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bot_id INTEGER NOT NULL,
		message TEXT NOT NULL,
		level TEXT DEFAULT 'info',
		create_time INTEGER NOT NULL
	)`, tableName)
	return db.Exec(createSQL).Error
}

// AddTgbotLog 添加日志到分表
func AddTgbotLog(botID int64, message, level string) error {
	now := time.Now()
	tableName := getTgbotLogTableName(now)

	// 确保表存在
	if err := createTgbotLogTable(tableName); err != nil {
		return err
	}

	log := model.TgbotLogs{
		BotID:      botID,
		Message:    message,
		Level:      level,
		CreateTime: now.Unix(),
	}

	return db.Table(tableName).Create(&log).Error
}

// DeleteTgbotLogsBeforeDate 删除指定日期之前的所有日志
func DeleteTgbotLogsBeforeDate(before time.Time) error {
	// 获取所有需要删除数据的表
	tables := getTgbotLogTableNamesInRange(time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local), before.Add(-24*time.Hour))
	for _, tableName := range tables {
		// 检查表是否存在
		var exists bool
		err := db.Raw("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name=?)", tableName).Scan(&exists).Error
		if err != nil {
			return err
		}
		if !exists {
			continue
		}
		// 删除表中所有数据（或者直接删除表）
		if err := db.Exec(fmt.Sprintf("DELETE FROM %s", tableName)).Error; err != nil {
			return err
		}
	}
	return nil
}

// DeleteTgbotLogsByBotID 删除指定 Bot 的所有日志（遍历所有表）
func DeleteTgbotLogsByBotID(botID int64) error {
	// 获取数据库中所有 tgbot_logs_* 表
	var tableNames []string
	err := db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'tgbot_logs_%'").Scan(&tableNames).Error
	if err != nil {
		return err
	}

	for _, tableName := range tableNames {
		if err := db.Table(tableName).Where("bot_id = ?", botID).Delete(&model.TgbotLogs{}).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetTgbotLogsByBotIDAndDateRange 根据 BotID 和日期范围查询日志
func GetTgbotLogsByBotIDAndDateRange(botID int64, start, end time.Time) ([]model.TgbotLogs, error) {
	var allLogs []model.TgbotLogs
	tables := getTgbotLogTableNamesInRange(start, end)

	for _, tableName := range tables {
		var exists bool
		err := db.Raw("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name=?)", tableName).Scan(&exists).Error
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}

		var logs []model.TgbotLogs
		err = db.Table(tableName).
			Where("bot_id = ?", botID).
			Where("create_time >= ? AND create_time <= ?", start.Unix(), end.Unix()).
			Order("create_time DESC").
			Find(&logs).Error
		if err != nil {
			return nil, err
		}
		allLogs = append(allLogs, logs...)
	}

	return allLogs, nil
}
