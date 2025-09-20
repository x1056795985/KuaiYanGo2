package cpsPayOrder

import (
	. "EFunc/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/global"
	"server/new/app/logic/common/log"
	"server/new/app/logic/webUser/cpsUser"
	"strings"

	m "server/new/app/models/common"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
	"time"
)

var L_cpsPayOrder cpsPayOrder

func init() {
	L_cpsPayOrder = cpsPayOrder{}

}

type cpsPayOrder struct {
}

// 要考虑线程安全,使用数据库 订单号唯一索引锁实现线程安全
func (j *cpsPayOrder) C处理佣金发放_线程安全(c *gin.Context, 参数 *m.PayParams) (err error) {

	//只有购卡直冲的订单才处理
	if 参数.ProcessingType != constant.D订单类型_购卡直冲 {
		return
	}
	var info struct {
		邀请人信息              DB.DB_User
		AppPromotionConfig dbm.DB_AppPromotionConfig
		cpsInfo            dbm.DB_CpsInfo
		cpsPayOrder        dbm.DB_CpsPayOrder
		上级                 dbm.DB_CpsInvitingRelation
		上上级                dbm.DB_CpsInvitingRelation
		卡类                 dbm.DB_KaClass
		cpsUser            dbm.DB_CpsUser
		有效邀请数量             int
		LogMoney           []DB.DB_LogMoney
		邀请人User            DB.DB_User
		邀请人上级User          DB.DB_User
	}
	var tx *gorm.DB
	if tempObj, ok := c.Get("tx"); ok {
		tx = tempObj.(*gorm.DB)
	} else {
		db := *global.GVA_DB
		tx = &db
	}

	_, err = service.NewCpsPayOrder(c, tx).InfoOrder(参数.PayOrder)
	if err == nil { //如果不报错,说明有值 其他线程已经处理过了, 先过滤一遍 然后再抢锁过滤
		return err
	}

	//获取该应用是否已开启了cps 可能会有多个符合时间的配置信息 只获取第一个
	数组_AppPromotionConfigs, err2 := service.NewAppPromotionConfig(c, tx).Infos(
		map[string]interface{}{
			"appId":         参数.E额外信息.Get("AppId").Int(),
			"promotionType": 1,
		})
	if err2 != nil && err2.Error() != "record not found" {
		//未配置活动
		return
	}
	局_当前时间戳 := time.Now().Unix()
	for i := range 数组_AppPromotionConfigs {
		if 数组_AppPromotionConfigs[i].StartTime < 局_当前时间戳 && 数组_AppPromotionConfigs[i].EndTime > 局_当前时间戳 {
			info.AppPromotionConfig = 数组_AppPromotionConfigs[i]
			break
		}
	}
	if info.AppPromotionConfig.Id == 0 {
		//无符合时间的活动
		return
	}
	info.cpsInfo, err = service.NewCpsInfo(c, tx).Info(info.AppPromotionConfig.TypeAssociatedId)
	if err != nil && err.Error() != "record not found" {
		global.GVA_LOG.Error("订单:"+参数.PayOrder+",佣金发放失败,获取cpsinfo"+strconv.Itoa(info.AppPromotionConfig.TypeAssociatedId)+"信息失败", zap.Any("err", err))
		return
	}

	//查询是否拥有邀请人   如果已设置过,需要删除,因为有唯一索引
	info.上级, info.上上级, err = service.NewCpsInvitingRelation(c, tx).Q取归属邀请人(参数.E额外信息.Get("AppId").Int(), 参数.Uid)
	if err != nil || info.上级.Id == 0 {
		return
	}
	//判断绑定关系是否超过了配置
	if info.上级.Status != 1 || time.Now().Unix()-info.上级.CreatedAt > info.cpsInfo.BindingDay*24*3600 {
		return
	}

	info.卡类, err = service.NewKaClass(c, tx).Info(参数.E额外信息.Get("KaClassId").Int())
	if err != nil {
		global.GVA_LOG.Error("订单:"+参数.PayOrder+",佣金发放失败,获取卡类"+strconv.Itoa(参数.E额外信息.Get("KaClassId").Int())+"信息失败0", zap.Any("err", err))
		return
	}
	//判断邀请人的级别
	info.有效邀请数量 = cpsUser.L_cpsUser.Q取有效邀请数量(c, 参数.E额外信息.Get("AppId").Int(), info.上级.InviterId)
	info.cpsUser, err = service.NewCpsUser(c, tx).Info(参数.E额外信息.Get("AppId").Int(), info.上级.InviterId)
	if err != nil {
		global.GVA_LOG.Error("订单:"+参数.PayOrder+",佣金发放失败,获取cpsUser"+strconv.Itoa(info.上级.InviterId)+"信息失败", zap.Any("err", err))
		return
	}

	if info.有效邀请数量 != info.cpsUser.Count { //更新一下缓存信息
		_, _ = service.NewCpsUser(c, tx).UpdateMap([]int{info.cpsUser.Id}, map[string]interface{}{"count": info.有效邀请数量})
	}

	//开始抢锁
	//基础信息
	info.cpsPayOrder.PayOrder = 参数.PayOrder
	info.cpsPayOrder.Time = time.Now().Unix()
	info.cpsPayOrder.AppId = 参数.E额外信息.Get("AppId").Int()
	info.cpsPayOrder.Uid = 参数.Uid
	info.cpsPayOrder.Rmb = info.卡类.Money //不能用订单的实付金额,而是用订单的卡类金额,因为代理有代理调价功能,可能导致实付金额和卡类金额不一致
	//邀请人信息
	info.cpsPayOrder.InviterId = info.上级.InviterId
	if info.有效邀请数量 >= info.cpsInfo.BronzeThreshold { //判断是否超过铜牌推广数量阈值
		info.cpsPayOrder.InviterDiscount = info.cpsInfo.BronzeKickback
	}
	if info.有效邀请数量 >= info.cpsInfo.SilverThreshold { //判断是否超过银牌推广数量阈值
		info.cpsPayOrder.InviterDiscount = info.cpsInfo.SilverKickback
	}
	if info.有效邀请数量 >= info.cpsInfo.GoldMedalThreshold { //判断是否超过金牌推广数量阈值
		info.cpsPayOrder.InviterDiscount = info.cpsInfo.GoldMedalKickback
	}

	info.cpsPayOrder.InviterRMB = Float64除int64(Float64乘int64(info.cpsPayOrder.Rmb, int64(info.cpsPayOrder.InviterDiscount)), 100, 2)
	info.cpsPayOrder.InviterStatus = constant.D订单状态_等待支付

	//上上级信息
	if info.上上级.Id != 0 {
		info.cpsPayOrder.GrandpaId = info.上上级.InviterId
		info.cpsPayOrder.GrandpaDiscount = info.cpsInfo.GrandsonKickback
		info.cpsPayOrder.GrandpaRMB = Float64除int64(Float64乘int64(info.cpsPayOrder.Rmb, int64(info.cpsPayOrder.GrandpaDiscount)), 100, 2)
		info.cpsPayOrder.GrandpaStatus = constant.D订单状态_等待支付
	}

	info.cpsPayOrder.Extra = "{}" //暂无额外信息
	//开始插入,抢到了唯一索引的线程才能继续执行
	_, err = service.NewCpsPayOrder(c, tx).Create(&info.cpsPayOrder)
	//判断是否为 唯一索引冲突导致的失败
	if err != nil && strings.Contains(err.Error(), "Error 1062") { //如果冲突,结束,因为被锁的线程已经插入了数据,这里结束
		return
	}
	if err != nil {
		global.GVA_LOG.Error("订单:"+参数.PayOrder+",佣金发放失败,插入cpsPayOrder信息失败", zap.Any("err", err))
		return
	}
	//更新支付状态 增加余额 增加累计收入缓存,都需要在事务内完成
	db := *global.GVA_DB
	err = db.Transaction(func(tx *gorm.DB) error {
		//处理邀请人的RMB增减
		if info.cpsPayOrder.InviterRMB > 0 { //只有大于0才执行
			err = tx.Model(DB.DB_User{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id=?", info.cpsPayOrder.InviterId).First(&info.邀请人User).Error
			if err != nil {
				return errors.Join(err, errors.New("邀请人账号不存在"))
			}
			err = tx.Model(DB.DB_User{}).Where("Id = ?", info.cpsPayOrder.InviterId).Update("Rmb", gorm.Expr("Rmb + ?", info.cpsPayOrder.InviterRMB)).Error
			if err != nil {
				return errors.Join(err, errors.New("邀请人支付佣金失败,请重试"))
			}
			var 局_新余额 float64
			err = tx.Model(DB.DB_User{}).Select("Rmb").Where("Id = ?", info.cpsPayOrder.InviterId).First(&局_新余额).Error
			if err != nil {
				return errors.Join(err, errors.New("邀请人支付佣金后读取新余额失败"))
			}
			//写入日志
			info.LogMoney = append(info.LogMoney, DB.DB_LogMoney{
				User:  info.邀请人User.User,
				Ip:    c.ClientIP(),
				Count: info.cpsPayOrder.InviterRMB,
				Note:  "订单:" + 参数.PayOrder + ",cps核心客户佣金|新余额≈" + Float64到文本(局_新余额, 2),
			})
			// 增加累计收入缓存
			err = tx.Model(dbm.DB_CpsUser{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("userId=?", info.cpsPayOrder.InviterId).First(&info.cpsUser).Error

			if err != nil {
				return errors.Join(err, errors.New("读取cps用户信息失败"))
			}
			info.cpsUser.CumulativeRMB = Float64加float64(info.cpsUser.CumulativeRMB, info.cpsPayOrder.InviterRMB, 2)
			err = tx.Model(dbm.DB_CpsUser{}).Where("appId=?", info.cpsPayOrder.AppId).Where("userId=?", info.cpsPayOrder.InviterId).Update("cumulativeRMB", info.cpsUser.CumulativeRMB).Error
			if err != nil {
				return errors.Join(err, errors.New("邀请人更新累计收入缓存失败"))
			}
		}
		//更新支付状态 //这里应该不用锁了, 因为只有抢到唯一索引的线程才能继续执行,这里不需要锁
		err = tx.Model(dbm.DB_CpsPayOrder{}).Where("id = ?", info.cpsPayOrder.Id).Update("inviterStatus", constant.D订单状态_成功).Error
		if err != nil {
			return errors.Join(err, errors.New("邀请人更新支付状态失败"))
		}

		return nil
	})

	if err != nil {
		//写入备注
		_, err = service.NewCpsPayOrder(c, tx).UpdateMap([]int{info.cpsPayOrder.Id}, map[string]interface{}{"Note": err.Error()})
		global.GVA_LOG.Error("订单:"+参数.PayOrder+",佣金发放失败,更新支付状态失败", zap.Any("err", err))
		return
	}

	//开始发放上上级佣金
	//更新支付状态 增加余额 增加累计收入缓存,都需要在事务内完成
	err = db.Transaction(func(tx *gorm.DB) error {
		//处理邀请人上级的RMB增减
		if info.cpsPayOrder.GrandpaRMB > 0 { //只有大于0才执行
			err = tx.Model(DB.DB_User{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id=?", info.cpsPayOrder.GrandpaId).First(&info.邀请人上级User).Error
			if err != nil {
				return errors.Join(err, errors.New("邀请人上级账号不存在"))
			}
			err = tx.Model(DB.DB_User{}).Where("Id = ?", info.cpsPayOrder.GrandpaId).Update("Rmb", Float64加float64(info.邀请人上级User.Rmb, info.cpsPayOrder.GrandpaRMB, 2)).Error
			if err != nil {
				return errors.Join(err, errors.New("邀请人上级支付佣金失败,请重试"))
			}
			var 局_新余额 float64
			err = tx.Model(DB.DB_User{}).Select("Rmb").Where("Id = ?", info.cpsPayOrder.GrandpaId).First(&局_新余额).Error
			if err != nil {
				return errors.Join(err, errors.New("邀请人上级支付佣金后读取新余额失败"))
			}
			//写入日志
			info.LogMoney = append(info.LogMoney, DB.DB_LogMoney{
				User:  info.邀请人上级User.User,
				Ip:    c.ClientIP(),
				Count: info.cpsPayOrder.GrandpaRMB,
				Note:  "订单:" + 参数.PayOrder + ",cps裂变客户佣金|新余额≈" + Float64到文本(局_新余额, 2),
			})
			// 增加累计收入缓存
			// 增加累计收入缓存
			err = tx.Model(dbm.DB_CpsUser{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("userId=?", info.cpsPayOrder.InviterId).First(&info.cpsUser).Error
			if err != nil {
				return errors.Join(err, errors.New("读取cps用户信息失败"))
			}
			info.cpsUser.CumulativeRMB = Float64加float64(info.cpsUser.CumulativeRMB, info.cpsPayOrder.InviterRMB, 2)
			err = tx.Model(dbm.DB_CpsUser{}).Where("appId=?", info.cpsPayOrder.AppId).Where("userId=?", info.cpsPayOrder.InviterId).Update("cumulativeRMB", info.cpsUser.CumulativeRMB).Error
			if err != nil {
				return errors.Join(err, errors.New("邀请人上级更新累计收入缓存失败"))
			}
		}
		//更新支付状态 //这里应该不用锁了, 因为只有抢到唯一索引的线程才能继续执行,这里不需要锁
		err = tx.Model(dbm.DB_CpsPayOrder{}).Where("id = ?", info.cpsPayOrder.Id).Update("grandpaStatus", constant.D订单状态_成功).Error
		if err != nil {
			return errors.Join(err, errors.New("邀请人上级更新支付状态失败"))
		}

		return nil
	})

	if err != nil {
		//写入备注
		_, err = service.NewCpsPayOrder(c, tx).UpdateMap([]int{info.cpsPayOrder.Id}, map[string]interface{}{"Note": err.Error()})
		global.GVA_LOG.Error("订单:"+参数.PayOrder+",佣金发放失败,更新支付状态失败", zap.Any("err", err))
		return
	}
	//输出日志
	if err = log.L_log.S输出日志(c, info.LogMoney); err != nil {
		global.GVA_LOG.Error("输出日志失败!", zap.Any("err", err))
	}
	return
}
