package rmbPayItem

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/smartwalle/alipay/v3"
	"net/http"
	"server/new/app/logic/agent/L_setting"
	"server/new/app/logic/common/rmbPay"
	m "server/new/app/models/common"
	"server/new/app/models/constant"
	"strconv"
	"time"
)

func init() {
	rmbPay.L_rmbPay.Z注册接口(pay_支付宝当面付)
}

var pay_支付宝当面付 支付宝当面付

type 支付宝当面付 struct {
}

func (j 支付宝当面付) Q取通道名称() string {
	return "支付宝当面付"
}

// 当无法通过订单号,获取订单信息时将循环每个接口,尝试获取订单号
func (j 支付宝当面付) Q取订单id(c *gin.Context, 参数 *m.PayParams) string {
	return ""
}
func (j 支付宝当面付) D订单创建(c *gin.Context, 参数 *m.PayParams) (response m.Request, err error) {
	var 局_支付配置 m.Z在线支付_支付宝当面付
	err = json.Unmarshal(参数.Z支付配置, &局_支付配置)

	if err != nil || !局_支付配置.Z支付宝当面付开关 {
		err = errors.New(局_支付配置.Z支付宝当面付显示名称 + "支付方式已关闭")
		return
	}

	if 局_支付配置.Z支付宝当面付商户私钥 == "" || 局_支付配置.Z支付宝当面付商户ID == "" || 局_支付配置.Z支付宝当面付公钥 == "" {
		err = errors.New(局_支付配置.Z支付宝当面付显示名称 + "服务端未配置参数")
		return
	}

	if 参数.ProcessingType == constant.D订单类型_余额充值 || 参数.ProcessingType == constant.D订单类型_积分充值 { //余额充值 和 积分充值判断单次最大金额
		if 参数.Rmb > float64(局_支付配置.Z支付宝当面付单次最大金额) {
			err = errors.New("支付金额必须小于" + strconv.Itoa(局_支付配置.Z支付宝当面付单次最大金额))
			return
		}
	}

	var privateKey = 局_支付配置.Z支付宝当面付商户私钥 // 必须，上一步中使用 RSA签名验签工具 生成的私钥
	client, err := alipay.New(局_支付配置.Z支付宝当面付商户ID, privateKey, true)
	if err != nil {
		err = errors.New(局_支付配置.Z支付宝当面付显示名称 + "支付商户私钥载入失败:" + err.Error())
		return
	}

	err = client.LoadAliPayPublicKey(局_支付配置.Z支付宝当面付公钥) // 加载支付宝公钥证书
	if err != nil {
		err = errors.New(局_支付配置.Z支付宝当面付显示名称 + "公钥载入失败:" + err.Error())
		return
	}

	var p = alipay.TradePreCreate{}
	p.NotifyURL = 参数.Y异步回调地址
	p.ReturnURL = rmbPay.L_rmbPay.Z支付订单回调关键字转换(局_支付配置.Z支付宝当面付同步回调url, 参数)
	p.Subject = 参数.S商品名称
	p.OutTradeNo = 参数.PayOrder
	p.TotalAmount = fmt.Sprintf("%.2f", 参数.Rmb)
	//p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	rsp, err := client.TradePreCreate(p)
	if err != nil {
		err = errors.New(局_支付配置.Z支付宝当面付显示名称 + "当面付支付Url获取失败:" + err.Error())
		return
	}
	if rsp.Content.Code != alipay.CodeSuccess {
		err = errors.New(局_支付配置.Z支付宝当面付显示名称 + "支付Url获取失败:" + rsp.Content.Msg + "|" + rsp.Content.SubMsg)
		return
	}

	var QRCode = rsp.Content.QRCode
	response = m.Request{
		Status:       1,
		PayQRCode:    QRCode,
		PayQRCodePNG: rmbPay.L_rmbPay.S生成二维码并转base64(QRCode),
		OrderId:      参数.PayOrder,
	}
	return
}
func (j 支付宝当面付) D订单退款(c *gin.Context, 参数 *m.PayParams) (err error) {
	支付配置 := 参数.Z支付配置s.Z在线支付_支付宝当面付
	var privateKey = 支付配置.Z支付宝当面付商户私钥 // 必须，上一步中使用 RSA签名验签工具 生成的私钥
	client, err := alipay.New(支付配置.Z支付宝当面付商户ID, privateKey, true)
	if err != nil {
		return errors.Join(err, errors.New("支付宝当面付退款商户私钥载入失败"))
	}

	err = client.LoadAliPayPublicKey(支付配置.Z支付宝当面付公钥) // 加载支付宝公钥证书
	if err != nil {
		if err != nil {
			return errors.Join(err, errors.New("支付宝当面付退款支付宝公钥载入失败"))
		}
	}
	var p = alipay.TradeRefund{}

	p.RefundAmount = fmt.Sprintf("%.2f", 参数.Rmb)
	p.OutTradeNo = 参数.PayOrder
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
func (j 支付宝当面付) D订单支付回调(c *gin.Context, 参数 *m.PayParams) (响应信息 string, 响应代码 int, err error) {
	defer func() {
		if err == nil {
			响应信息 = "success"
			响应代码 = http.StatusOK
		} else {
			响应信息 = "err"
			响应代码 = http.StatusInternalServerError
		}
	}()

	var 局_支付配置 m.Z在线支付_支付宝当面付
	if 参数.ReceivedUid == 0 {
		err = json.Unmarshal(参数.Z支付配置, &局_支付配置)
	} else {
		局_临时, err2 := L_setting.Q取代理在线支付信息(c, 参数.ReceivedUid)
		if err2 != nil {
			err = errors.Join(errors.New("Q取代理在线支付信息"), err2)
			return
		}
		局_支付配置 = 局_临时.Z在线支付_支付宝当面付
	}

	var privateKey = 局_支付配置.Z支付宝当面付商户私钥 // 必须，上一步中使用 RSA签名验签工具 生成的私钥
	client, err := alipay.New(局_支付配置.Z支付宝当面付商户ID, privateKey, true)
	if err != nil {
		err = errors.Join(err, errors.New(局_支付配置.Z支付宝当面付显示名称+"支付商户私钥载入失败"))
		return
	}
	err = client.LoadAliPayPublicKey(局_支付配置.Z支付宝当面付公钥) // 加载支付宝当面付公钥证书
	if err != nil {
		err = errors.Join(err, errors.New(局_支付配置.Z支付宝当面付显示名称+"支付商户公钥载入失败"))
		return
	}

	noti, err := client.GetTradeNotification(c.Request) //这里就会校验的
	//fmt.Println(c.Request.PostForm.Encode())
	//app_id=2021001159688744&auth_app_id=2021001159688744&buyer_id=2088022724614415&buyer_pay_amount=0.01&charset=utf-8&fund_bill_list=%5B%7B%22amount%22%3A%220.01%22%2C%22fundChannel%22%3A%22ALIPAYACCOUNT%22%7D%5D&gmt_create=202
	//3-05-16+11%3A14%3A37&gmt_payment=2023-05-16+11%3A14%3A48&invoice_amount=0.01&notify_id=2023051601222111448014411420706088&notify_time=2023-05-16+11%3A14%3A48&notify_type=trade_status_sync&out_trade_no=202305161113450001&poin
	//t_amount=0.00&receipt_amount=0.01&seller_id=2088422339120873&sign=AOGgQPzmHf1aTY695Ey39sxAni7J5EvZybD%2BOvBDfWMUSWRDAJm72Ciy4Rz3cxXYsfZO1t61qKKGVAjNoVDxAZfZdbZrKhk%2BFDRqM7n%2FODPdgI8pelo1NT4Af%2BGcYIF9zkhcmqHcpCJCMeh8yYAPdk
	//WkcTKWaGRwFAIELI9vd8DusrNegDLYKnPCrrNF1U4MSXAbhDXAnu5%2FONWBbWeedyY6xR5R%2BKWDnyWptcZaT8dJAWz23V3dVsH8vLMcv2Dx7q3SL7mQCiA3gAZuI0zitrIKfd7AybKQZD6Vjl%2FOEeyffnaE6D4kEiWOBSfXxwKr9uxPkcaFucoTw0ctWH3B8g%3D%3D&sign_type=RSA2&subject=%E7%94%A8%E6%88%B7aaaaaa%E5%85%85%E5%80%BC&total_amount=0.01&trade_no=2023051622001414411454464620&trade_status=TRADE_SUCCESS&version=1.0
	if err != nil {
		err = errors.Join(err, errors.New(局_支付配置.Z支付宝当面付显示名称+"支付异步回调被异常调用:"+c.Request.RequestURI+"|"+c.Request.PostForm.Encode()))
		return
	}

	//fmt.Printf("订单号:%s;状态:%s\n,%v", noti.OutTradeNo, noti.TradeStatus, noti)
	if 参数.PayOrder == noti.OutTradeNo && noti.TradeStatus == "TRADE_SUCCESS" {
		参数.PayOrder2 = noti.TradeNo
		err = 参数.E额外信息.Set("买家支付宝用户号", noti.BuyerId)
		err = 参数.E额外信息.Set("买家支付宝账号", noti.BuyerLogonId)
		err = 参数.E额外信息.Set("卖家支付宝用户号", noti.SellerId)
		err = 参数.E额外信息.Set("卖家支付宝账号", noti.SellerEmail)
	} else {
		err = errors.New(string(noti.TradeStatus))
	}
	return
}
func (j 支付宝当面付) D订单退款回调(c *gin.Context, 参数 *m.PayParams) (响应信息 string, 响应代码 int, err error) {
	return
}
