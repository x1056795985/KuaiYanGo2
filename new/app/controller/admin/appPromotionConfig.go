package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
	. "server/new/app/models/response"
	"server/new/app/service"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
	"time"
)

type AppPromotionConfig struct {
	Common.Common
}

func NewAppPromotionConfigController() *AppPromotionConfig {
	return &AppPromotionConfig{}
}

// GetList
func (C *AppPromotionConfig) GetList(c *gin.Context) {
	//{"AppId":2,"Type":2,"Size":10,"Page":1,"Status":1,"keywords":"1"}
	var 请求 struct {
		request.List
		AppId         int `json:"appId"`
		Status        int `json:"status"`
		PromotionType int `json:"promotionType"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	tx := *global.GVA_DB
	var S = service.NewAppPromotionConfig(c, &tx)
	var dataList []dbm.DB_AppPromotionConfig
	var 总数 int64
	var err error
	总数, dataList, err = S.GetList(请求.List, 请求.AppId, 请求.Status, 请求.PromotionType)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(GetList2{List: dataList, Count: 总数}, "操作成功", c)
	return

}

// Create
// @action 添加
// @show  2
func (C *AppPromotionConfig) Create(c *gin.Context) {
	var 请求 struct {
		Name          string `json:"name" binding:"required" zh:"活动名称"` //比如教师节活动
		AppId         int    `json:"appId" binding:"required" zh:"appId"`
		StartTime     int64  `json:"startTime" binding:"required" zh:"开始时间"`
		EndTime       int64  `json:"endTime" binding:"required" zh:"结束时间"`
		PromotionType int    `json:"promotionType" binding:"required" zh:"活动类型"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	if 请求.StartTime >= 请求.EndTime {
		response.FailWithMessage("开始时间必须早于结束时间", c)
		return
	}

	tx := *global.GVA_DB

	var info struct {
		AppInfo            DB.DB_AppInfo
		局_活动配置表ID          int
		Cps                dbm.DB_CpsInfo
		CheckInInfo        dbm.DB_CheckInInfo
		AppPromotionConfig dbm.DB_AppPromotionConfig
	}
	var err error
	info.AppInfo, err = service.NewAppInfo(c, &tx).Info(请求.AppId)
	if err != nil {
		response.FailWithMessage("AppId不存在", c)
		return
	}
	//做预检查
	if 请求.PromotionType == constant.H活动类型_cps && (info.AppInfo.AppType == 3 || info.AppInfo.AppType == 4) {
		response.FailWithMessage("卡号模式应用暂不支持该活动", c)
		return
	}
	if 请求.PromotionType == constant.H活动类型_签到 && (info.AppInfo.AppType == 3 || info.AppInfo.AppType == 4) {
		response.FailWithMessage("卡号模式应用暂不支持该活动", c)
		return
	}

	var S = service.NewAppPromotionConfig(c, &tx)

	//事务内处理
	err = tx.Transaction(func(tx *gorm.DB) error {
		//先创建活动表配置信息, 创建成功后, 再创建活动表数据
		switch 请求.PromotionType {
		default:
			return errors.New("活动类型错误")
		case constant.H活动类型_签到:
			info.CheckInInfo = dbm.DB_CheckInInfo{
				CreateTime:       time.Now().Unix(),
				UpdateTime:       time.Now().Unix(),
				ShareGivePoints:  10,
				InviteGivePoints: 88,
				CardClassList:    "[]",
			}

			_, err = service.NewCheckInInfo(c, tx).Create(&info.CheckInInfo)
			if err != nil {
				return err
			}
			info.局_活动配置表ID = info.CheckInInfo.Id

		case constant.H活动类型_cps:
			info.Cps = dbm.DB_CpsInfo{
				CreateTime:         time.Now().Unix(),
				UpdateTime:         time.Now().Unix(),
				BronzeThreshold:    0,
				BronzeKickback:     10,
				SilverThreshold:    10,
				SilverKickback:     20,
				GoldMedalThreshold: 20,
				GoldMedalKickback:  30,
				GrandsonKickback:   2,
				NarrowPic:          "",
				DetailPic:          "",
				BindingDay:         180,
			}
			_, err = service.NewCpsInfo(c, tx).Create(&info.Cps)
			if err != nil {
				return err
			}
			info.局_活动配置表ID = info.Cps.Id
		}

		_, err = S.Create(&dbm.DB_AppPromotionConfig{
			Name:             请求.Name,
			AppId:            请求.AppId,
			CreateTime:       time.Now().Unix(),
			UpdateTime:       time.Now().Unix(),
			StartTime:        请求.StartTime,
			EndTime:          请求.EndTime,
			PromotionType:    请求.PromotionType,
			TypeAssociatedId: info.局_活动配置表ID,
		})
		return err

	})

	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return // 确保退出函数
	}
	response.Ok(c) // 仅在无错误时返回成功
}

