package form

type TgbotPushMenuAdd struct {
	ID     int64  `form:"id"`
	Name   string `form:"name"`
	Params string `form:"params"`
	Status bool   `form:"status"`
}
