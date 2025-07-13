package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
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
		response.FailWithMessage("暂无网页用户中心配置,请点击保存初始化配置", c)
		return
	}
	response.OkWithData(info, c)
	return
}
