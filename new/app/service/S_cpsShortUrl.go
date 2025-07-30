package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

type CpsShortUrl struct {
	*BaseService[dbm.DB_CpsShortUrl] // 嵌入泛型基础服务
}

// NewCpsShortUrl 创建 CpsShortUrl 实例
func NewCpsShortUrl(c *gin.Context, db *gorm.DB) *CpsShortUrl {
	return &CpsShortUrl{
		BaseService: NewBaseService[dbm.DB_CpsShortUrl](c, db),
	}
}

// 查
func (s *CpsShortUrl) InfoShortUrl(ShortUrl string) (info dbm.DB_CpsShortUrl, err error) {
	tx := s.db.Model(dbm.DB_CpsShortUrl{}).Where("ShortUrl = ?", ShortUrl).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}
