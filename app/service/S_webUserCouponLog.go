package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/app/global"
	"server/app/models/dbm"
	"server/app/models/request"
	"server/app/utils"
)

type WebUserCouponLog struct {
	*BaseService[dbm.DB_WebUserCouponLog]
}

func NewWebUserCouponLog(c *gin.Context, db *gorm.DB) *WebUserCouponLog {
	return &WebUserCouponLog{
		BaseService: NewBaseService[dbm.DB_WebUserCouponLog](c, db),
	}
}

func (s *WebUserCouponLog) GetList(请求 request.List, AppId int, CouponId int, CouponUserId int, Uid int, EventType int, TimeRange []string) (int64, []dbm.DB_WebUserCouponLog, error) {
	db := s.db.Model(new(dbm.DB_WebUserCouponLog))
	if 请求.Page == 0 {
		请求.Page = 1
	}
	if AppId > 0 {
		db = db.Where("AppId = ?", AppId)
	}
	if CouponId > 0 {
		db = db.Where("CouponId = ?", CouponId)
	}
	if CouponUserId > 0 {
		db = db.Where("CouponUserId = ?", CouponUserId)
	}
	if Uid > 0 {
		db = db.Where("Uid = ?", Uid)
	}
	if EventType > 0 {
		db = db.Where("EventType = ?", EventType)
	}
	局_开始, 局_结束 := 解析时间范围(TimeRange)
	if 局_开始 > 0 && 局_结束 > 0 {
		db = db.Where("Time >= ? AND Time <= ?", 局_开始, 局_结束)
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1:
			db = db.Where("Id = ?", 请求.Keywords)
		case 2:
			db = db.Where("PayOrder LIKE ?", "%"+请求.Keywords+"%")
		case 3:
			db = db.Where("Operator LIKE ?", "%"+请求.Keywords+"%")
		}
	}
	var count int64
	if 请求.Count > 0 && 请求.Count <= 500000 {
		count = 请求.Count
	} else {
		if err := db.Count(&count).Error; err != nil {
			return 0, nil, err
		}
	}
	order := "Id DESC"
	if 请求.Order == 1 {
		order = "Id ASC"
	}
	var results []dbm.DB_WebUserCouponLog
	err := db.Order(order).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&results).Error
	if err != nil {
		global.GVA_LOG.Println(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}
	return count, results, err
}
