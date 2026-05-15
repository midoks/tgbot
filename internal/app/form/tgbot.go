package form

type TgbotAdd struct {
	ID   int64  `form:"id"`
	Gid  int64  `form:"gid"` // gid
	Name string `form:"name"`
	Type string `form:"type"`

	Token        string `form:"token"`
	ProxyScheme  string `form:"proxy_scheme"`
	ProxyValue   string `form:"proxy_value"`
	ListenEnable bool   `form:"listen_enable"`

	MenuFreq      int64 `form:"menu_freq"`
	MenuRelatedID int64 `form:"menu_related_id"`
	MenuSendID    int64 `form:"menu_send_id"`

	Status bool   `form:"status"` // status
	Mark   string `form:"mark"`   // mark
}

type TgbotList struct {
	Page
	Key string `form:"key"`
}

type TgbotLogList struct {
	Page

	Times     string `form:"times"`
	Type      string `form:"type"`
	Key       string `form:"key"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}
