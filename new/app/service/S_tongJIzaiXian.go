package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

type TongJIzaiXian struct {
	*BaseService[dbm.DB_TongJiZaiXian] // 嵌入泛型基础服务
}

// NewTongJIzaiXian 创建 TongJIzaiXian 实例
func NewTongJIzaiXian(c *gin.Context, db *gorm.DB) *TongJIzaiXian {
	return &TongJIzaiXian{
		BaseService: NewBaseService[dbm.DB_TongJiZaiXian](c, db),
	}
}
