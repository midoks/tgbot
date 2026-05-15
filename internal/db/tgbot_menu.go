package db

import (
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"tgbot/internal/app/form"
	"tgbot/internal/model"
)

func applyTgbotPushMenuFilters(query *gorm.DB, field form.TgbotList) *gorm.DB {
	if field.Key != "" {
		query = query.Where("word LIKE ?", "%"+field.Key+"%")
	}
	return query
}

func GetTgbotPushMenuByArgs(field form.TgbotList) ([]model.TgbotPushMenu, int64, error) {
	page := field.Page.Page
	size := field.Page.Limit

	baseQuery := applyTgbotPushMenuFilters(db.Model(&model.TgbotPushMenu{}), field)

	var count int64
	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get tgbot count")
	}

	var list []model.TgbotPushMenu
	if err := baseQuery.Order(columnName("create_time") + " desc").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed get tgbot banword list")
	}
	return list, count, nil
}

func GetTgbotPushMenuByID(id int64) (*model.TgbotPushMenu, error) {
	var u model.TgbotPushMenu
	if err := db.First(&u, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get tgbot data")
	}
	return &u, nil
}

func GetTgbotPushMenuByWord(word string) (*model.TgbotPushMenu, error) {
	var u model.TgbotPushMenu
	if err := db.Where("word = ?", word).First(&u).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get tgbot banword data")
	}
	return &u, nil
}

func TgbotPushMenuTriggerStatus(id int64) error {
	var data model.TgbotPushMenu
	if err := db.First(&data, id).Error; err != nil {
		return errors.Wrapf(err, "failed get tgbot")
	}
	if err := db.Model(&model.TgbotPushMenu{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      !data.Status,
			"update_time": time.Now().Unix(),
		}).Error; err != nil {
		return err
	}
	return nil
}

func DeleteTgbotPushMenuByID(id int64) error {
	var d model.TgbotPushMenu
	return db.Where("id = ?", id).Delete(&d).Error
}

func GetActiveTgbotPushMenu() ([]model.TgbotPushMenu, error) {
	var list []model.TgbotPushMenu
	if err := db.Where("status = ?", true).Find(&list).Error; err != nil {
		return nil, errors.Wrap(err, "failed get active banwords")
	}
	return list, nil
}
