package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/app/global"
	dbm "server/app/models/db"
	"server/app/models/request"
	"server/app/utils"
)

type LuckyWheelLog struct {
	*BaseService[dbm.DB_LuckyWheelLog]
}

func NewLuckyWheelLog(c *gin.Context, db *gorm.DB) *LuckyWheelLog {
	return &LuckyWheelLog{
		BaseService: NewBaseService[dbm.DB_LuckyWheelLog](c, db),
	}
}

// GetList 查询抽奖记录列表
func (s *LuckyWheelLog) GetList(请求 request.List2, appId, userId int, 开始时间, 结束时间 int64) (int64, []dbm.DB_LuckyWheelLog, error) {
	db := s.db.Model(new(dbm.DB_LuckyWheelLog))
	if appId > 0 {
		db = db.Where("appId = ?", appId)
	}
	if userId > 0 {
		db = db.Where("userId = ?", userId)
	}
	if 开始时间 > 0 {
		db = db.Where("createTime >= ?", 开始时间)
	}
	if 结束时间 > 0 {
		db = db.Where("createTime <= ?", 结束时间)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //用户名搜索
			局_临时userId := 0
			局_userInfo, err2 := NewUser(s.c, s.db).InfoName(请求.Keywords)
			if err2 == nil {
				局_临时userId = 局_userInfo.Id
			}
			db = db.Where("userId = ?", 局_临时userId)
		case 2: //卡类名称
			db = db.Where("kaClassName LIKE ?", "%"+请求.Keywords+"%")
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

	var results []dbm.DB_LuckyWheelLog
	err := db.Order(order).
		Limit(请求.Size).
		Offset((请求.Page - 1) * 请求.Size).
		Find(&results).Error

	if err != nil {
		global.GVA_LOG.Println(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}

	return count, results, err
}
