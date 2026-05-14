package tg

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tgbot/internal/app/common"
	"tgbot/internal/app/form"
	"tgbot/internal/db"
)

func Log(c *gin.Context) {
	data := common.CommonVer(c)

	c.HTML(http.StatusOK, "backend/tg/log/index.tmpl", data)
}

func LogList(c *gin.Context) {
	var field form.TgbotLogPage
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	result, count, err := db.GetTgbotLogListByArgs(field)
	if err != nil {
		common.ErrorResp(c, err, -1)
		return
	}
	common.SuccessLayuiResp(c, count, "ok", result)
}

func LogDelete(c *gin.Context) {
	var field form.ID
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	// 删除日志记录（需要根据实际需求实现）
	// 由于是分表，需要遍历所有表删除
	err := db.DeleteTgbotLogsByBotID(field.ID)
	if err != nil {
		common.ErrorResp(c, err, -1)
		return
	}
	common.SuccessResp(c)
}
