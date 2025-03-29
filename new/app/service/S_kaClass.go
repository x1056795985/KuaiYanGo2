package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	DB "server/structs/db"
)

type KaClass struct {
	db *gorm.DB
	c  *gin.Context
}

// NewKaClass 创建 KaClass 实例
func NewKaClass(c *gin.Context, db *gorm.DB) *KaClass {
	return &KaClass{
		db: db,
		c:  c,
	}
}

// 增
func (s *KaClass) Create(请求 *DB.DB_KaClass) (row int64, err error) {

	if 请求.Id > 0 {
		return 0, errors.New("添加用户不能有id值")
	}
	if 请求.AppId < 10000 {
		return 0, errors.New("AppId错误")
	}
	if 请求.Name == "" {
		return 0, errors.New("卡类名称不能为空")
	}

	if 请求.KaLength-len(请求.Prefix) < 10 {
		return 0, errors.New(`制卡可随机字符长度小于10,请增加卡长度或减少前缀长度`)
	}

	if 请求.VipNumber < 0 || 请求.VipTime < 0 || 请求.InviteCount < 0 || 请求.Num < 0 {
		return 0, errors.New(`时间点数积分次数值不能为为负数`)
	}

	if 请求.Money < -1 || 请求.AgentMoney < -1 {
		return 0, errors.New(`售价值不能为小于-1`)
	}
	tx := s.db.Model(DB.DB_KaClass{}).Create(请求)
	return tx.RowsAffected, tx.Error
}
func (s *KaClass) Info(id int) (info DB.DB_KaClass, err error) {
	tx := s.db.Model(DB.DB_KaClass{}).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *KaClass) Update(Id int, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(DB.DB_KaClass{}).Where("Id = ?", Id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}
