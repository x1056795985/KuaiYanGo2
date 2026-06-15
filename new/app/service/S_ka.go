package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	DB "server/structs/db"
)

const batchSize = 5000

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
	switch k := Id.(type) {
	case int:
		tx := s.db.Model(DB.DB_Ka{}).Where("Id = ?", k).Delete("")
		return tx.RowsAffected, tx.Error
	case []int:
		var total int64
		for i := 0; i < len(k); i += batchSize {
			end := i + batchSize
			if end > len(k) {
				end = len(k)
			}
			tx := s.db.Model(DB.DB_Ka{}).Where("Id IN ?", k[i:end]).Delete("")
			if tx.Error != nil {
				return total, tx.Error
			}
			total += tx.RowsAffected
		}
		return total, nil
	default:
		return 0, errors.New("错误的数据")
	}
}
