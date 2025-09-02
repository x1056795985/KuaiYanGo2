package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

type ShortUrl struct {
	*BaseService[dbm.DB_ShortUrl] // 嵌入泛型基础服务
}

// NewShortUrl 创建 ShortUrl 实例
func NewShortUrl(c *gin.Context, db *gorm.DB) *ShortUrl {
	return &ShortUrl{
		BaseService: NewBaseService[dbm.DB_ShortUrl](c, db),
	}
}

// 查
func (s *ShortUrl) InfoShortUrl(ShortUrl string) (info dbm.DB_ShortUrl, err error) {
	tx := s.db.Model(dbm.DB_ShortUrl{}).Where("ShortUrl = ?", ShortUrl).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 点击计数+1
func (s *ShortUrl) ClickCountUP(Id int, number int) (err error) {
	tx := s.db.Model(dbm.DB_ShortUrl{}).Where("id = ?", Id).Update("clickCount", gorm.Expr("clickCount + ?", number))
	err = tx.Error
	return
}
