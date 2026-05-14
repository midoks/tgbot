package op

import (
	"fmt"
	"time"

	"tgbot/internal/db"
)

// InitCleanTask 初始化清理任务
func InitCleanTask() {
	// 每天凌晨 0 点执行清理
	go func() {
		for {
			// 计算到下一次凌晨 0 点的时间
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			duration := next.Sub(now)
			// 等待到凌晨 0 点
			time.Sleep(duration)

			// 执行清理系统审计日志
			if err := CleanSysLogs(); err != nil {
				fmt.Printf("[%s] 清理过期系统审计日志失败: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
				SysLog(fmt.Sprintf("清理过期系统审计日志失败: %v", err))
			}
		}
	}()
}

// 清理过期的系统审计日志
func CleanSysLogs() error {
	setting, err := db.GetSysSettingByCode(db.SettingLog)
	if err != nil {
		return fmt.Errorf("获取系统审计日志失败: %v", err)
	}

	// 解析配置
	logConf, err := setting.GetLogValue()
	if err != nil {
		return fmt.Errorf("解析系统审计配置失败: %v", err)
	}

	// 默认 30 天
	days := logConf.SaveDay
	if days <= 0 {
		days = 30
	}

	// 执行清理
	fmt.Printf("[%s] 开始清理 %d 天之前的系统审计日志\n", time.Now().Format("2006-01-02 15:04:05"), days)
	if err := db.LogDeleteBeforeDays(int(days)); err != nil {
		return fmt.Errorf("删除过期系统审计日志失败: %v", err)
	}

	fmt.Printf("[%s] 清理 %d 天之前的系统审计日志完成\n", time.Now().Format("2006-01-02 15:04:05"), days)
	return nil
}
