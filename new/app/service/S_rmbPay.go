package service

import (
	"gorm.io/gorm"
	DB "server/structs/db"
)

// 在线列表 数据库处理
type RmbPayService struct {
	db *gorm.DB
}

// NewRmbPayService 创建 RmbPayService 实例
func NewRmbPayService(db *gorm.DB) *RmbPayService {
	return &RmbPayService{
		db: db,
	}
}

func (s *RmbPayService) Create(新订单 DB.DB_LogRMBPayOrder) (DB.DB_LogRMBPayOrder, error) {

	tx := s.db.Model(DB.DB_LogRMBPayOrder{}).Create(&新订单)

	if tx.Error != nil {
		return DB.DB_LogRMBPayOrder{}, tx.Error
	}

	return 新订单, nil
}

func (s *RmbPayService) Info(id int) (info DB.DB_LogRMBPayOrder, err error) {

	tx := s.db.Model(DB.DB_LogRMBPayOrder{}).Where("Id = ?", id).First(&info)

	if tx.Error != nil {
		err = tx.Error
	}
	return
}
func (s *RmbPayService) Info2(where map[string]interface{}) (info DB.DB_LogRMBPayOrder, err error) {

	tx := s.db.Model(DB.DB_LogRMBPayOrder{}).Where(where).First(&info)

	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *RmbPayService) Update(Id int, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(DB.DB_LogRMBPayOrder{}).Where("Id = ?", Id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}
