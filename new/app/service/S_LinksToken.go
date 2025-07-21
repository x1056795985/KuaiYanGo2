package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/Service/Ser_LinkUser"
	"server/new/app/models/request"
	DB "server/structs/db"
	"time"
)

// 在线列表 数据库处理
type LinksToken struct {
	db *gorm.DB
	c  *gin.Context
}

// NewLinksToken 创建 LinksToken 实例
func NewLinksToken(c *gin.Context, db *gorm.DB) *LinksToken {
	return &LinksToken{
		db: db,
		c:  c,
	}
}

// DeleteExpiredTokens 删除已过期的 token
func (s *LinksToken) S删除已过期的Token() error {
	// 删除已注销并 6 小时没活动的 token
	tx := s.db.Model(DB.DB_LinksToken{}).Where("Status = 2").Where("LastTime < ?", time.Now().Unix()-21600).Delete("")
	return tx.Error
}

// RevokeExpiredTokens 定时注销已过期的 token
func (s *LinksToken) Z注销已过期的Token() error {
	// 注销超时的 token
	tx := s.db.Model(DB.DB_LinksToken{}).Where("Status = 1").Where("LastTime + OutTime < ?", time.Now().Unix()).Updates(map[string]interface{}{"Status": 2, "LogoutCode": Ser_LinkUser.Z注销_心跳超时自动注销})
	return tx.Error
}

// 增
func (s *LinksToken) Create(info DB.DB_LinksToken) (row int64, err error) {
	//创建会自动重新赋值info.Id为新插入的数据id
	tx := s.db.Model(DB.DB_LinksToken{}).Create(&info)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *LinksToken) Delete(Id interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := Id.(type) {
	case int:
		tx2 = s.db.Model(DB.DB_LinksToken{}).Where("Id = ?", k).Delete("")
	case []int:
		tx2 = s.db.Model(DB.DB_LinksToken{}).Where("Id IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *LinksToken) GetList(请求 request.List, Status int) (int64, []DB.DB_LinksToken, error) {
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
	var 局_数组 []DB.DB_LinksToken
	tx = tx.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组)

	return 总数, 局_数组, tx.Error
}

// 查
func (s *LinksToken) Info(id int) (info DB.DB_LinksToken, err error) {
	tx := s.db.Model(DB.DB_LinksToken{}).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *LinksToken) Infos(where map[string]interface{}) (info []DB.DB_LinksToken, err error) {
	tx := s.db.Model(DB.DB_LinksToken{}).Where(where).Find(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 改
func (s *LinksToken) Update(id int, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(DB.DB_LinksToken{}).Where("Id = ?", id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// 可指定AppId,0为全部注销
func (s *LinksToken) Set批量注销Uid数组(UId []int, AppId int, 注销原因 int) (err error) {
	db := s.db.Model(DB.DB_LinksToken{}).Where("UId IN ? ", UId)
	if AppId != 0 {
		db.Where("LoginAppid =? ", AppId)
	}
	err = db.Updates(map[string]interface{}{"OutTime": 0, "Status": 2, "LogoutCode": 注销原因}).Error
	return
}
