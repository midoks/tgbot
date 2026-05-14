package model

type TgbotLogs struct {
	ID           int64  `json:"id" gorm:"primaryKey"` // unique key
	BotID        int64  `json:"bot_id"`               // bot id
	ChatID       int64  `json:"chat_id"`              // chat id
	ChatName     string `json:"chat_name"`            // chat name
	ChatType     string `json:"chat_type"`            // chat type: private, group, supergroup, channel
	UserID       int64  `json:"user_id"`              // user id
	FromUserName string `json:"from_user_name"`       // from user name
	MessageType  string `json:"message_type"`         // message type: text, photo, document, etc.
	Content      string `json:"content"`              // message content
	Op           string `json:"op"`                   // op: 0,正常;1:删除,2:封禁
	Level        string `json:"level"`                // log level: info, warn, error
	CreateTime   int64  `json:"create_time"`          // create time
}

// TableName 返回分表名称（不带后缀，由分表逻辑处理）
func (TgbotLogs) TableName() string {
	return "tgbot_logs"
}
