package webUserCoupon

import (
	. "EFunc/utils"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/app/global"
	"server/app/models/constant"
	"server/app/models/dbm"
	"server/app/service"
	"strconv"
	"time"
)

var L_webUserCoupon webUserCoupon

func init() {
	L_webUserCoupon = webUserCoupon{}
}

type webUserCoupon struct{}

type 优惠券计算结果 struct {
	CanUse         bool    `json:"canUse"`
	DisabledReason string  `json:"disabledReason"`
	Discount       float64 `json:"discount"`
	FinalAmount    float64 `json:"finalAmount"`
	TemplateName   string  `json:"templateName"`
	StatusText     string  `json:"statusText"`
}

func 当前数据库(c *gin.Context) *gorm.DB {
	if c != nil {
		if tx, ok := c.Get("tx"); ok {
			if db, ok2 := tx.(*gorm.DB); ok2 {
				return db
			}
		}
	}
	return global.GVA_DB
}

func 解析卡类ID列表(文本 string) (结果 []int) {
	if 文本 == "" {
		return []int{}
	}
	_ = json.Unmarshal([]byte(文本), &结果)
	return
}

func J解析卡类ID列表(文本 string) []int {
	return 解析卡类ID列表(文本)
}

func 卡类是否适用(卡类ID int, 允许列表 string) bool {
	ids := 解析卡类ID列表(允许列表)
	if len(ids) == 0 {
		return false
	}
	for _, v := range ids {
		if v == 卡类ID {
			return true
		}
	}
	return false
}

func 计算优惠金额(模板 dbm.DB_WebUserCoupon, 原价 float64) float64 {
	if 原价 <= 0 {
		return 0
	}
	var 优惠 float64
	switch 模板.Type {
	case constant.D网页用户优惠券类型_满减:
		优惠 = 模板.CouponValue
	case constant.D网页用户优惠券类型_折扣:
		优惠 = 原价 * (100 - 模板.CouponValue) / 100
		if 模板.MaxDiscountAmount > 0 && 优惠 > 模板.MaxDiscountAmount {
			优惠 = 模板.MaxDiscountAmount
		}
	}
	if 优惠 > 原价-0.01 {
		优惠 = 原价 - 0.01
	}
	return Float64加float64(优惠, 0, 2)
}

func (j *webUserCoupon) J计算优惠(c *gin.Context, couponUser dbm.DB_WebUserCouponUser, 原价 float64, 卡类ID int) (结果 优惠券计算结果, err error) {
	结果.FinalAmount = Float64加float64(原价, 0, 2)
	if couponUser.Id == 0 {
		结果.DisabledReason = "未选择优惠券"
		return
	}
	结果.TemplateName = couponUser.CouponName
	db := 当前数据库(c)
	var 模板 dbm.DB_WebUserCoupon
	if err = db.Model(dbm.DB_WebUserCoupon{}).Where("Id = ?", couponUser.CouponId).First(&模板).Error; err != nil {
		结果.DisabledReason = "优惠券模板不存在"
		return
	}
	模板.Type = couponUser.CouponType
	模板.CouponValue = couponUser.CouponValue
	模板.MinPayAmount = couponUser.MinPayAmount
	模板.MaxDiscountAmount = couponUser.MaxDiscountAmount
	if 模板.Status != constant.D网页用户优惠券模板状态_启用 {
		结果.DisabledReason = "模板已停用"
		return
	}
	now := time.Now().Unix()
	if 模板.ReceiveStartTime > 0 && 模板.ReceiveStartTime > now {
		结果.DisabledReason = "未开始"
		return
	}
	if 模板.ReceiveEndTime > 0 && 模板.ReceiveEndTime < now {
		结果.DisabledReason = "已结束"
		return
	}
	if 模板.UseStartTime > 0 && 模板.UseStartTime > now {
		结果.DisabledReason = "未到使用时间"
		return
	}
	if 模板.UseEndTime > 0 && 模板.UseEndTime < now {
		结果.DisabledReason = "已过期"
		return
	}
	if couponUser.UseStartTime > 0 && couponUser.UseStartTime > now {
		结果.DisabledReason = "未到使用时间"
		return
	}
	if couponUser.UseEndTime > 0 && couponUser.UseEndTime < now {
		结果.DisabledReason = "已过期"
		return
	}
	if !卡类是否适用(卡类ID, couponUser.KaClassIds) {
		结果.DisabledReason = "不适用于当前卡类"
		return
	}
	if 原价 < couponUser.MinPayAmount {
		结果.DisabledReason = "未满足满减门槛"
		return
	}
	优惠 := 计算优惠金额(模板, 原价)
	if 优惠 <= 0 {
		结果.DisabledReason = "优惠金额为0"
		return
	}
	结果.CanUse = true
	结果.Discount = 优惠
	结果.FinalAmount = Float64加float64(原价-优惠, 0, 2)
	结果.StatusText = j.状态文本(couponUser.Status)
	return
}

