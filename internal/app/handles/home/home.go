package home

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tgbot/internal/app/common"
	"tgbot/internal/db"
)

func Index(c *gin.Context) {
	data := common.FrontendCommonVer(c)

	// 获取 Bot 数量
	botCount, err := db.GetTgbotCount()
	if err != nil {
		botCount = 0
	}

	// 获取日志统计
	todayMsgCount, todayDeleteCount, last7DayCount, err := db.GetTgbotLogStats()
	if err != nil {
		todayMsgCount = 0
		todayDeleteCount = 0
		last7DayCount = 0
	}

	data["BotCount"] = botCount
	data["TodayMsgCount"] = todayMsgCount
	data["TodayDeleteCount"] = todayDeleteCount
	data["Last7DayCount"] = last7DayCount

	c.HTML(http.StatusOK, "home/index.tmpl", data)
}

func NotFound(c *gin.Context) {
	data := common.CommonVer(c)
	c.HTML(http.StatusOK, "404.tmpl", data)
}
