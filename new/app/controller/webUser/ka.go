package controller

import (
	"github.com/gin-gonic/gin"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	"server/new/app/logic/common/ka"
	DB "server/structs/db"
)

type Ka struct {
	Common.Common
}

func NewKaController() *Ka {
	return &Ka{}
}

// 卡号充值
func (C *Ka) UseKa(c *gin.Context) {
	var info = struct {
		ka       DB.DB_Ka
		likeInfo DB.DB_LinksToken
		appInfo  DB.DB_AppInfo
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	var 请求 struct {
		Ka         string `json:"ka" binding:"required" zh:"卡号"` // 用户名
		InviteUser string `json:"inviteUser" `                   // 推荐人
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	var err error
	err = ka.L_ka.K卡号充值_事务(c, info.appInfo.AppId, 请求.Ka, info.likeInfo.User, 请求.InviteUser)
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	response.OkWithDetailed(c, gin.H{"InviteUser": 请求.InviteUser != ""}, "成功")
	return
}
