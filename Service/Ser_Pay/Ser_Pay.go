package Ser_Pay

import (
	"EFunc/utils"
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"github.com/skip2/go-qrcode"
	"github.com/smartwalle/alipay/v3"
	"github.com/valyala/fastjson"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	WXutils "github.com/wechatpay-apiv3/wechatpay-go/utils"
	"net/url"
	"server/Service/Ser_Ka"
	"server/Service/Ser_RMBPayOrder"
	"server/Service/Ser_User"
	"server/new/app/logic/common/setting"
	DB "server/structs/db"
	utils2 "server/utils"
	"sort"
	"strconv"
	"strings"
	"time"
)

func 支付订单回调关键字转换(回调信息 string, 局_订单信息 DB.DB_LogRMBPayOrder) string {
	ReturnURL := strings.Replace(回调信息, "{OrderId}", 局_订单信息.PayOrder, -1)
	ReturnURL = strings.Replace(ReturnURL, "{OrderId2}", 局_订单信息.PayOrder2, -1)
	ReturnURL = strings.Replace(ReturnURL, "{User}", 局_订单信息.User, -1)
	ReturnURL = strings.Replace(ReturnURL, "{Type}", 局_订单信息.Type, -1)
	ReturnURL = strings.Replace(ReturnURL, "{ProcessingType}", strconv.Itoa(局_订单信息.ProcessingType), -1)
	ReturnURL = strings.Replace(ReturnURL, "{Extra}", 局_订单信息.Extra, -1)
	return ReturnURL
}

func Pay_取支付通道状态() gin.H {
	局_支付配置 := setting.Q在线支付配置()
	局map := gin.H{}

	if 局_支付配置.Z支付宝显示名称 != "" {
		局map[局_支付配置.Z支付宝显示名称] = 局_支付配置.Z支付宝开关
	} else {
		局map["支付宝PC"] = 局_支付配置.Z支付宝开关
	}

	if 局_支付配置.Z支付宝当面付显示名称 != "" {
		局map[局_支付配置.Z支付宝当面付显示名称] = 局_支付配置.Z支付宝当面付开关
	} else {
		局map["支付宝当面付"] = 局_支付配置.Z支付宝当面付开关
	}
	if 局_支付配置.Z支付宝H5显示名称 != "" {
		局map[局_支付配置.Z支付宝H5显示名称] = 局_支付配置.Z支付宝H5开关
	} else {
		局map["支付宝H5"] = 局_支付配置.Z支付宝H5开关
	}

	if 局_支付配置.W微信支付显示名称 != "" {
		局map[局_支付配置.W微信支付显示名称] = 局_支付配置.W微信支付开关
	} else {
		局map["微信支付"] = 局_支付配置.W微信支付开关
	}

	if 局_支付配置.X小叮当支付显示名称 != "" {
		局map[局_支付配置.X小叮当支付显示名称] = 局_支付配置.X小叮当支付开关
	} else {
		局map["小叮当"] = 局_支付配置.X小叮当支付开关
	}

	if 局_支付配置.H虎皮椒支付显示名称 != "" {
		局map[局_支付配置.H虎皮椒支付显示名称] = 局_支付配置.H虎皮椒支付开关
	} else {
		局map["虎皮椒"] = 局_支付配置.H虎皮椒支付开关
	}

	return 局map
}

func Pay_显示名称转原名(显示名称 string) string {
	局_支付配置 := setting.Q在线支付配置()
	//修改支付显示别名为原名称
	switch 显示名称 {
	case 局_支付配置.Z支付宝显示名称:
		return "支付宝PC"
	case 局_支付配置.Z支付宝H5显示名称:
		return "支付宝H5"
	case 局_支付配置.Z支付宝当面付显示名称:
		return "支付宝当面付"
	case 局_支付配置.W微信支付显示名称:
		return "微信支付"
	case 局_支付配置.X小叮当支付显示名称:
		return "小叮当"
	case 局_支付配置.H虎皮椒支付显示名称:
		return "虎皮椒"
	default:
		return 显示名称
	}
}

