package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

type CheckInLog struct {
	*BaseService[dbm.DB_CheckInLog] // 嵌入泛型基础服务
}

// NewcheckInLog 创建 checkInLog 实例
func NewCheckInLog(c *gin.Context, db *gorm.DB) *CheckInLog {
	return &CheckInLog{
		BaseService: NewBaseService[dbm.DB_CheckInLog](c, db),
	}
}

func (s *CheckInLog) Q取最后签到信息(appid, userid int) (dbm.DB_CheckInLog, error) {
	info := dbm.DB_CheckInLog{}
	tx := s.db.Model(dbm.DB_CheckInLog{}).Where("appId = ? and userId = ?", appid, userid).Order("createdAt desc").First(&info)

	return info, tx.Error
}
