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

	RelateMonitorGroupID int64 `form:"relate_monitor_group_id"`

	Timeout int    `form:"timeout"` // timeout
	Status  bool   `form:"status"`  // status
	Mark    string `form:"mark"`    // mark
}

type TgbotGroupAdd struct {
	ID       int64  `form:"id"`
	Name     string `form:"name"`
	RealTime bool   `form:"real_time"`
	Status   bool   `form:"status"`
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
