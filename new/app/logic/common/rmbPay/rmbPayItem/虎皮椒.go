package rmbPayItem

import (
	"EFunc/utils"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"github.com/valyala/fastjson"
	"io/ioutil"
	"net/http"
	"net/url"
	"server/new/app/logic/common/rmbPay"
	m "server/new/app/models/common"
	"server/new/app/models/constant"
	"sort"
	"strconv"
	"strings"
	"time"
)

func init() {
	rmbPay.L_rmbPay.Z注册接口(pay_虎皮椒)
}

var pay_虎皮椒 虎皮椒

type 虎皮椒 struct {
}

func (j 虎皮椒) Q取通道名称() string {
	return "虎皮椒"
}

// 当无法通过订单号,获取订单信息时将循环每个接口,尝试获取订单号
func (j 虎皮椒) Q取订单id(c *gin.Context, 参数 *m.PayParams) string {
	return ""
}
func (j 虎皮椒) D订单创建(c *gin.Context, 参数 *m.PayParams) (response m.Request, err error) {
	var 局_支付配置 m.Z在线支付_虎皮椒
	err = json.Unmarshal(参数.Z支付配置, &局_支付配置)

	if err != nil || !局_支付配置.H虎皮椒支付开关 {
		err = errors.New(局_支付配置.H虎皮椒支付显示名称 + "支付方式已关闭")
		return
	}
	if 局_支付配置.H虎皮椒appSecret == "" || 局_支付配置.H虎皮椒appId == "" {
		err = errors.New(局_支付配置.H虎皮椒支付显示名称 + "服务端未配置参数")
		return
	}
	if 参数.ProcessingType == constant.D订单类型_余额充值 || 参数.ProcessingType == constant.D订单类型_积分充值 { //余额充值 和 积分充值判断单次最大金额
		if 参数.Rmb > float64(局_支付配置.H虎皮椒单次最大金额) {
			err = errors.New("支付金额必须小于" + strconv.Itoa(局_支付配置.H虎皮椒单次最大金额))
			return
		}
	}

	局_网址 := `https://api.xunhupay.com/payment/do.html`
	if len(局_支付配置.H虎皮椒支付网关) > 10 {
		局_网址 = 局_支付配置.H虎皮椒支付网关
	}

	params := map[string]string{
		"version":        "1.1",
		"trade_order_id": 参数.PayOrder,
		"total_fee":      utils.Float64到文本(参数.Rmb, 2),
		"title":          参数.S商品名称,
		"notify_url":     参数.Y异步回调地址,
		"return_url":     局_支付配置.H虎皮椒同步回调url,
		"wap_name":       参数.S商品名称,
		"callback_url":   "",
		"time":           strconv.FormatInt(time.Now().Unix(), 10),
		"appid":          局_支付配置.H虎皮椒appId,
		"nonce_str":      strconv.FormatInt(time.Now().Unix(), 10),
	}
	data := url.Values{}
	for k, v := range params {
		data.Add(k, v)
	}
	data.Add("hash", j.Sign_虎皮椒(局_支付配置.H虎皮椒appSecret, params))
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
		err = errors.Join(err, errors.New("支付地址获取失败:"+局_请求结果.String()))
		return
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
		err = errors.Join(err, errors.New("支付信息解析失败:"+局_请求结果.String()))
		return
	}

	局_url := string(parse.GetStringBytes("url"))
	url_qrcode := string(parse.GetStringBytes("url_qrcode"))
	response = m.Request{
		Status:    1,
		PayURL:    局_url,
		PayQRCode: url_qrcode,
		OrderId:   参数.PayOrder,
	}
	if response.PayQRCode != "" {
		response.PayQRCodePNG = rmbPay.L_rmbPay.S生成二维码并转base64(response.PayQRCode)
	}
	return
}
func (j 虎皮椒) D订单退款(c *gin.Context, 参数 *m.PayParams) (err error) {

	return errors.New("支付类型不支持退款")
}
func (j 虎皮椒) D订单回调(c *gin.Context, 参数 *m.PayParams) (响应信息 string, 响应代码 int, err error) {
	defer func() {
		if err != nil {
			响应信息 = "success"
			响应代码 = http.StatusOK
		} else {
			响应信息 = "err"
			响应代码 = http.StatusInternalServerError
		}
	}()

	var 局_支付配置 m.Z在线支付_虎皮椒
	err = json.Unmarshal(参数.Z支付配置, &局_支付配置)
	data, err := ioutil.ReadAll(c.Request.Body)

	//trade_order_id=202401052126060001&total_fee=0.01&transaction_id=4200001926202401054012972507&open_order_id=20241873681&order_title=%E7%94%A8%E6%88%B7%3Aaaaaaa_%E8%B4%AD%E5%8D%A1%E7%9B%B4%E5%86%B2&status=OD&nonce_str=4217541016&time=1704461215&appid=201906157675&hash=9bdfdb7f14a8bf0db85139a17e6b9372
	//fmt.Printf(string(data))
	if err != nil {
		err = errors.Join(err, errors.New("虎皮椒异步回调反序列化失败:"+c.Request.RequestURI+"|"+string(data)))
		return
	}

	formData, err := url.ParseQuery(string(data))
	if err != nil {
		err = errors.Join(err, errors.New("虎皮椒异步回调反序列化失败:"+c.Request.RequestURI+"|"+string(data)))
		return
	}
	callbackData := make(map[string]string)
	for key, value := range formData {
		callbackData[key] = value[0]
	}

	// 验证签名
	appSecret := 局_支付配置.H虎皮椒appSecret
	hash := j.SignHuPiJiao(callbackData, appSecret)
	if callbackData["hash"] != hash {
		err = errors.New("虎皮椒异步回调签名错误:" + c.Request.RequestURI + "|" + string(data))
		return
	}

	// 商户订单ID
	tradeOrderID, ok2 := callbackData["trade_order_id"]
	if !ok2 {
		err = errors.New("虎皮椒异步回调订单参数错误:" + c.Request.RequestURI + "|" + string(data))
		return
	}
	/*	*       'trade_order_id'，商户网站订单ID
		'total_fee',订单支付金额
		'transaction_id',//支付平台订单ID
			'order_date',//支付时间
			'plugins',//自定义插件ID,与支付请求时一致
			'status'=>'OD'//订单状态，OD已支付，WP未支付*/
	if tradeOrderID == 参数.PayOrder && callbackData["status"] == "OD" {
		//这里是支付成功的回调
		参数.PayOrder2, _ = callbackData["transaction_id"]
		err = 参数.E额外信息.Set("订单支付金额", callbackData["total_fee"])
	} else {
		err = errors.New(c.Request.RequestURI + "|" + string(data))
	}
	return
}

// Sign 签名方法
func (j 虎皮椒) Sign_虎皮椒(appSecret string, params map[string]string) string {
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

// 虎皮椒签名计算

func (j 虎皮椒) SignHuPiJiao(params map[string]string, appSecret string) string {
	var data string
	keys := make([]string, 0, 0)
	for key, _ := range params {
		if key != "hash" {
			keys = append(keys, key)
		}

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
