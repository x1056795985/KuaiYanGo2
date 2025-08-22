package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

type CpsUser struct {
	*BaseService[dbm.DB_CpsUser] // 嵌入泛型基础服务
}

// NewcpsUser 创建 cpsUser 实例
func NewCpsUser(c *gin.Context, db *gorm.DB) *CpsUser {
	return &CpsUser{
		BaseService: NewBaseService[dbm.DB_CpsUser](c, db),
	}
}
