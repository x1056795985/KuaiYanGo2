package Ser_RMBPayOrder //生成单号
import (
	"errors"
	"fmt"
	"server/Service/Ser_Ka"
	"server/Service/Ser_User"
	"server/global"
	DB "server/structs/db"
	"strconv"
	"sync"
	"time"
)

var (
	// 逻辑中使用的某个变量
	集_订单当前秒计数 int
	集_订单当前时间戳 int64
	// 与变量对应的使用互斥锁
	集_互斥锁_订单号 sync.Mutex
)

// 生成18位订单号  线程安全
// 年月日时分秒0001计数 每秒9999订单内没问题
func Get获取新订单号() string {

	集_互斥锁_订单号.Lock()
	当前时间戳 := time.Now().Unix()
	if 当前时间戳 == 集_订单当前时间戳 {
		集_订单当前秒计数++
	} else {
		集_订单当前时间戳 = 当前时间戳
		集_订单当前秒计数 = 1
	}
	局_计数 := 集_订单当前秒计数
	集_互斥锁_订单号.Unlock()

	var 最终订单号 string = time.Unix(当前时间戳, 0).Format("20060102150405")
	if 局_计数 < 10 {
		最终订单号 += "000" + strconv.Itoa(局_计数)
	} else if 局_计数 < 100 {
		最终订单号 += "00" + strconv.Itoa(局_计数)
	} else if 局_计数 < 1000 {
		最终订单号 += "0" + strconv.Itoa(局_计数)
	} else if 局_计数 < 10000 {
		最终订单号 += strconv.Itoa(局_计数)
	} else {
		fmt.Println("恭喜生成订单号大于每秒1w建议更换算法")
	}

	return 最终订单号
}

const D订单状态_等待支付 = 1
const D订单状态_已付待处理 = 2
const D订单状态_成功 = 3
const D订单状态_退款中 = 4
const D订单状态_退款失败 = 5
const D订单状态_退款成功 = 6
const D订单状态_已关闭 = 7

// Order更新订单状态
// 1  '等待支付'  2  '已付待充' 3 '充值成功' 4 退款中 5 ? 退款失败" : 6退款成功
func Order更新订单状态(订单号 string, 状态值 int) bool {
	if 订单号 == "" {
		return false
	}
	err := global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).Where("PayOrder = ?", 订单号).Update("Status", 状态值).Error
	if err != nil {
		global.GVA_LOG.Error(订单号 + "Order更新订单状态失败:" + err.Error())
		return false
	}
	return true
}

func Order更新订单备注_批量(订单号 []string, 备注 string) error {
	if len(订单号) == 0 {
		return errors.New("订单号数组不能为空")
	}
	err := global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).Where("PayOrder IN ?", 订单号).Update("Note", 备注).Error
	return err

}

var C处理类型 = map[int]string{
	0: "余额充值",
	1: "购卡直冲",
	2: "积分充值",
	3: "支付购卡",
}

// Uid类型 1账号 2卡号
func Order订单创建(Uid, Uid类型 int, Rmb float64, 支付类型, 订单备注, Ip string, 处理类型 int, 额外信息 string) (DB.DB_LogRMBPayOrder, error) {
	var 新订单 DB.DB_LogRMBPayOrder
	新订单.Id = 0
	新订单.Uid = Uid
	新订单.UidType = Uid类型
	if 新订单.UidType == 2 {
		新订单.User = Ser_Ka.Id取卡号(新订单.Uid)
	} else {
		新订单.User = Ser_User.Id取User(新订单.Uid)
	}

	新订单.Status = 1
	新订单.Time = time.Now().Unix()
	新订单.Ip = Ip
	新订单.Type = 支付类型
	新订单.ProcessingType = 处理类型
	新订单.Extra = 额外信息
	新订单.Rmb = Rmb
	新订单.Note = 订单备注
	新订单.PayOrder = Get获取新订单号()
	err := global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).Create(&新订单).Error
	if err != nil {
		return DB.DB_LogRMBPayOrder{}, err
	}
	return 新订单, err
}

func Order取订单详细(订单号 string) (DB.DB_LogRMBPayOrder, bool) {
	if 订单号 == "" {
		return DB.DB_LogRMBPayOrder{}, false
	}
	var 局订单信息 DB.DB_LogRMBPayOrder
	err := global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).Where("PayOrder = ?", 订单号).First(&局订单信息).Error

	return 局订单信息, err == nil
}
func Order取订单详细_第三方订单(第三方订单 string) (DB.DB_LogRMBPayOrder, bool) {
	if 第三方订单 == "" {
		return DB.DB_LogRMBPayOrder{}, false
	}
	var 局订单信息 DB.DB_LogRMBPayOrder
	err := global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).Where("PayOrder2 = ?", 第三方订单).First(&局订单信息).Error

	return 局订单信息, err == nil
}