const D订单_处理类型_余额充值 = 0
const D订单_处理类型_购卡直冲 = 1
const D订单_处理类型_积分充值 = 2
const D订单_处理类型_支付购卡 = 3

// Uid类型 1账号 2卡号
// 0 余额充值 1 购卡直冲 2 积分充值  3 支付购卡
func Pay_支付宝Pc_订单创建(Uid, Uid类型 int, 支付金额 float64, ip string, 处理类型 int, 处理类型额外信息 string) (error, gin.H) {
	局_支付配置 := setting.Q在线支付配置()
	if !局_支付配置.Z支付宝开关 {
		return errors.New(局_支付配置.Z支付宝显示名称 + "支付方式已关闭"), gin.H{}
	}

	var privateKey = 局_支付配置.Z支付宝商户私钥 // 必须，上一步中使用 RSA签名验签工具 生成的私钥
	client, err := alipay.New(局_支付配置.Z支付宝商户ID, privateKey, true)
	if err != nil {
		return errors.New(局_支付配置.Z支付宝显示名称 + "商户私钥载入失败:" + err.Error()), gin.H{}
	}

	err = client.LoadAliPayPublicKey(局_支付配置.Z支付宝公钥) // 加载支付宝公钥证书
	if err != nil {
		return errors.New(局_支付配置.Z支付宝显示名称 + "公钥载入失败:" + err.Error()), gin.H{}
	}
	if 支付金额 <= 0 {
		return errors.New("支付金额必须大于0"), gin.H{}
	}
	if 处理类型 == 0 || 处理类型 == 2 { //余额充值 和 积分充值判断单次最大金额
		if 支付金额 > float64(局_支付配置.Z支付宝单次最大金额) {
			return errors.New("支付金额必须小于" + strconv.Itoa(局_支付配置.Z支付宝单次最大金额)), gin.H{}
		}
	}

	局_订单信息, err := Ser_RMBPayOrder.Order订单创建(Uid, Uid类型, 支付金额, "支付宝PC", "", ip, 处理类型, 处理类型额外信息)
	局_用户提示信息, err2 := 取提示信息(局_订单信息, Uid, Uid类型)
	if err2 != nil {
		return err2, gin.H{}
	}

	var p = alipay.TradePagePay{}
	p.NotifyURL = setting.Q系统设置().X系统地址 + "/WebApi/PayAliNotify"
	p.ReturnURL = 支付订单回调关键字转换(局_支付配置.Z支付宝同步回调url, 局_订单信息)
	p.Subject = 局_用户提示信息
	p.OutTradeNo = 局_订单信息.PayOrder
	p.TotalAmount = fmt.Sprintf("%.2f", 局_订单信息.Rmb)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	url2, err := client.TradePagePay(p)
	if err != nil {
		return errors.New(局_支付配置.Z支付宝显示名称 + "支付Url获取失败:" + err.Error()), gin.H{}
	}
	var payURL = url2.String()

	return nil, gin.H{"Status": 1, "PayURL": payURL, "OrderId": 局_订单信息.PayOrder}
}

