package model

type TgbotBanWord struct {
	ID         int64  `json:"id" gorm:"primaryKey"` // unique key
	Word       string `json:"word"`                 // word
	Status     bool   `json:"status"`               // status
	CreateTime int64  `json:"create_time"`          // create time
	UpdateTime int64  `json:"update_time"`          // update_time
}
