package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	"server/structs/Http/response"
)

type AppInfoWebUser struct {
	Common.Common
}

func NewAppInfoWebUserController() *AppInfoWebUser {
	return &AppInfoWebUser{}
}

// 修改app排序
func (C *AppInfoWebUser) GetInfo(c *gin.Context) {
	var 请求 struct {
		Id int `json:"id"`
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	tx := *global.GVA_DB
	var S = service.NewAppInfoWebUser(c, &tx)

	info, err := S.Info(请求.Id)
	if err != nil {
		info = dbm.DB_AppInfoWebUser{
			Id:           请求.Id,
			Status:       2,
			CaptchaLogin: 3,
			UrlDownload:  "https://www.fnkuaiyan.com/",
		}
		_, _ = S.Create(info)

	}
	response.OkWithData(info, c)
	return
}
