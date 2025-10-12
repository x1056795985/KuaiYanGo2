package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/global"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
	"server/utils"
	"strconv"
)

type CpsPayOrder struct {
	*BaseService[dbm.DB_CpsPayOrder] // 嵌入泛型基础服务
}

func NewCpsPayOrder(c *gin.Context, db *gorm.DB) *CpsPayOrder {
	return &CpsPayOrder{
		BaseService: NewBaseService[dbm.DB_CpsPayOrder](c, db),
	}
}
func (s *CpsPayOrder) InfoOrder(order string) (info dbm.DB_CpsPayOrder, err error) {
	tx := s.db.Model(new(dbm.DB_CpsPayOrder)).Where("payOrder = ?", order).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *CpsPayOrder) Q好友订单(appId, 邀请人id int, 数量限制 int) (infos []dbm.DB_CpsPayOrder, err error) {

	tx := s.db.Model(new(dbm.DB_CpsPayOrder)).
		Where("appId = ?", appId).
		Where("inviterId = ?", 邀请人id).
		Order("id desc")

	if 数量限制 > 0 {
		tx = tx.Limit(数量限制).
			Find(&infos)
	} else {
		tx = tx.Find(&infos)
	}

	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *CpsPayOrder) Q裂变订单(appId, 邀请人id int, 数量限制 int) (infos []dbm.DB_CpsPayOrder, err error) {

	tx := s.db.Model(new(dbm.DB_CpsPayOrder)).
		Where("appId = ?", appId).
		Where("grandpaId = ?", 邀请人id).
		Order("id desc")

	if 数量限制 > 0 {
		tx = tx.Limit(数量限制).
			Find(&infos)
	} else {
		tx = tx.Find(&infos)
	}

	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// 优化查询链式操作
func (s *CpsPayOrder) GetList(请求 request.List, rangeTime []string) (int64, []dbm.DB_CpsPayOrder, error) {
	// 创建查询构建器
	db := s.db.Model(new(dbm.DB_CpsPayOrder))

	// 关键字搜索
	if 请求.Keywords != "" && 请求.Type == 1 {
		db = db.Where("Id = ?", 请求.Keywords)
	}

	if rangeTime != nil && len(rangeTime) == 2 && rangeTime[0] != "" && rangeTime[1] != "" {
		开始时间, _ := strconv.ParseInt(rangeTime[0], 10, 64)
		结束时间, _ := strconv.ParseInt(rangeTime[1], 10, 64)
		db.Where("time > ?", 开始时间).Where("time < ?", 结束时间+86400)
	}
	// 优化计数逻辑
	var count int64
	if 请求.Count > 0 && 请求.Count <= 500000 {
		count = 请求.Count
	} else {
		if err := db.Count(&count).Error; err != nil {
			return 0, nil, err
		}
	}

	// 排序处理
	order := "Id DESC"
	if 请求.Order == 1 {
		order = "Id ASC"
	}

	// 分页查询
	var results []dbm.DB_CpsPayOrder
	err := db.Order(order).
		Limit(请求.Size).
		Offset((请求.Page - 1) * 请求.Size).
		Find(&results).Error

	if err != nil {
		global.GVA_LOG.Error(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}

	return count, results, err
}
