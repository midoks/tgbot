package model

type Tgbot struct {
	ID            int64  `json:"id" gorm:"primaryKey"`                  // unique key
	Name          string `json:"name" gorm:"unique" binding:"required"` // name
	Token         string `json:"token"`                                 // token
	Mark          string `json:"mark"`                                  // mark
	Status        bool   `json:"status"`                                // status
	ProxyScheme   string `json:"proxy_scheme"`                          // proxy_scheme
	ProxyValue    string `json:"proxy_value"`                           // proxy_value
	ListenEnable  bool   `json:"listen_enable"`                         // listen_enable
	IsDeleted     int    `json:"is_deleted" gorm:"default:0"`           // is_deleted
	MenuFreq      int64  `json:"menu_freq"`                             // menu_freq
	MenuRelatedID int64  `json:"menu_related_id"`                       // menu_related_id
	UpdateTime    int64  `json:"update_time"`                           // update_time
	CreateTime    int64  `json:"create_time"`                           // create_time
}