func (j *webUserCoupon) 状态文本(status int) string {
	switch status {
	case constant.D网页用户优惠券状态_未使用:
		return "未使用"
	case constant.D网页用户优惠券状态_已锁定:
		return "已锁定"
	case constant.D网页用户优惠券状态_已使用:
		return "已使用"
	case constant.D网页用户优惠券状态_已过期:
		return "已过期"
	case constant.D网页用户优惠券状态_已作废:
		return "已作废"
	}
	return "未知"
}

func (j *webUserCoupon) S状态文本(status int) string {
	return j.状态文本(status)
}

func (j *webUserCoupon) 事件类型文本(eventType int) string {
	switch eventType {
	case constant.D网页用户优惠券流水类型_领取:
		return "领取"
	case constant.D网页用户优惠券流水类型_锁定:
		return "锁定"
	case constant.D网页用户优惠券流水类型_释放:
		return "释放"
	case constant.D网页用户优惠券流水类型_使用:
		return "使用"
	case constant.D网页用户优惠券流水类型_过期:
		return "过期"
	case constant.D网页用户优惠券流水类型_作废:
		return "作废"
	}
	return "未知"
}

func (j *webUserCoupon) 模板是否可领取(模板 dbm.DB_WebUserCoupon) (bool, string) {
	now := time.Now().Unix()
	if 模板.Status != constant.D网页用户优惠券模板状态_启用 {
		return false, "模板已停用"
	}
	if 模板.TotalCount > 0 && 模板.ReceivedCount >= 模板.TotalCount {
		return false, "已领完"
	}
	if 模板.ReceiveStartTime > 0 && 模板.ReceiveStartTime > now {
		return false, "未开始"
	}
	if 模板.ReceiveEndTime > 0 && 模板.ReceiveEndTime < now {
		return false, "已结束"
	}
	if 模板.UseEndTime > 0 && 模板.UseEndTime < now {
		return false, "已过期"
	}
	return true, ""
}

func (j *webUserCoupon) M模板是否可领取(模板 dbm.DB_WebUserCoupon) (bool, string) {
	return j.模板是否可领取(模板)
}

