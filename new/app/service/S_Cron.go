package service

import (
	"errors"
	"gorm.io/gorm"
	"server/global"
	"server/new/app/models/db"
	"server/new/app/models/request"
	"server/utils"
)

type S_Cron struct {
}

// NewCronService 创建 NewCronService 实例
func NewCronService(db *gorm.DB) *S_Cron {
	return &S_Cron{}
}

func (s *S_Cron) Info(tx *gorm.DB, Id int) (db.DB_Cron, error) {
	var value db.DB_Cron
	err := tx.Model(db.DB_Cron{}).Where("Id =?", Id).First(&value).Error
	return value, err
}

func (s *S_Cron) Update(tx *gorm.DB, value db.DB_Cron) error {
	err := tx.Model(db.DB_Cron{}).Where("Id = ?", value.Id).Updates(&value).Error
	if err != nil {

	}
	return err
}
func (s *S_Cron) Create(tx *gorm.DB, value db.DB_Cron) error {
	err := tx.Model(db.DB_Cron{}).Create(&value).Error
	return err
}

// 删除 支持 数组,和id
func (s *S_Cron) Delete(tx *gorm.DB, Id interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := Id.(type) {
	case int:
		tx2 = tx.Model(db.DB_Cron{}).Where("Id = ?", k).Delete("")
	case []int:
		tx2 = tx.Model(db.DB_Cron{}).Where("Id IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *S_Cron) GetList(tx *gorm.DB, 请求 request.List, Status int) (int64, []db.DB_Cron, error) {

	局_DB := tx.Model(db.DB_Cron{})

	if Status > 0 {
		局_DB.Where("Status = ?", Status)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2: //任务名称
			局_DB.Where("Name LIKE ? ", "%"+请求.Keywords+"%")
		}
	}
	var 总数 int64
	//Count(&总数) 必须放在where 后面 不然值会被清0
	if 请求.Count > 500000 { //如果数据大于50万 直接使用,不重新查询了 优化速度
		总数 = 请求.Count
	} else {
		局_DB.Count(&总数)
	}
	//处理排序
	switch 请求.Order {
	default:
		局_DB.Order("Id ASC")
	case 2:
		局_DB.Order("Id DESC")
	}
	var 局_数组 []db.DB_Cron
	err := 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组).Error
	if err != nil {
		global.GVA_LOG.Error(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}
	return 总数, 局_数组, err
}

// 批量维护删除
func (s *S_Cron) DeleteType(tx *gorm.DB, Type int) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch Type {
	case 1: //删除全部
		tx2 = tx.Model(db.DB_Cron{}).Where("Id > 0").Delete("")
	default:
		return 0, errors.New("类型错误")
	}
	return tx2.RowsAffected, tx2.Error
}

// GetAllInfo 获取全部任务信息
func (s *S_Cron) GetAllInfo(tx *gorm.DB, status int) ([]db.DB_Cron, error) {
	var value = []db.DB_Cron{}
	var tx2 *gorm.DB
	tx2 = tx.Model(db.DB_Cron{})
	if status > 0 {
		tx2 = tx2.Where("Status = ?", status)
	}
	err := tx2.Find(&value).Error

	return value, err
}
