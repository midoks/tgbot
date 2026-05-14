package form

type SubMenu struct {
	Number int64  `form:"number"`
	Name   string `form:"name"`
	Link   string `form:"link"`
}

type SubSettingMenu struct {
	Number int64  `form:"number"`
	Name   string `form:"name"`
	Link   string `form:"link"`
	Type   string `form:"type"`
}

type Page struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

type ID struct {
	ID int64 `form:"id"`
}

type IDs struct {
	Ids string `json:"ids" binding:"required"`
}

type DatabaseCommon struct {
	TableName string `form:"table_name"`
}

// TgbotLogPage 日志查询参数
type TgbotLogPage struct {
	Page  int    `form:"page"`
	Limit int    `form:"limit"`
	Times string `form:"times"`  // 日期范围
	Type  string `form:"type"`   // 查询类型
	Key   string `form:"key"`    // 关键词
	BotID int64  `form:"bot_id"` // 机器人ID
}
