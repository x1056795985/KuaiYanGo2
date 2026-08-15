package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/app/models/dbm"
)

type LuckyWheelInfo struct {
	*BaseService[dbm.DB_LuckyWheelInfo]
}

func NewLuckyWheelInfo(c *gin.Context, db *gorm.DB) *LuckyWheelInfo {
	return &LuckyWheelInfo{
		BaseService: NewBaseService[dbm.DB_LuckyWheelInfo](c, db),
	}
}
