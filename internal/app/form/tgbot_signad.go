package form

type TgbotSignAd struct {
	ID           int64  `form:"id"`
	UserID       int64  `form:"user_id"`
	FromUserName string `form:"from_user_name"`
	Status       bool   `form:"status"` // status
}
