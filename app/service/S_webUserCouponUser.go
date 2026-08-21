package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/app/global"
	"server/app/models/dbm"
	"server/app/models/request"
	"server/app/utils"
	"strconv"
)

type WebUserCouponUser struct {
	*BaseService[dbm.DB_WebUserCouponUser]
}

func NewWebUserCouponUser(c *gin.Context, db *gorm.DB) *WebUserCouponUser {
	return &WebUserCouponUser{
		BaseService: NewBaseService[dbm.DB_WebUserCouponUser](c, db),
	}
}

func (s *WebUserCouponUser) GetList(请求 request.List, AppId int, Uid int, Status int, ReceiveTime []string, UseTime []string) (int64, []dbm.DB_WebUserCouponUser, error) {
	db := s.db.Model(new(dbm.DB_WebUserCouponUser))
	if 请求.Page == 0 {
		请求.Page = 1
	}
	if AppId > 0 {
		db = db.Where("AppId = ?", AppId)
	}
	if Uid > 0 {
		db = db.Where("Uid = ?", Uid)
	}
	if Status > 0 {
		db = db.Where("Status = ?", Status)
	}
	局_领取开始, 局_领取结束 := 解析时间范围(ReceiveTime)
	if 局_领取开始 > 0 && 局_领取结束 > 0 {
		db = db.Where("ReceiveTime >= ? AND ReceiveTime <= ?", 局_领取开始, 局_领取结束)
	}
	局_使用开始, 局_使用结束 := 解析时间范围(UseTime)
	if 局_使用开始 > 0 && 局_使用结束 > 0 {
		db = db.Where("UseTime >= ? AND UseTime <= ?", 局_使用开始, 局_使用结束)
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1:
			db = db.Where("Id = ?", 请求.Keywords)
		case 2:
			db = db.Where("CouponName LIKE ?", "%"+请求.Keywords+"%")
		case 3:
			db = db.Where("PayOrder LIKE ?", "%"+请求.Keywords+"%")
		case 4:
			if uid, err := strconv.Atoi(请求.Keywords); err == nil && uid > 0 {
				db = db.Where("Uid = ?", uid)
			}
		case 5:
			db = db.Where("Uid IN (?)", s.db.Model(new(dbm.DB_User)).Select("Id").Where("User LIKE ?", "%"+请求.Keywords+"%"))
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
	var results []dbm.DB_WebUserCouponUser
	err := db.Order(order).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&results).Error
	if err != nil {
		global.GVA_LOG.Println(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}
	return count, results, err
}

func 解析时间范围(时间数组 []string) (int64, int64) {
	if len(时间数组) != 2 {
		return 0, 0
	}
	局_开始, 局_开始错误 := strconv.ParseInt(时间数组[0], 10, 64)
	if 局_开始错误 != nil {
		return 0, 0
	}
	局_结束, 局_结束错误 := strconv.ParseInt(时间数组[1], 10, 64)
	if 局_结束错误 != nil {
		return 0, 0
	}
	return 局_开始, 局_结束
}
