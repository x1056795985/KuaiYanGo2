package rmbPay

import (
	. "EFunc/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/encoding/gjson"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	App服务 "server/Service/Ser_AppInfo"
	"server/global"
	"server/new/app/logic/agent/L_setting"
	"server/new/app/logic/common/agent"
	"server/new/app/logic/common/ka"
	"server/new/app/logic/common/log"
	"server/new/app/logic/common/setting"
	"server/new/app/logic/webUser/cpsPayOrder"
	m "server/new/app/models/common"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"server/utils/Qqwry"
	"strconv"
	"sync"
	"time"
)

var L_rmbPay rmbPay

func init() {
	L_rmbPay = rmbPay{
		已注册通道: make(map[string]RmbPayItem, 10),
		Map订单类型: map[int]string{
			constant.D订单类型_余额充值: "余额充值",
			constant.D订单类型_购卡直冲: "购卡直冲",
			constant.D订单类型_积分充值: "积分充值",
			constant.D订单类型_支付购卡: "支付购卡",
		},
	}

}

type rmbPay struct {
	// 逻辑中使用的某个变量
	订单号计数  int
	订单号初始值 int
	订单号时间戳 int64
	// 与变量对应的使用互斥锁
	锁       sync.Mutex
	已注册通道   map[string]RmbPayItem
	Map订单类型 map[int]string
}

// 注册通道接口
type RmbPayItem interface {
	D订单创建(c *gin.Context, 参数 *m.PayParams) (req m.Request, err error)
	D订单退款(c *gin.Context, 参数 *m.PayParams) (err error)
	D订单支付回调(c *gin.Context, 参数 *m.PayParams) (响应信息 string, 响应代码 int, err error)
	D订单退款回调(c *gin.Context, 参数 *m.PayParams) (响应信息 string, 响应代码 int, err error)
	Q取通道名称() string
	Q取订单id(c *gin.Context, 参数 *m.PayParams) string
}

func (j *rmbPay) Z注册接口(通道 RmbPayItem) {
	j.已注册通道[通道.Q取通道名称()] = 通道
}

func (j *rmbPay) D订单创建(c *gin.Context, 参数 m.PayParams) (req m.Request, err error) {

	参数.Z支付配置s = setting.Q在线支付配置()
	参数.Z支付配置, _ = json.Marshal(&参数.Z支付配置s)

	参数.Type = j.Pay_显示名称转原名(参数.Type)

	局_通道, ok := j.已注册通道[参数.Type]
	if !ok {
		err = errors.New("支付方式未配置")
		return
	}

	if 参数.Rmb <= 0 {
		err = errors.New("支付金额必须大于0")
		return
	}
	参数.PayOrder = j.获取新订单号()

	参数.Y异步回调地址 = setting.Q系统设置().X系统地址 + "/webApi/payNotify/" + 参数.PayOrder

	switch 参数.UidType {
	default:
		err = errors.New("UidType错误")
		return
	case 1:
		tx := *global.GVA_DB
		if info, err2 := service.NewUser(c, &tx).Info(参数.Uid); err2 == nil {
			参数.User = info.User
		}
	case 2:
		tx := *global.GVA_DB
		if info, err2 := service.NewKa(c, &tx).Info(参数.Uid); err2 == nil {
			参数.User = info.Name
		}
	}

	if 参数.User == "" {
		if 参数.ProcessingType == constant.D订单类型_余额充值 || 参数.ProcessingType == constant.D订单类型_购卡直冲 {
			err = errors.New("用户名不能为空")
			return
		}
	}
	tx := *global.GVA_DB
	var 局_通道数据 m.Request

	参数.S商品名称 = App服务.AppId取应用名称(参数.E额外信息.Get("AppId").Int()) + j.Q取提示信息(&参数)

	if 参数.ReceivedUid > 0 && agent.L_agent.Id功能权限检测(c, 参数.ReceivedUid, DB.D代理功能_代收款) {
		var 局代理Info DB.DB_User
		var 代理在线支付信息 m.Z在线支付
		if 局代理Info, err = service.NewUser(c, &tx).Info(参数.ReceivedUid); err == nil {
			if 代理在线支付信息, err = L_setting.Q取代理在线支付信息(c, 参数.ReceivedUid); err == nil {
				局_代收款金额 := 参数.Rmb + L_rmbPay.Pay_指定Uid待支付金额(c, 参数.ReceivedUid)
				if 局代理Info.Rmb > 局_代收款金额 {
					参数.Z支付配置s = 代理在线支付信息
					参数.Z支付配置, _ = json.Marshal(&参数.Z支付配置s)
					局_通道数据, err = 局_通道.D订单创建(c, &参数)
					global.GVA_LOG.Error(参数.PayOrder+"代付订单创建失败", zap.Error(err))
					if err == nil {
						goto 下单成功
					}
				} else {
					err = errors.New("代理余额不足(" + Float64到文本(局_代收款金额, 2) + "=未关闭代收款订单总额)")
				}
			}
		}

		参数.Z支付配置s = setting.Q在线支付配置()
		参数.Z支付配置, _ = json.Marshal(&参数.Z支付配置s)
		参数.ReceivedUid = 0
		if err != nil {
			参数.E额外信息.Set("代收款err", err.Error())
		}
	} else {
		参数.ReceivedUid = 0
	}

	局_通道数据, err = 局_通道.D订单创建(c, &参数)
	if err != nil {
		return
	}
下单成功:
	参数.Status = constant.D订单状态_等待支付
	参数.Time = time.Now().Unix()
	参数.Ip = c.ClientIP()
	参数.Extra = "{}"
	if 字节数组, err2 := json.Marshal(参数.E额外信息); err2 == nil {
		参数.Extra = string(字节数组)
	}

	s := service.NewRmbPayService(&tx)
	_, err = s.Create(参数.DB_LogRMBPayOrder)
	if err != nil {
		return
	}
	req = 局_通道数据
	return
}

