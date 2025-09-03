package controller

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	"server/new/app/logic/common/ka"
	"server/new/app/logic/common/log"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"time"
)

type CheckInInfo struct {
	Common.Common
}

func NewCheckInInfoController() *CheckInInfo {
	return &CheckInInfo{}
}

type cardClassListItem struct {
	Id     int    `json:"id"`
	Points int    `json:"p"`
	Name   string `json:"name"`
}

func (C *CheckInInfo) Info(c *gin.Context) {
	var err error
	var info = struct {
		appInfo             DB.DB_AppInfo
		likeInfo            DB.DB_LinksToken
		CheckInInfo         dbm.DB_CheckInInfo
		AppPromotionConfigs []dbm.DB_AppPromotionConfig
		AppPromotionConfig  dbm.DB_AppPromotionConfig
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	tx := *global.GVA_DB

	if 临时, ok := c.Get("AppPromotionConfig"); ok {
		info.AppPromotionConfig = 临时.(dbm.DB_AppPromotionConfig)
	} else {
		response.FailWithMessage(c, "未开启签到活动")
		return
	}

	info.CheckInInfo, err = service.NewCheckInInfo(c, &tx).Info(info.AppPromotionConfig.TypeAssociatedId)
	if err != nil && err.Error() != "record not found" {
		response.FailWithMessage(c, err.Error())
		return
	}

	var 响应 struct {
		Id               int                 `json:"id"  `
		CreateTime       int64               `json:"createTime" `
		UpdateTime       int64               `json:"updateTime" `
		ShareGivePoints  int                 `json:"shareGivePoints"`
		InviteGivePoints int                 `json:"inviteGivePoints"`
		CardClassList    []cardClassListItem `json:"cardClassList"`
	}
	响应.Id = info.CheckInInfo.Id
	响应.CreateTime = info.CheckInInfo.CreateTime
	响应.UpdateTime = info.CheckInInfo.UpdateTime
	响应.ShareGivePoints = info.CheckInInfo.ShareGivePoints
	响应.InviteGivePoints = info.CheckInInfo.InviteGivePoints

	var 临时CardClassList []cardClassListItem
	_ = json.Unmarshal([]byte(info.CheckInInfo.CardClassList), &临时CardClassList)
	响应.CardClassList = make([]cardClassListItem, 0, len(info.CheckInInfo.CardClassList))
	for _, v := range 临时CardClassList {
		卡类info, err2 := service.NewKaClass(c, &tx).Info(v.Id)
		if err2 != nil {
			continue
		}
		响应.CardClassList = append(响应.CardClassList, cardClassListItem{
			Id:     v.Id,
			Name:   卡类info.Name,
			Points: v.Points,
		})
	}

	response.OkWithData(c, gin.H{
		"appPromotionConfig": info.AppPromotionConfig,
		"checkIn":            响应,
	})
	return
}

func (C *CheckInInfo) RedeemReward(c *gin.Context) {
	var 请求 struct {
		Id int `json:"id"  binding:"required" zh:"兑换卡类"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	var err error
	var info = struct {
		appInfo             DB.DB_AppInfo
		likeInfo            DB.DB_LinksToken
		CheckInInfo         dbm.DB_CheckInInfo
		AppPromotionConfigs []dbm.DB_AppPromotionConfig
		AppPromotionConfig  dbm.DB_AppPromotionConfig
		KaClass             dbm.DB_KaClass
		checkInUser         dbm.DB_CheckInUser
		LogMoney            []DB.DB_LogMoney
		LogVipNumber        []DB.DB_LogVipNumber
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	db := *global.GVA_DB

	if 临时, ok := c.Get("AppPromotionConfig"); ok {
		info.AppPromotionConfig = 临时.(dbm.DB_AppPromotionConfig)
	} else {
		response.FailWithMessage(c, "未开启签到活动")
		return
	}

	info.CheckInInfo, err = service.NewCheckInInfo(c, &db).Info(info.AppPromotionConfig.TypeAssociatedId)
	if err != nil && err.Error() != "record not found" {
		response.FailWithMessage(c, err.Error())
		return
	}

	var 临时CardClassList []cardClassListItem
	_ = json.Unmarshal([]byte(info.CheckInInfo.CardClassList), &临时CardClassList)
	局_增加签到分 := 0
	//判断这个卡类是否在活动可兑换卡类里
	for _, v := range 临时CardClassList {
		if v.Id == 请求.Id {
			var 局临时 dbm.DB_KaClass
			局临时, err = service.NewKaClass(c, &db).Info(v.Id)
			if err == nil {
				info.KaClass = 局临时
				局_增加签到分 = v.Points
			}
		}
	}
	if info.KaClass.Id == 0 {
		response.FailWithMessage(c, "未找到该商品卡类")
		return
	}
	info.checkInUser, err = service.NewCheckInUser(c, &db).Info(info.appInfo.AppId, info.likeInfo.Uid)
	if err != nil {
		response.FailWithMessage(c, "用户签到分信息不存在")
		return
	}
	if info.checkInUser.CheckInScore < 局_增加签到分 {
		response.FailWithMessage(c, "用户签到分不足")
		return
	}

	//开始执行卡类充值 事务
	err = db.Transaction(func(tx *gorm.DB) error {

		// 加锁重新查签到分
		err = tx.Model(dbm.DB_CheckInUser{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id = ?", info.checkInUser.Id).First(&info.checkInUser).Error
		if err != nil {
			return err
		}

		if info.checkInUser.CheckInScore-局_增加签到分 < 0 {
			return errors.New("用户签到分不足")
		}
		//增加签到积分记录
		_, err = service.NewCheckInScoreLog(c, tx).Create(&dbm.DB_CheckInScoreLog{
			Id:           0,
			AppId:        info.appInfo.AppId,
			UserId:       info.likeInfo.Uid,
			CreatedAt:    time.Now().Unix(),
			Number:       int64(-局_增加签到分),
			Msg:          "兑换:" + info.KaClass.Name,
			NumberBefore: info.checkInUser.CheckInScore,
			NumberAfter:  info.checkInUser.CheckInScore - 局_增加签到分,
		})
		if err != nil {
			return err
		}
		_, err = service.NewCheckInUser(c, tx).UpdateMap([]int{info.checkInUser.Id}, map[string]interface{}{
			"checkInScore": info.checkInUser.CheckInScore - 局_增加签到分,
		})
		if err != nil {
			return err
		}
		//签到分日志和扣除完毕, 开始执行卡类充值
		if err = ka.L_ka.K卡类直冲_事务(c, info.KaClass.Id, info.likeInfo.Uid); err != nil {
			return err
		}
		if 临时数据, ok := c.Get("logMoney"); ok { //判断是否有rmb充值的日志
			局_临时 := 临时数据.(DB.DB_LogMoney)
			局_临时.Note = "签到分兑换," + 局_临时.Note
			info.LogMoney = append(info.LogMoney, 局_临时)
		}
		if 临时数据, ok := c.Get("logVipNumber"); ok { //判断是否有积分充值的日志
			局_临时 := 临时数据.(DB.DB_LogVipNumber)
			局_临时.Note = "签到分兑换," + 局_临时.Note
			info.LogVipNumber = append(info.LogVipNumber, 局_临时)
		}
		return err
	})
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	if err = log.L_log.S输出日志(c, info.LogMoney); err != nil {
		global.GVA_LOG.Error("输出日志失败!", zap.Any("err", err))
	}
	if err = log.L_log.S输出日志(c, info.LogVipNumber); err != nil {
		global.GVA_LOG.Error("输出日志失败!", zap.Any("err", err))
	}
	response.Ok(c)
	return
}