// Uid类型 1账号 2卡号
// 0 余额充值 1 购卡直冲 2 积分充值  3 支付购卡
func Pay_支付宝H5_订单创建(Uid, Uid类型 int, 支付金额 float64, ip string, 处理类型 int, 处理类型额外信息 string) (error, gin.H) {
	局_支付配置 := setting.Q在线支付配置()
	if !局_支付配置.Z支付宝H5开关 {
		return errors.New(局_支付配置.Z支付宝H5显示名称 + "支付方式已关闭"), gin.H{}
	}

	var privateKey = 局_支付配置.Z支付宝H5商户私钥 // 必须，上一步中使用 RSA签名验签工具 生成的私钥
	client, err := alipay.New(局_支付配置.Z支付宝H5商户ID, privateKey, true)
	if err != nil {
		return errors.New(局_支付配置.Z支付宝H5显示名称 + "商户私钥载入失败:" + err.Error()), gin.H{}
	}

	err = client.LoadAliPayPublicKey(局_支付配置.Z支付宝H5公钥) // 加载支付宝手机网站公钥证书
	if err != nil {
		return errors.New(局_支付配置.Z支付宝H5显示名称 + "公钥载入失败:" + err.Error()), gin.H{}
	}

	if 支付金额 <= 0 {
		return errors.New("支付金额必须大于0"), gin.H{}
	}
	if 处理类型 == 0 || 处理类型 == 2 { //余额充值 和 积分充值判断单次最大金额
		if 支付金额 > float64(局_支付配置.Z支付宝H5单次最大金额) {
			return errors.New("支付金额必须小于" + strconv.Itoa(局_支付配置.Z支付宝H5单次最大金额)), gin.H{}
		}
	}

	局_订单信息, err := Ser_RMBPayOrder.Order订单创建(Uid, Uid类型, 支付金额, "支付宝H5", "", ip, 处理类型, 处理类型额外信息)
	局_用户提示信息, err2 := 取提示信息(局_订单信息, Uid, Uid类型)
	if err2 != nil {
		return err2, gin.H{}
	}

	var p = alipay.TradeWapPay{}
	p.NotifyURL = setting.Q系统设置().X系统地址 + "/WebApi/PayAliNotifyH5"

	p.ReturnURL = 支付订单回调关键字转换(局_支付配置.Z支付宝H5同步回调url, 局_订单信息)
	p.Subject = 局_用户提示信息
	p.OutTradeNo = 局_订单信息.PayOrder
	p.TotalAmount = fmt.Sprintf("%.2f", 局_订单信息.Rmb)
	p.ProductCode = "QUICK_WAP_WAY"

	url2, err := client.TradeWapPay(p)
	if err != nil {
		return errors.New(局_支付配置.Z支付宝H5显示名称 + "支付Url获取失败:" + err.Error()), gin.H{}
	}
	var payURL = url2.String()

	return nil, gin.H{"Status": 1, "PayURL": payURL, "OrderId": 局_订单信息.PayOrder}
}

