package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
)

type AppUser struct {
	Common.Common
}

func NewAppUserController() *AppUser {
	return &AppUser{}
}

func (C *AppUser) GetAppUserInfo(c *gin.Context) {
	var err error
	var info = struct {
		appInfo   DB.DB_AppInfo
		likeInfo  DB.DB_LinksToken
		appUser   DB.DB_AppUser
		userClass DB.DB_UserClass
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)

	tx := *global.GVA_DB

	info.appUser, err = service.NewAppUser(c, &tx, info.appInfo.AppId).Info(info.likeInfo.Uid)
	if err != nil {
		response.FailWithMessage(c, "用户不存在")
		return
	}

	info.userClass, err = service.NewUserClass(c, &tx).Info(info.appUser.UserClassId)
	if err != nil {
		info.userClass.Name = "已删类型id" + strconv.Itoa(info.appUser.UserClassId)
		info.userClass.Id = info.appUser.UserClassId
	}

	response.OkWithData(c, gin.H{
		"Uid":             info.appUser.Uid,
		"User":            info.likeInfo.User,
		"Key":             info.appUser.Key,
		"VipTime":         info.appUser.VipTime,
		"VipNumber":       info.appUser.VipNumber,
		"Status":          info.appUser.Status,
		"MaxOnline":       info.appUser.MaxOnline,
		"AgentUid":        info.appUser.AgentUid,
		"LoginTime":       info.likeInfo.LoginTime,
		"LoginIp":         info.likeInfo.Ip,
		"RegisterTime":    info.appUser.RegisterTime,
		"UserClassId":     info.appUser.UserClassId,
		"UserClassName":   info.userClass.Name,
		"UserClassMark":   info.userClass.Mark,
		"UserClassWeight": info.userClass.Weight,
	})
	return

}
