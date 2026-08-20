package controller

import (
	. "EFunc/utils"
	"time"

	"github.com/gin-gonic/gin"

	"server/app/controller/Common"
	"server/app/controller/Common/response"
	"server/app/global"
	"server/app/logic/common/kaClassUpPrice"
	webUserCouponLogic "server/app/logic/common/webUserCoupon"
	"server/app/models/constant"
	"server/app/models/dbm"
	"server/app/service"
)

type WebUserCouponUserApi struct{ Common.Common }

func NewWebUserCouponUserControllerApi() *WebUserCouponUserApi { return &WebUserCouponUserApi{} }

func (C *WebUserCouponUserApi) GetReceiveList(c *gin.Context) {
	var info struct {
		likeInfo dbm.DB_LinksToken
		appInfo  dbm.DB_AppInfo
	}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	db := *global.GVA_DB
	var templates []dbm.DB_WebUserCoupon
	if err := db.Where("AppId = ?", info.appInfo.AppId).Order("Id DESC").Find(&templates).Error; err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	list := make([]gin.H, 0, len(templates))
	for _, item := range templates {
		canReceive, reason := webUserCouponLogic.L_webUserCoupon.M模板是否可领取(item)
		var receivedCount int64
		_ = db.Model(dbm.DB_WebUserCouponUser{}).Where("CouponId = ? AND AppId = ? AND Uid = ?", item.Id, item.AppId, info.likeInfo.Uid).Count(&receivedCount).Error
		if item.PerUserReceiveLimit > 0 && int(receivedCount) >= item.PerUserReceiveLimit {
			canReceive = false
			reason = "已达到个人领取上限"
		}
		list = append(list, gin.H{
			"id":                    item.Id,
			"name":                  item.Name,
			"type":                  item.Type,
			"couponValue":           item.CouponValue,
			"minPayAmount":          item.MinPayAmount,
			"maxDiscountAmount":     item.MaxDiscountAmount,
			"kaClassIds":            webUserCouponLogic.J解析卡类ID列表(item.KaClassIds),
			"totalCount":            item.TotalCount,
			"receivedCount":         item.ReceivedCount,
			"perUserReceiveLimit":   item.PerUserReceiveLimit,
			"receiveStartTime":      item.ReceiveStartTime,
			"receiveEndTime":        item.ReceiveEndTime,
			"useStartTime":          item.UseStartTime,
			"useEndTime":            item.UseEndTime,
			"note":                  item.Note,
			"canReceive":            canReceive,
			"receiveDisabledReason": reason,
		})
	}
	response.OkWithData(c, list)
}

func (C *WebUserCouponUserApi) Receive(c *gin.Context) {
	var req struct {
		CouponId int `json:"couponId" binding:"required"`
	}
	if !C.ToJSON(c, &req) {
		return
	}
	var info struct {
		likeInfo dbm.DB_LinksToken
		appInfo  dbm.DB_AppInfo
	}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	item, err := webUserCouponLogic.L_webUserCoupon.L领取(c, info.appInfo.AppId, info.likeInfo.Uid, req.CouponId)
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	response.OkWithDetailed(c, item, "领取成功")
}

func (C *WebUserCouponUserApi) GetMyList(c *gin.Context) {
	var req struct {
		Status int `json:"status"`
	}
	if !C.ToJSON(c, &req) {
		return
	}
	var info struct {
		likeInfo dbm.DB_LinksToken
		appInfo  dbm.DB_AppInfo
	}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	db := global.GVA_DB.Where("AppId = ? AND Uid = ?", info.appInfo.AppId, info.likeInfo.Uid)
	switch req.Status {
	case 1:
		db = db.Where("Status = ?", constant.D网页用户优惠券状态_未使用)
	case 2:
		db = db.Where("Status = ?", constant.D网页用户优惠券状态_已使用)
	case 3:
		db = db.Where("Status IN ?", []int{constant.D网页用户优惠券状态_已过期, constant.D网页用户优惠券状态_已作废})
	}
	var coupons []dbm.DB_WebUserCouponUser
	if err := db.Order("Id DESC").Find(&coupons).Error; err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	list := make([]gin.H, 0, len(coupons))
	for _, item := range coupons {
		list = append(list, gin.H{
			"id":                item.Id,
			"couponId":          item.CouponId,
			"couponName":        item.CouponName,
			"couponType":        item.CouponType,
			"couponValue":       item.CouponValue,
			"minPayAmount":      item.MinPayAmount,
			"maxDiscountAmount": item.MaxDiscountAmount,
			"kaClassIds":        webUserCouponLogic.J解析卡类ID列表(item.KaClassIds),
			"status":            item.Status,
			"statusText":        webUserCouponLogic.L_webUserCoupon.S状态文本(item.Status),
			"receiveTime":       item.ReceiveTime,
			"useStartTime":      item.UseStartTime,
			"useEndTime":        item.UseEndTime,
			"payOrder":          item.PayOrder,
			"useTime":           item.UseTime,
			"discountAmount":    item.DiscountAmount,
		})
	}
	response.OkWithData(c, list)
}

func (C *WebUserCouponUserApi) GetPayKaCouponList(c *gin.Context) {
	var req struct {
		KaClassId int `json:"kaClassId" binding:"required"`
	}
	if !C.ToJSON(c, &req) {
		return
	}
	var info struct {
		likeInfo dbm.DB_LinksToken
		appInfo  dbm.DB_AppInfo
		appUser  dbm.DB_AppUser
		kaClass  dbm.DB_KaClass
	}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	db := *global.GVA_DB
	info.kaClass, _ = service.NewKaClass(c, &db).Info(req.KaClassId)
	info.appUser, _ = service.NewAppUser(c, &db, info.appInfo.AppId).InfoUid(info.likeInfo.Uid)
	if info.kaClass.Id == 0 || info.kaClass.AppId != info.appInfo.AppId {
		response.FailWithMessage(c, "卡类不存在")
		return
	}
	originalAmount := info.kaClass.Money
	if info.appUser.Id > 0 {
		if markup, _, err := kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, info.kaClass.Id, info.appUser.AgentUid); err == nil {
			originalAmount = Float64加float64(originalAmount, markup, 2)
		}
	}
	var coupons []dbm.DB_WebUserCouponUser
	if err := db.Where("AppId = ? AND Uid = ? AND Status IN ?", info.appInfo.AppId, info.likeInfo.Uid, []int{constant.D网页用户优惠券状态_未使用, constant.D网页用户优惠券状态_已锁定}).Order("Id DESC").Find(&coupons).Error; err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	list := make([]gin.H, 0, len(coupons))
	for _, item := range coupons {
		result, _ := webUserCouponLogic.L_webUserCoupon.J计算优惠(c, item, originalAmount, req.KaClassId)
		list = append(list, gin.H{
			"id":             item.Id,
			"couponName":     item.CouponName,
			"couponType":     item.CouponType,
			"couponValue":    item.CouponValue,
			"status":         item.Status,
			"statusText":     webUserCouponLogic.L_webUserCoupon.S状态文本(item.Status),
			"canUse":         result.CanUse,
			"disabledReason": result.DisabledReason,
			"discount":       result.Discount,
			"originalAmount": originalAmount,
			"finalAmount":    result.FinalAmount,
			"useEndTime":     item.UseEndTime,
		})
	}
	response.OkWithDetailed(c, gin.H{
		"originalAmount": originalAmount,
		"list":           list,
		"serverTime":     time.Now().Unix(),
	}, "操作成功")
}
