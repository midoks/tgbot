package db

import (
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"tgbot/internal/app/form"
	"tgbot/internal/model"
)

func applyTgbotBanwordFilters(query *gorm.DB, field form.TgbotList) *gorm.DB {
	if field.Key != "" {
		query = query.Where("word LIKE ?", "%"+field.Key+"%")
	}
	return query
}

func GetTgbotBanwordListByArgs(field form.TgbotList) ([]model.TgbotBanWord, int64, error) {
	page := field.Page.Page
	size := field.Page.Limit

	baseQuery := applyTgbotBanwordFilters(db.Model(&model.TgbotBanWord{}), field)

	var count int64
	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get tgbot count")
	}

	var list []model.TgbotBanWord
	if err := baseQuery.Order(columnName("create_time") + " desc").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed get tgbot banword list")
	}
	return list, count, nil
}

func GetTgbotBanwordByID(id int64) (*model.TgbotBanWord, error) {
	var u model.TgbotBanWord
	if err := db.First(&u, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get tgbot data")
	}
	return &u, nil
}

func TgbotBanwordTriggerStatus(id int64) error {
	var data model.TgbotBanWord
	if err := db.First(&data, id).Error; err != nil {
		return errors.Wrapf(err, "failed get tgbot")
	}
	if err := db.Model(&model.TgbotBanWord{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      !data.Status,
			"update_time": time.Now().Unix(),
		}).Error; err != nil {
		return err
	}
	return nil
}

func DeleteTgbotBanwordByID(id int64) error {
	var d model.TgbotBanWord
	return db.Where("id = ?", id).Delete(&d).Error
}

func GetActiveTgbotBanwords() ([]model.TgbotBanWord, error) {
	var list []model.TgbotBanWord
	if err := db.Where("status = ?", true).Find(&list).Error; err != nil {
		return nil, errors.Wrap(err, "failed get active banwords")
	}
	return list, nil
}
