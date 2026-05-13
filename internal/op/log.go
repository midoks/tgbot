package op

import (
	// "fmt"
	// "time"

	// "github.com/pkg/errors"
	// "gorm.io/gorm"

	"tgbot/internal/db"
	// "tgbot/internal/model"
	// utils "tgbot/internal/utils"
)

func AddLog(uid int64, content string) error {
	return db.AddLog(nil, uid, content)
}

func SysLog(content string) error {
	return db.AddLog(nil, 0, content)
}