func (j *rmbPay) D订单退款(c *gin.Context, 参数 m.PayParams, 追回资产 bool, 备注 string) (err error) {
	var info struct {
		user         DB.DB_User
		Agent        DB.DB_User
		LogMoney     []DB.DB_LogMoney
		LogVipNumber []DB.DB_LogVipNumber
		卡类详情         dbm.DB_KaClass
		软件用户详情       DB.DB_AppUser
		app详情        DB.DB_AppInfo
	}

	db := *global.GVA_DB
	if 参数.DB_LogRMBPayOrder, err = service.NewRmbPayService(&db).Info2(map[string]interface{}{"PayOrder": 参数.PayOrder}); err != nil {
		err = errors.New("订单不存在")
		return
	}

	if 参数.Status != constant.D订单状态_成功 {
		err = errors.New("仅成功状态订单可操作支持退款")
		return
	}

	//禁止退款走管理员设置
	参数.Z支付配置s = setting.Q在线支付配置()
	参数.Z支付配置, _ = json.Marshal(&参数.Z支付配置s)
	参数.E额外信息, _ = gjson.LoadJson(参数.Extra)
	if 参数.Z支付配置s.J禁止退款 {
		err = errors.New("已禁止退款,请手动前往服务器数据库,修改配置信息文件 禁止退款:true")
		return
	}
	//判断是否为代收款如果是代收款读取代收用户id
	if 参数.ReceivedUid > 0 {
		参数.Z支付配置s = m.Z在线支付{} //重新清零数据防止下边读取失败,依然使用系统配置
		if 参数.Z支付配置s, err = L_setting.Q取代理在线支付信息(c, 参数.ReceivedUid); err == nil {
			参数.Z支付配置, _ = json.Marshal(&参数.Z支付配置s)
		}
	}

	局_通道, ok := j.已注册通道[参数.Type]
	if !ok {
		err = errors.New("支付方式未配置")
		return
	}

	err = db.Transaction(func(tx *gorm.DB) (err error) {

		//加锁重新查
		err = tx.Model(DB.DB_LogRMBPayOrder{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id = ?", 参数.Id).First(&参数.DB_LogRMBPayOrder).Error
		if err != nil {
			return errors.New("订单不存在")
		}
		if 参数.Status != constant.D订单状态_成功 { //重新确认一次订单状态
			err = errors.New("仅成功状态订单可操作支持退款")
			return
		}
		参数.Status = constant.D订单状态_退款中

		if 追回资产 && 参数.ProcessingType == constant.D订单类型_余额充值 { //追回余额
			if 参数.UidType == 1 { //只有账号模式的应用,才能扣余额
				err = tx.Model(DB.DB_User{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id = ?", 参数.Uid).First(&info.user).Error
				if err != nil {
					return errors.New("用户不存在,无法减余额")
				}
				err = tx.Model(DB.DB_User{}).Where("Id = ?", 参数.Uid).Update("Rmb", gorm.Expr("RMB - ?", 参数.Rmb)).Error
				if err != nil {
					return errors.New("减余额失败")
				}
				info.LogMoney = append(info.LogMoney, DB.DB_LogMoney{
					User:  info.user.User,
					Time:  time.Now().Unix(),
					Ip:    c.ClientIP(),
					Count: Float64取负值(参数.Rmb),
					Note:  fmt.Sprintf("管理员操作退款,订单:%s,扣除用户余额%s|新余额≈%s", 参数.PayOrder, Float64到文本(参数.Rmb, 2), Float64到文本(info.user.Rmb, 2)),
				})
			}
		}
		if 追回资产 && 参数.ProcessingType == constant.D订单类型_购卡直冲 { //购卡直冲  追回卡类时间余额 积分
			info.卡类详情, err = service.NewKaClass(c, tx).Info(参数.E额外信息.Get("KaClassId").Int())
			if err != nil {
				err = errors.New("追回资产时,发现卡类id" + 参数.E额外信息.Get("KaClassId").String() + "已不存在")
				return
			}
			if 参数.UidType == 1 && info.卡类详情.RMb != 0 { //只有账号模式的应用,才能扣余额
				err = tx.Model(DB.DB_User{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id = ?", 参数.Uid).First(&info.user).Error
				if err != nil {
					return errors.New("追回资产时,发现用户不存在,无法减余额")
				}
				err = tx.Model(DB.DB_User{}).Where("Id = ?", 参数.Uid).Update("Rmb", gorm.Expr("RMB - ?", 参数.Rmb)).Error
				if err != nil {
					return errors.New("减余额失败")
				}
				info.LogMoney = append(info.LogMoney, DB.DB_LogMoney{
					User:  info.user.User,
					Time:  time.Now().Unix(),
					Ip:    c.ClientIP(),
					Count: Float64取负值(参数.Rmb),
					Note:  fmt.Sprintf("管理员操作退款,订单:%s,扣除用户余额%s|新余额≈%s", 参数.PayOrder, Float64到文本(参数.Rmb, 2), Float64到文本(info.user.Rmb, 2)),
				})
			}

			err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+参数.E额外信息.Get("AppId").String()).Clauses(clause.Locking{Strength: "UPDATE"}).Where("Uid = ?", 参数.E额外信息.Get("AppUserUid").Int()).First(&info.软件用户详情).Error
			if err != nil {
				return errors.New("应用:" + 参数.E额外信息.Get("AppId").String() + "软件用户id" + 参数.E额外信息.Get("AppUserUid").String() + "已不存在")
			}
			info.软件用户详情.VipTime -= info.卡类详情.VipTime
			info.软件用户详情.VipNumber -= info.卡类详情.VipNumber
			_, err = service.NewAppUser(c, tx, 参数.E额外信息.Get("AppId").Int()).UpdateUid(info.软件用户详情.Uid, map[string]interface{}{
				"VipTime":   info.软件用户详情.VipTime,
				"VipNumber": info.软件用户详情.VipNumber,
			})
			if err != nil {
				return err
			}
			if info.卡类详情.VipTime != 0 {
				局_is计点 := App服务.App是否为计点(参数.E额外信息.Get("AppId").Int())
				info.LogVipNumber = append(info.LogVipNumber, DB.DB_LogVipNumber{
					User:  参数.User,
					AppId: 参数.E额外信息.Get("AppId").Int(),
					Type:  S三元(局_is计点, constant.Log_type_点数, constant.Log_type_时间),
					Time:  time.Now().Unix(),
					Ip:    c.ClientIP(),
					Count: Float64取负值(Int64到Float64(info.卡类详情.VipTime)),
					Note:  fmt.Sprintf("管理员操作退款,订单:%s,扣除软件用户"+S三元(局_is计点, "点数", "会员时间"), 参数.PayOrder),
				})
			}
			if info.卡类详情.VipNumber != 0 {
				info.LogVipNumber = append(info.LogVipNumber, DB.DB_LogVipNumber{
					User:  参数.User,
					AppId: 参数.E额外信息.Get("AppId").Int(),
					Type:  constant.Log_type_积分,
					Time:  time.Now().Unix(),
					Ip:    c.ClientIP(),
					Count: Float64取负值(info.卡类详情.VipNumber),
					Note:  fmt.Sprintf("管理员操作退款,订单:%s,扣除软件用户积分|新积分≈%s", 参数.PayOrder, Float64到文本(info.软件用户详情.VipNumber, 2)),
				})
			}
		}
		if 追回资产 && 参数.ProcessingType == constant.D订单类型_积分充值 { //追回积分
			if info.app详情, err = service.NewAppInfo(c, tx).Info(参数.E额外信息.Get("AppId").Int()); err != nil {
				return errors.Join(err, errors.New(fmt.Sprintf("AppId:%d取详情失败", 参数.E额外信息.Get("AppId").Int())))
			}
			局_增加积分 := Float64乘int64(参数.Rmb, int64(info.app详情.RmbToVipNumber))
			err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+参数.E额外信息.Get("AppId").String()).Clauses(clause.Locking{Strength: "UPDATE"}).Where("Uid = ?", 参数.E额外信息.Get("AppUserUid").Int()).First(&info.软件用户详情).Error
			if err != nil {
				return errors.New("应用:" + 参数.E额外信息.Get("AppId").String() + "软件用户id" + 参数.E额外信息.Get("AppUserUid").String() + "已不存在")
			}
			info.软件用户详情.VipNumber -= 局_增加积分
			_, err = service.NewAppUser(c, tx, 参数.E额外信息.Get("AppId").Int()).UpdateUid(info.软件用户详情.Uid, map[string]interface{}{
				"VipNumber": info.软件用户详情.VipNumber,
			})
			if err != nil {
				return errors.Join(err, errors.New("扣除积分失败"))
			}

			info.LogVipNumber = append(info.LogVipNumber, DB.DB_LogVipNumber{
				User:  参数.User,
				AppId: 参数.E额外信息.Get("AppId").Int(),
				Type:  constant.Log_type_积分,
				Time:  time.Now().Unix(),
				Ip:    c.ClientIP(),
				Count: 局_增加积分,
				Note:  fmt.Sprintf("管理员操作退款,订单:%s,扣除软件用户积分(积分rmb比例:%d)|新积分≈%s", 参数.PayOrder, info.app详情.RmbToVipNumber, Float64到文本(info.软件用户详情.VipNumber, 2)),
			})

		}

		if 追回资产 && 参数.ProcessingType == constant.D订单类型_支付购卡 { //追回积分
			c.Set("tx", tx)
			err = ka.L_ka.K卡号追回(c, 参数.E额外信息.Get("Id").Int(), c.GetString("User"))
			if err != nil {
				return err
			}
			if 局_临时, ok2 := c.Get("info.LogVipNumber"); ok2 {
				info.LogVipNumber = append(info.LogVipNumber, 局_临时.([]DB.DB_LogVipNumber)...)
			}
			if 局_临时, ok2 := c.Get("info.LogMoney"); ok2 {
				info.LogMoney = append(info.LogMoney, 局_临时.([]DB.DB_LogMoney)...)
			}
		}
		//判断是否有代理分成,如果有代理分成,对应扣除
		for 索引 := range 参数.E额外信息.Len("分成详细") {
			局_uid := 参数.E额外信息.Get("分成详细." + strconv.Itoa(索引) + ".Uid").Int()
			局_金额 := 参数.E额外信息.Get("分成详细." + strconv.Itoa(索引) + ".S实际分成金额").Float64()
			var 代理详情 DB.DB_User
			代理详情, err = service.NewUser(c, tx).Info(局_uid)
			if err != nil {
				return errors.Join(err, errors.New("代理"+代理详情.User+",不存在,无法扣除分成"))
			}
			err = tx.Model(DB.DB_User{}).Where("Id = ?", 局_uid).Update("Rmb", gorm.Expr("RMB - ?", 局_金额)).Error

			if err != nil {
				return errors.Join(err, errors.New("代理分成用户扣除分成失败"))
			}
			代理详情.Rmb = Float64减float64(代理详情.Rmb, 局_金额, 2)
			info.LogMoney = append(info.LogMoney, DB.DB_LogMoney{
				User:  代理详情.User,
				Time:  time.Now().Unix(),
				Ip:    c.ClientIP(),
				Count: Float64取负值(局_金额),
				Note:  fmt.Sprintf("管理员操作用户退款,订单:%s,扣除订单分成%s|新余额≈%s", 参数.PayOrder, Float64到文本(局_金额, 2), Float64到文本(代理详情.Rmb, 2)),
			})
		}
		//err = 参数.E额外信息.Set("卡类金额", 局_卡类信息.Money)
		//err = 参数.E额外信息.Set("调价详情", 调价信息列表)
		//err = 参数.E额外信息.Set("总调价", 总调价)
		for 索引 := range 参数.E额外信息.Len("调价详情") {
			局_uid := 参数.E额外信息.Get("调价详情." + strconv.Itoa(索引) + ".AgentId").Int()
			局_金额 := 参数.E额外信息.Get("分成详细." + strconv.Itoa(索引) + ".Markup").Float64()
			var 代理详情 DB.DB_User
			代理详情, err = service.NewUser(c, tx).Info(局_uid)
			if err != nil {
				return errors.Join(err, errors.New("代理"+代理详情.User+",不存在,无法扣除调价分成"))
			}
			err = tx.Model(DB.DB_User{}).Where("Id = ?", 局_uid).Update("Rmb", gorm.Expr("RMB - ?", 局_金额)).Error
			if err != nil {
				return errors.Join(err, errors.New("代理调价用户扣除调价失败"))
			}
			代理详情.Rmb = Float64减float64(代理详情.Rmb, 局_金额, 2)
			info.LogMoney = append(info.LogMoney, DB.DB_LogMoney{
				User:  代理详情.User,
				Time:  time.Now().Unix(),
				Ip:    c.ClientIP(),
				Count: Float64取负值(局_金额),
				Note:  fmt.Sprintf("管理员操作用户退款,订单:%s,扣除订单代理调价%s|新余额≈%s", 参数.PayOrder, Float64到文本(局_金额, 2), Float64到文本(代理详情.Rmb, 2)),
			})
		}

		参数.Status = constant.D订单状态_退款成功
		参数.E额外信息.Set("退款时间", time.Now().Format("2006-01-02 15:04:05"))
		参数.Extra = 参数.E额外信息.String()
		data := map[string]interface{}{
			"Status": 参数.Status,
			"Extra":  参数.Extra,
		}
		if 备注 != "" {
			data["Note"] = 备注
		}
		if err = tx.Model(DB.DB_LogRMBPayOrder{}).Where("Id = ?", 参数.Id).Updates(data).Error; err != nil {
			return errors.Join(err, errors.New("订单状态更新失败"))
		}
		参数.Y异步回调地址 = setting.Q系统设置().X系统地址 + "/webApi/payNotify2/" + 参数.PayOrder //微信可能用到

		if err == nil && 参数.ReceivedUid > 0 { //如果是代收款订单, 要恢复已扣的余额
			err = tx.Model(DB.DB_User{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id = ?", 参数.ReceivedUid).First(&info.user).Error
			if err != nil {
				return errors.New("代收款用户不存在,无法恢复余额")
			}
			err = tx.Model(DB.DB_User{}).Where("Id = ?", 参数.ReceivedUid).Update("Rmb", gorm.Expr("RMB + ?", 参数.Rmb)).Error
			if err != nil {
				return errors.New("代收款用户恢复余额失败")
			}
			err = tx.Model(DB.DB_User{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id = ?", 参数.ReceivedUid).First(&info.Agent).Error
			if err != nil {
				return errors.New("代收款用户不存在,无法恢复余额")
			}
			info.LogMoney = append(info.LogMoney, DB.DB_LogMoney{
				User:  info.Agent.User,
				Time:  time.Now().Unix(),
				Ip:    c.ClientIP(),
				Count: 参数.Rmb,
				Note:  fmt.Sprintf("管理员操作代收款订单id:%s,第三方订单:%s,退款,恢复代扣余额%s|新余额≈%s", 参数.PayOrder, 参数.PayOrder2, Float64到文本(参数.Rmb, 2), Float64到文本(info.Agent.Rmb, 2)),
			})
		}
		err = 局_通道.D订单退款(c, &参数) //最后处理因为不可恢复,所以退款结果作为最终条件

		return err
	})
	//最后写出日志
	if err == nil {
		c.Set("tx", &db)
		if err = log.L_log.S输出日志(c, info.LogMoney); err != nil {
			global.GVA_LOG.Error("输出日志失败!", zap.Any("err", err))
		}
		if err = log.L_log.S输出日志(c, info.LogVipNumber); err != nil {
			global.GVA_LOG.Error("输出日志失败!", zap.Any("err", err))
		}
	}

	return
}

func (j *rmbPay) D订单回调(c *gin.Context) (响应信息 string, 响应代码 int) {
	响应代码 = 200
	响应信息 = "订单不存在"
	var 参数 m.PayParams
	参数.Z支付配置s = setting.Q在线支付配置()
	参数.Z支付配置, _ = json.Marshal(&参数.Z支付配置s)

	orderId := c.Param("order")
	if orderId == "" || orderId == "123456" {
		for i, _ := range j.已注册通道 {
			if orderId = j.已注册通道[i].Q取订单id(c, &参数); orderId != "" {
				break
			}
		}
	}

	if orderId == "" {
		return
	}
	tx := *global.GVA_DB
	s := service.NewRmbPayService(&tx)

	var err error
	参数.DB_LogRMBPayOrder, err = s.Info2(gin.H{"PayOrder": orderId})
	参数.E额外信息, _ = gjson.LoadJson(参数.Extra)
	if err != nil || 参数.Status != constant.D订单状态_等待支付 {
		return
	}

	局_通道, ok := j.已注册通道[参数.Type]
	if !ok {
		err = errors.New("支付方式未配置")
		return
	}

	响应信息, 响应代码, err = 局_通道.D订单支付回调(c, &参数)
	if err != nil {
		global.GVA_LOG.Error("订单回调失败!", zap.Any("err", err))
		return
	}
	err = 参数.E额外信息.Set("回调时间", time.Now().Unix())
	err = 参数.E额外信息.Set("回调ip", c.ClientIP())
	err = 参数.E额外信息.Set("回调ua", c.GetHeader("User-Agent"))
	参数.Extra = 参数.E额外信息.String()

	db := *global.GVA_DB
	//先加锁修改为待处理
	err = db.Transaction(func(tx *gorm.DB) error {
		var 局_订单信息 DB.DB_LogRMBPayOrder
		err = tx.Model(DB.DB_LogRMBPayOrder{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("Id=?", 参数.Id).
			First(&局_订单信息).Error //加锁再查一次
		if err != nil || 局_订单信息.Status != constant.D订单状态_等待支付 { //有其他线程抢到任务了并处理了
			return err
		}
		//确定订单状态没问题,而且数据已经是行锁状态,开始更新状态等信息
		参数.Status = constant.D订单状态_已付待处理
		err = tx.Model(DB.DB_LogRMBPayOrder{}).
			Where("Id=?", 参数.Id).
			Updates(map[string]interface{}{
				"PayOrder":       参数.PayOrder,
				"PayOrder2":      参数.PayOrder2,
				"User":           参数.User,
				"Uid":            参数.Uid,
				"UidType":        参数.UidType,
				"Status":         参数.Status,
				"ProcessingType": 参数.ProcessingType,
				"Extra":          参数.Extra,
				"Rmb":            参数.Rmb, //小叮当可能改变实际支付金额
			}).Error
		return err //提交事务自动解锁
	})
	if err != nil {
		return
	}

	_ = j.Z支付成功_后处理(c, &参数)
	_ = cpsPayOrder.L_cpsPayOrder.C处理佣金发放_线程安全(c, &参数)

	return
}
func (j *rmbPay) D订单退款回调(c *gin.Context) (响应信息 string, 响应代码 int) {
	响应代码 = 200
	响应信息 = "订单不存在"
	var 参数 m.PayParams
	参数.Z支付配置s = setting.Q在线支付配置()
	参数.Z支付配置, _ = json.Marshal(&参数.Z支付配置s)
	参数.E额外信息, _ = gjson.LoadJson(参数.Extra)
	orderId := c.Param("order")
	if orderId == "" {
		for i, _ := range j.已注册通道 {
			if orderId = j.已注册通道[i].Q取订单id(c, &参数); orderId != "" {
				break
			}
		}
	}

	if orderId == "" {
		return
	}
	tx := *global.GVA_DB
	s := service.NewRmbPayService(&tx)

	var err error
	参数.DB_LogRMBPayOrder, err = s.Info2(gin.H{"PayOrder": orderId})
	if err != nil || 参数.Status != constant.D订单状态_退款中 {
		return
	}

	局_通道, ok := j.已注册通道[参数.Type]
	if !ok {
		err = errors.New("支付方式未配置")
		return
	}

	响应信息, 响应代码, err = 局_通道.D订单退款回调(c, &参数)
	if err != nil {
		参数.Status = constant.D订单状态_退款失败
		参数.Note = 参数.Note + err.Error()
	} else {
		参数.Status = constant.D订单状态_退款成功
	}
	err = 参数.E额外信息.Set("退款回调时间", time.Now().Unix())
	err = 参数.E额外信息.Set("退款回调ip", c.ClientIP())
	err = 参数.E额外信息.Set("退款回调ua", c.GetHeader("User-Agent"))
	参数.Extra = 参数.E额外信息.String()

	db := *global.GVA_DB
	//先加锁修改为待处理
	err = db.Model(DB.DB_LogRMBPayOrder{}).
		Where("Id=?", 参数.Id).
		Updates(map[string]interface{}{
			"Status": 参数.Status,
			"Extra":  参数.Extra,
		}).Error
	return
}
func (j *rmbPay) Z支付成功_后处理(c *gin.Context, 参数 *m.PayParams) (err error) {
	if 参数.Status != constant.D订单状态_已付待处理 {
		return
	}
	var info struct {
		LogMoney     []DB.DB_LogMoney
		LogVipNumber []DB.DB_LogVipNumber
		LogKa        []DB.DB_LogKa
		/*
			user用户详情 DB.DB_User*/
		app用户详情 DB.DB_AppUser
		卡类详情    dbm.DB_KaClass
		卡号详情    DB.DB_Ka
		app详情   DB.DB_AppInfo
	}
	var 临时数据 interface{}
	var ok bool

	//这里就是转账了,需要开启事务保证
	db := *global.GVA_DB
	//先加锁修改为待处理
	err = db.Transaction(func(tx *gorm.DB) error {
		//重新加锁,确定状态
		c.Set("tx", tx)
		defer delete(c.Keys, "tx")

		var 局_订单信息 DB.DB_LogRMBPayOrder
		err = tx.Model(DB.DB_LogRMBPayOrder{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("Id=?", 参数.Id).
			First(&局_订单信息).Error //加锁再查一次
		if err != nil || 局_订单信息.Status != constant.D订单状态_已付待处理 { //有其他线程抢到任务了并处理了
			return err
		}
		参数.DB_LogRMBPayOrder = 局_订单信息
		参数.E额外信息, _ = gjson.LoadJson(参数.Extra)

		局_订单信息.Status = constant.D订单状态_已付待处理
		switch 参数.ProcessingType {
		default:
			return errors.New("ProcessingType错误")
		case constant.D订单类型_余额充值: //0
			err = tx.Model(DB.DB_User{}).Where("Id = ?", 参数.Uid).Update("RMB", gorm.Expr("RMB + ?", 参数.Rmb)).Error
			if err != nil {
				return errors.Join(err, errors.New(strconv.Itoa(参数.Uid)+"Id余额增加失败"))
			}
			var 局_新余额 float64
			err = tx.Model(DB.DB_User{}).Select("Rmb").Where("Id=?", 参数.Uid).First(&局_新余额).Error
			if err != nil {
				return errors.Join(err, errors.New(strconv.Itoa(参数.Uid)+"新余额读取失败"))
			}
			info.LogMoney = append(info.LogMoney, DB.DB_LogMoney{
				User:  参数.User,
				Time:  time.Now().Unix(),
				Ip:    参数.Ip,
				Count: 参数.Rmb,
				Note:  fmt.Sprintf("余额充值支付订单:%s,付款成功|新余额≈%v", 参数.PayOrder, Float64到文本(局_新余额, 2)),
			})
		case constant.D订单类型_购卡直冲: //1
			var 卡类ID, AppUserUid int

			if 卡类ID = 参数.E额外信息.Get("KaClassId").Int(); 卡类ID == 0 {
				return errors.New("订单id:%s,扩展信息KaClassId不正确")
			}

			if AppUserUid = 参数.E额外信息.Get("AppUserUid").Int(); AppUserUid == 0 {
				return errors.New("订单id:%s,扩展信息AppUserUid不正确")
			}

			if err = ka.L_ka.K卡类直冲_事务(c, 卡类ID, AppUserUid); err != nil {
				return err
			}
			if 临时数据, ok = c.Get("logMoney"); ok { //判断是否有rmb充值的日志
				info.LogMoney = append(info.LogMoney, 临时数据.(DB.DB_LogMoney))
				info.LogMoney[len(info.LogMoney)-1].Note = "购卡直冲支付订单:" + 参数.PayOrder + info.LogMoney[len(info.LogMoney)-1].Note
			}

			if 临时数据, ok = c.Get("logVipNumber"); ok { //判断是否有积分充值的日志
				info.LogVipNumber = append(info.LogVipNumber, 临时数据.(DB.DB_LogVipNumber))
				info.LogVipNumber[len(info.LogVipNumber)-1].Note = "购卡直冲支付订单:" + 参数.PayOrder + info.LogVipNumber[len(info.LogVipNumber)-1].Note
			}

			if 临时数据, ok = c.Get("info.app详情"); ok {
				info.app详情 = 临时数据.(DB.DB_AppInfo)
			}
			if 临时数据, ok = c.Get("info.卡类详情"); ok {
				info.卡类详情 = 临时数据.(dbm.DB_KaClass)
			}
			if text, ok2 := c.Get("info.app用户详情"); ok2 {
				info.app用户详情 = text.(DB.DB_AppUser)
			}
			参数.E额外信息.Get("卡类ID", 卡类ID)
			参数.E额外信息.Get("卡类名称", info.卡类详情.Name)
			参数.E额外信息.Get("应用", info.app详情.AppName)
			参数.Note = 参数.Note + "充值卡类ID:" + strconv.Itoa(卡类ID) + ",应用:" + info.app详情.AppName + ",卡类:" + info.卡类详情.Name
			参数.E额外信息.Set("AgentUid", info.app用户详情.AgentUid)

			//判断代理是否有分成,如果有进行处理
			if err = j.代理分成(c, 参数, 参数.E额外信息.Get("卡类金额").Float64()); err != nil {
				return err
			} else {
				if 临时数据, ok = c.Get("LogMoney"); ok && 临时数据 != nil {
					info.LogMoney = append(info.LogMoney, 临时数据.([]DB.DB_LogMoney)...)
				}
			}
		case constant.D订单类型_支付购卡: //3
			//没有订单信息没有Uid,用户名,需要修改
			if 参数.E额外信息.Get("KaClassId").Int() == 0 {
				return errors.New("扩展信息KaClassId不正确")
			}
			if info.卡类详情, err = service.NewKaClass(c, tx).Info(参数.E额外信息.Get("KaClassId").Int()); err != nil {
				return errors.Join(err, errors.New(fmt.Sprintf("卡类:%d取详情失败", 参数.E额外信息.Get("KaClassId").Int())))
			}
			if info.app详情, err = service.NewAppInfo(c, tx).Info(info.卡类详情.AppId); err != nil {
				return errors.Join(err, errors.New(fmt.Sprintf("AppId:%d取详情失败", 参数.E额外信息.Get("AppId").Int())))
			}

			info.卡号详情, err = ka.L_ka.Ka单卡创建(c, info.卡类详情.Id, "系统自动", "支付购卡订单ID:"+参数.PayOrder, "", 0)
			if err != nil {
				return errors.Join(err, errors.New("卡号创建失败"))
			}

			err = 参数.E额外信息.Set("Id", info.卡号详情.Id)
			err = 参数.E额外信息.Set("卡号", info.卡号详情.Name)
			err = 参数.E额外信息.Set("卡类", info.卡类详情.Name)
			err = 参数.E额外信息.Set("应用", info.app详情.AppName)

			参数.Note = 参数.Note + "购卡:" + info.卡号详情.Name + ",应用:" + info.app详情.AppName + ",卡类:" + info.卡类详情.Name
			局_文本 := fmt.Sprintf("支付购卡订单ID:%s,卡类:%d,消费:%.2f)", 参数.PayOrder, info.卡号详情.KaClassId, 参数.Rmb)
			info.LogKa = append(info.LogKa, DB.DB_LogKa{
				User:     "支付购卡",
				UserType: constant.Log_卡操作用户_系统自动,
				Ka:       info.卡类详情.Name,
				KaType:   constant.Log_卡操作_增,
				Time:     time.Now().Unix(),
				Ip:       参数.Ip,
				Note:     局_文本,
			})
			if text, ok2 := c.Get("info.app用户详情"); ok2 {
				info.app用户详情 = text.(DB.DB_AppUser)
			}

			if info.app用户详情.AgentUid != 0 {
				参数.E额外信息.Set("AgentUid", info.app用户详情.AgentUid)
			} else {
				//支付购卡,如果用户没登陆,可能没有用户代理标志,就需要使用在线代理标志
				参数.E额外信息.Set("AgentUid", 参数.E额外信息.Get("在线信息AgentUid").Int())
			}

			//判断代理是否有分成,如果有进行处理
			if err = j.代理分成(c, 参数, 参数.E额外信息.Get("卡类金额").Float64()); err != nil {
				return err
			} else {
				if 临时数据, ok = c.Get("LogMoney"); ok {
					info.LogMoney = append(info.LogMoney, 临时数据.([]DB.DB_LogMoney)...)
				}
			}
		}
		//判断是否为代收款
		if 参数.ReceivedUid > 0 {
			var 局_info DB.DB_User
			//加锁再查一次 锁定数值 防止并发数据错误
			err = tx.Model(DB.DB_User{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id=?", 参数.ReceivedUid).First(&局_info).Error
			if err != nil {
				return errors.Join(err, errors.New(strconv.Itoa(参数.ReceivedUid)+"代理信息读取失败"))
			}
			//只有有信任度的代理,才可以代收款,所以可以扣到一定值的负数
			err = tx.Model(DB.DB_User{}).Where("Id = ?", 参数.ReceivedUid).Update("RMB", gorm.Expr("RMB - ?", 参数.Rmb)).Error
			if err != nil {
				return errors.Join(err, errors.New(strconv.Itoa(参数.ReceivedUid)+"Id余额减少失败"))
			}
			//再查一次
			err = tx.Model(DB.DB_User{}).Where("Id=?", 参数.ReceivedUid).First(&局_info).Error
			if err != nil {
				return errors.Join(err, errors.New(strconv.Itoa(参数.ReceivedUid)+"新余额读取失败"))
			}
			str := fmt.Sprintf("用户%s,%s订单ID:%s,第三方订单ID:%s,%s代收款:¥%s ,|新余额≈%s", 参数.User, j.Map订单类型[参数.ProcessingType], 参数.PayOrder, 参数.PayOrder2, 参数.Type, Float64到文本(参数.Rmb, 2), Float64到文本(局_info.Rmb, 2))
			info.LogMoney = append(info.LogMoney, DB.DB_LogMoney{
				User:  局_info.User,
				Time:  time.Now().Unix(),
				Ip:    参数.Ip,
				Count: Float64取负值(参数.Rmb),
				Note:  str,
			})
		}
		//如果能走到这里说明上面处理成功了, 订单状态改为成功
		参数.Status = constant.D订单状态_成功
		参数.Extra = 参数.E额外信息.String()
		err = tx.Model(DB.DB_LogRMBPayOrder{}).
			Where("Id=?", 参数.Id).
			Updates(map[string]interface{}{
				"Status": constant.D订单状态_成功,
				"Extra":  参数.Extra,
				"Note":   参数.Note,
			}).Error
		return err //最后一步提交事务
	})
	if err != nil { //如果有错误,只修改备注,然后等人工处理
		参数.Note = 参数.Note + err.Error()
		err = db.Model(DB.DB_LogRMBPayOrder{}).
			Where("Id=?", 参数.Id).
			Updates(map[string]interface{}{"Note": 参数.Note + err.Error()}).Error
		if err != nil {
			global.GVA_LOG.Error("更新数据库失败!", zap.Any("err", err))
		}
	} else {
		//最后写出日志
		if err = log.L_log.S输出日志(c, info.LogKa); err != nil {
			global.GVA_LOG.Error("输出日志失败!", zap.Any("err", err))
		}
		if err = log.L_log.S输出日志(c, info.LogMoney); err != nil {
			global.GVA_LOG.Error("输出日志失败!", zap.Any("err", err))
		}
		if err = log.L_log.S输出日志(c, info.LogVipNumber); err != nil {
			global.GVA_LOG.Error("输出日志失败!", zap.Any("err", err))
		}
	}

	return err
}
func (j *rmbPay) 代理分成(c *gin.Context, 参数 *m.PayParams, AgentMoney float64) (err error) {
	var info struct {
		LogMoney []DB.DB_LogMoney
	}
	//下边这两个可空
	var AgentUid int
	AgentUid = 参数.E额外信息.Get("AgentUid").Int()

	if AgentUid > 0 && AgentMoney > 0 {
		var tx *gorm.DB
		if tempObj, ok := c.Get("tx"); ok {
			tx = tempObj.(*gorm.DB)
		} else {
			db := *global.GVA_DB
			tx = &db
		}
		var 局_价格组成 struct {
			总调价  float64
			调价详情 []dbm.DB_KaClassUpPrice
			购买数量 int64

			卡类金额 float64
		}
		//err = 参数.E额外信息.Set("卡类金额", 局_卡类信息.Money)
		//err = 参数.E额外信息.Set("调价详情", 调价信息列表)
		//err = 参数.E额外信息.Set("总调价", 总调价)
		局_价格组成.总调价 = 参数.E额外信息.Get("总调价").Float64()
		局_价格组成.购买数量 = 1
		局_价格组成.卡类金额 = 参数.E额外信息.Get("卡类金额").Float64()
		err = 参数.E额外信息.Get("调价详情").Scan(&局_价格组成.调价详情)
		//先分成 代理调价信息的价格
		if 局_价格组成.总调价 > 0 {
			for _, v := range 局_价格组成.调价详情 {
				分成金额 := Float64乘int64(v.Markup, 局_价格组成.购买数量) //有多少卡就分多少个
				err = tx.Model(DB.DB_User{}).Where("Id = ?", v.AgentId).Update("RMB", gorm.Expr("RMB + ?", 分成金额)).Error
				if err != nil {
					return errors.Join(err, fmt.Errorf("代理分成失败,请检查原因%d,%s", v.AgentId, Float64到文本(分成金额, 2)))
				}
				var 局_userInfo DB.DB_User
				err = tx.Model(DB.DB_User{}).Where("Id = ?", v.AgentId).Find(&局_userInfo).Error
				if err != nil {
					return errors.Join(err, fmt.Errorf("代理分成后,读取代理数据失败请检查原因%d,%s", v.AgentId, Float64到文本(分成金额, 2)))
				}

				// 构建日志记录
				var 局_临时日志 DB.DB_LogMoney
				局_临时日志.Time = time.Now().Unix()
				局_临时日志.Ip = c.ClientIP() + " " + Qqwry.Ip查信息2(c.ClientIP())
				局_临时日志.User = 局_userInfo.User
				局_临时日志.Count = 分成金额
				局_临时文本1, 局_临时文本2 := "", ""
				if agent.L_agent.Id功能权限检测(c, 局_userInfo.Id, DB.D代理功能_查看归属软件用户) {
					局_临时文本1 = 参数.User
					局_临时文本2 = 参数.Note
				}
				局_日志前缀 := fmt.Sprintf("用户%s%s%s订单ID:%s", 局_临时文本1, j.Map订单类型[参数.ProcessingType], 局_临时文本2, 参数.PayOrder)
				局_临时日志.Note = 局_日志前缀 + fmt.Sprintf("调价分成:¥%s(%s*%d),|新余额≈%s",
					Float64到文本(分成金额, 2),
					Float64到文本(v.Markup, 2),
					局_价格组成.购买数量, Float64到文本(局_userInfo.Rmb, 2))
				info.LogMoney = append(info.LogMoney, 局_临时日志)
			}
		}
		//代理分成
		//开始分利润 20240202 mark处理重构以后改事务
		代理分成数据, err3 := agent.L_agent.D代理分成计算(c, AgentUid, AgentMoney)
		if err3 == nil {
			for 局_索引 := range 代理分成数据 {
				d := 代理分成数据[局_索引] //太长了,放个变量里
				err = tx.Model(DB.DB_User{}).Where("Id = ?", d.Uid).Update("RMB", gorm.Expr("RMB + ?", d.S实际分成金额)).Error
				if err != nil {
					return errors.Join(err, errors.New(strconv.Itoa(d.Uid)+"Id余额增加失败"))
				}
				var 局_新余额 float64
				err = tx.Model(DB.DB_User{}).Select("Rmb").Where("Id=?", d.Uid).First(&局_新余额).Error
				if err != nil {
					return errors.Join(err, errors.New(strconv.Itoa(d.Uid)+"新余额读取失败"))
				}
				局_临时文本1 := S三元(agent.L_agent.Id功能权限检测(c, d.Uid, DB.D代理功能_查看归属软件用户), 参数.User, "")
				局_临时文本2 := S三元(agent.L_agent.Id功能权限检测(c, d.Uid, DB.D代理功能_查看归属软件用户), 参数.Note, "")
				str := fmt.Sprintf("用户%s%s%s订单ID:%s,分成:¥%s (¥%s(实价)*(%d%%-%d%%)),|新余额≈%s", 局_临时文本1, j.Map订单类型[参数.ProcessingType], 局_临时文本2, 参数.PayOrder, Float64到文本(d.S实际分成金额, 2), Float64到文本(AgentMoney, 2), d.F分成百分比, d.F分给下级百分比, Float64到文本(局_新余额, 2))

				info.LogMoney = append(info.LogMoney, DB.DB_LogMoney{
					User:  d.User,
					Time:  time.Now().Unix(),
					Ip:    参数.Ip,
					Count: d.S实际分成金额,
					Note:  str,
				})
			}
		}
		// 分成结束============== 记录分成情况, 后续退款对应扣除
		参数.E额外信息.Set("分成详细", 代理分成数据)
	}
	if info.LogMoney != nil {
		c.Set("LogMoney", info.LogMoney)
	}

	return
}

func (j *rmbPay) Pay_取支付通道状态() gin.H {
	局_数组 := j.Pay_取支付通道基本信息()
	局map := make(gin.H, len(局_数组))

	for _, v := range 局_数组 {
		if v.Alias != "" {
			局map[v.Alias] = v.Status
		} else {
			局map[v.Name] = v.Status
		}
	}
	return 局map
}

type 支付通道基本信息 struct {
	Id     int    `json:"Id"`
	Name   string `json:"Name"`
	Alias  string `json:"Alias"`  //显示名称
	Status bool   `json:"Status"` //开关
	RMB    int    `json:"RMB"`    //最大金额
}

func (j *rmbPay) Pay_取支付通道基本信息() []支付通道基本信息 {
	支付配置 := setting.Q在线支付配置()
	if &支付配置 == nil {
		return []支付通道基本信息{}
	}
	支付通道列表 := []支付通道基本信息{
		{Id: 1, Name: "支付宝PC", Alias: 支付配置.Z支付宝显示名称, Status: 支付配置.Z支付宝开关, RMB: 支付配置.Z支付宝单次最大金额},
		{Id: 2, Name: "支付宝当面付", Alias: 支付配置.Z支付宝当面付显示名称, Status: 支付配置.Z支付宝当面付开关, RMB: 支付配置.Z支付宝单次最大金额},
		{Id: 3, Name: "支付宝H5", Alias: 支付配置.Z支付宝H5显示名称, Status: 支付配置.Z支付宝H5开关, RMB: 支付配置.Z支付宝单次最大金额},
		{Id: 4, Name: "微信支付", Alias: 支付配置.W微信支付显示名称, Status: 支付配置.W微信支付开关, RMB: 支付配置.W微信支付单次最大金额},
		{Id: 5, Name: "小叮当", Alias: 支付配置.X小叮当支付显示名称, Status: 支付配置.X小叮当支付开关, RMB: 支付配置.X小叮当单次最大金额},
		{Id: 6, Name: "虎皮椒", Alias: 支付配置.H虎皮椒支付显示名称, Status: 支付配置.H虎皮椒支付开关, RMB: 支付配置.H虎皮椒单次最大金额},
		{Id: 7, Name: "易支付", Alias: 支付配置.Y易支付显示名称, Status: 支付配置.Y易支付开关, RMB: 支付配置.Y易支付最大金额},
		{Id: 8, Name: "易支付2", Alias: 支付配置.Y易支付2显示名称, Status: 支付配置.Y易支付2开关, RMB: 支付配置.Y易支付2最大金额},
	}
	return 支付通道列表
}

func (j *rmbPay) Pay_指定Uid待支付金额(c *gin.Context, Uid int) (金额 float64) {
	// 开启事务,检测上层是否有事务,如果有直接使用,没有就创建一个
	var tx *gorm.DB
	if tempObj, ok := c.Get("tx"); ok {
		tx = tempObj.(*gorm.DB)
	} else {
		db := *global.GVA_DB
		tx = &db
	}
	//获取该uid 等待支付的金额总数
	err := tx.Model(DB.DB_LogRMBPayOrder{}).Select("sum(Rmb) as Rmb").Where("ReceivedUid=? and Status=?", Uid, constant.D订单状态_等待支付).First(&金额).Error
	if err != nil {
		//如果出错,就返回0   报错一般是rmb字段为null 但是给的变量类型为float64  暂不影响,以后再查
		//global.GVA_LOG.Error("获取指定Uid待支付金额!", zap.Any("err", err))
	}
	return
}
