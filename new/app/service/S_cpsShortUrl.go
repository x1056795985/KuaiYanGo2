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

// 点击计数+1
func (s *CpsShortUrl) ClickCountUP(Id int, number int) (err error) {
	tx := s.db.Model(dbm.DB_CpsShortUrl{}).Where("id = ?", Id).Update("clickCount", gorm.Expr("clickCount + ?", number))
	err = tx.Error
	return
}