// Uid类型 1账号 2卡号
// 0 余额充值 1 购卡直冲 2 应用积分充值
func Pay_支付宝当面付_订单创建(Uid, Uid类型 int, 支付金额 float64, ip string, 处理类型 int, 处理类型额外信息 string) (error, gin.H) {
	局_支付配置 := setting.Q在线支付配置()
	if !局_支付配置.Z支付宝当面付开关 {
		return errors.New(局_支付配置.Z支付宝当面付显示名称 + "支付方式已关闭"), gin.H{}
	}

	if 局_支付配置.Z支付宝当面付商户私钥 == "" || 局_支付配置.Z支付宝当面付商户ID == "" || 局_支付配置.Z支付宝当面付公钥 == "" {
		return errors.New(局_支付配置.Z支付宝当面付显示名称 + "服务端未配置参数"), gin.H{}
	}

	var privateKey = 局_支付配置.Z支付宝当面付商户私钥 // 必须，上一步中使用 RSA签名验签工具 生成的私钥
	client, err := alipay.New(局_支付配置.Z支付宝当面付商户ID, privateKey, true)
	if err != nil {
		return errors.New(局_支付配置.Z支付宝当面付显示名称 + "支付商户私钥载入失败:" + err.Error()), gin.H{}
	}

	err = client.LoadAliPayPublicKey(局_支付配置.Z支付宝当面付公钥) // 加载支付宝公钥证书
	if err != nil {
		return errors.New(局_支付配置.Z支付宝当面付显示名称 + "公钥载入失败:" + err.Error()), gin.H{}
	}

	if 支付金额 <= 0 {
		return errors.New("支付金额必须大于0"), gin.H{}
	}
	if 处理类型 == 0 || 处理类型 == 2 { //余额充值 和 积分充值判断单次最大金额
		if 支付金额 > float64(局_支付配置.Z支付宝当面付单次最大金额) {
			return errors.New("支付金额必须小于" + strconv.Itoa(局_支付配置.Z支付宝当面付单次最大金额)), gin.H{}
		}
	}

	局_订单信息, err := Ser_RMBPayOrder.Order订单创建(Uid, Uid类型, 支付金额, "支付宝当面付", "", ip, 处理类型, 处理类型额外信息)

	局_用户提示信息, err2 := 取提示信息(局_订单信息, Uid, Uid类型)
	if err2 != nil {
		return err2, gin.H{}
	}

	var p = alipay.TradePreCreate{}
	p.NotifyURL = setting.Q系统设置().X系统地址 + "/WebApi/PayAliNotifyDangMianFu"
	p.ReturnURL = 支付订单回调关键字转换(局_支付配置.Z支付宝当面付同步回调url, 局_订单信息)
	p.Subject = 局_用户提示信息
	p.OutTradeNo = 局_订单信息.PayOrder
	p.TotalAmount = fmt.Sprintf("%.2f", 局_订单信息.Rmb)
	//p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	rsp, err := client.TradePreCreate(p)
	if err != nil {
		return errors.New(局_支付配置.Z支付宝当面付显示名称 + "当面付支付Url获取失败:" + err.Error()), gin.H{}
	}
	if rsp.Content.Code != alipay.CodeSuccess {
		return errors.New(局_支付配置.Z支付宝当面付显示名称 + "支付Url获取失败:" + rsp.Content.Msg + "|" + rsp.Content.SubMsg), gin.H{}
	}

	var QRCode = rsp.Content.QRCode

	return nil, gin.H{"Status": 1, "PayQRCode": QRCode, "PayQRCodePNG": 生成二维码并转base64(QRCode), "OrderId": 局_订单信息.PayOrder}

}

func 取提示信息(局_订单信息 DB.DB_LogRMBPayOrder, Uid, Uid类型 int) (string, error) {

	局_用户名 := ""
	局_用户名类型提示 := "账号"
	if Uid类型 == 2 {
		局_用户名 = Ser_Ka.Id取卡号(Uid)
		局_用户名类型提示 = "卡号"
	} else {
		局_用户名 = Ser_User.Id取User(Uid)
	}

	if 局_用户名 == "" && 局_订单信息.ProcessingType == D订单_处理类型_支付购卡 {
		return "支付购卡:" + 局_用户名 + "_" + Ser_RMBPayOrder.C处理类型[局_订单信息.ProcessingType], nil
	}

	if 局_用户名 == "" {
		return "", errors.New(局_用户名类型提示 + "不存在")
	}

	return "用户:" + 局_用户名 + "_" + Ser_RMBPayOrder.C处理类型[局_订单信息.ProcessingType], nil
}
func 生成二维码并转base64(内容 string) string {
	局_二维码base64 := ""
	png, err := qrcode.Encode(内容, qrcode.Medium, 256)
	if err == nil {
		局_二维码base64 = base64.StdEncoding.EncodeToString(png)
	}
	return 局_二维码base64
}

