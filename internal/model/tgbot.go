package model

type Tgbot struct {
	ID           int64  `json:"id" gorm:"primaryKey"`                  // unique key
	Name         string `json:"name" gorm:"unique" binding:"required"` // name
	Token        string `json:"token"`                                 // token
	Params       string `json:"params"`                                // params
	Rate         string `json:"rate"`                                  // rate
	Mark         string `json:"mark"`                                  // mark
	IsOn         string `json:"is_on"`                                 // is_on
	Status       bool   `json:"status"`                                // status
	ProxyScheme  string `json:"proxy_scheme"`                          // proxy_scheme
	ProxyValue   string `json:"proxy_value"`                           // proxy_value
	ListenEnable bool   `json:"listen_enable"`                         // listen_enable
	IsDeleted    int    `json:"is_deleted" gorm:"default:0"`           // is_deleted
	UpdateTime   int64  `json:"update_time"`                           // update_time
	CreateTime   int64  `json:"create_time"`                           // create_time
}

type TgBotCommonParams struct {
	SendID                 string `json:"send_id"`
	TelegramListenStrategy string `json:"telegram_listen_strategy"`
	RelateMonitorGroupID   int64  `form:"relate_monitor_group_id"`
}
