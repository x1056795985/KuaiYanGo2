package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

type LogUserActive struct {
	*BaseService[dbm.DB_LogUserActive] // 嵌入泛型基础服务
}

// NewUserActive 创建 UserActive 实例
func NewLogUserActive(c *gin.Context, db *gorm.DB) *LogUserActive {
	return &LogUserActive{
		BaseService: NewBaseService[dbm.DB_LogUserActive](c, db),
	}
}
