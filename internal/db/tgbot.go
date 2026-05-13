package db

import (
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"tgbot/internal/app/form"
	"tgbot/internal/model"
)

func applyTgbotFilters(query *gorm.DB, field form.TgbotList) *gorm.DB {
	if field.Key != "" {
		query = query.Where("name LIKE ?", "%"+field.Key+"%").Or("mark LIKE ?", "%"+field.Key+"%")
	}
	return query
}

func GetTgbotListByArgs(field form.TgbotList) ([]model.Tgbot, int64, error) {
	page := field.Page.Page
	size := field.Page.Limit

	baseQuery := applyTgbotFilters(db.Model(&model.Tgbot{}), field)

	var count int64
	if err := baseQuery.Where("is_deleted=?", 0).Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get tgbot count")
	}

	var list []model.Tgbot
	if err := baseQuery.Order(columnName("create_time")+" desc").Where("is_deleted=?", 0).Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed get tgbot list")
	}
	return list, count, nil
}

func GetTgbotList(page, size int) ([]model.Tgbot, int64, error) {
	mm := db.Model(&model.Tgbot{})
	var count int64
	if err := mm.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get tgbot count")
	}

	var list []model.Tgbot
	if err := db.Order(columnName("id")).Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, errors.WithStack(err)
	}
	return list, count, nil
}

func TgbotDelete(id int64) error {
	return db.Delete(&model.Tgbot{}, id).Error
}

func TgbotSoftDeleteByID(id int64) error {
	if err := db.Model(&model.Tgbot{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted":  1,
			"update_time": time.Now().Unix(),
		}).Error; err != nil {
		return err
	}
	return nil
}

func TgbotDeleteByIDs(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return db.Delete(&model.Tgbot{}, ids).Error
}

func GetTgbotByID(id int64) (*model.Tgbot, error) {
	var u model.Tgbot
	if err := db.First(&u, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get tgbot data")
	}
	return &u, nil
}

func GetTgbotDeletedID() (int64, error) {
	var d model.Tgbot
	if err := db.Order(columnName("create_time")).Where("is_deleted=?", 1).First(&d).Error; err != nil {
		return 0, errors.Wrap(err, "failed get deleted tgbot")
	}
	return d.ID, nil
}

func TgbotUpdate(id int64, updates map[string]interface{}) error {
	return db.Model(&model.Tgbot{}).Where("id = ?", id).Updates(updates).Error
}

func TgbotCreate(m *model.Tgbot) error {
	return db.Create(m).Error
}

func TgbotSave(m *model.Tgbot) error {
	return db.Save(m).Error
}

func TgbotGetByName(name string) (*model.Tgbot, error) {
	var m model.Tgbot
	if err := db.Where("name = ?", name).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func TgbotGetAll() ([]model.Tgbot, error) {
	var list []model.Tgbot
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func TgbotCount() (int64, error) {
	var count int64
	if err := db.Model(&model.Tgbot{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
