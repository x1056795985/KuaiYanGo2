package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
)

type AppInfoWebUser struct {
	db *gorm.DB
	c  *gin.Context
}

// NewAppInfoWebUser 创建 AppInfoWebUser 实例
func NewAppInfoWebUser(c *gin.Context, db *gorm.DB) *AppInfoWebUser {
	return &AppInfoWebUser{
		db: db,
		c:  c,
	}
}

// 增
func (s *AppInfoWebUser) Create(info dbm.DB_AppInfoWebUser) (row int64, err error) {
	tx := s.db.Model(dbm.DB_AppInfoWebUser{}).Create(&info)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *AppInfoWebUser) Delete(id interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := id.(type) {
	case int:
		tx2 = s.db.Model(dbm.DB_AppInfoWebUser{}).Where("id = ?", k).Delete("")
	case []int:
		tx2 = s.db.Model(dbm.DB_AppInfoWebUser{}).Where("id IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *AppInfoWebUser) GetList(请求 request.List, Status int) (int64, []dbm.DB_AppInfoWebUser, error) {
	tx := s.db
	if Status > 0 {
		tx = tx.Where("Status = ?", Status)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			tx = tx.Where("id = ?", 请求.Keywords)
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
		tx = tx.Order("id ASC")
	case 2:
		tx = tx.Order("id DESC")
	}
	var 局_数组 []dbm.DB_AppInfoWebUser
	tx = tx.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组)

	return 总数, 局_数组, tx.Error
}

// 查
func (s *AppInfoWebUser) Info(id int) (info dbm.DB_AppInfoWebUser, err error) {
	tx := s.db.Model(dbm.DB_AppInfoWebUser{}).Where("id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *AppInfoWebUser) Infos(where map[string]interface{}) (info []dbm.DB_AppInfoWebUser, err error) {
	tx := s.db.Model(dbm.DB_AppInfoWebUser{}).Where(where).Scan(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 改
func (s *AppInfoWebUser) Update(id int, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(dbm.DB_AppInfoWebUser{}).Where("id = ?", id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}