// Uid类型 1账号 2卡号
// 0 余额充值 1 购卡直冲 2 应用积分充值
func Pay_微信Pc_订单创建(Uid, Uid类型 int, 支付金额 float64, ip string, 处理类型 int, 处理类型额外信息 string) (error, gin.H) {
	局_支付配置 := setting.Q在线支付配置()
	if !局_支付配置.W微信支付开关 {
		return errors.New(局_支付配置.W微信支付显示名称 + "支付方式已关闭"), gin.H{}
	}

	if 支付金额 <= 0 {
		return errors.New("支付金额必须大于0"), gin.H{}
	}
	if 处理类型 == 0 || 处理类型 == 2 { //余额充值 和 积分充值判断单次最大金额
		if 支付金额 > float64(局_支付配置.W微信支付单次最大金额) {
			return errors.New("支付金额必须小于" + strconv.Itoa(局_支付配置.W微信支付单次最大金额)), gin.H{}
		}
	}

	局_订单信息, err := Ser_RMBPayOrder.Order订单创建(Uid, Uid类型, 支付金额, "微信支付", "", ip, 处理类型, 处理类型额外信息)
	局_用户提示信息, err2 := 取提示信息(局_订单信息, Uid, Uid类型)
	if err2 != nil {
		return err2, gin.H{}
	}
	var (
		mchID                      string = 局_支付配置.W微信支付商户ID    // 商户号
		mchCertificateSerialNumber string = 局_支付配置.W微信支付商户证书序列号 // 商户证书序列号
		mchAPIv3Key                string = 局_支付配置.W微信支付商户v3密钥  // 商户APIv3密钥
	)

	// 使用 utils 提供的函数从本地文件中加载商户私钥，商户私钥会用来生成请求的签名

	mchPrivateKey, err := WXutils.LoadPrivateKey(局_支付配置.W微信支付商户证书串)
	if err != nil {
		return errors.New(局_支付配置.W微信支付显示名称 + "Url微信支付商户证书串加载失败"), gin.H{}
	}

	ctx := context.Background()
	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, mchCertificateSerialNumber, mchPrivateKey, mchAPIv3Key),
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		return errors.New(局_支付配置.W微信支付显示名称 + "创建错误失败请重试"), gin.H{}
	}

	svc := native.NativeApiService{Client: client}
	resp, _, err := svc.Prepay(ctx,
		native.PrepayRequest{
			Appid:         core.String(局_支付配置.W微信支付AppId),
			Mchid:         core.String(mchID),
			Description:   core.String(局_用户提示信息),
			OutTradeNo:    core.String(局_订单信息.PayOrder),
			TimeExpire:    core.Time(time.Now().Add(time.Second * time.Duration(300))),
			Attach:        core.String(局_用户提示信息),
			NotifyUrl:     core.String(setting.Q系统设置().X系统地址 + "/WebApi/PayWxNotify"),
			GoodsTag:      core.String("WXG"),
			LimitPay:      []string{},
			SupportFapiao: core.Bool(false),
			Amount: &native.Amount{
				Currency: core.String("CNY"),
				Total:    core.Int64(int64(int(局_订单信息.Rmb * 100))),
			},
			Detail: &native.Detail{
				CostPrice: core.Int64(608800),
				GoodsDetail: []native.GoodsDetail{{
					GoodsName:        core.String(局_用户提示信息),
					MerchantGoodsId:  core.String(局_订单信息.PayOrder),
					Quantity:         core.Int64(1),
					UnitPrice:        core.Int64(int64(int(局_订单信息.Rmb * 100))),
					WechatpayGoodsId: core.String("1001"),
				}},
				InvoiceId: core.String(局_订单信息.PayOrder),
			},
			SettleInfo: &native.SettleInfo{
				ProfitSharing: core.Bool(false),
			},
		},
	)

	if err != nil {
		// 处理错误
		return errors.New(局_支付配置.W微信支付显示名称 + "支付Url获取失败:" + err.(*core.APIError).Body), gin.H{}
	}
	// 处理返回结果

	return nil, gin.H{"Status": 1, "PayQRCode": resp.CodeUrl, "PayQRCodePNG": 生成二维码并转base64(*resp.CodeUrl), "OrderId": 局_订单信息.PayOrder}

}