func (j *webUserCoupon) 领取(c *gin.Context, appId, uid, couponId int) (用户券 dbm.DB_WebUserCouponUser, err error) {
	db := 当前数据库(c)
	err = db.Transaction(func(tx *gorm.DB) error {
		var 模板 dbm.DB_WebUserCoupon
		if err2 := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id = ? AND AppId = ?", couponId, appId).First(&模板).Error; err2 != nil {
			return err2
		}
		可用, 原因 := j.模板是否可领取(模板)
		if !可用 {
			return errors.New(原因)
		}
		var 已领取数 int64
		_ = tx.Model(dbm.DB_WebUserCouponUser{}).Where("CouponId = ? AND Uid = ? AND AppId = ?", couponId, uid, appId).Count(&已领取数).Error
		if 模板.PerUserReceiveLimit > 0 && int(已领取数) >= 模板.PerUserReceiveLimit {
			return errors.New("已超过个人领取上限")
		}
		用户券 = dbm.DB_WebUserCouponUser{
			CouponId:          模板.Id,
			AppId:             appId,
			Uid:               uid,
			CouponName:        模板.Name,
			CouponType:        模板.Type,
			CouponValue:       模板.CouponValue,
			MinPayAmount:      模板.MinPayAmount,
			MaxDiscountAmount: 模板.MaxDiscountAmount,
			KaClassIds:        模板.KaClassIds,
			Status:            constant.D网页用户优惠券状态_未使用,
			ReceiveTime:       time.Now().Unix(),
			UseStartTime:      模板.UseStartTime,
			UseEndTime:        模板.UseEndTime,
			Source:            1,
			Note:              "用户领取",
		}
		if _, err2 := service.NewWebUserCouponUser(c, tx).Create(&用户券); err2 != nil {
			return err2
		}
		if _, err2 := service.NewWebUserCoupon(c, tx).UpdateMap([]int{模板.Id}, map[string]interface{}{
			"receivedCount": gorm.Expr("receivedCount + ?", 1),
			"updateTime":    time.Now().Unix(),
		}); err2 != nil {
			return err2
		}
		_, _, err2 := j.记录流水(tx, c, appId, uid, 模板.Id, 用户券.Id, constant.D网页用户优惠券流水类型_领取, "", 0, "user", "用户领取")
		return err2
	})
	return
}

func (j *webUserCoupon) L领取(c *gin.Context, appId, uid, couponId int) (dbm.DB_WebUserCouponUser, error) {
	return j.领取(c, appId, uid, couponId)
}

func (j *webUserCoupon) F批量发放(c *gin.Context, appId, couponId int, uids []int, note string, force bool) (int, error) {
	if appId <= 0 || couponId <= 0 {
		return 0, errors.New("应用和优惠券不能为空")
	}
	if len(uids) == 0 {
		return 0, errors.New("至少选择一名用户")
	}

	局_唯一Uid := make([]int, 0, len(uids))
	局_已存在Uid := make(map[int]bool, len(uids))
	for _, uid := range uids {
		if uid <= 0 || 局_已存在Uid[uid] {
			continue
		}
		局_已存在Uid[uid] = true
		局_唯一Uid = append(局_唯一Uid, uid)
	}
	if len(局_唯一Uid) == 0 {
		return 0, errors.New("至少选择一名有效用户")
	}

	局_已发放数量 := 0
	局_数据库 := 当前数据库(c)
	err := 局_数据库.Transaction(func(tx *gorm.DB) error {
		var 局_模板 dbm.DB_WebUserCoupon
		if err2 := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id = ? AND AppId = ?", couponId, appId).First(&局_模板).Error; err2 != nil {
			return errors.New("优惠券不存在或不属于当前应用")
		}
		局_可用, 局_原因 := j.模板是否可领取(局_模板)
		if !局_可用 {
			return errors.New("优惠券当前不可发放：" + 局_原因)
		}
		if 局_模板.TotalCount > 0 && 局_模板.ReceivedCount+len(局_唯一Uid) > 局_模板.TotalCount {
			return errors.New("优惠券剩余库存不足")
		}

		局_操作人 := service.NewAdmin(c, tx).Id取User(c.GetInt("Uid"))
		if 局_操作人 == "" {
			局_操作人 = "admin"
		}
		局_说明 := "管理员发放"
		if note != "" {
			局_说明 += "：" + note
		}
		局_现在 := time.Now().Unix()
		for _, uid := range 局_唯一Uid {
			if _, err2 := service.NewAppUser(c, tx, appId).InfoUid(uid); err2 != nil {
				return errors.New("用户UID不存在：" + strconv.Itoa(uid))
			}
			var 局_已领取数量 int64
			if err2 := tx.Model(dbm.DB_WebUserCouponUser{}).Where("CouponId = ? AND AppId = ? AND Uid = ?", couponId, appId, uid).Count(&局_已领取数量).Error; err2 != nil {
				return err2
			}
			if !force && 局_模板.PerUserReceiveLimit > 0 && int(局_已领取数量) >= 局_模板.PerUserReceiveLimit {
				return errors.New("用户UID " + strconv.Itoa(uid) + " 已达到该优惠券的领取上限")
			}

			局_用户券 := dbm.DB_WebUserCouponUser{
				CouponId:          局_模板.Id,
				AppId:             appId,
				Uid:               uid,
				CouponName:        局_模板.Name,
				CouponType:        局_模板.Type,
				CouponValue:       局_模板.CouponValue,
				MinPayAmount:      局_模板.MinPayAmount,
				MaxDiscountAmount: 局_模板.MaxDiscountAmount,
				KaClassIds:        局_模板.KaClassIds,
				Status:            constant.D网页用户优惠券状态_未使用,
				ReceiveTime:       局_现在,
				UseStartTime:      局_模板.UseStartTime,
				UseEndTime:        局_模板.UseEndTime,
				Source:            2,
				Note:              局_说明,
			}
			if _, err2 := service.NewWebUserCouponUser(c, tx).Create(&局_用户券); err2 != nil {
				return err2
			}
			if _, err2 := service.NewWebUserCouponLog(c, tx).Create(&dbm.DB_WebUserCouponLog{
				AppId:        appId,
				CouponId:     局_模板.Id,
				CouponUserId: 局_用户券.Id,
				Uid:          uid,
				EventType:    constant.D网页用户优惠券流水类型_领取,
				Operator:     局_操作人,
				Time:         局_现在,
				Note:         局_说明,
			}); err2 != nil {
				return err2
			}
			局_已发放数量++
		}
		return tx.Model(dbm.DB_WebUserCoupon{}).Where("Id = ?", 局_模板.Id).Updates(map[string]interface{}{
			"receivedCount": gorm.Expr("receivedCount + ?", 局_已发放数量),
			"updateTime":    局_现在,
		}).Error
	})
	if err != nil {
		return 0, err
	}
	return 局_已发放数量, nil
}

