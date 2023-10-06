package Ser_RMBPayOrder //生成单号
import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/smartwalle/alipay/v3"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	WXutils "github.com/wechatpay-apiv3/wechatpay-go/utils"
	"log"
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

// Order更新订单状态
// 1  '等待支付'  2  '已付待充' 3 '充值成功' 4 退款中 5 ? 退款失败" : 6退款成功
func Order更新订单状态和第三方订单号(订单号 string, 状态值 int, 第三方订单号 string) bool {
	if 订单号 == "" {
		return false
	}
	err := global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).Where("PayOrder = ?", 订单号).Updates(
		map[string]interface{}{
			"PayOrder2": 第三方订单号,
			"Status":    状态值,
		}).Error

	if err != nil {
		global.GVA_LOG.Error(订单号 + "Order更新订单状态失败:" + err.Error())
		return false
	}
	return true
}

// Order更新订单备注

func Order更新订单备注(订单号 string, 备注 string) bool {
	if 订单号 == "" {
		return false
	}
	err := global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).Where("PayOrder = ?", 订单号).Update("Note", 备注).Error
	if err != nil {
		global.GVA_LOG.Error(订单号 + "Order更新订单备注失败:" + err.Error())
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
func Order更新订单第三方订单ID(订单号 string, PayOrder2 string) bool {
	if 订单号 == "" {
		return false
	}
	err := global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).Where("PayOrder = ?", 订单号).Update("PayOrder2", PayOrder2).Error
	if err != nil {
		global.GVA_LOG.Error(订单号 + "Order更新订单备注失败:" + err.Error())
		return false
	}
	return true
}

func Order更新订单备注和扩展信息(订单号 string, 备注, 扩展信息 string) bool {
	if 订单号 == "" {
		return false
	}

	err := global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).Where("PayOrder = ?", 订单号).Updates(
		map[string]interface{}{
			"Extra": 扩展信息,
			"Note":  备注,
		}).Error
	if err != nil {
		global.GVA_LOG.Error(订单号 + "Order更新订单注和扩展失败:" + err.Error())
		return false
	}
	return true
}

// Order取订单状态
// 1  '等待支付'  2  '已付待充' 3 '充值成功' 4 退款中 5 ? 退款失败" : 6退款成功

func Order取订单状态(订单号 string) int {
	if 订单号 == "" {
		return 0
	}
	var 状态值 int
	_ = global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).Where("PayOrder = ?", 订单号).Select("Status").First(&状态值).Error

	return 状态值
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

