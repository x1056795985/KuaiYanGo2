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
	//创建会自动重新赋值info.AppId为新插入的数据AppId
	tx := s.db.Model(DB.DB_AppInfo{}).Create(&info)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *AppInfo) Delete(AppId interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := AppId.(type) {
	case int:
		tx2 = s.db.Model(DB.DB_AppInfo{}).Where("AppId = ?", k).Delete("")
	case []int:
		tx2 = s.db.Model(DB.DB_AppInfo{}).Where("AppId IN ?", k).Delete("")
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
		case 1: //AppId
			tx = tx.Where("AppId = ?", 请求.Keywords)
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
		tx = tx.Order("AppId ASC")
	case 2:
		tx = tx.Order("AppId DESC")
	}
	var 局_数组 []DB.DB_AppInfo
	tx = tx.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组)

	return 总数, 局_数组, tx.Error
}

// 查
func (s *AppInfo) Info(AppId int) (info DB.DB_AppInfo, err error) {
	tx := s.db.Model(DB.DB_AppInfo{}).Where("AppId = ?", AppId).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *AppInfo) Infos(where map[string]interface{}) (info []DB.DB_AppInfo, err error) {
	tx := s.db.Model(DB.DB_AppInfo{}).Where(where).Scan(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 改
func (s *AppInfo) Update(AppId int, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(DB.DB_AppInfo{}).Where("AppId = ?", AppId).Updates(&数据)
	return tx.RowsAffected, tx.Error
}
