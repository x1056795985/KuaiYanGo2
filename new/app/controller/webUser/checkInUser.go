package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"time"
)

type CheckInUser struct {
	Common.Common
}

func NewCheckInUserController() *CheckInUser {
	return &CheckInUser{}
}

func (C *CheckInUser) Info(c *gin.Context) {
	var err error
	var info = struct {
		appInfo     DB.DB_AppInfo
		likeInfo    DB.DB_LinksToken
		checkInUser dbm.DB_CheckInUser
		最后签到信息      dbm.DB_CheckInLog
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	//查询是否拥有邀请人   如果已设置过,需要删除,因为有唯一索引
	tx := *global.GVA_DB
	info.checkInUser, err = service.NewCheckInUser(c, &tx).Info(info.appInfo.AppId, info.likeInfo.Uid)
	//判断是否存在,如果不存在,插入默认数据
	if err != nil && err.Error() == "record not found" {
		tx = *global.GVA_DB
		info.checkInUser.UserId = info.likeInfo.Uid
		info.checkInUser.AppId = info.appInfo.AppId
		info.checkInUser.CreatedAt = time.Now().Unix()
		info.checkInUser.UpdatedAt = info.checkInUser.CreatedAt
		_, err = service.NewCheckInUser(c, &tx).Create(&info.checkInUser)
	}
	info.最后签到信息, err = service.NewCheckInLog(c, &tx).Q取最后签到信息(info.appInfo.AppId, info.likeInfo.Uid)
	if err != nil {
		//没有记录,忽略即可
	}
	var 响应 struct {
		dbm.DB_CheckInUser
		TodayCheckIn bool `json:"todayCheckIn"` //今日是否签到
	}
	响应.DB_CheckInUser = info.checkInUser
	响应.TodayCheckIn = false

	局_今天唯一标记 := time.Now().Format("20060102")
	局_昨天天唯一标记 := time.Now().AddDate(0, 0, -1).Format("20060102")

	if 局_今天唯一标记 == info.最后签到信息.Day {
		响应.TodayCheckIn = true
	}

	// 如果今天未签到,判断昨天是否签到
	if 响应.TodayCheckIn == false {
		//判断昨天或今天是否签到,
		if 局_昨天天唯一标记 != info.最后签到信息.Day {
			响应.ContinuousDay = 0 //断签到
			_, err = service.NewCheckInUser(c, &tx).UpdateMap([]int{info.checkInUser.Id}, map[string]interface{}{"continuousDay": 0})
		}
	}

	response.OkWithData(c, 响应)
}