// Order更新订单金额, 有些验证会修改实际收款金额
func Order更新订单金额(订单号 string, 金额 float64) bool {
	if 订单号 == "" {
		return false
	}
	err := global.GVA_DB.Model(DB.DB_LogRMBPayOrder{}).Where("PayOrder = ?", 订单号).Update("Rmb", 金额).Error
	if err != nil {
		return false
	}
	return true
}
func Order_退款_支付宝PC(订单信息 DB.DB_LogRMBPayOrder) error {
	var privateKey = global.GVA_CONFIG.Z在线支付.Z支付宝商户私钥 // 必须，上一步中使用 RSA签名验签工具 生成的私钥
	client, err := alipay.New(global.GVA_CONFIG.Z在线支付.Z支付宝商户ID, privateKey, true)
	if err != nil {
		return errors.New("支付宝pc退款商户私钥载入失败")
	}

	err = client.LoadAliPayPublicKey(global.GVA_CONFIG.Z在线支付.Z支付宝公钥) // 加载支付宝公钥证书
	if err != nil {
		if err != nil {
			return errors.New("支付宝pc退款支付宝公钥载入失败")
		}
	}
	var p = alipay.TradeRefund{}

	p.RefundAmount = fmt.Sprintf("%.2f", 订单信息.Rmb)
	p.OutTradeNo = 订单信息.PayOrder
	p.OutRequestNo = strconv.FormatInt(time.Now().Unix(), 10)
	rsp, err := client.TradeRefund(p)
	if err != nil {
		fmt.Printf("%v", err.Error())
		return err
	}
	fmt.Printf("%v", rsp.Content)
	//{40004 Business Failed ACQ.TRADE_HAS_CLOSE 交易已经关闭      0.00  [] 0.00  []}
	//{40004 Business Failed ACQ.TRADE_NOT_EXIST 交易不存在      0.00  [] 0.00  []}
	//{40004 Business Failed ACQ.REASON_TRADE_REFUND_FEE_ERR 退款金额无效  202305161100260001    0.00  [] 0.00  []}
	//{10000 Success   2023051622001414411454629611 202305161100260001 156******66 2088022724614415 Y 0.01  [] 0.00  []}

	if rsp.Content.Code == "10000" {
		return nil
	}
	return errors.New(rsp.Content.SubMsg)
}
func Order_退款_支付宝当面付(订单信息 DB.DB_LogRMBPayOrder) error {
	var privateKey = global.GVA_CONFIG.Z在线支付.Z支付宝当面付商户私钥 // 必须，上一步中使用 RSA签名验签工具 生成的私钥
	client, err := alipay.New(global.GVA_CONFIG.Z在线支付.Z支付宝当面付商户ID, privateKey, true)
	if err != nil {
		return errors.New("支付宝当面付退款商户私钥载入失败")
	}

	err = client.LoadAliPayPublicKey(global.GVA_CONFIG.Z在线支付.Z支付宝当面付公钥) // 加载支付宝公钥证书
	if err != nil {
		if err != nil {
			return errors.New("支付宝当面付退款支付宝公钥载入失败")
		}
	}
	var p = alipay.TradeRefund{}

	p.RefundAmount = fmt.Sprintf("%.2f", 订单信息.Rmb)
	p.OutTradeNo = 订单信息.PayOrder
	p.OutRequestNo = strconv.FormatInt(time.Now().Unix(), 10)
	rsp, err := client.TradeRefund(p)
	if err != nil {
		fmt.Printf("%v", err.Error())
		return err
	}
	fmt.Printf("%v", rsp.Content)
	//{40004 Business Failed ACQ.TRADE_HAS_CLOSE 交易已经关闭      0.00  [] 0.00  []}
	//{40004 Business Failed ACQ.TRADE_NOT_EXIST 交易不存在      0.00  [] 0.00  []}
	//{40004 Business Failed ACQ.REASON_TRADE_REFUND_FEE_ERR 退款金额无效  202305161100260001    0.00  [] 0.00  []}
	//{10000 Success   2023051622001414411454629611 202305161100260001 156******66 2088022724614415 Y 0.01  [] 0.00  []}

	if rsp.Content.Code == "10000" {
		return nil
	}
	return errors.New(rsp.Content.SubMsg)
}
func Order_退款_微信支付(订单信息 DB.DB_LogRMBPayOrder) error {
	var (
		mchID                      string = global.GVA_CONFIG.Z在线支付.W微信支付商户ID    // 商户号
		mchCertificateSerialNumber string = global.GVA_CONFIG.Z在线支付.W微信支付商户证书序列号 // 商户证书序列号
		mchAPIv3Key                string = global.GVA_CONFIG.Z在线支付.W微信支付商户v3密钥  // 商户APIv3密钥
	)

	// 使用 utils 提供的函数从本地文件中加载商户私钥，商户私钥会用来生成请求的签名

	mchPrivateKey, err := WXutils.LoadPrivateKey(global.GVA_CONFIG.Z在线支付.W微信支付商户证书串)
	if err != nil {
		return errors.New("微信支付商户证书串载入失败")
	}

	ctx := context.Background()
	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, mchCertificateSerialNumber, mchPrivateKey, mchAPIv3Key),
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		return errors.New("创建微信退款错误失败请重试")
	}
	//TransactionId: core.String(订单信息.PayOrder),
	svc := refunddomestic.RefundsApiService{Client: client}
	resp, result, err := svc.Create(ctx,
		refunddomestic.CreateRequest{

			OutTradeNo:   core.String(订单信息.PayOrder),
			OutRefundNo:  core.String(订单信息.PayOrder),
			Reason:       core.String("协商退款"),
			NotifyUrl:    core.String(global.GVA_CONFIG.X系统设置.X系统地址 + "/WebApi/PayWxRefundsNotify"),
			FundsAccount: refunddomestic.REQFUNDSACCOUNT_AVAILABLE.Ptr(),
			Amount: &refunddomestic.AmountReq{
				Currency: core.String("CNY"),
				Refund:   core.Int64(int64(int(订单信息.Rmb * 100))),
				Total:    core.Int64(int64(int(订单信息.Rmb * 100))),
			},
		},
	)

	if err != nil {
		// 处理错误
		return errors.New(err.(*core.APIError).Message)
	} else {
		// 处理返回结果
		log.Printf("status=%d resp=%s", result.Response.StatusCode, resp)

		//{RefundId:50300705822023051734540677508, OutRefundNo:202305171343250001, TransactionId:4200001829202305172758614786, OutTradeNo:202305171343250001, Channel:ORIGINAL, UserReceivedAcco
		//unt:支付用户零钱, SuccessTime:<nil>, CreateTime:2023-05-17 14:07:34 +0800 CST, Status:PROCESSING, FundsAccount:AVAILABLE, Amount:Amount{Total:2, Refund:2, From:[], PayerTotal:2, PayerRefund:2, SettlementRefund:2, SettlementTotal:2, DiscountRefund:0, Currency:CNY}, PromotionDetail:[]}
		//{RefundId:50300705822023051734540677508, OutRefundNo:202305171343250001, TransactionId:4200001829202305172758614786, OutTradeNo:202305171343250001, Channel:ORIGINAL, UserReceivedAcco
		//unt:支付用户零钱, SuccessTime:2023-05-17 14:07:41 +0800 CST, CreateTime:2023-05-17 14:07:34 +0800 CST, Status:SUCCESS, FundsAccount:AVAILABLE, Amount:Amount{Total:2, Refund:2, From:[], PayerTotal:2, PayerRefund:2, SettlementRefund:2, SettlementTotal:2, DiscountRefund:0, Currency:CNY}, PromotionDetail:[]}
		return nil
	}
}

