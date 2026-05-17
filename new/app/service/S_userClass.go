package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/new/app/models/request"
	DB "server/structs/db"
	"strconv"
)

type UserClass struct {
	db *gorm.DB
	c  *gin.Context
}

// NewUserClass 创建 UserClass 实例
func NewUserClass(c *gin.Context, db *gorm.DB) *UserClass {
	return &UserClass{
		db: db,
		c:  c,
	}
}

// 增
func (s *UserClass) Create(info DB.DB_UserClass) (row int64, err error) {
	//创建会自动重新赋值info.Id为新插入的数据id
	tx := s.db.Model(DB.DB_UserClass{}).Create(&info)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *UserClass) Delete(Id interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := Id.(type) {
	case int:
		tx2 = s.db.Model(DB.DB_UserClass{}).Where("Id = ?", k).Delete("")
	case []int:
		tx2 = s.db.Model(DB.DB_UserClass{}).Where("Id IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *UserClass) GetList(请求 request.List, Status int) (int64, []DB.DB_UserClass, error) {
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
	var 局_数组 []DB.DB_UserClass
	tx = tx.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组)

	return 总数, 局_数组, tx.Error
}

// 查
func (s *UserClass) Info(id int) (info DB.DB_UserClass, err error) {
	tx := s.db.Model(DB.DB_UserClass{}).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 改
func (s *UserClass) Update(Id int, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(DB.DB_UserClass{}).Where("Id = ?", Id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// GetListByAppId 按AppId过滤的用户类型列表
func (s *UserClass) GetListByAppId(AppId int, 请求 request.List) (int64, []DB.DB_UserClass, error) {
	tx := s.db.Model(DB.DB_UserClass{}).Where("AppId = ?", AppId)

	if 请求.Order == 1 {
		tx = tx.Order("Id ASC")
	} else {
		tx = tx.Order("Id DESC")
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			tx = tx.Where("Id = ?", 请求.Keywords)
		}
	}

	var 总数 int64
	if 请求.Count > 500000 {
		总数 = 请求.Count
	} else {
		tx.Count(&总数)
	}

	var dataList []DB.DB_UserClass
	err := tx.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&dataList).Error
	return 总数, dataList, err
}

// IsIdExists 用户类型Id是否存在
func (s *UserClass) IsIdExists(id int) bool {
	var count int64
	s.db.Model(DB.DB_UserClass{}).Select("1").Where("Id = ?", id).Count(&count)
	return count > 0
}

// IsNameExists 用户类型名称是否存在
func (s *UserClass) IsNameExists(AppId int, Name string) bool {
	var count int64
	s.db.Model(DB.DB_UserClass{}).Where("AppId = ? AND Name = ?", AppId, Name).Count(&count)
	return count > 0
}

// IsMarkExists 整数代号是否存在
func (s *UserClass) IsMarkExists(AppId int, Mark int) bool {
	var count int64
	s.db.Model(DB.DB_UserClass{}).Where("AppId = ? AND Mark = ?", AppId, Mark).Count(&count)
	return count > 0
}

// IsMarkExistsCount 整数代号存在数量（排除指定ID）
func (s *UserClass) IsMarkExistsCount(AppId int, Mark int, 排除ID []int) int64 {
	var count int64
	tx := s.db.Model(DB.DB_UserClass{}).Where("AppId = ?", AppId).Where("Mark = ?", Mark)
	if len(排除ID) > 0 {
		tx = tx.Where("Id NOT IN ?", 排除ID)
	}
	tx.Count(&count)
	return count
}

// GetIdNameList 取id和名字map列表
func (s *UserClass) GetIdNameList(AppId int) (map[string]string, error) {
	var list []DB.DB_UserClass
	err := s.db.Model(DB.DB_UserClass{}).Select("Id, Name").Where("AppId = ?", AppId).Find(&list).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(list))
	for _, v := range list {
		result[strconv.Itoa(v.Id)] = v.Name
	}
	return result, nil
}

// DeleteByAppIdAndIds 按AppId和Id数组删除，同时更新AppUser中的用户分类为0
func (s *UserClass) DeleteByAppIdAndIds(AppId int, Ids []int) (int64, error) {
	// 删除UserClass
	tx := s.db.Model(DB.DB_UserClass{}).Where("AppId = ?", AppId).Where("Id IN ?", Ids).Delete("")
	if tx.Error != nil {
		return 0, tx.Error
	}
	影响行数 := tx.RowsAffected

	// 修改软件中已删除的用户类型为0
	s.db.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(AppId)).
		Select("UserClassId").Where("UserClassId IN ?", Ids).Update("UserClassId", 0)

	return 影响行数, nil
}
