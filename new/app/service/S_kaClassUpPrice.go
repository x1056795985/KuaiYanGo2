package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/global"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
	"server/utils"
)

type KaClassUpPrice struct {
	db *gorm.DB
	c  *gin.Context
}

// NewKaClassUpPrice 创建 KaClassUpPrice 实例
func NewKaClassUpPrice(c *gin.Context, db *gorm.DB) *KaClassUpPrice {
	return &KaClassUpPrice{
		db: db,
		c:  c,
	}
}
func (s *KaClassUpPrice) Create(value *dbm.DB_KaClassUpPrice) error {
	err := s.db.Model(dbm.DB_KaClassUpPrice{}).Create(&value).Error
	return err
}
func (s *KaClassUpPrice) Info(id int) (info dbm.DB_KaClassUpPrice, err error) {
	tx := s.db.Model(dbm.DB_KaClassUpPrice{}).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *KaClassUpPrice) Info2(where map[string]interface{}) (info dbm.DB_KaClassUpPrice, err error) {
	tx := s.db.Model(dbm.DB_KaClassUpPrice{}).Where(where).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *KaClassUpPrice) Infos(where map[string]interface{}) (info []dbm.DB_KaClassUpPrice, err error) {
	tx := s.db.Model(dbm.DB_KaClassUpPrice{}).Where(where).Find(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}
func (s *KaClassUpPrice) Infos2(where map[string]interface{}) (info []dbm.DB_KaClassUpPrice, err error) {
	tx := s.db.Model(dbm.DB_KaClassUpPrice{}).Where(where).Where("Markup > 0").Find(&info)

	if tx.Error != nil {
		err = tx.Error
	}
	return
}
func (s *KaClassUpPrice) Update(Id int, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(dbm.DB_KaClassUpPrice{}).Where("Id = ?", Id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *KaClassUpPrice) Delete(Id interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := Id.(type) {
	case int:
		tx2 = s.db.Model(dbm.DB_KaClassUpPrice{}).Where("Id = ?", k).Delete("")
	case []int:
		tx2 = s.db.Model(dbm.DB_KaClassUpPrice{}).Where("Id IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *KaClassUpPrice) GetList(请求 request.List) (int64, []dbm.DB_KaClassUpPrice, error) {

	局_DB := s.db.Model(dbm.DB_KaClassUpPrice{})
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			局_DB.Where("KaClassId = ?", 请求.Keywords)
		case 2: //用户名
			局_DB.Where("AgentId = ? ", 请求.Keywords)
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
	var 局_数组 []dbm.DB_KaClassUpPrice
	err := 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组).Error
	if err != nil {
		global.GVA_LOG.Error(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}
	return 总数, 局_数组, err
}
