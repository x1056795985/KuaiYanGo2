package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	DB "server/structs/db"
)

type PublicData struct {
	db *gorm.DB
	c  *gin.Context
}

// NewPublicData 创建 PublicData 实例
func NewPublicData(c *gin.Context, db *gorm.DB) *PublicData {
	return &PublicData{
		db: db,
		c:  c,
	}
}

// 增
func (s *PublicData) Create(info DB.DB_PublicData) (row int64, err error) {
	//创建会自动重新赋值info.Id为新插入的数据id
	tx := s.db.Model(DB.DB_PublicData{}).Create(&info)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *PublicData) Delete(Id interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := Id.(type) {
	case int:
		tx2 = s.db.Model(DB.DB_PublicData{}).Where("Id = ?", k).Delete("")
	case []int:
		tx2 = s.db.Model(DB.DB_PublicData{}).Where("Id IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 删除 支持 数组,和id
func (s *PublicData) Delete2(where map[string]interface{}) (影响行数 int64, error error) {

	tx := s.db.Model(DB.DB_PublicData{}).Where(where).Delete("")

	return tx.RowsAffected, tx.Error
}

// 查
func (s *PublicData) Info(id int) (info DB.DB_PublicData, err error) {
	tx := s.db.Model(DB.DB_PublicData{}).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *PublicData) Info2(where map[string]interface{}) (info DB.DB_PublicData, err error) {
	tx := s.db.Model(DB.DB_PublicData{}).Where(where).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *PublicData) Infos(where map[string]interface{}) (info []DB.DB_PublicData, err error) {
	tx := s.db.Model(DB.DB_PublicData{}).Where(where).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 改
func (s *PublicData) Update(where map[string]interface{}, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(DB.DB_PublicData{}).Where(where).Updates(&数据)
	return tx.RowsAffected, tx.Error
}
