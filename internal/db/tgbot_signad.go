package db

import (
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"tgbot/internal/app/form"
	"tgbot/internal/model"
)

func applyTgbotSignadFilters(query *gorm.DB, field form.TgbotList) *gorm.DB {
	if field.Key != "" {
		query = query.Where("word LIKE ?", "%"+field.Key+"%")
	}
	return query
}

func GetTgbotSignadListByArgs(field form.TgbotList) ([]model.TgbotSignAd, int64, error) {
	page := field.Page.Page
	size := field.Page.Limit

	baseQuery := applyTgbotBanwordFilters(db.Model(&model.TgbotSignAd{}), field)

	var count int64
	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get tgbot count")
	}

	var list []model.TgbotSignAd
	if err := baseQuery.Order(columnName("create_time") + " desc").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed get tgbot signad list")
	}
	return list, count, nil
}

func GetTgbotSignadByID(id int64) (*model.TgbotSignAd, error) {
	var u model.TgbotSignAd
	if err := db.First(&u, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get tgbot signad data")
	}
	return &u, nil
}

func GetTgbotSignadByUserID(user_id int64) (*model.TgbotSignAd, error) {
	var u model.TgbotSignAd
	if err := db.Where("user_id = ?", user_id).First(&u).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get tgbot signad data")
	}
	return &u, nil
}

func TgbotSignadTriggerStatus(id int64) error {
	var data model.TgbotSignAd
	if err := db.First(&data, id).Error; err != nil {
		return errors.Wrapf(err, "failed get tgbot signad")
	}
	if err := db.Model(&model.TgbotSignAd{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      !data.Status,
			"update_time": time.Now().Unix(),
		}).Error; err != nil {
		return err
	}
	return nil
}

func DeleteTgbotSignadByID(id int64) error {
	var d model.TgbotSignAd
	return db.Where("id = ?", id).Delete(&d).Error
}

func GetActiveTgbotSignad() ([]model.TgbotSignAd, error) {
	var list []model.TgbotSignAd
	if err := db.Where("status = ?", true).Find(&list).Error; err != nil {
		return nil, errors.Wrap(err, "failed get active signad")
	}
	return list, nil
}
