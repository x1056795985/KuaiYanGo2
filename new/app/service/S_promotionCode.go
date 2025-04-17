package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
)

type PromotionCode struct {
	db *gorm.DB
	c  *gin.Context
}

// NewPromotionCode 创建 PromotionCode 实例
func NewPromotionCode(c *gin.Context, db *gorm.DB) *PromotionCode {
	return &PromotionCode{
		db: db,
		c:  c,
	}
}

// 增
func (s *PromotionCode) Create(info dbm.DB_PromotionCode) (row int64, err error) {
	tx := s.db.Model(dbm.DB_PromotionCode{}).Create(&info)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *PromotionCode) Delete(Id interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := Id.(type) {
	case int:
		tx2 = s.db.Model(dbm.DB_PromotionCode{}).Where("Id = ?", k).Delete("")
	case []int:
		tx2 = s.db.Model(dbm.DB_PromotionCode{}).Where("Id IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *PromotionCode) GetList(请求 request.List, Status int) (int64, []dbm.DB_PromotionCode, error) {
	tx := s.db
	if Status > 0 {
		tx = tx.Where("Status = ?", Status)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			tx = tx.Where("Id = ?", 请求.Keywords)
		case 2: //任务名称
			tx = tx.Where("Name LIKE ? ", "%"+请求.Keywords+"%")
		}
	}
	var 总数 int64
	//Count(&总数) 必须放在where 后面 不然值会被清0
	if 请求.Count > 500000 { //如果数据大于50万 直接使用,不重新查询了 优化速度
		总数 = 请求.Count
	} else {
		tx.Count(&总数)
	}
	//处理排序
	switch 请求.Order {
	default:
		tx = tx.Order("Id ASC")
	case 2:
		tx = tx.Order("Id DESC")
	}
	var 局_数组 []dbm.DB_PromotionCode
	tx = tx.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组)

	return 总数, 局_数组, tx.Error
}

// 查
func (s *PromotionCode) Info(id int) (info dbm.DB_PromotionCode, err error) {
	tx := s.db.Model(dbm.DB_PromotionCode{}).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *PromotionCode) Info2(where map[string]interface{}) (info dbm.DB_PromotionCode, err error) {
	tx := s.db.Model(dbm.DB_PromotionCode{}).Where(where).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 改
func (s *PromotionCode) Update(id int, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(dbm.DB_PromotionCode{}).Where("Id = ?", id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// 保存
func (s *PromotionCode) Save(info dbm.DB_PromotionCode) (row int64, err error) {
	tx := s.db.Model(dbm.DB_PromotionCode{}).Where("Id = ?", info.Id).Save(&info)
	return tx.RowsAffected, tx.Error
}
