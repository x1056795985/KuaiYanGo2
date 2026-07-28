package userSafetyApi

import (
	utils2 "EFunc/utils"
	"github.com/gin-gonic/gin"
	"github.com/songzhibin97/gkit/tools/rand_string"
	"server/new/app/controller/userSafetyApi/response"
	"server/new/app/global"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	"server/new/app/utils"
	"server/new/app/utils/Qqwry"
	"strings"
	"time"
)

func UserApi_GetToken(c *gin.Context) {
	ctx := utils.Q取上下文(c)
	var DB_links_user dbm.DB_LinksToken
	DB_links_user.User = "游客"
	DB_links_user.Ip = c.ClientIP()
	省市, 运行商, err := Qqwry.Ip查信息(DB_links_user.Ip)
	if err == nil && 省市 != "" {
		DB_links_user.IPCity = 省市 + " " + 运行商
	}
	DB_links_user.Status = 1
	DB_links_user.LoginTime = time.Now().Unix()
	DB_links_user.OutTime = ctx.AppInfo.OutTime //退出时间 半小时
	DB_links_user.LastTime = DB_links_user.LoginTime

	DB_links_user.Token = strings.ToUpper(rand_string.RandomLetter(32))
	DB_links_user.LoginAppid = ctx.AppInfo.AppId       //管理员后台代号1
	DB_links_user.CryptoKeyAes = utils2.W文本_取随机字符串(24) //通讯key
	db := global.GVA_DB

	_, err = service.NewLinksToken(c, db).Create(DB_links_user)
	if err != nil {
		response.Fail(c, constant.Status_SQl错误)
		return
	}
	// 回复json结构体
	type 响应 struct {
		Token        string `json:"Token"`
		CryptoKeyAes string `json:"CryptoKeyAes"`
		IP           string `json:"IP"`
	}
	//这里吧成功的状态
	response.OkData(c, 响应{Token: DB_links_user.Token, CryptoKeyAes: DB_links_user.CryptoKeyAes, IP: c.ClientIP()})
}