// Uid类型 1账号 2卡号
// 0 余额充值 1 购卡直冲 2 应用积分充值
func Pay_小叮当_订单创建(Uid, Uid类型 int, 支付金额 float64, ip string, 处理类型 int, 处理类型额外信息 string) (error, gin.H) {
	局_支付配置 := setting.Q在线支付配置()
	if !局_支付配置.X小叮当支付开关 {
		return errors.New(局_支付配置.X小叮当支付显示名称 + "支付方式已关闭"), gin.H{}
	}

	if 支付金额 <= 0 {
		return errors.New("支付金额必须大于0"), gin.H{}
	}
	if 处理类型 == 0 || 处理类型 == 2 { //余额充值 和 积分充值判断单次最大金额
		if 支付金额 > float64(局_支付配置.Z支付宝单次最大金额) {
			return errors.New("支付金额必须小于" + strconv.Itoa(局_支付配置.Z支付宝单次最大金额)), gin.H{}
		}
	}

	局_订单信息, err := Ser_RMBPayOrder.Order订单创建(Uid, Uid类型, 支付金额, "小叮当", "", ip, 处理类型, 处理类型额外信息)
	局_用户提示信息, err2 := 取提示信息(局_订单信息, Uid, Uid类型)
	if err2 != nil {
		return err2, gin.H{}
	}

	局_网址 := `https://gateway.xddpay.com`
	Http请求 := req.SetRedirectPolicy(req.NoRedirectPolicy()).R()

	//考虑了一下还是项支付宝一样显示url,比较好
	/*	局_网址 += `?format={format}`
		Http请求.SetPathParam(`format`, `json`)*/
	values := url.Values{}
	values.Set("order_no", 局_订单信息.PayOrder)
	values.Set("subject", 局_用户提示信息)
	values.Set("pay_type", strconv.Itoa(局_支付配置.X小叮当支付类型))
	values.Set("money", utils.Float64到文本(局_订单信息.Rmb, 2))
	values.Set("app_id", 局_支付配置.X小叮当app_id)
	values.Set("extra", "")

	局_sign := utils2.Md5String(fmt.Sprintf("order_no=%s&subject=%s&pay_type=%s&money=%s&app_id=%s&extra=%s&%s",
		values.Get("order_no"),
		values.Get("subject"),
		values.Get("pay_type"),
		values.Get("money"),
		values.Get("app_id"),
		values.Get("extra"),
		局_支付配置.X小叮当接口密钥,
	))

	values.Set("sign", 局_sign)
	post数据 := values.Encode()
	Http请求.SetBodyString(post数据)
	Http请求.SetHeader("Content-Type", "application/x-www-form-urlencoded")

	var 局_请求结果 *req.Response

	for i := 0; i < 3; i++ { // 重试三次防止意外
		局_请求结果, err = Http请求.Post(局_网址)
		if len(局_请求结果.Bytes()) > 0 || err != nil {
			break
		}
	}

	//<html><head><title>Object moved</title></head><body>
	//<h2>Object moved to <a href="/pay/?order_no=ttt123465&amp;subject=YuEr&amp;pay_type=43&amp;money=0.01&amp;app_id=17088&amp;extra=89757&amp;sign=271971B0E9DBFFD85DDC89083FBAB844&amp;xddpay_order=20230907145540699737&amp;user_id=8109&amp;token=3E8F15B80BDC4FC75F798CABD654921B">here</a>.</h2>
	//</body></html>

	jsonStr := strings.Replace(utils.W文本_取出中间文本(局_请求结果.String(), `Object moved to <a href="`, `"`), "&amp;", "&", -1)

	if jsonStr == "" {
		return errors.New("支付地址获取失败:" + 局_请求结果.String()), gin.H{}
	}

	// 处理返回结果
	return nil, gin.H{"Status": 1, "OrderId": 局_订单信息.PayOrder, "PayURL": 局_网址 + jsonStr}

}

