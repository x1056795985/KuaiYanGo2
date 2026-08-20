package controller

import (
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/gin-gonic/gin"

	"server/app/controller/Common"
	"server/app/global"
	webUserCouponLogic "server/app/logic/common/webUserCoupon"
	"server/app/models/constant"
	"server/app/models/dbm"
	"server/app/models/old/response"
	"server/app/models/request"
	. "server/app/models/response"
	"server/app/service"
)

type WebUserCoupon struct{ Common.Common }

func NewWebUserCouponController() *WebUserCoupon { return &WebUserCoupon{} }

func (C *WebUserCoupon) GetList(c *gin.Context) {
	var req struct {
		request.List2
		AppId      int `json:"appId"`
		Status     int `json:"status"`
		CouponType int `json:"couponType"`
	}
	if !C.ToJSON(c, &req) {
		return
	}
	db := *global.GVA_DB
	count, list, err := service.NewWebUserCoupon(c, &db).GetList(request.List{
		Page: req.Page, Size: req.Size, Type: req.Type, Keywords: req.Keywords, Order: req.Order, Count: req.Count,
	}, req.AppId, req.Status, req.CouponType)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(GetList2{List: list, Count: count}, "操作成功", c)
}

func (C *WebUserCoupon) Info(c *gin.Context) {
	var req request.Id2
	if !C.ToJSON(c, &req) {
		return
	}
	db := *global.GVA_DB
	info, err := service.NewWebUserCoupon(c, &db).Info(req.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var ids []int
	_ = json.Unmarshal([]byte(info.KaClassIds), &ids)
	response.OkWithDetailed(gin.H{"info": info, "kaClassIds": ids}, "操作成功", c)
}

func (C *WebUserCoupon) Create(c *gin.Context) {
	var req struct {
		AppId               int     `json:"appId" binding:"required"`
		Name                string  `json:"name" binding:"required"`
		Type                int     `json:"type" binding:"required"`
		CouponValue         float64 `json:"couponValue" binding:"required"`
		MinPayAmount        float64 `json:"minPayAmount"`
		MaxDiscountAmount   float64 `json:"maxDiscountAmount"`
		KaClassIds          []int   `json:"kaClassIds" binding:"required"`
		TotalCount          int     `json:"totalCount"`
		PerUserReceiveLimit int     `json:"perUserReceiveLimit"`
		ReceiveStartTime    int64   `json:"receiveStartTime" binding:"required"`
		ReceiveEndTime      int64   `json:"receiveEndTime" binding:"required"`
		UseStartTime        int64   `json:"useStartTime" binding:"required"`
		UseEndTime          int64   `json:"useEndTime" binding:"required"`
		Status              int     `json:"status"`
		Note                string  `json:"note"`
	}
	if !C.ToJSON(c, &req) {
		return
	}
	if err := validateCouponTemplate(c, req.AppId, req.Type, req.CouponValue, req.MaxDiscountAmount, req.KaClassIds, req.ReceiveStartTime, req.ReceiveEndTime, req.UseStartTime, req.UseEndTime); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.PerUserReceiveLimit <= 0 {
		req.PerUserReceiveLimit = 1
	}
	if req.Status == 0 {
		req.Status = constant.D网页用户优惠券模板状态_启用
	}
	kaIds, _ := json.Marshal(req.KaClassIds)
	db := *global.GVA_DB
	_, err := service.NewWebUserCoupon(c, &db).Create(&dbm.DB_WebUserCoupon{
		AppId:               req.AppId,
		Name:                req.Name,
		Type:                req.Type,
		CouponValue:         req.CouponValue,
		MinPayAmount:        req.MinPayAmount,
		MaxDiscountAmount:   req.MaxDiscountAmount,
		KaClassIds:          string(kaIds),
		TotalCount:          req.TotalCount,
		PerUserReceiveLimit: req.PerUserReceiveLimit,
		ReceiveStartTime:    req.ReceiveStartTime,
		ReceiveEndTime:      req.ReceiveEndTime,
		UseStartTime:        req.UseStartTime,
		UseEndTime:          req.UseEndTime,
		Status:              req.Status,
		Note:                req.Note,
		CreateTime:          time.Now().Unix(),
		UpdateTime:          time.Now().Unix(),
	})
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

func (C *WebUserCoupon) Update(c *gin.Context) {
	var req struct {
		Id                  int     `json:"id" binding:"required"`
		Name                string  `json:"name" binding:"required"`
		Type                int     `json:"type" binding:"required"`
		CouponValue         float64 `json:"couponValue" binding:"required"`
		MinPayAmount        float64 `json:"minPayAmount"`
		MaxDiscountAmount   float64 `json:"maxDiscountAmount"`
		KaClassIds          []int   `json:"kaClassIds" binding:"required"`
		TotalCount          int     `json:"totalCount"`
		PerUserReceiveLimit int     `json:"perUserReceiveLimit"`
		ReceiveStartTime    int64   `json:"receiveStartTime" binding:"required"`
		ReceiveEndTime      int64   `json:"receiveEndTime" binding:"required"`
		UseStartTime        int64   `json:"useStartTime" binding:"required"`
		UseEndTime          int64   `json:"useEndTime" binding:"required"`
		Status              int     `json:"status"`
		Note                string  `json:"note"`
	}
	if !C.ToJSON(c, &req) {
		return
	}
	db := *global.GVA_DB
	oldInfo, err := service.NewWebUserCoupon(c, &db).Info(req.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var receivedCount int64
	_ = db.Model(dbm.DB_WebUserCouponUser{}).Where("CouponId = ?", req.Id).Count(&receivedCount).Error
	if receivedCount > 0 {
		if oldInfo.Type != req.Type || oldInfo.CouponValue != req.CouponValue || oldInfo.MinPayAmount != req.MinPayAmount || oldInfo.MaxDiscountAmount != req.MaxDiscountAmount || !sameCouponKaClassIds(oldInfo.KaClassIds, req.KaClassIds) {
			response.FailWithMessage("已有领取记录后不可修改结算规则", c)
			return
		}
	}
	if err := validateCouponTemplate(c, oldInfo.AppId, req.Type, req.CouponValue, req.MaxDiscountAmount, req.KaClassIds, req.ReceiveStartTime, req.ReceiveEndTime, req.UseStartTime, req.UseEndTime); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	kaIds, _ := json.Marshal(req.KaClassIds)
	_, err = service.NewWebUserCoupon(c, &db).UpdateMap([]int{req.Id}, map[string]interface{}{
		"name":                req.Name,
		"type":                req.Type,
		"couponValue":         req.CouponValue,
		"minPayAmount":        req.MinPayAmount,
		"maxDiscountAmount":   req.MaxDiscountAmount,
		"kaClassIds":          string(kaIds),
		"totalCount":          req.TotalCount,
		"perUserReceiveLimit": req.PerUserReceiveLimit,
		"receiveStartTime":    req.ReceiveStartTime,
		"receiveEndTime":      req.ReceiveEndTime,
		"useStartTime":        req.UseStartTime,
		"useEndTime":          req.UseEndTime,
		"status":              req.Status,
		"note":                req.Note,
		"updateTime":          time.Now().Unix(),
	})
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

func (C *WebUserCoupon) SetStatus(c *gin.Context) {
	var req struct {
		Id     int `json:"id" binding:"required"`
		Status int `json:"status" binding:"required"`
	}
	if !C.ToJSON(c, &req) {
		return
	}
	if req.Status != constant.D网页用户优惠券模板状态_启用 && req.Status != constant.D网页用户优惠券模板状态_停用 {
		response.FailWithMessage("状态只能设置为启用或停用", c)
		return
	}
	db := *global.GVA_DB
	_, err := service.NewWebUserCoupon(c, &db).UpdateMap([]int{req.Id}, map[string]interface{}{"status": req.Status, "updateTime": time.Now().Unix()})
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

func (C *WebUserCoupon) Delete(c *gin.Context) {
	var req request.Id2
	if !C.ToJSON(c, &req) {
		return
	}
	db := *global.GVA_DB
	var receivedCount int64
	_ = db.Model(dbm.DB_WebUserCouponUser{}).Where("CouponId = ?", req.Id).Count(&receivedCount).Error
	if receivedCount > 0 {
		_, err := service.NewWebUserCoupon(c, &db).UpdateMap([]int{req.Id}, map[string]interface{}{"status": constant.D网页用户优惠券模板状态_停用, "updateTime": time.Now().Unix()})
		if err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
		response.OkWithMessage("已有领取记录，已改为停用", c)
		return
	}
	_, err := service.NewWebUserCoupon(c, &db).Delete(req.Id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

func (C *WebUserCoupon) BatchGrant(c *gin.Context) {
	var 请求 struct {
		AppId    int    `json:"appId" binding:"required"`
		CouponId int    `json:"couponId" binding:"required"`
		Uids     []int  `json:"uids" binding:"required"`
		Note     string `json:"note"`
		Force    bool   `json:"force"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	数量, err := webUserCouponLogic.L_webUserCoupon.F批量发放(c, 请求.AppId, 请求.CouponId, 请求.Uids, 请求.Note, 请求.Force)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"count": 数量}, "已成功发放优惠券", c)
}

func validateCouponTemplate(c *gin.Context, appId int, couponType int, couponValue float64, maxDiscount float64, kaClassIds []int, receiveStart int64, receiveEnd int64, useStart int64, useEnd int64) error {
	if len(kaClassIds) == 0 {
		return errors.New("至少选择一个卡类")
	}
	if receiveStart >= receiveEnd || useStart >= useEnd {
		return errors.New("时间范围错误")
	}
	if couponValue <= 0 {
		return errors.New("优惠值必须大于0")
	}
	if couponType == constant.D网页用户优惠券类型_折扣 && maxDiscount <= 0 {
		return errors.New("折扣券必须设置最大优惠金额")
	}
	db := *global.GVA_DB
	for _, id := range kaClassIds {
		info, err := service.NewKaClass(c, &db).Info(id)
		if err != nil || info.AppId != appId {
			return errors.New("存在不属于当前应用的卡类")
		}
	}
	return nil
}

func sameCouponKaClassIds(saved string, input []int) bool {
	var savedIds []int
	if json.Unmarshal([]byte(saved), &savedIds) != nil || len(savedIds) != len(input) {
		return false
	}
	sort.Ints(savedIds)
	inputIds := append([]int(nil), input...)
	sort.Ints(inputIds)
	for index := range savedIds {
		if savedIds[index] != inputIds[index] {
			return false
		}
	}
	return true
}
