package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
	. "server/new/app/models/response"
	"server/new/app/service"
	"server/structs/Http/response"
)

type AppPromotionConfig struct {
	Common.Common
}

func NewAppPromotionConfigController() *AppPromotionConfig {
	return &AppPromotionConfig{}
}

// GetList
func (C *AppPromotionConfig) GetList(c *gin.Context) {
	//{"AppId":2,"Type":2,"Size":10,"Page":1,"Status":1,"keywords":"1"}
	var 请求 struct {
		request.List
		AppId         int `json:"appId"`
		Status        int `json:"status"`
		PromotionType int `json:"promotionType"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	tx := *global.GVA_DB
	var S = service.NewAppPromotionConfig(c, &tx)
	var dataList []dbm.DB_AppPromotionConfig
	var 总数 int64
	var err error
	总数, dataList, err = S.GetList(请求.List, 请求.AppId, 请求.Status, 请求.PromotionType)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(GetList2{List: dataList, Count: 总数}, "操作成功", c)
	return
}

func (C *AppPromotionConfig) Info(c *gin.Context) {
	var 请求 request.Id
	if !C.ToJSON(c, &请求) {
		return
	}

	tx := *global.GVA_DB
	var S = service.NewAppPromotionConfig(c, &tx)
	var info dbm.DB_AppPromotionConfig
	info, err := S.Info(请求.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)

	} else {
		response.OkWithDetailed(info, "操作成功", c)
	}

}