func (j *webUserCoupon) 锁定(c *gin.Context, appId, uid, couponUserId, kaClassId int, 原价 float64, payOrder string) (用户券 dbm.DB_WebUserCouponUser, 优惠 float64, err error) {
	db := 当前数据库(c)
	err = db.Transaction(func(tx *gorm.DB) error {
		if err2 := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id = ? AND AppId = ? AND Uid = ?", couponUserId, appId, uid).First(&用户券).Error; err2 != nil {
			return err2
		}
		if 用户券.Status != constant.D网页用户优惠券状态_未使用 {
			return errors.New("优惠券不可用")
		}
		结果, err2 := j.J计算优惠(c, 用户券, 原价, kaClassId)
		if err2 != nil {
			return err2
		}
		if !结果.CanUse {
			return errors.New(结果.DisabledReason)
		}
		优惠 = 结果.Discount
		用户券.Status = constant.D网页用户优惠券状态_已锁定
		用户券.LockTime = time.Now().Unix()
		用户券.LockExpireTime = time.Now().Unix() + constant.D网页用户优惠券锁定秒数
		用户券.PayOrder = payOrder
		用户券.DiscountAmount = 优惠
		if err2 := tx.Model(dbm.DB_WebUserCouponUser{}).Where("Id = ? AND Status = ?", 用户券.Id, constant.D网页用户优惠券状态_未使用).Updates(map[string]interface{}{
			"status":         constant.D网页用户优惠券状态_已锁定,
			"lockTime":       用户券.LockTime,
			"lockExpireTime": 用户券.LockExpireTime,
			"payOrder":       payOrder,
			"discountAmount": 优惠,
		}).Error; err2 != nil {
			return err2
		}
		_, _, err2 = j.记录流水(tx, c, appId, uid, 用户券.CouponId, 用户券.Id, constant.D网页用户优惠券流水类型_锁定, payOrder, 优惠, "user", "下单锁定")
		return err2
	})
	return
}

func (j *webUserCoupon) S锁定(c *gin.Context, appId, uid, couponUserId, kaClassId int, 原价 float64, payOrder string) (dbm.DB_WebUserCouponUser, float64, error) {
	return j.锁定(c, appId, uid, couponUserId, kaClassId, 原价, payOrder)
}