func Q取余额充值_支付宝PC支付(局_用户Id int, 充值金额 float64, ip string) (gin.H, error) {

	if !global.GVA_CONFIG.Z在线支付.Z支付宝开关 {
		return gin.H{}, errors.New("当前支付方式已关闭")
	}

	局_用户名 := Ser_User.Id取User(局_用户Id)
	if 局_用户名 == "" {
		return gin.H{}, errors.New("要充值的用户不存在")
	}
	var privateKey = global.GVA_CONFIG.Z在线支付.Z支付宝商户私钥 // 必须，上一步中使用 RSA签名验签工具 生成的私钥
	client, err := alipay.New(global.GVA_CONFIG.Z在线支付.Z支付宝商户ID, privateKey, true)
	if err != nil {
		return gin.H{}, errors.New("系统加载支付宝配置私钥失败,请检查参数")
	}
	err = client.LoadAliPayPublicKey(global.GVA_CONFIG.Z在线支付.Z支付宝公钥) // 加载支付宝公钥证书
	if err != nil {
		return gin.H{}, errors.New("系统加载支付宝配置公钥失败,请检查参数")
	}

	if 充值金额 <= 0 || 充值金额 > float64(global.GVA_CONFIG.Z在线支付.Z支付宝单次最大金额) {
		return gin.H{}, errors.New(fmt.Sprintf("充值金额必须大于0且小于%.2f", global.GVA_CONFIG.Z在线支付.Z支付宝单次最大金额))
	}

	局_余额订单信息, err := Order订单创建(局_用户Id, 1, 充值金额, "支付宝PC", "", ip, 0, "")

	var p = alipay.TradePagePay{}
	p.NotifyURL = global.GVA_CONFIG.X系统设置.X系统地址 + "/WebApi/PayAliNotify"
	p.ReturnURL = global.GVA_CONFIG.Z在线支付.Z支付宝同步回调url
	p.Subject = "用户:" + 局_用户名 + "_充值余额"
	p.OutTradeNo = 局_余额订单信息.PayOrder
	p.TotalAmount = fmt.Sprintf("%.2f", 局_余额订单信息.Rmb)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	url, err := client.TradePagePay(p)
	if err != nil {
		return gin.H{}, errors.New("获取支付链接失败:" + err.Error())
	}
	var payURL = url.String()
	// 这个 payURL 即 是用于支付的 URL，可将输出的内容复制，到浏览器中访问该 URL 即可打开支付页面。
	return gin.H{"PayURL": payURL, "OrderId": 局_余额订单信息.PayOrder, "Status": 1}, nil
}

