package tg

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"tgbot/internal/app/common"
	"tgbot/internal/app/form"
	"tgbot/internal/db"
	"tgbot/internal/model"
)

func Banword(c *gin.Context) {
	data := common.CommonVer(c)
	c.HTML(http.StatusOK, "backend/tg/banword/index.tmpl", data)
}

func BanwordAdd(c *gin.Context) {
	id := c.Query("id")
	idint, _ := strconv.ParseInt(id, 10, 64)

	data := common.CommonVer(c)
	data["id"] = id

	banword_data, err := db.GetTgbotBanwordByID(idint)
	if err == nil {
		data["Data"] = banword_data
	}
	c.HTML(http.StatusOK, "backend/tg/banword/add.tmpl", data)
}

func PostBanwordAdd(c *gin.Context) {
	var field form.TgbotBanwordAdd
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	common_data := &model.TgbotBanWord{
		Word:       field.Word,
		Status:     field.Status,
		CreateTime: time.Now().Unix(),
	}

	if field.ID > 0 {
		_, err := db.GetTgbotBanwordByID(field.ID)
		if err == nil {
			common_data.UpdateTime = time.Now().Unix()
			if err := db.GetDb().Model(&model.TgbotBanWord{}).Where("id = ?", field.ID).Updates(common_data).Error; err != nil {
				common.ErrorResp(c, err, -1)
				return
			}
		}
	} else {

		if err := db.GetDb().Create(common_data).Error; err != nil {
			common.ErrorResp(c, err, -1)
			return
		}

	}

	common.SuccessResp(c)
}

func BanwordList(c *gin.Context) {
	var field form.TgbotList
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	result, count, err := db.GetTgbotBanwordListByArgs(field)
	if err != nil {
		common.ErrorResp(c, err, -1)
		return
	}
	common.SuccessLayuiResp(c, count, "ok", result)
}

func TgbotBanwordTriggerStatus(c *gin.Context) {
	var field form.ID
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	err := db.TgbotBanwordTriggerStatus(field.ID)
	if err != nil {
		common.ErrorResp(c, err, -1)
		return
	}
	common.SuccessResp(c)
}

func BanwordDelete(c *gin.Context) {
	var field form.ID
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	err := db.DeleteTgbotBanwordByID(field.ID)
	if err != nil {
		common.ErrorResp(c, err, -1)
		return
	}
	common.SuccessResp(c)
}
