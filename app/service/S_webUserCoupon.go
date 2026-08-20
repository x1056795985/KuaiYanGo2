package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/app/global"
	"server/app/models/dbm"
	"server/app/models/request"
	"server/app/utils"
)

type WebUserCoupon struct {
	*BaseService[dbm.DB_WebUserCoupon]
}

func NewWebUserCoupon(c *gin.Context, db *gorm.DB) *WebUserCoupon {
	return &WebUserCoupon{
		BaseService: NewBaseService[dbm.DB_WebUserCoupon](c, db),
	}
}

func (s *WebUserCoupon) GetList(请求 request.List, AppId int, Status int, Type int) (int64, []dbm.DB_WebUserCoupon, error) {
	db := s.db.Model(new(dbm.DB_WebUserCoupon))
	if 请求.Page == 0 {
		请求.Page = 1
	}
	if AppId > 0 {
		db = db.Where("AppId = ?", AppId)
	}
	if Status > 0 {
		db = db.Where("Status = ?", Status)
	}
	if Type > 0 {
		db = db.Where("Type = ?", Type)
	}
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1:
			db = db.Where("Id = ?", 请求.Keywords)
		case 2:
			db = db.Where("Name LIKE ?", "%"+请求.Keywords+"%")
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
	var results []dbm.DB_WebUserCoupon
	err := db.Order(order).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&results).Error
	if err != nil {
		global.GVA_LOG.Println(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}
	return count, results, err
}
