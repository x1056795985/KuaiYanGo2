package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/global"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
	DB "server/structs/db"
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

	//关键字筛选
	if 请求.Keywords != "" {
		var 用户详情 DB.DB_User

		switch 请求.Type {
		case 1: //消费者
			tx := s.db.Model(new(dbm.DB_CpsPayOrder))
			_ = tx.Model(DB.DB_User{}).Where("User =?", 请求.Keywords).First(&用户详情).Error
			db.Where("uid = ?", 用户详情.Id)
		case 2: //备注
			db.Where("note LIKE ?", "%"+请求.Keywords+"%")
		case 3: //邀请人
			tx := s.db.Model(new(dbm.DB_CpsPayOrder))
			_ = tx.Model(DB.DB_User{}).Where("User =?", 请求.Keywords).First(&用户详情).Error
			db.Where("inviterId = ? ", 用户详情.Id)
		case 4: //邀请上级
			tx := s.db.Model(new(dbm.DB_CpsPayOrder))
			_ = tx.Model(DB.DB_User{}).Where("User =?", 请求.Keywords).First(&用户详情).Error
			db.Where("grandpaId = ? ", 用户详情.Id)
		case 5: //订单编号
			db.Where("payOrder LIKE ?", "%"+请求.Keywords+"%")
		case 6: //金额
			db.Where("Rmb = ? ", 请求.Keywords)
		case 7: //用户名
			tx := s.db.Model(new(dbm.DB_CpsPayOrder))
			_ = tx.Model(DB.DB_User{}).Where("User =?", 请求.Keywords).First(&用户详情).Error
			db.Where("(inviterId = ? OR uid = ? OR grandpaId = ?)", 用户详情.Id, 用户详情.Id, 用户详情.Id)
		}
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
