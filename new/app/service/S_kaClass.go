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
func (s *KaClass) Create(请求 *dbm.DB_KaClass) (row int64, err error) {

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
	tx := s.db.Model(dbm.DB_KaClass{}).Create(请求)
	return tx.RowsAffected, tx.Error
}
func (s *KaClass) Info(id int) (info dbm.DB_KaClass, err error) {
	tx := s.db.Model(dbm.DB_KaClass{}).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}
func (s *KaClass) Infos(where map[string]interface{}) (info []dbm.DB_KaClass, err error) {
	tx := s.db.Model(dbm.DB_KaClass{}).Where(where).Find(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}
func (s *KaClass) Update(Id int, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(dbm.DB_KaClass{}).Where("Id = ?", Id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// 获取列表
func (s *KaClass) GetList(请求 request.List, AppId int, ids []int) (int64, []dbm.DB_KaClass, error) {

	局_DB := s.db.Model(dbm.DB_KaClass{})
	//直接限制代理ids
	if len(ids) > 0 {
		局_DB.Where("Id in (?)", ids)
	}

	if AppId > 0 {
		局_DB.Where("AppId = ?", AppId)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2: //用户名
			局_DB.Where("Name LIKE ? ", "%"+请求.Keywords+"%")
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
	var 局_数组 []dbm.DB_KaClass
	err := 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组).Error
	if err != nil {
		global.GVA_LOG.Error(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}
	return 总数, 局_数组, err
}