func (j *webUserCoupon) 绑定订单(c *gin.Context, appId, uid, couponUserId int, payOrder string) error {
	if couponUserId <= 0 || payOrder == "" {
		return nil
	}
	db := 当前数据库(c)
	return db.Model(dbm.DB_WebUserCouponUser{}).Where("Id = ? AND AppId = ? AND Uid = ? AND Status = ?", couponUserId, appId, uid, constant.D网页用户优惠券状态_已锁定).Update("PayOrder", payOrder).Error
}

func (j *webUserCoupon) B绑定订单(c *gin.Context, appId, uid, couponUserId int, payOrder string) error {
	return j.绑定订单(c, appId, uid, couponUserId, payOrder)
}

func (j *webUserCoupon) 支付成功后处理(c *gin.Context, appId, uid, couponUserId int, payOrder string, discount float64) error {
	if couponUserId <= 0 {
		return nil
	}
	db := 当前数据库(c)
	return db.Transaction(func(tx *gorm.DB) error {
		var 用户券 dbm.DB_WebUserCouponUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id = ? AND AppId = ? AND Uid = ?", couponUserId, appId, uid).First(&用户券).Error; err != nil {
			return err
		}
		if 用户券.Status != constant.D网页用户优惠券状态_已锁定 {
			return errors.New("优惠券未处于锁定状态")
		}
		if 用户券.PayOrder != "" && 用户券.PayOrder != payOrder {
			return errors.New("优惠券关联订单不匹配")
		}
		_, _, err := j.记录流水(tx, c, appId, uid, 用户券.CouponId, couponUserId, constant.D网页用户优惠券流水类型_使用, payOrder, discount, "system", "订单支付成功")
		return err
	})
}

func (j *webUserCoupon) Z支付成功后处理(c *gin.Context, appId, uid, couponUserId int, payOrder string, discount float64) error {
	return j.支付成功后处理(c, appId, uid, couponUserId, payOrder, discount)
}

func (j *webUserCoupon) 释放锁定(c *gin.Context, appId, uid, couponUserId int, payOrder string, note string) error {
	if couponUserId <= 0 {
		return nil
	}
	db := 当前数据库(c)
	return db.Transaction(func(tx *gorm.DB) error {
		_, _, err := j.记录流水(tx, c, appId, uid, 0, couponUserId, constant.D网页用户优惠券流水类型_释放, payOrder, 0, "system", note)
		return err
	})
}

func (j *webUserCoupon) S释放锁定(c *gin.Context, appId, uid, couponUserId int, payOrder string, note string) error {
	return j.释放锁定(c, appId, uid, couponUserId, payOrder, note)
}

func (j *webUserCoupon) 作废(c *gin.Context, appId, uid, couponUserId int, note string) error {
	db := 当前数据库(c)
	return db.Transaction(func(tx *gorm.DB) error {
		_, _, err := j.记录流水(tx, c, appId, uid, 0, couponUserId, constant.D网页用户优惠券流水类型_作废, "", 0, "admin", note)
		return err
	})
}

func (j *webUserCoupon) Z作废(c *gin.Context, appId, uid, couponUserId int, note string) error {
	return j.作废(c, appId, uid, couponUserId, note)
}

func (j *webUserCoupon) 过期处理(c *gin.Context) error {
	if global.GVA_DB == nil {
		return nil
	}
	db := 当前数据库(c)
	now := time.Now().Unix()
	return db.Transaction(func(tx *gorm.DB) error {
		var 锁定券 []dbm.DB_WebUserCouponUser
		if err := tx.Where("Status = ? AND LockExpireTime > 0 AND LockExpireTime < ?", constant.D网页用户优惠券状态_已锁定, now).Find(&锁定券).Error; err != nil {
			return err
		}
		for _, v := range 锁定券 {
			if v.PayOrder != "" {
				var 订单 dbm.DB_LogRMBPayOrder
				if err := tx.Where("PayOrder = ?", v.PayOrder).First(&订单).Error; err == nil {
					if 订单.Status == constant.D订单状态_已付待处理 || 订单.Status == constant.D订单状态_成功 {
						continue
					}
				}
			}
			_, _, _ = j.记录流水(tx, c, v.AppId, v.Uid, v.CouponId, v.Id, constant.D网页用户优惠券流水类型_释放, v.PayOrder, 0, "system", "锁定超时释放")
		}
		var 过期券 []dbm.DB_WebUserCouponUser
		if err := tx.Where("Status = ? AND UseEndTime > 0 AND UseEndTime < ?", constant.D网页用户优惠券状态_未使用, now).Find(&过期券).Error; err != nil {
			return err
		}
		for _, v := range 过期券 {
			_, _, _ = j.记录流水(tx, c, v.AppId, v.Uid, v.CouponId, v.Id, constant.D网页用户优惠券流水类型_过期, "", 0, "system", "使用期过期")
		}
		return nil
	})
}

