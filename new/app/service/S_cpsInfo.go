package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

type CpsInfo struct {
	*BaseService[dbm.DB_CpsInfo] // 嵌入泛型基础服务
}

// NewCpsInfo 创建 CpsInfo 实例
func NewCpsInfo(c *gin.Context, db *gorm.DB) *CpsInfo {
	return &CpsInfo{
		BaseService: NewBaseService[dbm.DB_CpsInfo](c, db),
	}
}

//// 可添加特殊方法（需要时）
//func (s *CpsInfo) GetList() {
//	// 通过 s.c 访问上下文
//	// 通过 s.db 访问数据库
//}
