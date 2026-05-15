package entity

import (
	"tgbot/internal/model"
)

// TgbotLogWithSignad 带广告标记的日志
type TgbotLogWithSignad struct {
	model.TgbotLogs
	IsSignad bool `json:"is_signad"`
}
