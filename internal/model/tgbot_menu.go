package model

type TgbotPushMenu struct {
	ID         int64  `json:"id" gorm:"primaryKey"` // unique key
	Name       string `json:"name"`                 // from user name
	Params     string `json:"params"`               // param
	Status     bool   `json:"status"`               // status
	CreateTime int64  `json:"create_time"`          // create time
	UpdateTime int64  `json:"update_time"`          // update_time
}
