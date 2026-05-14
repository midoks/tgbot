package model

type AdminRecipientsTasks struct {
	ID           int64  `json:"id" gorm:"primaryKey"`
	RecipientID  int64  `json:"recipient_id"`
	MediaID      int64  `json:"media_id"`
	Name         string `json:"name"`
	Period       string `json:"period"`
	LastSendTime int64  `json:"last_send_time"`
	Status       int    `json:"status"`
	IsDeleted    int    `json:"is_deleted" gorm:"default:0"`
	CreateTime   int64  `json:"create_time"`
	UpdateTime   int64  `json:"update_time"`
}