func Q取余额充值_微信支付支付(局_用户Id int, 充值金额 float64, ip string) (gin.H, error) {

	局_用户名 := Ser_User.Id取User(局_用户Id)
	if 局_用户名 == "" {
		return gin.H{}, errors.New("要充值的用户不存在")
	}

	if 充值金额 <= 0 || 充值金额 > float64(global.GVA_CONFIG.Z在线支付.Z支付宝单次最大金额) {
		return gin.H{}, errors.New(fmt.Sprintf("充值金额必须大于0且小于%.2f", global.GVA_CONFIG.Z在线支付.Z支付宝单次最大金额))
	}

	局_余额订单信息, err := Order订单创建(局_用户Id, 1, 充值金额, "微信支付", "", ip, 0, "")

	var (
		mchID                      string = global.GVA_CONFIG.Z在线支付.W微信支付商户ID    // 商户号
		mchCertificateSerialNumber string = global.GVA_CONFIG.Z在线支付.W微信支付商户证书序列号 // 商户证书序列号
		mchAPIv3Key                string = global.GVA_CONFIG.Z在线支付.W微信支付商户v3密钥  // 商户APIv3密钥
	)

	// 使用 utils 提供的函数从本地文件中加载商户私钥，商户私钥会用来生成请求的签名

	mchPrivateKey, err := WXutils.LoadPrivateKey(global.GVA_CONFIG.Z在线支付.W微信支付商户证书串)
	if err != nil {
		return gin.H{}, errors.New("微信支付支付Url微信支付商户证书串加载失败")
	}

	ctx := context.Background()
	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, mchCertificateSerialNumber, mchPrivateKey, mchAPIv3Key),
	}
	client, err := core.NewClient(ctx, opts...)

	if err != nil {
		return gin.H{}, errors.New("创建微信支付错误失败::" + err.Error())
	}

	svc := native.NativeApiService{Client: client}
	resp, _, err := svc.Prepay(ctx,
		native.PrepayRequest{
			Appid:         core.String(global.GVA_CONFIG.Z在线支付.W微信支付AppId),
			Mchid:         core.String(mchID),
			Description:   core.String("用户:" + 局_用户名 + "_充值余额"),
			OutTradeNo:    core.String(局_余额订单信息.PayOrder),
			TimeExpire:    core.Time(time.Now().Add(time.Second * time.Duration(300))),
			Attach:        core.String("用户:" + 局_用户名 + "_充值余额"),
			NotifyUrl:     core.String(global.GVA_CONFIG.X系统设置.X系统地址 + "/WebApi/PayWxNotify"),
			GoodsTag:      core.String("WXG"),
			LimitPay:      []string{},
			SupportFapiao: core.Bool(false),
			Amount: &native.Amount{
				Currency: core.String("CNY"),
				Total:    core.Int64(int64(int(局_余额订单信息.Rmb * 100))),
			},
			Detail: &native.Detail{
				CostPrice: core.Int64(608800),
				GoodsDetail: []native.GoodsDetail{{
					GoodsName:        core.String("用户:" + 局_用户名 + "_充值余额"),
					MerchantGoodsId:  core.String(局_余额订单信息.PayOrder),
					Quantity:         core.Int64(1),
					UnitPrice:        core.Int64(int64(int(局_余额订单信息.Rmb * 100))),
					WechatpayGoodsId: core.String("1001"),
				}},
				InvoiceId: core.String(局_余额订单信息.PayOrder),
			},
			SettleInfo: &native.SettleInfo{
				ProfitSharing: core.Bool(false),
			},
		},
	)

	if err != nil {
		return gin.H{}, errors.New("微信支付支付Url获取失败:" + err.(*core.APIError).Body)
	}
	// 处理返回结果
	// 这个 payURL 即 是用于支付的 URL，可将输出的内容复制，到浏览器中访问该 URL 即可打开支付页面。
	/*	log.Printf("status=%d", result.Response.StatusCode)
		log.Printf("status=%v", resp)*/
	return gin.H{"WxPayURL": resp.CodeUrl, "OrderId": 局_余额订单信息.PayOrder, "Status": 1}, nil
}
