package model

type TgbotLogs struct {
	ID         int64  `json:"id" gorm:"primaryKey"` // unique key
	BotID      int64  `json:"bot_id"`               // bot id
	Message    string `json:"message"`              // log message
	Level      string `json:"level"`                // log level: info, warn, error
	CreateTime int64  `json:"create_time"`          // create time
}

// TableName 返回分表名称（不带后缀，由分表逻辑处理）
func (TgbotLogs) TableName() string {
	return "tgbot_logs"
}
