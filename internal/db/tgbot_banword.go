package db

import (
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

func DeleteTgbotBanwordByID(id int64) error {
	var d model.TgbotBanWord
	return db.Where("id = ?", id).Delete(&d).Error
}
