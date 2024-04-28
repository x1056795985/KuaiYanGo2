package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	DB "server/structs/db"
)

type Ka struct {
	db *gorm.DB
	c  *gin.Context
}

// NewKa 创建 Ka 实例
func NewKa(c *gin.Context, db *gorm.DB) *Ka {
	return &Ka{
		db: db,
		c:  c,
	}
}

func (s *Ka) Info(id int) (info DB.DB_Ka, err error) {
	tx := s.db.Model(DB.DB_Ka{}).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *Ka) Info2(where map[string]interface{}) (info DB.DB_Ka, err error) {
	tx := s.db.Model(DB.DB_Ka{}).Where(where).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *Ka) Update(id int, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(DB.DB_Ka{}).Where("id = ?", id).Create(&数据)
	return tx.RowsAffected, tx.Error
}
