package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	db "server/structs/db"

	"server/new/app/models/request"
)

type TaskPoolData struct {
	db *gorm.DB
	c  *gin.Context
}

// NewTaskPoolData 创建 TaskPoolData 实例
func NewTaskPoolData(c *gin.Context, db *gorm.DB) *TaskPoolData {
	return &TaskPoolData{
		db: db,
		c:  c,
	}
}

// 增
func (s *TaskPoolData) Create(info db.DB_TaskPoolData) (row int64, err error) {
	tx := s.db.Model(db.DB_TaskPoolData{}).Create(&info)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *TaskPoolData) Delete(Uuid interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := Uuid.(type) {
	case int:
		tx2 = s.db.Model(db.DB_TaskPoolData{}).Where("Uuid = ?", k).Delete("")
	case []int:
		tx2 = s.db.Model(db.DB_TaskPoolData{}).Where("Uuid IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *TaskPoolData) GetList(请求 request.List, Tid, SubmitAppId, SubmitUid int) (int64, []db.DB_TaskPoolData, error) {
	tx := s.db
	if Tid > 0 {
		tx = tx.Where("Tid = ?", Tid)
	}

	if SubmitUid > 0 {
		tx = tx.Where("SubmitAppId = ?", Tid)
	}

	if SubmitAppId > 0 {
		tx = tx.Where("SubmitAppId = ?", Tid)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //SubmitData
			tx = tx.Where("SubmitData LIKE ? ", "%"+请求.Keywords+"%")
		case 2: //ReturnData
			tx = tx.Where("ReturnData LIKE ? ", "%"+请求.Keywords+"%")
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
		tx = tx.Order("TimeStart ASC")
	case 2:
		tx = tx.Order("TimeStart DESC")
	}
	var 局_数组 []db.DB_TaskPoolData
	tx = tx.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组)

	return 总数, 局_数组, tx.Error
}

// 查
func (s *TaskPoolData) Info(Uuid string) (info db.DB_TaskPoolData, err error) {
	tx := s.db.Model(db.DB_TaskPoolData{}).Where("Uuid = ?", Uuid).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *TaskPoolData) Info2(where map[string]interface{}) (info db.DB_TaskPoolData, err error) {
	tx := s.db.Model(db.DB_TaskPoolData{}).Where(where).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 改
func (s *TaskPoolData) Update(Uuid string, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(db.DB_TaskPoolData{}).Where("Uuid = ?", Uuid).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// 保存
func (s *TaskPoolData) Save(info db.DB_TaskPoolData) (row int64, err error) {
	tx := s.db.Model(db.DB_TaskPoolData{}).Where("Uuid = ?", info.Uuid).Save(&info)
	return tx.RowsAffected, tx.Error
}
