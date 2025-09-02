package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

type CheckInTaskLog struct {
	*BaseService[dbm.DB_CheckInTaskLog] // 嵌入泛型基础服务
}

// NewcheckInTaskLog 创建 checkInTaskLog 实例
func NewCheckInTaskLog(c *gin.Context, db *gorm.DB) *CheckInTaskLog {
	return &CheckInTaskLog{
		BaseService: NewBaseService[dbm.DB_CheckInTaskLog](c, db),
	}
}

// 查
func (s *CheckInTaskLog) Info(appId, id, task int, day string) (info dbm.DB_CheckInTaskLog, err error) {
	tx := s.db.Model(new(dbm.DB_CheckInTaskLog)).
		Where("userId = ?", id).
		Where("appId = ?", appId).
		Where("task = ?", task).
		Where("day = ?", day).
		First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}
