package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/app/models/db"
)

type LuckyWheelUser struct {
	*BaseService[dbm.DB_LuckyWheelUser]
}

func NewLuckyWheelUser(c *gin.Context, db *gorm.DB) *LuckyWheelUser {
	return &LuckyWheelUser{
		BaseService: NewBaseService[dbm.DB_LuckyWheelUser](c, db),
	}
}

// Info 查询用户大转盘数据
func (s *LuckyWheelUser) Info(appId, userId int) (info dbm.DB_LuckyWheelUser, err error) {
	tx := s.db.Model(new(dbm.DB_LuckyWheelUser)).Where("userId = ?", userId).Where("appId = ?", appId).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}
