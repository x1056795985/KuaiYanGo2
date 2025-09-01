package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

type CheckInInfo struct {
	*BaseService[dbm.DB_CheckInInfo] // 嵌入泛型基础服务
}

// NewCpsInfo 创建 CpsInfo 实例
func NewCheckInInfo(c *gin.Context, db *gorm.DB) *CheckInInfo {
	return &CheckInInfo{
		BaseService: NewBaseService[dbm.DB_CheckInInfo](c, db),
	}
}

//// 可添加特殊方法（需要时）
//func (s *CpsInfo) GetList() {
//	// 通过 s.c 访问上下文
//	// 通过 s.db 访问数据库
//}
