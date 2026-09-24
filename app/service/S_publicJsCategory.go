package service

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/app/models/dbm"
)

type PublicJsCategory struct {
	db *gorm.DB
	c  *gin.Context
}

// NewPublicJsCategory 创建 PublicJsCategory 实例
func NewPublicJsCategory(c *gin.Context, db *gorm.DB) *PublicJsCategory {
	return &PublicJsCategory{
		db: db,
		c:  c,
	}
}

// Q取列表 取全部分类列表(Sort升序)
func (s *PublicJsCategory) Q取列表() []dbm.DB_PublicJsCategory {
	var 局_列表 []dbm.DB_PublicJsCategory
	err := s.db.Model(dbm.DB_PublicJsCategory{}).Order("Sort ASC, Id ASC").Find(&局_列表).Error
	if err != nil {
		return []dbm.DB_PublicJsCategory{}
	}
	return 局_列表
}

// Id是否存在 分类Id是否存在
func (s *PublicJsCategory) Id是否存在(Id int) bool {
	var Count int64
	result := s.db.Model(dbm.DB_PublicJsCategory{}).Select("1").Where("Id=?", Id).First(&Count)
	return result.Error == nil
}

// Name取Id 按分类名取Id,不存在返回0
func (s *PublicJsCategory) Name取Id(Name string) int {
	if Name == "" {
		return 0
	}
	var Id int
	s.db.Model(dbm.DB_PublicJsCategory{}).Select("Id").Where("Name=?", Name).First(&Id)
	return Id
}

// P取分类函数数量统计 取每个分类下的函数数量 key=CategoryId (CategoryId=0为未分类)
func (s *PublicJsCategory) P取分类函数数量() map[int]int64 {
	type 结构_分类函数统计 struct {
		CategoryId int   `gorm:"column:CategoryId"`
		Count      int64 `gorm:"column:Count"`
	}
	var 局_统计列表 []结构_分类函数统计
	err := s.db.Model(dbm.DB_PublicJs{}).Select("CategoryId, COUNT(*) as Count").Group("CategoryId").Find(&局_统计列表).Error
	if err != nil {
		return map[int]int64{}
	}
	局_map := make(map[int]int64, len(局_统计列表)+1)
	for _, v := range 局_统计列表 {
		局_map[v.CategoryId] = v.Count
	}
	return 局_map
}

// Q取子孙Ids含自身 递归取分类自身及所有子孙分类Id
func (s *PublicJsCategory) Q取子孙Ids含自身(Id int) []int {
	var 局_全部分类 []dbm.DB_PublicJsCategory
	err := s.db.Model(dbm.DB_PublicJsCategory{}).Find(&局_全部分类).Error
	if err != nil {
		return []int{Id}
	}
	局_结果 := []int{Id}
	局_待处理 := []int{Id}
	for len(局_待处理) > 0 {
		局_当前 := 局_待处理[0]
		局_待处理 = 局_待处理[1:]
		for _, v := range 局_全部分类 {
			if v.ParentId == 局_当前 {
				局_结果 = append(局_结果, v.Id)
				局_待处理 = append(局_待处理, v.Id)
			}
		}
	}
	return 局_结果
}

// Q是否自身或子孙分类 判断 待检查父Id 是否是 Id 自身或其子孙(防止把自己挂到子孙下形成环)
func (s *PublicJsCategory) Q是否自身或子孙分类(Id int, 待检查父Id int) bool {
	if 待检查父Id <= 0 {
		return false
	}
	if Id == 待检查父Id {
		return true
	}
	for _, v := range s.Q取子孙Ids含自身(Id) {
		if v == 待检查父Id {
			return true
		}
	}
	return false
}

// D删除 删除分类及其整个子孙树,分类树下所有函数移入未分类 (多表事务)
func (s *PublicJsCategory) D删除(Id int) (移动数量 int64, err error) {
	局_待删Ids := s.Q取子孙Ids含自身(Id)
	err = s.db.Transaction(func(tx *gorm.DB) error {
		//分类树下所有函数移入未分类
		结果 := tx.Model(dbm.DB_PublicJs{}).Where("CategoryId IN ?", 局_待删Ids).Update("CategoryId", 0)
		if 结果.Error != nil {
			return 结果.Error
		}
		移动数量 = 结果.RowsAffected
		//删除分类子树
		结果 = tx.Where("Id IN ?", 局_待删Ids).Delete(&dbm.DB_PublicJsCategory{})
		if 结果.Error != nil {
			return 结果.Error
		}
		if 结果.RowsAffected == 0 {
			return errors.New("分类不存在或已被删除")
		}
		return nil
	})
	if err != nil {
		移动数量 = 0
	}
	return 移动数量, err
}

// P批量修改分类 批量修改函数归属分类
func (s *PublicJsCategory) P批量修改分类(Id []int, CategoryId int) error {
	return s.db.Model(dbm.DB_PublicJs{}).Where("Id in ?", Id).Update("CategoryId", CategoryId).Error
}
