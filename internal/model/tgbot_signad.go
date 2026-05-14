package model

type TgbotSignAd struct {
	ID           int64  `json:"id" gorm:"primaryKey"` // unique key
	UserID       int64  `json:"user_id"`              // user id
	FromUserName string `json:"from_user_name"`       // from user name
	Status       bool   `json:"status"`               // status
	CreateTime   int64  `json:"create_time"`          // create time
	UpdateTime   int64  `json:"update_time"`          // update_time
}
