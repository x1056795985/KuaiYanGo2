package service

import (
	"errors"
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
func (s *Ka) InfoKa(Name string) (info DB.DB_Ka, err error) {
	tx := s.db.Model(DB.DB_Ka{}).Where("Name = ?", Name).First(&info)
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

func (s *Ka) Infos(where map[string]interface{}) (info []DB.DB_Ka, err error) {
	tx := s.db.Model(DB.DB_Ka{}).Where(where).Find(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}
func (s *Ka) Update(Id int, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(DB.DB_Ka{}).Where("Id = ?", Id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *Ka) Delete(Id interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := Id.(type) {
	case int:
		tx2 = s.db.Model(DB.DB_Ka{}).Where("Id = ?", k).Delete("")
	case []int:
		tx2 = s.db.Model(DB.DB_Ka{}).Where("Id IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}
