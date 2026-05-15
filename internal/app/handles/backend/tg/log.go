package tg

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"tgbot/internal/app/common"
	"tgbot/internal/app/form"
	"tgbot/internal/db"
	"tgbot/internal/model"
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

func LogSignad(c *gin.Context) {
	var field form.TgbotSignAd
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	common_data := &model.TgbotSignAd{
		UserID:       field.UserID,
		FromUserName: field.FromUserName,
		Status:       true,
		CreateTime:   time.Now().Unix(),
	}

	if field.ID > 0 {
		_, err := db.GetTgbotSignadByID(field.ID)
		if err == nil {
			common_data.UpdateTime = time.Now().Unix()
			if err := db.GetDb().Model(&model.TgbotSignAd{}).Where("id = ?", field.ID).Updates(common_data).Error; err != nil {
				common.ErrorResp(c, err, -1)
				return
			}
		}
	} else {
		_, err := db.GetTgbotSignadByUserID(field.UserID)
		if err == nil {
			common.ErrorResp(c, errors.New("已经存在!"), -2)
		}
		if err := db.GetDb().Create(common_data).Error; err != nil {
			common.ErrorResp(c, err, -1)
			return
		}
	}

	common.SuccessResp(c)
}
