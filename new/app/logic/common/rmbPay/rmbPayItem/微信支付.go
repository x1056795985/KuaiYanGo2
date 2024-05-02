package rmbPayItem

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastjson"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	WXutils "github.com/wechatpay-apiv3/wechatpay-go/utils"
	"log"
	"net/http"
	"server/new/app/logic/common/rmbPay"
	m "server/new/app/models/common"
	"server/new/app/models/constant"
	"strconv"
	"time"
)

func init() {
	rmbPay.L_rmbPay.Z注册接口(pay_微信支付)
}

var pay_微信支付 微信支付

type 微信支付 struct {
}

func (j 微信支付) Q取通道名称() string {
	return "微信支付"
}

// 当无法通过订单号,获取订单信息时将循环每个接口,尝试获取订单号
func (j 微信支付) Q取订单id(c *gin.Context, 参数 *m.PayParams) string {
	return ""
}
func (j 微信支付) D订单创建(c *gin.Context, 参数 *m.PayParams) (response m.Request, err error) {
	var 局_支付配置 m.Z在线支付_微信支付
	err = json.Unmarshal(参数.Z支付配置, &局_支付配置)

	if err != nil || !局_支付配置.W微信支付开关 {
		err = errors.New(局_支付配置.W微信支付显示名称 + "支付方式已关闭")
		return
	}

	if 参数.ProcessingType == constant.D订单类型_余额充值 || 参数.ProcessingType == constant.D订单类型_积分充值 { //余额充值 和 积分充值判断单次最大金额
		if 参数.Rmb > float64(局_支付配置.W微信支付单次最大金额) {
			err = errors.New("支付金额必须小于" + strconv.Itoa(局_支付配置.W微信支付单次最大金额))
			return
		}
	}

	var (
		mchID                      string = 局_支付配置.W微信支付商户ID    // 商户号
		mchCertificateSerialNumber string = 局_支付配置.W微信支付商户证书序列号 // 商户证书序列号
		mchAPIv3Key                string = 局_支付配置.W微信支付商户v3密钥  // 商户APIv3密钥
	)

	// 使用 utils 提供的函数从本地文件中加载商户私钥，商户私钥会用来生成请求的签名

	mchPrivateKey, err := WXutils.LoadPrivateKey(局_支付配置.W微信支付商户证书串)
	if err != nil {
		err = errors.Join(err, errors.New("加载微信支付商户证书串失败"))
		return
	}

	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, mchCertificateSerialNumber, mchPrivateKey, mchAPIv3Key),
	}
	client, err := core.NewClient(c, opts...)
	if err != nil {
		err = errors.Join(err, errors.New("微信支付创建错误失败请重试"))
		return
	}
	svc := native.NativeApiService{Client: client}
	resp, _, err := svc.Prepay(c,
		native.PrepayRequest{
			Appid:         core.String(局_支付配置.W微信支付AppId),
			Mchid:         core.String(mchID),
			Description:   core.String(参数.S商品名称),
			OutTradeNo:    core.String(参数.PayOrder),
			TimeExpire:    core.Time(time.Now().Add(time.Second * time.Duration(300))),
			Attach:        core.String(参数.S商品名称),
			NotifyUrl:     core.String(参数.Y异步回调地址),
			GoodsTag:      core.String("WXG"),
			LimitPay:      []string{},
			SupportFapiao: core.Bool(false),
			Amount: &native.Amount{
				Currency: core.String("CNY"),
				Total:    core.Int64(int64(int(参数.Rmb * 100))),
			},
			Detail: &native.Detail{
				CostPrice: core.Int64(608800),
				GoodsDetail: []native.GoodsDetail{{
					GoodsName:        core.String(参数.S商品名称),
					MerchantGoodsId:  core.String(参数.PayOrder),
					Quantity:         core.Int64(1),
					UnitPrice:        core.Int64(int64(int(参数.Rmb * 100))),
					WechatpayGoodsId: core.String("1001"),
				}},
				InvoiceId: core.String(参数.PayOrder),
			},
			SettleInfo: &native.SettleInfo{
				ProfitSharing: core.Bool(false),
			},
		},
	)

	if err != nil {
		//err.(*core.APIError).Body
		err = errors.Join(err, errors.New("支付Url获取失败"))
		// 处理错误
		return //errors.New(局_支付配置.W微信支付显示名称 + "支付Url获取失败:" + err.(*core.APIError).Body), gin.H{}
	}
	// 处理返回结果
	var QRCode = *resp.CodeUrl
	response = m.Request{
		Status:       1,
		PayQRCode:    QRCode,
		PayQRCodePNG: rmbPay.L_rmbPay.S生成二维码并转base64(QRCode),
		OrderId:      参数.PayOrder,
	}
	return
}
func (j 微信支付) D订单退款(c *gin.Context, 参数 *m.PayParams) (err error) {

	var (
		mchID                      string = 参数.Z支付配置s.W微信支付商户ID    // 商户号
		mchCertificateSerialNumber string = 参数.Z支付配置s.W微信支付商户证书序列号 // 商户证书序列号
		mchAPIv3Key                string = 参数.Z支付配置s.W微信支付商户v3密钥  // 商户APIv3密钥
	)

	// 使用 utils 提供的函数从本地文件中加载商户私钥，商户私钥会用来生成请求的签名

	mchPrivateKey, err := WXutils.LoadPrivateKey(参数.Z支付配置s.W微信支付商户证书串)
	if err != nil {
		return errors.New("微信支付商户证书串载入失败")
	}

	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, mchCertificateSerialNumber, mchPrivateKey, mchAPIv3Key),
	}
	client, err := core.NewClient(c, opts...)
	if err != nil {
		return errors.New("创建微信退款错误失败请重试")
	}
	//TransactionId: core.String(订单信息.PayOrder),
	svc := refunddomestic.RefundsApiService{Client: client}
	resp, result, err := svc.Create(c,
		refunddomestic.CreateRequest{
			OutTradeNo:   core.String(参数.PayOrder),
			OutRefundNo:  core.String(参数.PayOrder),
			Reason:       core.String("协商退款"),
			NotifyUrl:    core.String(参数.Y异步回调地址),
			FundsAccount: refunddomestic.REQFUNDSACCOUNT_AVAILABLE.Ptr(),
			Amount: &refunddomestic.AmountReq{
				Currency: core.String("CNY"),
				Refund:   core.Int64(int64(int(参数.Rmb * 100))),
				Total:    core.Int64(int64(int(参数.Rmb * 100))),
			},
		},
	)

	if err != nil {
		// 处理错误
		return errors.New(err.(*core.APIError).Message)
	} else {
		// 处理返回结果
		参数.Status = constant.D订单状态_退款中 //微信需要等回调才能知道是否退款成功
		log.Printf("status=%d resp=%s", result.Response.StatusCode, resp)

		//{RefundId:50300705822023051734540677508, OutRefundNo:202305171343250001, TransactionId:4200001829202305172758614786, OutTradeNo:202305171343250001, Channel:ORIGINAL, UserReceivedAcco
		//unt:支付用户零钱, SuccessTime:<nil>, CreateTime:2023-05-17 14:07:34 +0800 CST, Status:PROCESSING, FundsAccount:AVAILABLE, Amount:Amount{Total:2, Refund:2, From:[], PayerTotal:2, PayerRefund:2, SettlementRefund:2, SettlementTotal:2, DiscountRefund:0, Currency:CNY}, PromotionDetail:[]}
		//{RefundId:50300705822023051734540677508, OutRefundNo:202305171343250001, TransactionId:4200001829202305172758614786, OutTradeNo:202305171343250001, Channel:ORIGINAL, UserReceivedAcco
		//unt:支付用户零钱, SuccessTime:2023-05-17 14:07:41 +0800 CST, CreateTime:2023-05-17 14:07:34 +0800 CST, Status:SUCCESS, FundsAccount:AVAILABLE, Amount:Amount{Total:2, Refund:2, From:[], PayerTotal:2, PayerRefund:2, SettlementRefund:2, SettlementTotal:2, DiscountRefund:0, Currency:CNY}, PromotionDetail:[]}
		return nil
	}
}
func (j 微信支付) D订单支付回调(c *gin.Context, 参数 *m.PayParams) (响应信息 string, 响应代码 int, err error) {
	defer func() {
		if err == nil {
			响应信息 = "success"
			响应代码 = http.StatusOK
		} else {
			响应信息 = AckFail
			响应代码 = http.StatusInternalServerError
		}
	}()

	var 局_支付配置 m.Z在线支付_微信支付
	err = json.Unmarshal(参数.Z支付配置, &局_支付配置)
	var 局_微信响应 微信回调响应
	err = c.ShouldBindJSON(&局_微信响应)
	if err != nil {
		err = errors.New("微信回调参数解析失败")
		return
	}

	plaintext, err := WXutils.DecryptAES256GCM(
		局_支付配置.W微信支付商户v3密钥,
		局_微信响应.Resource.AssociatedData,
		局_微信响应.Resource.Nonce, 局_微信响应.Resource.Ciphertext,
	)

	if err != nil {
		err = errors.Join(err, errors.New("微信密钥参数加载失败"))
		return
	}

	//{"mchid":"1613740956","appid":"wxeb886f382a7a71be","out_trade_no":"202305171129350001","transaction_id":"4200001827202305179902405083","trade_type":"NATIVE","trade_state":"SUCCESS","trade_state_desc":"支付成功","bank_type":"OTHERS","attach":"用户:aaaaaa_充值余额","success_time":"2023-05-17T11:30:21+08:00","payer":{"openid":"o-qvM6nT1T6lvh5wY-BK08oAdHpI"},"amount":{"total":1,"payer_total":1,"currency":"CNY","payer_currency":"CNY"}}
	//fmt.Printf("微信支付回调:  %v\n %s\n", 局_微信响应, plaintext)
	局_回调, err := fastjson.Parse(plaintext)
	if err != nil {
		err = errors.Join(err, errors.New("微信支付回调请求主体解析失败"))
		return
	}

	if string(局_回调.GetStringBytes("out_trade_no")) == 参数.PayOrder && string(局_回调.GetStringBytes("trade_state")) == "SUCCESS" {
		//这里是支付成功的回调
		参数.PayOrder2 = string(局_回调.GetStringBytes("transaction_id"))
		err = 参数.E额外信息.Set("买家openid", string(局_回调.GetStringBytes("payer", "openid")))
	} else {
		err = errors.New(string(局_回调.GetStringBytes("trade_state")))
	}
	return
}

