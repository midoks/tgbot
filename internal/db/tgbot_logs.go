package db

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"tgbot/internal/app/form"
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

// parseDateTime 解析多种日期时间格式
func parseDateTime(s string) (time.Time, error) {
	// 尝试多种格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
		"2006/01/02",
	}

	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse datetime: %s", s)
}

// createTgbotLogTable 如果表不存在则创建，如果存在则添加缺失的字段
func createTgbotLogTable(tableName string) error {
	// 检查表是否存在
	var exists bool
	err := db.Raw("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name=?)", tableName).Scan(&exists).Error
	if err != nil {
		return err
	}

	if !exists {
		// 创建表
		createSQL := fmt.Sprintf(`
		CREATE TABLE %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bot_id INTEGER NOT NULL,
			chat_id INTEGER NOT NULL,
			chat_name TEXT,
			chat_type TEXT,
			user_id INTEGER,
			from_user_name TEXT,
			message_type TEXT,
			content TEXT,
			op INTEGER DEFAULT '0',
			level TEXT DEFAULT 'info',
			create_time INTEGER NOT NULL
		)`, tableName)
		return db.Exec(createSQL).Error
	}

	// 表已存在，检查并添加缺失的字段
	fieldsToAdd := []string{
		"chat_id INTEGER NOT NULL DEFAULT 0",
		"chat_name TEXT",
		"chat_type TEXT",
		"user_id INTEGER DEFAULT 0",
		"from_user_name TEXT",
		"message_type TEXT",
		"content TEXT",
		"op INTEGER DEFAULT '0'",
	}

	for _, fieldDef := range fieldsToAdd {
		// 获取字段名（第一个空格之前的部分）
		fieldName := ""
		for _, c := range fieldDef {
			if c == ' ' {
				break
			}
			fieldName += string(c)
		}

		// 检查字段是否存在
		query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
		rows, err := db.Raw(query).Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		fieldExists := false
		for rows.Next() {
			var cid int
			var name string
			var typ string
			var notNull int
			var dfltValue interface{}
			var pk int
			err := rows.Scan(&cid, &name, &typ, &notNull, &dfltValue, &pk)
			if err != nil {
				return err
			}
			if name == fieldName {
				fieldExists = true
				break
			}
		}

		if !fieldExists {
			// 添加字段
			alterSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, fieldDef)
			if err := db.Exec(alterSQL).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// AddTgbotLog 添加日志到分表（完整字段版本）
func AddTgbotLog(botID, chatID, userID int64, chatName, chatType, fromUserName, messageType, content string, op int, level string) error {
	now := time.Now()
	tableName := getTgbotLogTableName(now)

	// 确保表存在
	if err := createTgbotLogTable(tableName); err != nil {
		return err
	}

	log := model.TgbotLogs{
		BotID:        botID,
		ChatID:       chatID,
		ChatName:     chatName,
		ChatType:     chatType,
		UserID:       userID,
		FromUserName: fromUserName,
		MessageType:  messageType,
		Content:      content,
		Op:           op,
		Level:        level,
		CreateTime:   now.Unix(),
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

// DeleteTgbotLogsByChatID 删除指定聊天的所有日志
func DeleteTgbotLogsByChatID(chatID int64) error {
	var tableNames []string
	err := db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'tgbot_logs_%'").Scan(&tableNames).Error
	if err != nil {
		return err
	}

	for _, tableName := range tableNames {
		if err := db.Table(tableName).Where("chat_id = ?", chatID).Delete(&model.TgbotLogs{}).Error; err != nil {
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

// GetTgbotLogListByArgs 根据条件查询日志列表（支持分页）
func GetTgbotLogListByArgs(field form.TgbotLogPage) ([]model.TgbotLogs, int64, error) {
	// 解析日期范围
	var start, end time.Time
	var err error

	if field.Times != "" {
		// 假设格式为 "2026-05-01 - 2026-05-14" 或带时间的格式
		parts := strings.Split(field.Times, " - ")
		if len(parts) == 2 {
			start, err = parseDateTime(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, 0, err
			}
			end, err = parseDateTime(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, 0, err
			}
			// 结束日期设为当天结束时间
			end = end.AddDate(0, 0, 1).Add(-time.Nanosecond)
		}
	} else {
		// 默认查询最近7天
		end = time.Now()
		start = end.AddDate(0, 0, -7)
	}

	// 获取日期范围内的所有表
	tables := getTgbotLogTableNamesInRange(start, end)

	var allLogs []model.TgbotLogs
	for _, tableName := range tables {
		var exists bool
		err := db.Raw("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name=?)", tableName).Scan(&exists).Error
		if err != nil {
			return nil, 0, err
		}
		if !exists {
			continue
		}

		var logs []model.TgbotLogs
		dbQuery := db.Table(tableName)

		// 根据查询类型和关键词过滤
		if field.Type != "" && field.Key != "" {
			switch field.Type {
			case "bot_id":
				botID, _ := strconv.ParseInt(field.Key, 10, 64)
				dbQuery = dbQuery.Where("bot_id = ?", botID)
			case "chat_id":
				chatID, _ := strconv.ParseInt(field.Key, 10, 64)
				dbQuery = dbQuery.Where("chat_id = ?", chatID)
			case "chat_name":
				dbQuery = dbQuery.Where("chat_name LIKE ?", "%"+field.Key+"%")
			case "from_user_name":
				dbQuery = dbQuery.Where("from_user_name LIKE ?", "%"+field.Key+"%")
			case "message_type":
				dbQuery = dbQuery.Where("message_type = ?", field.Key)
			case "content":
				dbQuery = dbQuery.Where("content LIKE ?", "%"+field.Key+"%")
			default:
				// 模糊搜索所有文本字段
				dbQuery = dbQuery.Where("chat_name LIKE ? OR from_user_name LIKE ? OR content LIKE ?",
					"%"+field.Key+"%", "%"+field.Key+"%", "%"+field.Key+"%")
			}
		}

		// 按BotID过滤
		if field.BotID > 0 {
			dbQuery = dbQuery.Where("bot_id = ?", field.BotID)
		}

		// 按时间范围过滤
		dbQuery = dbQuery.Where("create_time >= ? AND create_time <= ?", start.Unix(), end.Unix())

		err = dbQuery.Order("create_time DESC").Find(&logs).Error
		if err != nil {
			return nil, 0, err
		}
		allLogs = append(allLogs, logs...)
	}

	// 计算总数
	total := int64(len(allLogs))

	// 分页处理
	if field.Page > 0 && field.Limit > 0 {
		offset := (field.Page - 1) * field.Limit
		if offset < len(allLogs) {
			endIdx := offset + field.Limit
			if endIdx > len(allLogs) {
				endIdx = len(allLogs)
			}
			allLogs = allLogs[offset:endIdx]
		} else {
			allLogs = []model.TgbotLogs{}
		}
	}

	return allLogs, total, nil
}
