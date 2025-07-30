package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/global"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
	"server/utils"
	"time"
)

type AppPromotionConfig struct {
	*BaseService[dbm.DB_AppPromotionConfig] // 嵌入泛型基础服务
}

func NewAppPromotionConfig(c *gin.Context, db *gorm.DB) *AppPromotionConfig {
	return &AppPromotionConfig{
		BaseService: NewBaseService[dbm.DB_AppPromotionConfig](c, db),
	}
}

// 增
func (s *AppPromotionConfig) Create(请求 *dbm.DB_AppPromotionConfig) (row int64, err error) {

	if 请求.Id > 0 {
		return 0, errors.New("创建不能有id值")
	}
	if 请求.AppId < 10000 {
		return 0, errors.New("AppId错误")
	}
	if 请求.Name == "" {
		return 0, errors.New("名称不能为空")
	}

	tx := s.db.Model(dbm.DB_AppPromotionConfig{}).Create(请求)
	return tx.RowsAffected, tx.Error
}

// 获取列表
func (s *AppPromotionConfig) GetList(请求 request.List, AppId int, Status int, PromotionType int) (int64, []dbm.DB_AppPromotionConfig, error) {

	局_DB := s.db.Model(dbm.DB_AppPromotionConfig{})
	if AppId > 0 {
		局_DB.Where("AppId = ?", AppId)
	}

	switch Status {
	case 1:
		//<el-option key="1" label="未开始" :value="1"/>
		局_DB.Where("startTime > ?", time.Now().Unix())
	case 2:
		//<el-option key="2" label="活动中" :value="2"/>
		局_DB.Where("startTime < ?", time.Now().Unix()).Where("endTime > ?", time.Now().Unix())
	case 3:
		//<el-option key="3" label="已结束" :value="3"/>
		局_DB.Where("endTime < ?", time.Now().Unix())
	case 4: //
		//<el-option key="4" label="活动中和即将开始" :value="3"/>

		局_DB.Where("startTime < ?", time.Now().Unix()+86400).Where("endTime > ?", time.Now().Unix())
	}

	if PromotionType > 0 {
		局_DB.Where("PromotionType = ?", PromotionType)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			局_DB.Where("Id = ?", 请求.Keywords)
		case 2: //用户名
			局_DB.Where("Name LIKE ? ", "%"+请求.Keywords+"%")
		}
	}
	var 总数 int64
	//Count(&总数) 必须放在where 后面 不然值会被清0
	if 请求.Count > 500000 { //如果数据大于50万 直接使用,不重新查询了 优化速度
		总数 = 请求.Count
	} else {
		局_DB.Count(&总数)
	}

	//处理排序
	if 请求.Order == 1 {
		局_DB.Order("Sort DESC, Id ASC")
	} else {
		局_DB.Order("Sort DESC, Id DESC")
	}

	var 局_数组 []dbm.DB_AppPromotionConfig
	err := 局_DB.Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&局_数组).Error
	if err != nil {
		global.GVA_LOG.Error(utils.Q取包名结构体方法(s) + ":" + err.Error())
	}
	return 总数, 局_数组, err
}
