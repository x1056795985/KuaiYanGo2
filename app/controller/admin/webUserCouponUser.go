package controller

import (
	"github.com/gin-gonic/gin"
	"server/app/controller/Common"
	"server/app/global"
	webUserCouponLogic "server/app/logic/common/webUserCoupon"
	"server/app/models/old/response"
	"server/app/models/request"
	. "server/app/models/response"
	"server/app/service"
)

type WebUserCouponUser struct {
	Common.Common
}

func NewWebUserCouponUserController() *WebUserCouponUser {
	return &WebUserCouponUser{}
}

func (C *WebUserCouponUser) GetList(c *gin.Context) {
	var 请求 struct {
		request.List2
		AppId       int      `json:"appId"`
		Uid         int      `json:"uid"`
		Status      int      `json:"status"`
		ReceiveTime []string `json:"receiveTime"`
		UseTime     []string `json:"useTime"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	tx := *global.GVA_DB
	总数, dataList, err := service.NewWebUserCouponUser(c, &tx).GetList(request.List{
		Page: 请求.Page, Size: 请求.Size, Type: 请求.Type, Keywords: 请求.Keywords, Order: 请求.Order, Count: 请求.Count,
	}, 请求.AppId, 请求.Uid, 请求.Status, 请求.ReceiveTime, 请求.UseTime)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(GetList2{List: dataList, Count: 总数}, "操作成功", c)
}

func (C *WebUserCouponUser) Void(c *gin.Context) {
	var 请求 struct {
		Id   int    `json:"id" binding:"required"`
		Note string `json:"note"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	tx := *global.GVA_DB
	info, err := service.NewWebUserCouponUser(c, &tx).Info(请求.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if info.Status == 3 {
		response.FailWithMessage("已使用优惠券不可作废", c)
		return
	}
	if err = webUserCouponLogic.L_webUserCoupon.Z作废(c, info.AppId, info.Uid, info.Id, 请求.Note); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

type WebUserCouponLog struct {
	Common.Common
}

func NewWebUserCouponLogController() *WebUserCouponLog {
	return &WebUserCouponLog{}
}

func (C *WebUserCouponLog) GetList(c *gin.Context) {
	var 请求 struct {
		request.List2
		AppId        int      `json:"appId"`
		CouponId     int      `json:"couponId"`
		CouponUserId int      `json:"couponUserId"`
		Uid          int      `json:"uid"`
		EventType    int      `json:"eventType"`
		TimeRange    []string `json:"timeRange"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	tx := *global.GVA_DB
	总数, dataList, err := service.NewWebUserCouponLog(c, &tx).GetList(request.List{
		Page: 请求.Page, Size: 请求.Size, Type: 请求.Type, Keywords: 请求.Keywords, Order: 请求.Order, Count: 请求.Count,
	}, 请求.AppId, 请求.CouponId, 请求.CouponUserId, 请求.Uid, 请求.EventType, 请求.TimeRange)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(GetList2{List: dataList, Count: 总数}, "操作成功", c)
}
