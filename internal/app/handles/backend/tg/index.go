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

func Home(c *gin.Context) {
	data := common.CommonVer(c)
	c.HTML(http.StatusOK, "backend/tg/index.tmpl", data)
}

func Add(c *gin.Context) {
	data := common.CommonVer(c)
	data["id"] = c.Query("id")
	if data["id"] != "" {
		qid, err := strconv.ParseInt(data["id"].(string), 10, 64)
		if err == nil {
			tgbot_data, err := db.GetTgbotByID(qid)
			if err == nil {
				data["Data"] = tgbot_data
			}
		}
	}
	c.HTML(http.StatusOK, "backend/tg/add.tmpl", data)
}

func List(c *gin.Context) {
	var field form.TgbotList
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	result, count, err := db.GetTgbotListByArgs(field)
	if err != nil {
		common.ErrorResp(c, err, -1)
		return
	}
	common.SuccessLayuiResp(c, count, "ok", result)
}

func PostAdd(c *gin.Context) {
	var field form.TgbotAdd
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	common_data := &model.Tgbot{
		Name:         field.Name,
		Mark:         field.Mark,
		Token:        field.Token,
		ProxyScheme:  field.ProxyScheme,
		ProxyValue:   field.ProxyValue,
		ListenEnable: field.ListenEnable,
		Status:       field.Status,
		CreateTime:   time.Now().Unix(),
	}

	if field.ID != 0 {
		_, err := db.GetTgbotByID(field.ID)
		if err == nil {
			common_data.UpdateTime = time.Now().Unix()
			if err := db.GetDb().Model(&model.Tgbot{}).Where("id = ?", field.ID).Updates(common_data).Error; err != nil {
				common.ErrorResp(c, err, -1)
				return
			}
		}
	} else {
		delete_id, err := db.GetTgbotDeletedID()
		if err == nil {
			field.ID = delete_id
			if err := db.GetDb().Model(&model.Tgbot{}).Where("id = ?", field.ID).Update("is_deleted", 0).Error; err != nil {
				common.ErrorResp(c, err, -1)
				return
			}
			if err := db.GetDb().Model(&model.Tgbot{}).Where("id = ?", field.ID).Updates(common_data).Error; err != nil {
				common.ErrorResp(c, err, -1)
				return
			}
			common_data.ID = field.ID
			common_data.IsDeleted = 0
		} else {
			if err := db.GetDb().Create(common_data).Error; err != nil {
				common.ErrorResp(c, err, -1)
				return
			}
		}
	}
	common.SuccessResp(c)
}

func PostSoftDelete(c *gin.Context) {
	var field form.ID
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	err := db.TgbotSoftDeleteByID(field.ID)
	if err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	var data model.Tgbot
	if err := db.GetDb().First(&data, field.ID).Error; err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	common.SuccessResp(c)
}

func TgbotTriggerStatus(c *gin.Context) {
	var field form.ID
	if err := c.ShouldBind(&field); err != nil {
		common.ErrorResp(c, err, -1)
		return
	}

	err := db.TgbotTriggerStatus(field.ID)
	if err == nil {
		common.SuccessResp(c)
		return
	}
	common.ErrorResp(c, err, -1)
}

func Details(c *gin.Context) {
	id := c.Query("id")
	idint, _ := strconv.ParseInt(id, 10, 64)
	tgbot_data, _ := db.GetTgbotByID(idint)

	data := common.CommonVer(c)
	data["id"] = id
	data["Data"] = tgbot_data
	c.HTML(http.StatusOK, "backend/tg/details.tmpl", data)
}
