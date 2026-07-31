package controller

import (
	"github.com/gin-gonic/gin"
	"server/app/controller/Common"
	"server/app/controller/Common/response"
	"server/app/global"
	dbm "server/app/models/db"
	"server/app/models/request"
	. "server/app/models/response"
	"server/app/service"
)

type CheckInScoreLog struct {
	Common.Common
}

func NewCheckInScoreLogController() *CheckInScoreLog {
	return &CheckInScoreLog{}
}

func (C *CheckInScoreLog) GetList(c *gin.Context) {
	var 请求 struct {
		request.List2
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	var err error
	var info = struct {
		appInfo  dbm.DB_AppInfo
		likeInfo dbm.DB_LinksToken
		积分日志     []dbm.DB_CheckInScoreLog
		总数       int64
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	//查询是否拥有邀请人   如果已设置过,需要删除,因为有唯一索引
	db := *global.GVA_DB

	info.总数, info.积分日志, err = service.NewCheckInScoreLog(c, &db).GetList(请求.List2, info.appInfo.AppId, info.likeInfo.Uid, 0, 0)

	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}

	response.OkWithData(c, GetList2{List: info.积分日志, Count: info.总数})
}
