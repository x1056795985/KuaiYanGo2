package controller

import (
	"EFunc/utils"
	"crypto/sha256"
	"github.com/gin-gonic/gin"
	"server/app/controller/Common"
	"server/app/controller/Common/response"
	"server/app/global"
	checkInLogic "server/app/logic/webUser/checkIn"
	dbm "server/app/models/db"
	"server/app/service"
	"strconv"
	"time"
)

type CheckInLog struct {
	Common.Common
}

func NewCheckInLogController() *CheckInLog {
	return &CheckInLog{}
}

func (C *CheckInLog) Create(c *gin.Context) {
	var 请求 struct {
		Time int    `json:"time"  binding:"required" zh:"时间"`
		Sign string `json:"sign"  binding:"required,min=32,max=32" zh:"签名"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	var err error
	var info = struct {
		appInfo     dbm.DB_AppInfo
		likeInfo    dbm.DB_LinksToken
		checkInUser dbm.DB_CheckInUser
		最后签到信息      dbm.DB_CheckInLog
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	//查询是否拥有邀请人   如果已设置过,需要删除,因为有唯一索引
	db := *global.GVA_DB
	info.checkInUser, err = service.NewCheckInUser(c, &db).Info(info.appInfo.AppId, info.likeInfo.Uid)
	if err != nil {
		response.FailWithMessage(c, "来源错误")
		return
	}
	info.最后签到信息, err = service.NewCheckInLog(c, &db).Q取最后签到信息(info.appInfo.AppId, info.likeInfo.Uid)
	if err != nil {
		info.最后签到信息.Day = "" //没有记录 默认赋值0
	}
	局_今天唯一标记 := time.Now().Format("20060102")
	局_昨天天唯一标记 := time.Now().AddDate(0, 0, -1).Format("20060102")
	if 局_今天唯一标记 == info.最后签到信息.Day {
		response.FailWithMessage(c, "今天已签到,明日再来吧")
		return
	}
	//获取连续签到天数
	局_连续签到天数 := info.checkInUser.ContinuousDay
	局_增加签到分 := 局_连续签到天数
	if 局_昨天天唯一标记 != info.最后签到信息.Day {
		局_连续签到天数 = 0 //断签到
	}
	if 局_增加签到分 > 6 {
		局_增加签到分 = 6 //最多只能获取7分
	}
	局_增加签到分 += 1
	//签到风控算法校验
	ua := c.Request.Header.Get("User-Agent")
	ua = strconv.Itoa(info.checkInUser.ContinuousDay) + ua
	sha := sha256.New()
	sha.Write([]byte(ua))
	hashBytes := sha.Sum(nil)

	// 取前16个字节并转为十六进制字符串
	签到风控 := utils.Z字节集_字节集到十六进制(hashBytes[:16])
	if 签到风控 != 请求.Sign {
		response.FailWithMessage(c, "签到失败,时间错误.")
		return
	}

	info.checkInUser, err = checkInLogic.Q签到_执行(c, &db, info.appInfo.AppId, info.likeInfo.Uid, info.checkInUser, 局_连续签到天数, 局_增加签到分, 局_今天唯一标记)

	if err != nil {
		response.FailWithMessage(c, "签到失败,请重试.")
		return
	}

	response.OkWithData(c, gin.H{
		"score": info.checkInUser.CheckInScore,
		"count": 局_增加签到分,
	})
}
