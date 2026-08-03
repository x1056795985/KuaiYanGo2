package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/app/models/db"
)

type LuckyWheelInfo struct {
	*BaseService[dbm.DB_LuckyWheelInfo]
}

func NewLuckyWheelInfo(c *gin.Context, db *gorm.DB) *LuckyWheelInfo {
	return &LuckyWheelInfo{
		BaseService: NewBaseService[dbm.DB_LuckyWheelInfo](c, db),
	}
}
