package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/app/models/dbm"
	"server/app/models/request"
)

// 网关转发 数据库处理
type Gateway struct {
	db *gorm.DB
	c  *gin.Context
}

// NewGateway 创建 Gateway 实例
func NewGateway(c *gin.Context, db *gorm.DB) *Gateway {
	return &Gateway{
		db: db,
		c:  c,
	}
}

// GetList 分页列表,Type 1名称 2Url 3备注
func (j *Gateway) GetList(请求 request.List) (int64, []dbm.DB_Gateway, error) {
	db := j.db.Model(dbm.DB_Gateway{})
	if 请求.Keywords != "" && 请求.Type > 0 {
		switch 请求.Type {
		case 1:
			db = db.Where("Name like ?", "%"+请求.Keywords+"%")
		case 2:
			db = db.Where("Url like ?", "%"+请求.Keywords+"%")
		case 3:
			db = db.Where("Remark like ?", "%"+请求.Keywords+"%")
		}
	}
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, nil, err
	}
	order := "Id DESC"
	if 请求.Order == 1 {
		order = "Id ASC"
	}
	if 请求.Page == 0 {
		请求.Page = 1
	}
	if 请求.Size == 0 {
		请求.Size = 10
	}
	var list []dbm.DB_Gateway
	err := db.Order(order).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&list).Error
	return count, list, err
}

// Info 单条
func (j *Gateway) Info(id int) (info dbm.DB_Gateway, err error) {
	err = j.db.Model(dbm.DB_Gateway{}).Where("Id = ?", id).First(&info).Error
	return
}

// InfoName 按名称查(查重)
func (j *Gateway) InfoName(name string) (info dbm.DB_Gateway, err error) {
	err = j.db.Model(dbm.DB_Gateway{}).Where("Name = ?", name).First(&info).Error
	return
}

// Create 新增
func (j *Gateway) Create(info *dbm.DB_Gateway) (row int64, err error) {
	tx := j.db.Create(info)
	if tx.Error == nil {
		row = tx.RowsAffected
	}
	return
}

// Update 更新指定字段
func (j *Gateway) Update(id int, 数据 map[string]interface{}) (row int64, err error) {
	tx := j.db.Model(dbm.DB_Gateway{}).Where("Id = ?", id).Updates(数据)
	if tx.Error == nil {
		row = tx.RowsAffected
	}
	return
}

// Delete 批量删除
func (j *Gateway) Delete(Ids []int) (影响行数 int64, err error) {
	if len(Ids) == 0 {
		return 0, errors.New("未选择要删除的网关")
	}
	tx := j.db.Where("Id in ?", Ids).Delete(dbm.DB_Gateway{})
	return tx.RowsAffected, tx.Error
}