func (j *webUserCoupon) G过期处理(c *gin.Context) error {
	return j.过期处理(c)
}

func (j *webUserCoupon) 记录流水(tx *gorm.DB, c *gin.Context, appId, uid, couponId, couponUserId, eventType int, payOrder string, amount float64, operator, note string) (模板 dbm.DB_WebUserCoupon, 用户券 dbm.DB_WebUserCouponUser, err error) {
	if tx == nil {
		tx = 当前数据库(c)
	}
	switch eventType {
	case constant.D网页用户优惠券流水类型_领取:
		if err = tx.Model(dbm.DB_WebUserCouponUser{}).Where("Id = ?", couponUserId).First(&用户券).Error; err != nil {
			return
		}
		if err = tx.Model(dbm.DB_WebUserCoupon{}).Where("Id = ?", couponId).First(&模板).Error; err != nil {
			return
		}
	case constant.D网页用户优惠券流水类型_锁定, constant.D网页用户优惠券流水类型_释放, constant.D网页用户优惠券流水类型_使用, constant.D网页用户优惠券流水类型_过期, constant.D网页用户优惠券流水类型_作废:
		if err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id = ? AND AppId = ? AND Uid = ?", couponUserId, appId, uid).First(&用户券).Error; err != nil {
			return
		}
		couponId = 用户券.CouponId
		if couponId > 0 {
			_ = tx.Model(dbm.DB_WebUserCoupon{}).Where("Id = ?", couponId).First(&模板).Error
		}
	}
	switch eventType {
	case constant.D网页用户优惠券流水类型_领取:
		if 用户券.Id == 0 {
			用户券 = dbm.DB_WebUserCouponUser{
				CouponId:          模板.Id,
				AppId:             appId,
				Uid:               uid,
				CouponName:        模板.Name,
				CouponType:        模板.Type,
				CouponValue:       模板.CouponValue,
				MinPayAmount:      模板.MinPayAmount,
				MaxDiscountAmount: 模板.MaxDiscountAmount,
				KaClassIds:        模板.KaClassIds,
				Status:            constant.D网页用户优惠券状态_未使用,
				ReceiveTime:       time.Now().Unix(),
				UseStartTime:      模板.UseStartTime,
				UseEndTime:        模板.UseEndTime,
				Source:            1,
				Note:              note,
			}
			if _, err = service.NewWebUserCouponUser(c, tx).Create(&用户券); err != nil {
				return
			}
			if _, err = service.NewWebUserCoupon(c, tx).UpdateMap([]int{模板.Id}, map[string]interface{}{
				"receivedCount": gorm.Expr("receivedCount + ?", 1),
				"updateTime":    time.Now().Unix(),
			}); err != nil {
				return
			}
		}
	case constant.D网页用户优惠券流水类型_锁定:
		if 用户券.Status != constant.D网页用户优惠券状态_已锁定 {
			if err = tx.Model(dbm.DB_WebUserCouponUser{}).Where("Id = ? AND Status = ?", 用户券.Id, constant.D网页用户优惠券状态_未使用).Updates(map[string]interface{}{
				"status":         constant.D网页用户优惠券状态_已锁定,
				"lockTime":       time.Now().Unix(),
				"lockExpireTime": time.Now().Unix() + constant.D网页用户优惠券锁定秒数,
				"payOrder":       payOrder,
				"discountAmount": amount,
			}).Error; err != nil {
				return
			}
			用户券.Status = constant.D网页用户优惠券状态_已锁定
			用户券.LockTime = time.Now().Unix()
			用户券.LockExpireTime = time.Now().Unix() + constant.D网页用户优惠券锁定秒数
			用户券.PayOrder = payOrder
			用户券.DiscountAmount = amount
		}
	case constant.D网页用户优惠券流水类型_释放:
		if 用户券.Status == constant.D网页用户优惠券状态_已锁定 {
			if 用户券.PayOrder != "" && payOrder != "" && 用户券.PayOrder != payOrder {
				return
			}
			if err = tx.Model(dbm.DB_WebUserCouponUser{}).Where("Id = ?", 用户券.Id).Updates(map[string]interface{}{
				"status":         constant.D网页用户优惠券状态_未使用,
				"payOrder":       "",
				"lockTime":       0,
				"lockExpireTime": 0,
				"discountAmount": 0,
			}).Error; err != nil {
				return
			}
			用户券.Status = constant.D网页用户优惠券状态_未使用
		}
	case constant.D网页用户优惠券流水类型_使用:
		if 用户券.Status == constant.D网页用户优惠券状态_已锁定 {
			if 用户券.PayOrder != "" && payOrder != "" && 用户券.PayOrder != payOrder {
				return
			}
			if err = tx.Model(dbm.DB_WebUserCouponUser{}).Where("Id = ? AND Status = ?", 用户券.Id, constant.D网页用户优惠券状态_已锁定).Updates(map[string]interface{}{
				"status":         constant.D网页用户优惠券状态_已使用,
				"payOrder":       payOrder,
				"useTime":        time.Now().Unix(),
				"discountAmount": amount,
			}).Error; err != nil {
				return
			}
			用户券.Status = constant.D网页用户优惠券状态_已使用
			用户券.UseTime = time.Now().Unix()
			用户券.PayOrder = payOrder
			用户券.DiscountAmount = amount
			if couponId > 0 {
				_, err = service.NewWebUserCoupon(c, tx).UpdateMap([]int{couponId}, map[string]interface{}{
					"usedCount":  gorm.Expr("usedCount + ?", 1),
					"updateTime": time.Now().Unix(),
				})
				if err != nil {
					return
				}
			}
		}
	case constant.D网页用户优惠券流水类型_过期:
		if 用户券.Status == constant.D网页用户优惠券状态_未使用 || 用户券.Status == constant.D网页用户优惠券状态_已锁定 {
			if err = tx.Model(dbm.DB_WebUserCouponUser{}).Where("Id = ?", 用户券.Id).Updates(map[string]interface{}{
				"status":         constant.D网页用户优惠券状态_已过期,
				"payOrder":       "",
				"lockTime":       0,
				"lockExpireTime": 0,
			}).Error; err != nil {
				return
			}
			用户券.Status = constant.D网页用户优惠券状态_已过期
		}
	case constant.D网页用户优惠券流水类型_作废:
		if 用户券.Status == constant.D网页用户优惠券状态_未使用 || 用户券.Status == constant.D网页用户优惠券状态_已锁定 {
			if err = tx.Model(dbm.DB_WebUserCouponUser{}).Where("Id = ?", 用户券.Id).Updates(map[string]interface{}{
				"status":         constant.D网页用户优惠券状态_已作废,
				"payOrder":       "",
				"lockTime":       0,
				"lockExpireTime": 0,
			}).Error; err != nil {
				return
			}
			用户券.Status = constant.D网页用户优惠券状态_已作废
		}
	}
	_, _ = service.NewWebUserCouponLog(c, tx).Create(&dbm.DB_WebUserCouponLog{
		AppId:        appId,
		CouponId:     couponId,
		CouponUserId: 用户券.Id,
		Uid:          uid,
		EventType:    eventType,
		PayOrder:     payOrder,
		Amount:       amount,
		Operator:     operator,
		Time:         time.Now().Unix(),
		Note:         note,
	})
	return
}
