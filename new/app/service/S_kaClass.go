package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	DB "server/structs/db"
)

type KaClass struct {
	db *gorm.DB
	c  *gin.Context
}

// NewKaClass 创建 KaClass 实例
func NewKaClass(c *gin.Context, db *gorm.DB) *KaClass {
	return &KaClass{
		db: db,
		c:  c,
	}
}

func (s *KaClass) Info(id int) (info DB.DB_KaClass, err error) {
	tx := s.db.Model(DB.DB_KaClass{}).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *KaClass) Update(Id int, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(DB.DB_KaClass{}).Where("Id = ?", Id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}
