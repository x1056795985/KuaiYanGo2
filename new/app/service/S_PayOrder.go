package service

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/global"
	"server/new/app/models/request"
	DB "server/structs/db"
	"server/utils"
)

type PayOrder struct {
	*BaseService[DB.DB_LogRMBPayOrder] // 嵌入泛型基础服务
}

func NewPayOrder(c *gin.Context, db *gorm.DB) *PayOrder {
	return &PayOrder{
		BaseService: NewBaseService[DB.DB_LogRMBPayOrder](c, db),
	}
}

// 优化查询链式操作
func (s *PayOrder) GetList(请求 request.List) (int64, []DB.DB_LogRMBPayOrder, error) {
	// 创建查询构建器
	db := s.db.Model(new(DB.DB_LogRMBPayOrder))
	if 请求.Page == 0 {
		请求.Page = 1
	}

	// 关键字搜索
	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: // id
			db = db.Where("Id = ?", 请求.Keywords)
		case 2: //uid 只能查账号模式
			db = db.Where("UidType = 1").Where("Uid = ?", D到整数(请求.Keywords))
		case 3: //uid 只能查卡号模式
			db = db.Where("UidType = 2").Where("Uid = ?", D到整数(请求.Keywords))
		}
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
	var results []DB.DB_LogRMBPayOrder
	err := db.Order(order).
		Limit(请求.Size).
		Offset((请求.Page - 1) * 请求.Size).
		Find(&results).Error

	if err != nil {
		global.GVA_LOG.Error(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}

	return count, results, err
}