type 微信回调响应 struct {
	ID           string    `json:"id"`
	CreateTime   time.Time `json:"create_time"`
	ResourceType string    `json:"resource_type"`
	EventType    string    `json:"event_type"`
	Summary      string    `json:"summary"`
	Resource     struct {
		OriginalType   string `json:"original_type"`
		Algorithm      string `json:"algorithm"`
		Ciphertext     string `json:"ciphertext"`
		AssociatedData string `json:"associated_data"`
		Nonce          string `json:"nonce"`
	} `json:"resource"`
}

const (
	AckFail = `<xml><return_code><![CDATA[FAIL]]></return_code></xml>`
)

func (j 微信支付) D订单退款回调(c *gin.Context, 参数 *m.PayParams) (响应信息 string, 响应代码 int, err error) {
	defer func() {
		if err == nil {
			响应信息 = "success"
			响应代码 = http.StatusOK
		} else {
			响应信息 = AckFail
			响应代码 = http.StatusInternalServerError
		}
	}()
	var 局_微信响应 微信回调响应
	plaintext, err := WXutils.DecryptAES256GCM(
		参数.Z支付配置s.W微信支付商户v3密钥,
		局_微信响应.Resource.AssociatedData,
		局_微信响应.Resource.Nonce, 局_微信响应.Resource.Ciphertext,
	)

	if err != nil {
		err = errors.Join(err, errors.New("微信支付退款回调解密失败"))
		return
	}

	//{"mchid":"1613740956","out_trade_no":"202305171129350001","transaction_id":"4200001827202305179902405083","out_refund_no":"202305171129350001","refund_id":"50302406042023051734540094143","refund_status":"SUCCESS","success_time":"2023-05-17T14:15:26+08:00","amount":{"total":1,"refund":1,"payer_total":1,"payer_refund":1},"user_received_account":"支付用户零钱"}
	//fmt.Printf("微信支付退款回调:  %v\n %s\n", 局_微信响应, plaintext)

	局_回调, err := fastjson.Parse(plaintext)
	if err != nil {
		err = errors.Join(err, errors.New("微信支付退款回调解析失败"))
		return
	}

	if string(局_回调.GetStringBytes("out_trade_no")) == 参数.PayOrder && string(局_回调.GetStringBytes("refund_status")) == "SUCCESS" {
		err = 参数.E额外信息.Set("退款金额", string(局_回调.GetStringBytes("amount.refund")))
		err = 参数.E额外信息.Set("退款位置", string(局_回调.GetStringBytes("user_received_account")))
	} else {
		err = errors.New("微信支付退款回调失败" + 局_微信响应.Summary)
		//这里是退款失败的回调
	}
	return
}
