package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/new/app/models/request"
	DB "server/structs/db"
)

type AppInfo struct {
	db *gorm.DB
	c  *gin.Context
}

// NewAppInfo 创建 AppInfo 实例
func NewAppInfo(c *gin.Context, db *gorm.DB) *AppInfo {
	return &AppInfo{
		db: db,
		c:  c,
	}
}

// 增
func (s *AppInfo) Create(info DB.DB_AppInfo) (row int64, err error) {
	//创建会自动重新赋值info.Id为新插入的数据id
	tx := s.db.Model(DB.DB_AppInfo{}).Create(&info)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *AppInfo) Delete(Id interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := Id.(type) {
	case int:
		tx2 = s.db.Model(DB.DB_AppInfo{}).Where("Id = ?", k).Delete("")
	case []int:
		tx2 = s.db.Model(DB.DB_AppInfo{}).Where("Id IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *AppInfo) GetList(请求 request.List, Status int) (int64, []DB.DB_AppInfo, error) {
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
	var 局_数组 []DB.DB_AppInfo
	tx = tx.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组)

	return 总数, 局_数组, tx.Error
}

// 查
func (s *AppInfo) Info(id int) (info DB.DB_AppInfo, err error) {
	tx := s.db.Model(DB.DB_AppInfo{}).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 改
func (s *AppInfo) Update(id int, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(DB.DB_AppInfo{}).Where("id = ?", id).Create(&数据)
	return tx.RowsAffected, tx.Error
}
