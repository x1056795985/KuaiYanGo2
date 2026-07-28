package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
)

const Js类型_公共函数 = 1
const Js类型_任务池Hook函数 = 2
const Js类型_ApiHook函数 = 3

type PublicJs struct {
	db *gorm.DB
	c  *gin.Context
}

// NewPublicJs 创建 PublicJs 实例
func NewPublicJs(c *gin.Context, db *gorm.DB) *PublicJs {
	return &PublicJs{
		db: db,
		c:  c,
	}
}

// Id是否存在 公共JS函数Id是否存在
func (s *PublicJs) Id是否存在(Id int) bool {
	var Count int64
	result := s.db.Model(dbm.DB_PublicJs{}).Select("1").Where("Id=?", Id).First(&Count)
	return result.Error == nil
}

// Name是否存在 公共JS函数Name是否存在
func (s *PublicJs) Name是否存在(AppId int, Name string) bool {

	var Count int64
	s.db.Model(dbm.DB_PublicJs{}).Select("1").Where("AppId=?", AppId).Where("Name=?", Name).Take(&Count)
	return Count > 0
}

// Q取值 按AppId和Name取公共JS函数值
func (s *PublicJs) Q取值(AppId int, Name string) string {
	var value string
	s.db.Model(dbm.DB_PublicJs{}).Select("Value").Where("AppId=?", AppId).Where("Name=?", Name).First(&value)
	return value
}

// Q取值2 按Id取公共JS函数完整信息
func (s *PublicJs) Q取值2(id int) (dbm.DB_PublicJs, error) {
	var value dbm.DB_PublicJs
	err := s.db.Model(dbm.DB_PublicJs{}).Where("Id=?", id).First(&value).Error
	return value, err
}

// Name取Id 按AppId数组和Name取公共JS函数Id
func (s *PublicJs) Name取Id(AppId []int, Name string) int {
	if Name == "" {
		return 0
	}
	var Id int

	s.db.Model(dbm.DB_PublicJs{}).Select("Id").Where("AppId IN ?", AppId).Where("Name=?", Name).First(&Id)
	return Id
}

// Z置值 按id修改公共JS函数值
func (s *PublicJs) Z置值(id int, Value string) error {
	return s.db.Model(dbm.DB_PublicJs{}).Select("Value").Where("id=?", id).Update("Value", Value).Error
}

// Z置值2 修改公共JS函数(涉及文件IO,实际实现委托给logic/common/publicJs)
// 此方法涉及磁盘文件IO+数据库+缓存三重操作，必须放到logic层保持事务一致性
// func (s *PublicJs) Z置值2(PublicJs dbm.DB_PublicJs) error

// C创建 创建公共JS函数(涉及文件IO,实际实现委托给logic/common/publicJs)
// func (s *PublicJs) C创建(PublicJs dbm.DB_PublicJs) error

// P批量修改IsVip 批量修改公共JS函数IsVip
func (s *PublicJs) P批量修改IsVip(Id []int, IsVip int) error {
	return s.db.Model(dbm.DB_PublicJs{}).Where("Id in ?", Id).Update("IsVip", IsVip).Error
}

// P取全部公共函数名称 按Appid取全部公共函数名称
func (s *PublicJs) P取全部公共函数名称(Appid int) []string {
	var 局_PublicJs []string
	err := s.db.Model(dbm.DB_PublicJs{}).Select("Name").Where("AppId=?", Appid).Find(&局_PublicJs).Error
	if err != nil {
		return []string{}
	}
	return 局_PublicJs
}

// 确保 errors 包被使用
var _ = errors.New
