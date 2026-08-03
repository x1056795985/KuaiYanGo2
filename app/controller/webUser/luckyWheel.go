package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"server/app/controller/Common"
	"server/app/controller/Common/response"
	"server/app/global"
	luckyWheelLogic "server/app/logic/webUser/luckyWheel"
	dbm "server/app/models/db"
	"server/app/service"
	"time"
)

type LuckyWheel struct {
	Common.Common
}

func NewLuckyWheelController() *LuckyWheel {
	return &LuckyWheel{}
}

type prizeListItem struct {
	KaClassId    int    `json:"kaClassId"`
	Probability int    `json:"probability"`
	Name         string `json:"name"`
}

// Info 获取大转盘活动信息
func (C *LuckyWheel) Info(c *gin.Context) {
	var err error
	var info = struct {
		appInfo            dbm.DB_AppInfo
		likeInfo           dbm.DB_LinksToken
		LuckyWheelInfo     dbm.DB_LuckyWheelInfo
		AppPromotionConfig dbm.DB_AppPromotionConfig
		luckyWheelUser     dbm.DB_LuckyWheelUser
		prizeList          []prizeListItem
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	tx := *global.GVA_DB

	if 临时, ok := c.Get("AppPromotionConfig"); ok {
		info.AppPromotionConfig = 临时.(dbm.DB_AppPromotionConfig)
	} else {
		response.FailWithMessage(c, "未开启大转盘活动")
		return
	}

	info.LuckyWheelInfo, err = service.NewLuckyWheelInfo(c, &tx).Info(info.AppPromotionConfig.TypeAssociatedId)
	if err != nil && err.Error() != "record not found" {
		response.FailWithMessage(c, err.Error())
		return
	}

	// 解析奖品列表,并补充卡类名称
	if info.LuckyWheelInfo.PrizeList == "" {
		info.LuckyWheelInfo.PrizeList = "[]"
	}
	_ = json.Unmarshal([]byte(info.LuckyWheelInfo.PrizeList), &info.prizeList)
	局_概率总和 := 0
	for i, v := range info.prizeList {
		局_概率总和 += v.Probability
		if v.KaClassId > 0 {
			卡类info, err2 := service.NewKaClass(c, &tx).Info(v.KaClassId)
			if err2 == nil {
				info.prizeList[i].Name = 卡类info.Name
			}
		}
	}
	// 概率不足10000,自动追加"谢谢参与"
	if 局_概率总和 < 10000 {
		info.prizeList = append(info.prizeList, prizeListItem{
			KaClassId:   0,
			Probability: 10000 - 局_概率总和,
			Name:        "谢谢参与",
		})
	}

	// 自动领取每日免费次数
	_ = luckyWheelLogic.L领取每日免费(c, &tx, info.appInfo, info.likeInfo, info.AppPromotionConfig)

	// 查询用户数据
	info.luckyWheelUser, _ = service.NewLuckyWheelUser(c, &tx).Info(info.appInfo.AppId, info.likeInfo.Uid)

	局_今日 := time.Now().Format("20060102")
	局_今日剩余免费 := 0
	if info.luckyWheelUser.DailyFreeDate != 局_今日 {
		局_今日剩余免费 = info.LuckyWheelInfo.DailyFreeCount
	} else {
		局_今日剩余免费 = info.LuckyWheelInfo.DailyFreeCount - info.luckyWheelUser.DailyFreeUsed
		if 局_今日剩余免费 < 0 {
			局_今日剩余免费 = 0
		}
	}

	response.OkWithData(c, gin.H{
		"appPromotionConfig": info.AppPromotionConfig,
		"luckyWheel":         info.LuckyWheelInfo,
		"prizeList":          info.prizeList,
		"remainCount":        info.luckyWheelUser.RemainCount,
		"totalInviteCount":   info.luckyWheelUser.TotalInviteCount,
		"todayFreeRemain":    局_今日剩余免费,
	})
	return
}

// Draw 执行抽奖
func (C *LuckyWheel) Draw(c *gin.Context) {
	var info = struct {
		appInfo            dbm.DB_AppInfo
		likeInfo           dbm.DB_LinksToken
		AppPromotionConfig dbm.DB_AppPromotionConfig
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)

	if 临时, ok := c.Get("AppPromotionConfig"); ok {
		info.AppPromotionConfig = 临时.(dbm.DB_AppPromotionConfig)
	} else {
		response.FailWithMessage(c, "未开启大转盘活动")
		return
	}

	中奖索引, 中奖奖品, err := luckyWheelLogic.L抽奖_执行(c, global.GVA_DB, info.appInfo, info.likeInfo, info.AppPromotionConfig)
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}

	response.OkWithData(c, gin.H{
		"winIndex": 中奖索引,
		"kaClassId": 中奖奖品.KaClassId,
		"name":      中奖奖品.Name,
	})
	return
}

// LogList 获取我的抽奖记录
func (C *LuckyWheel) LogList(c *gin.Context) {
	var info = struct {
		appInfo  dbm.DB_AppInfo
		likeInfo dbm.DB_LinksToken
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	tx := *global.GVA_DB

	logs, err := service.NewLuckyWheelLog(c, &tx).Infos(map[string]interface{}{
		"appId":  info.appInfo.AppId,
		"userId": info.likeInfo.Uid,
	})
	if err != nil && err.Error() != "record not found" {
		response.FailWithMessage(c, err.Error())
		return
	}

	response.OkWithData(c, gin.H{
		"list": logs,
	})
	return
}
