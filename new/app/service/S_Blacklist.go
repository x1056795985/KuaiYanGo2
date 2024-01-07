package service

import (
	"errors"
	"gorm.io/gorm"
	"server/global"
	"server/new/app/models/db"
	"server/new/app/models/request"
	"server/utils"
	"time"
)

type S_Blacklist struct {
}

const 黑名单_ = "黑名单_"

func (s *S_Blacklist) Info(tx *gorm.DB, Id int) (db.DB_Blacklist, error) {
	var value db.DB_Blacklist
	err := tx.Model(db.DB_Blacklist{}).Where("Id =?", Id).First(&value).Error
	return value, err
}
func (s *S_Blacklist) CountAdd1(tx *gorm.DB, Id int) (db.DB_Blacklist, error) {
	var value db.DB_Blacklist
	err := tx.Model(db.DB_Blacklist{}).Where("Id =?", Id).Update("Count", gorm.Expr("Count + 1")).Error
	return value, err
}

// 读取黑名单key 高频访问,其他接口都为这个让路
func (s *S_Blacklist) InfoItemKey(tx *gorm.DB, ItemKey string) ([]db.DB_Blacklist, error) {
	if 局_临时, ok := global.H缓存.Get(黑名单_ + ItemKey); ok { //高频
		return 局_临时.([]db.DB_Blacklist), nil
	}

	var value = []db.DB_Blacklist{}
	err := tx.Model(db.DB_Blacklist{}).Where("ItemKey = ?", ItemKey).Find(&value).Error
	global.H缓存.Set(黑名单_+ItemKey, value, time.Hour*720) //保存一个月
	return value, err
}
func (s *S_Blacklist) Update(tx *gorm.DB, value db.DB_Blacklist) error {
	err := tx.Model(db.DB_Blacklist{}).Where("ItemKey = ?", value.ItemKey).Updates(&value).Error
	if _, ok := global.H缓存.Get(黑名单_ + value.ItemKey); ok {
		global.H缓存.Delete(黑名单_ + value.ItemKey)
	}
	return err
}
func (s *S_Blacklist) Create(tx *gorm.DB, value db.DB_Blacklist) error {
	if value.Time == 0 {
		value.Time = time.Now().Unix()
	}
	err := tx.Model(db.DB_Blacklist{}).Create(&value).Error
	if _, ok := global.H缓存.Get(黑名单_ + value.ItemKey); ok { //黑名单添加也需要增加缓存。因为如果有缓存的情况加,增加一个同名黑名单,就不会读取这个新的,只会读取以前缓存的记录
		global.H缓存.Delete(黑名单_ + value.ItemKey)
	}
	return err
}
func (s *S_Blacklist) Delete(tx *gorm.DB, Id interface{}) (影响行数 int64, error error) {
	var ItemKey []string
	var tx2 *gorm.DB
	switch k := Id.(type) {
	case int:
		tx.Model(db.DB_Blacklist{}).Select("ItemKey").Where("Id = ?", k).Find(&ItemKey)
		tx2 = tx.Model(db.DB_Blacklist{}).Where("Id = ?", k).Delete("")
	case []int:
		tx.Model(db.DB_Blacklist{}).Select("ItemKey").Where("Id IN ?", k).Find(&ItemKey)
		tx2 = tx.Model(db.DB_Blacklist{}).Where("Id IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	for 索引, _ := range ItemKey {
		if _, ok := global.H缓存.Get(黑名单_ + ItemKey[索引]); ok {
			global.H缓存.Delete(黑名单_ + ItemKey[索引])
		}
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *S_Blacklist) GetList(tx *gorm.DB, 请求 request.List, AppId int) (int64, []db.DB_Blacklist, error) {

	局_DB := tx.Model(db.DB_Blacklist{})

	if AppId > 0 {
		局_DB.Where("AppId = ?", AppId)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2: //用户名
			局_DB.Where("ItemKey LIKE ? ", "%"+请求.Keywords+"%")
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
	var 局_数组 []db.DB_Blacklist
	err := 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组).Error
	if err != nil {
		global.GVA_LOG.Error(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}
	return 总数, 局_数组, err
}

func (s *S_Blacklist) DeleteType(tx *gorm.DB, Type int) (影响行数 int64, error error) {
	var ItemKey []string
	var tx2 *gorm.DB
	switch Type {
	case 1: //删除全部
		tx.Model(db.DB_Blacklist{}).Select("ItemKey").Where("Id > 0").Find(&ItemKey)
		tx2 = tx.Model(db.DB_Blacklist{}).Where("Id > 0").Delete("")
	default:
		return 0, errors.New("类型错误")
	}

	for 索引, _ := range ItemKey {
		if _, ok := global.H缓存.Get(黑名单_ + ItemKey[索引]); ok {
			global.H缓存.Delete(黑名单_ + ItemKey[索引])
		}
	}
	return tx2.RowsAffected, tx2.Error

}