// @action 删除
// @show  2
func (C *AppPromotionConfig) Delete(c *gin.Context) {
	var 请求 request.Ids
	if !C.ToJSON(c, &请求) {
		return
	}

	tx := *global.GVA_DB
	var err error
	var 影响行数 int64

	// 使用事务包裹删除操作
	err = tx.Transaction(func(tx *gorm.DB) error {
		// 1. 先查询需要删除的活动配置
		var configs []dbm.DB_AppPromotionConfig
		if err = tx.Where("id IN (?)", 请求.Ids).Find(&configs).Error; err != nil {
			return err
		}

		// 2. 删除关联的CpsInfo记录（仅限类型1的活动）
		var cpsIds []int
		for _, config := range configs {
			if config.PromotionType == 1 {
				cpsIds = append(cpsIds, config.TypeAssociatedId)
			}
		}

		if len(cpsIds) > 0 {
			if err = tx.Where("id IN (?)", cpsIds).Delete(&dbm.DB_CpsInfo{}).Error; err != nil {
				return err
			}
		}

		// 3. 删除主表记录
		result := tx.Where("id IN (?)", 请求.Ids).Delete(&dbm.DB_AppPromotionConfig{})
		if result.Error != nil {
			return result.Error
		}

		影响行数 = result.RowsAffected
		return nil
	})

	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
}

// Update
// @action 更新
// @show  2
func (C *AppPromotionConfig) Update(c *gin.Context) {
	var 请求 dbm.DB_AppPromotionConfig
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Id <= 0 {
		response.FailWithMessage("Id必须大于0", c)
		return
	}
	tx := *global.GVA_DB

	_, err := service.NewAppPromotionConfig(c, &tx).UpdateMap([]int{请求.Id}, map[string]interface{}{
		"name":       请求.Name,
		"updateTime": time.Now().Unix(),
		"startTime":  请求.StartTime,
		"endTime":    请求.EndTime,
	})
	if err != nil {
		response.FailWithMessage(err.Error(), c)

	} else {
		response.OkWithMessage("操作成功", c)
	}

}

// Info
// @action 查询
// @show  2
func (C *AppPromotionConfig) Info(c *gin.Context) {
	var 请求 request.Id
	if !C.ToJSON(c, &请求) {
		return
	}

	tx := *global.GVA_DB
	var S = service.NewAppPromotionConfig(c, &tx)
	var info dbm.DB_AppPromotionConfig
	info, err := S.Info(请求.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)

	} else {
		response.OkWithDetailed(info, "操作成功", c)
	}

}

// @action 查询
// @show  2
func (C *AppPromotionConfig) Sort(c *gin.Context) {
	var 请求 struct {
		request.Ids
		Sort int64 `json:"sort" `
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	tx := *global.GVA_DB
	var S = service.NewAppPromotionConfig(c, &tx)
	总数, err := S.UpdateWhere(map[string]interface{}{"id": 请求.Ids.Ids}, map[string]interface{}{"sort": 请求.Sort})

	if err != nil {
		response.FailWithMessage(err.Error(), c)
	} else {
		response.OkWithDetailed(总数, "操作成功", c)
	}

}
func (C *AppPromotionConfig) Reset(c *gin.Context) {
	var 请求 struct {
		Id int `json:"id" binding:"required,min=1"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	tx := *global.GVA_DB
	var S = service.NewAppPromotionConfig(c, &tx)
	var info dbm.DB_AppPromotionConfig
	info, err := S.Info(请求.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var 局_数量_邀请关系, 局_数量_用户数量, 局_数量_佣金订单 int64

	switch info.PromotionType {
	default:
		response.FailWithMessage("不支持的推广类型", c)
		return
	case constant.H活动类型_cps: //cps
		//删除应用邀请关系
		局_数量_邀请关系, err = service.NewCpsInvitingRelation(c, &tx).DeleteWhere(map[string]interface{}{"inviteeAppId": info.AppId})
		//重置 用户邀请数量和金额
		局_数量_用户数量, err = service.NewCpsUser(c, &tx).UpdateWhere(map[string]interface{}{"appId": info.AppId}, map[string]interface{}{"count": 0, "cumulativeRMB": 0})
		//删除佣金订单
		局_数量_佣金订单, err = service.NewCpsPayOrder(c, &tx).DeleteWhere(map[string]interface{}{"appId": info.AppId})
	case constant.H活动类型_签到: //
		//清空用户签到分
		局_数量_用户数量, err = service.NewCheckInUser(c, &tx).UpdateWhere(map[string]interface{}{"appId": info.AppId}, map[string]interface{}{"count": 0, "cumulativeRMB": 0})
		//删除该应用的签到记录
		局_数量_用户数量, err = service.NewCheckInLog(c, &tx).DeleteWhere(map[string]interface{}{"appId": info.AppId})
		//删除该应用用户的积分记录
		局_数量_用户数量, err = service.NewCheckInScoreLog(c, &tx).DeleteWhere(map[string]interface{}{"appId": info.AppId})

	}

	if err != nil {
		response.FailWithMessage(err.Error(), c)
	} else {
		局_提示信息 := fmt.Sprintf("重置成功,删除了%d个应用邀请关系,%d个用户,%d个佣金订单", 局_数量_邀请关系, 局_数量_用户数量, 局_数量_佣金订单)

		response.OkWithDetailed(info, 局_提示信息, c)
	}

}
