package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/new/app/models/request"
	DB "server/structs/db"
	"strconv"
)

type User struct {
	db *gorm.DB
	c  *gin.Context
}

// NewUser 创建 User 实例
func NewUser(c *gin.Context, db *gorm.DB) *User {
	return &User{
		db: db,
		c:  c,
	}
}

// 增
func (s *User) Create(info *DB.DB_User) (row int64, err error) {
	//创建会自动重新赋值info.Id为新插入的数据id
	tx := s.db.Model(DB.DB_User{}).Create(info)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *User) Delete(Id interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := Id.(type) {
	case int:
		tx2 = s.db.Model(DB.DB_User{}).Where("Id = ?", k).Delete("")
	case []int:
		tx2 = s.db.Model(DB.DB_User{}).Where("Id IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *User) GetList(请求 request.List, Status int) (int64, []DB.DB_User, error) {
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
	var 局_数组 []DB.DB_User
	tx = tx.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组)

	return 总数, 局_数组, tx.Error
}

// 查
func (s *User) Info(id int) (info DB.DB_User, err error) {
	tx := s.db.Model(DB.DB_User{}).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *User) Info2(where map[string]interface{}) (info DB.DB_User, err error) {
	tx := s.db.Model(DB.DB_User{}).Where(where).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *User) Infos(where map[string]interface{}) (info []DB.DB_User, err error) {
	tx := s.db.Model(DB.DB_User{}).Where(where).Find(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 查
func (s *User) InfoName(name string) (info DB.DB_User, err error) {
	tx := s.db.Model(DB.DB_User{}).Where("User = ?", name).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 改
func (s *User) Update(id int, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(DB.DB_User{}).Where("Id = ?", id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

func (s *User) Id取Uid_批量(AppId int, id []int) []int {
	var Uid []int
	s.db.Raw("SELECT `Uid` FROM `db_AppUser_"+strconv.Itoa(AppId)+"` WHERE `Id` IN  ? ", id).Scan(&Uid)
	return Uid
}