// Uid类型 1账号 2卡号
// 0 余额充值 1 购卡直冲 2 应用积分充值
func Pay_虎皮椒_订单创建(Uid, Uid类型 int, 支付金额 float64, ip string, 处理类型 int, 处理类型额外信息 string) (error, gin.H) {
	局_支付配置 := setting.Q在线支付配置()
	if !局_支付配置.H虎皮椒支付开关 {
		return errors.New(局_支付配置.H虎皮椒支付显示名称 + "支付方式已关闭"), gin.H{}
	}

	if 支付金额 <= 0 {
		return errors.New("支付金额必须大于0"), gin.H{}
	}
	if 处理类型 == 0 || 处理类型 == 2 { //余额充值 和 积分充值判断单次最大金额
		if 支付金额 > float64(局_支付配置.Z支付宝单次最大金额) {
			return errors.New("支付金额必须小于" + strconv.Itoa(局_支付配置.Z支付宝单次最大金额)), gin.H{}
		}
	}

	局_订单信息, err := Ser_RMBPayOrder.Order订单创建(Uid, Uid类型, 支付金额, "虎皮椒", "", ip, 处理类型, 处理类型额外信息)
	局_用户提示信息, err2 := 取提示信息(局_订单信息, Uid, Uid类型)
	if err2 != nil {
		return err2, gin.H{}
	}
	局_网址 := `https://api.xunhupay.com/payment/do.html`
	if len(局_支付配置.H虎皮椒支付网关) > 10 {
		局_网址 = 局_支付配置.H虎皮椒支付网关
	}

	params := map[string]string{
		"version":        "1.1",
		"trade_order_id": 局_订单信息.PayOrder,
		"total_fee":      utils.Float64到文本(局_订单信息.Rmb, 2),
		"title":          局_用户提示信息,
		"notify_url":     setting.Q系统设置().X系统地址 + "/WebApi/PayHuPiJiaoNotify",
		"return_url":     局_支付配置.H虎皮椒同步回调url,
		"wap_name":       局_用户提示信息,
		"callback_url":   "",
		"time":           strconv.FormatInt(time.Now().Unix(), 10),
		"appid":          局_支付配置.H虎皮椒appId,
		"nonce_str":      strconv.FormatInt(time.Now().Unix(), 10),
	}
	data := url.Values{}
	for k, v := range params {
		data.Add(k, v)
	}
	data.Add("hash", Sign_虎皮椒(局_支付配置.H虎皮椒appSecret, params))
	post数据 := data.Encode()
	Http请求 := req.SetRedirectPolicy(req.NoRedirectPolicy()).R()
	Http请求.SetBodyString(post数据)
	Http请求.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	var 局_请求结果 *req.Response
	for i := 0; i < 3; i++ { // 重试三次防止意外
		局_请求结果, err = Http请求.Post(局_网址)
		if len(局_请求结果.Bytes()) > 0 || err != nil {
			break
		}
	}
	if err != nil || 局_请求结果.String() == "" {
		return errors.New("支付地址获取失败:" + 局_请求结果.String()), gin.H{}
	}

	/*	{
		"openid":"2019081202",
		"url":"https:\/\/api.xunhupay.com\/alipay\/pay\/index.html?id=20351731&nonce_str=3642452019&time=1522390464&appid=20146122002&hash=ef07fb856239c6066a8c84c21835e047",
		"errcode":0,
		"errmsg":"success!",
		"hash":"3a91e22ee359c914b0788c6007377638"
	}*/

	parse, err := fastjson.Parse(局_请求结果.String())
	if err != nil {
		return errors.New("支付信息解析失败:" + 局_请求结果.String()), gin.H{}
	}

	局_url := string(parse.GetStringBytes("url"))
	url_qrcode := string(parse.GetStringBytes("url_qrcode"))

	// 处理返回结果
	return nil, gin.H{"Status": 1, "OrderId": 局_订单信息.PayOrder, "PayURL": 局_url, "url_qrcode": url_qrcode}

}

// Sign 签名方法
func Sign_虎皮椒(appSecret string, params map[string]string) string {
	var data string
	keys := make([]string, 0, 0)
	for key, _ := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	//拼接
	for _, k := range keys {
		data = fmt.Sprintf("%s%s=%s&", data, k, params[k])
	}
	data = strings.Trim(data, "&")
	data = fmt.Sprintf("%s%s", data, appSecret)
	m := md5.New()
	m.Write([]byte(data))
	sign := fmt.Sprintf("%x", m.Sum(nil))
	return sign
}
