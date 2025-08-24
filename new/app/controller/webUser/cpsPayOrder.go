package controller

import (
	"EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
)

type CpsPayOrder struct {
	Common.Common
}

func NewCpsPayOrderController() *CpsPayOrder {
	return &CpsPayOrder{}
}

func (C *CpsPayOrder) List(c *gin.Context) {
	var info = struct {
		appInfo  DB.DB_AppInfo
		likeInfo DB.DB_LinksToken
		数组好友订单   []dbm.DB_CpsPayOrder
		数组裂变订单   []dbm.DB_CpsPayOrder
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	tx := *global.GVA_DB
	info.数组好友订单, _ = service.NewCpsPayOrder(c, &tx).Q好友订单(info.appInfo.AppId, info.likeInfo.Uid, 50)
	info.数组裂变订单, _ = service.NewCpsPayOrder(c, &tx).Q裂变订单(info.appInfo.AppId, info.likeInfo.Uid, 50)

	type 订单信息简 struct {
		Order      string  `json:"order"`
		Time       int64   `json:"time"`
		Rmb        float64 `json:"rmb"`        // 订单实际付款金额佣金
		Commission float64 `json:"commission"` //佣金
	}

	var 响应data好友订单 = make([]订单信息简, 0, len(info.数组好友订单))

	for _, v := range info.数组好友订单 {
		响应data好友订单 = append(响应data好友订单, 订单信息简{
			utils.W文本_去除敏感信息(v.PayOrder),
			v.Time,
			v.Rmb,
			v.InviterRMB,
		})
	}
	var 响应data裂变订单 = make([]订单信息简, 0, len(info.数组裂变订单))

	for _, v := range info.数组裂变订单 {
		响应data裂变订单 = append(响应data裂变订单, 订单信息简{
			utils.W文本_去除敏感信息(v.PayOrder),
			v.Time,
			v.Rmb,
			v.GrandpaRMB,
		})
	}

	response.OkWithData(c, gin.H{"friend": 响应data好友订单, "fission": 响应data裂变订单})
}
