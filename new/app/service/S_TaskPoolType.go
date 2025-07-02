package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	db "server/structs/db"

	"server/new/app/models/request"
)

type TaskPoolType struct {
	db *gorm.DB
	c  *gin.Context
}

// NewTaskPoolType 创建 TaskPoolType 实例
func NewTaskPoolType(c *gin.Context, db *gorm.DB) *TaskPoolType {
	return &TaskPoolType{
		db: db,
		c:  c,
	}
}

// 增
func (s *TaskPoolType) Create(info db.TaskPool_类型) (row int64, err error) {
	tx := s.db.Model(db.TaskPool_类型{}).Create(&info)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *TaskPoolType) Delete(id interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := id.(type) {
	case string:
		tx2 = s.db.Model(db.TaskPool_类型{}).Where("id = ?", k).Delete("")
	case []string:
		tx2 = s.db.Model(db.TaskPool_类型{}).Where("id IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *TaskPoolType) GetList(请求 request.List) (int64, []db.TaskPool_类型, error) {
	tx := s.db.Model(db.TaskPool_类型{})

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //name
			tx = tx.Where("Id = ? ", 请求.Keywords)
		case 2: //ID
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
	var 局_数组 []db.TaskPool_类型
	tx = tx.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组)

	return 总数, 局_数组, tx.Error
}

// 查
func (s *TaskPoolType) Info(Id string) (info db.TaskPool_类型, err error) {
	tx := s.db.Model(db.TaskPool_类型{}).Where("Id = ?", Id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *TaskPoolType) Info2(where map[string]interface{}) (info db.TaskPool_类型, err error) {
	tx := s.db.Model(db.TaskPool_类型{}).Where(where).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 改
func (s *TaskPoolType) Update(Id int, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(db.TaskPool_类型{}).Where("Id = ?", Id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// 保存
func (s *TaskPoolType) Save(info db.TaskPool_类型) (row int64, err error) {
	tx := s.db.Model(db.TaskPool_类型{}).Where("Id = ?", info.Id).Save(&info)
	return tx.RowsAffected, tx.Error
}
