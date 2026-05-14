package form

type TgbotBanwordAdd struct {
	ID     int64  `form:"id"`
	Word   string `form:"word"`
	Status bool   `form:"status"` // status
}
