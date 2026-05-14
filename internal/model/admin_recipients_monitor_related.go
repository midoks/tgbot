package model

type AdminRecipientsMonitorRelated struct {
	ID          int64  `json:"id" gorm:"primaryKey"`
	RecipientID string `json:"recipient_id"`
	MonitorGid  int64  `json:"monitor_gid"`
	Status      int    `json:"status"`
	CreateTime  int64  `json:"create_time"`
	UpdateTime  int64  `json:"update_time"`
}
