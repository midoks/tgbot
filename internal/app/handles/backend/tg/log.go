package tg

import (
	"net/http"
	"strconv"

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
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	result, count, _ := db.GetTgbotList(page, limit)
	common.SuccessLayuiResp(c, count, "ok", result)
}

func LogDelete(c *gin.Context) {
	var field form.ID
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	err := db.TgbotDelete(field.ID)
	if err != nil {
		common.ErrorResp(c, err, -1)
		return
	}
	common.SuccessResp(c)
}
