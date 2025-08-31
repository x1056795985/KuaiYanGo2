package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/global"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
	"server/utils"
)

type CheckInScoreLog struct {
	*BaseService[dbm.DB_CheckInScoreLog] // 嵌入泛型基础服务
}

// NewcheckInScoreLog 创建 checkInScoreLog 实例
func NewCheckInScoreLog(c *gin.Context, db *gorm.DB) *CheckInScoreLog {
	return &CheckInScoreLog{
		BaseService: NewBaseService[dbm.DB_CheckInScoreLog](c, db),
	}
}

// 优化查询链式操作
func (s *CheckInScoreLog) GetList(请求 request.List2, appId, userId int) (int64, []dbm.DB_CheckInScoreLog, error) {
	// 创建查询构建器
	db := s.db.Model(new(dbm.DB_CheckInScoreLog))

	db = db.Where("appId = ? and userId=?", appId, userId)
	// 关键字搜索
	if 请求.Keywords != "" && 请求.Type == 1 {
		db = db.Where("Id = ?", 请求.Keywords)
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
	var results []dbm.DB_CheckInScoreLog
	err := db.Order(order).
		Limit(请求.Size).
		Offset((请求.Page - 1) * 请求.Size).
		Find(&results).Error

	if err != nil {
		global.GVA_LOG.Error(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}

	return count, results, err
}
