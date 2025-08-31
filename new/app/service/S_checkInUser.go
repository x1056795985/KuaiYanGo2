package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

type CheckInUser struct {
	*BaseService[dbm.DB_CheckInUser] // 嵌入泛型基础服务
}

// NewcheckInUser 创建 checkInUser 实例
func NewCheckInUser(c *gin.Context, db *gorm.DB) *CheckInUser {
	return &CheckInUser{
		BaseService: NewBaseService[dbm.DB_CheckInUser](c, db),
	}
}

// 查
func (s *CheckInUser) Info(appId, id int) (info dbm.DB_CheckInUser, err error) {
	tx := s.db.Model(new(dbm.DB_CheckInUser)).Where("userId = ?", id).Where("appId = ?", appId).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}
