package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

type CpsCode struct {
	*BaseService[dbm.DB_CpsCode] // 嵌入泛型基础服务
}

// NewcpsCode 创建 cpsCode 实例
func NewCpsCode(c *gin.Context, db *gorm.DB) *CpsCode {
	return &CpsCode{
		BaseService: NewBaseService[dbm.DB_CpsCode](c, db),
	}
}

// 查
func (s *CpsCode) InfoCode(Code string) (info dbm.DB_CpsCode, err error) {
	tx := s.db.Model(new(dbm.DB_CpsCode)).Where("cpsCode = ?", Code).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *CpsCode) InfoUserId(UserId int) (info dbm.DB_CpsCode, err error) {
	tx := s.db.Model(new(dbm.DB_CpsCode)).Where("userId = ?", UserId).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}
