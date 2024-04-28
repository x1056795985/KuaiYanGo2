package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/new/app/models/request"
	DB "server/structs/db"
	"time"
)

type LogMoney struct {
	db *gorm.DB
	c  *gin.Context
}

// NewLogMoney 创建 LogMoney 实例
func NewLogMoney(c *gin.Context, db *gorm.DB) *LogMoney {
	return &LogMoney{
		db: db,
		c:  c,
	}
}

// 增
func (s *LogMoney) Create(info DB.DB_LogMoney) (row int64, err error) {
	//创建会自动重新赋值info.Id为新插入的数据id
	info.Time = int(time.Now().Unix())
	tx := s.db.Model(DB.DB_LogMoney{}).Create(&info)
	return tx.RowsAffected, tx.Error
}

// 删除 支持 数组,和id
func (s *LogMoney) Delete(Id interface{}) (影响行数 int64, error error) {
	var tx2 *gorm.DB
	switch k := Id.(type) {
	case int:
		tx2 = s.db.Model(DB.DB_LogMoney{}).Where("Id = ?", k).Delete("")
	case []int:
		tx2 = s.db.Model(DB.DB_LogMoney{}).Where("Id IN ?", k).Delete("")
	default:
		return 0, errors.New("错误的数据")
	}
	return tx2.RowsAffected, tx2.Error
}

// 获取列表
func (s *LogMoney) GetList(请求 request.List, Status int) (int64, []DB.DB_LogMoney, error) {
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
	var 局_数组 []DB.DB_LogMoney
	tx = tx.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组)

	return 总数, 局_数组, tx.Error
}

// 查
func (s *LogMoney) Info(id int) (info DB.DB_LogMoney, err error) {
	tx := s.db.Model(DB.DB_LogMoney{}).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 改
func (s *LogMoney) Update(id int, 数据 map[string]interface{}) (row int64, err error) {

	tx := s.db.Model(DB.DB_LogMoney{}).Where("id = ?", id).Create(&数据)
	return tx.RowsAffected, tx.Error
}
