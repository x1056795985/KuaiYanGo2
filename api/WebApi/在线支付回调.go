package WebApi

import (
	"EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/smartwalle/alipay/v3"
	"github.com/valyala/fastjson"
	WXutils "github.com/wechatpay-apiv3/wechatpay-go/utils"
	"net/http"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Ka"
	"server/Service/Ser_Log"
	"server/Service/Ser_Pay"
	"server/Service/Ser_RMBPayOrder"
	"server/Service/Ser_User"
	"server/global"
	DB "server/structs/db"
	utils2 "server/utils"
	"strconv"
	"strings"
	"time"
)

// 支付宝PC支付异步回调 Notify - 支付成功后会回调这里;我们可以用来修改订单状态等等
func PayAliNotify(c *gin.Context) {
	var privateKey = global.GVA_CONFIG.Z在线支付.Z支付宝商户私钥 // 必须，上一步中使用 RSA签名验签工具 生成的私钥
	client, err := alipay.New(global.GVA_CONFIG.Z在线支付.Z支付宝商户ID, privateKey, true)
	if err != nil {
		go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, "系统PayAliNotify", "系统内部", global.X系统信息.B版本号当前, "PayAliNotify回调商户私钥载入失败:"+err.Error(), c.ClientIP())
		// 开始时间
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	err = client.LoadAliPayPublicKey(global.GVA_CONFIG.Z在线支付.Z支付宝公钥) // 加载支付宝公钥证书
	if err != nil {
		go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, "系统PayAliNotify", "系统内部", global.X系统信息.B版本号当前, "PayAliNotify回调商户公钥载入失败:"+err.Error(), c.ClientIP())
		// 开始时间
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	noti, err := client.GetTradeNotification(c.Request) //这里就会校验的
	//fmt.Println(c.Request.PostForm.Encode())
	/*
		app_id=2021001159688744&auth_app_id=2021001159688744&buyer_id=2088022724614415&buyer_pay_amount=0.01&charset=utf-8&fund_bill_list=%5B%7B%22amount%22%3A%220.01%22%2C%22fundChannel%22%3A%22ALIPAYACCOUNT%22%7D%5D&gmt_create=202
			3-05-16+11%3A14%3A37&gmt_payment=2023-05-16+11%3A14%3A48&invoice_amount=0.01&notify_id=2023051601222111448014411420706088&notify_time=2023-05-16+11%3A14%3A48&notify_type=trade_status_sync&out_trade_no=202305161113450001&poin
			t_amount=0.00&receipt_amount=0.01&seller_id=2088422339120873&sign=AOGgQPzmHf1aTY695Ey39sxAni7J5EvZybD%2BOvBDfWMUSWRDAJm72Ciy4Rz3cxXYsfZO1t61qKKGVAjNoVDxAZfZdbZrKhk%2BFDRqM7n%2FODPdgI8pelo1NT4Af%2BGcYIF9zkhcmqHcpCJCMeh8yYAPdk
			WkcTKWaGRwFAIELI9vd8DusrNegDLYKnPCrrNF1U4MSXAbhDXAnu5%2FONWBbWeedyY6xR5R%2BKWDnyWptcZaT8dJAWz23V3dVsH8vLMcv2Dx7q3SL7mQCiA3gAZuI0zitrIKfd7AybKQZD6Vjl%2FOEeyffnaE6D4kEiWOBSfXxwKr9uxPkcaFucoTw0ctWH3B8g%3D%3D&sign_type=RSA2&subject=%E7%94%A8%E6%88%B7aaaaaa%E5%85%85%E5%80%BC&total_amount=0.01&trade_no=2023051622001414411454464620&trade_status=TRADE_SUCCESS&version=1.0

	*/
	if err != nil {
		局_boyd := c.Request.PostForm.Encode()

		Ser_Log.Log_写风控日志(0, Ser_Log.Log风控类型_Api异常调用, "WebApi", c.ClientIP(), "支付宝PC异步回调被异常调用:"+c.Request.RequestURI+"|"+局_boyd)
		c.Status(http.StatusInternalServerError)
		return
	}

	fmt.Printf("订单号:%s;状态:%s\n,%v", noti.OutTradeNo, noti.TradeStatus, noti)
	if noti.TradeStatus == "TRADE_SUCCESS" {
		局_订单详细信息, ok := Ser_RMBPayOrder.Order取订单详细(noti.OutTradeNo)
		if ok && 局_订单详细信息.Status == 1 { //有订单且订单信息为未支付
			局_订单详细信息.PayOrder2 = noti.TradeNo
			Z支付成功_后处理(局_订单详细信息)
		}
	}

	c.String(http.StatusOK, "success")
	return
}

// 支付宝当面付支付异步回调 Notify - 支付成功后会回调这里;我们可以用来修改订单状态等等
func PayAliNotify_当面付(c *gin.Context) {
	var privateKey = global.GVA_CONFIG.Z在线支付.Z支付宝当面付商户私钥 // 必须，上一步中使用 RSA签名验签工具 生成的私钥
	client, err := alipay.New(global.GVA_CONFIG.Z在线支付.Z支付宝当面付商户ID, privateKey, true)
	if err != nil {
		go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, "系统PayAliNotify当面付", "系统内部", global.X系统信息.B版本号当前, "PayAliNotify当面付回调商户私钥载入失败:"+err.Error(), c.ClientIP())
		// 开始时间
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	err = client.LoadAliPayPublicKey(global.GVA_CONFIG.Z在线支付.Z支付宝当面付公钥) // 加载支付宝当面付公钥证书
	if err != nil {
		go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, "系统PayAliNotify当面付", "系统内部", global.X系统信息.B版本号当前, "PayAliNotify回调商户公钥载入失败:"+err.Error(), c.ClientIP())
		// 开始时间
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	noti, err := client.GetTradeNotification(c.Request) //这里就会校验的
	//fmt.Println(c.Request.PostForm.Encode())
	//app_id=2021001159688744&auth_app_id=2021001159688744&buyer_id=2088022724614415&buyer_pay_amount=0.01&charset=utf-8&fund_bill_list=%5B%7B%22amount%22%3A%220.01%22%2C%22fundChannel%22%3A%22ALIPAYACCOUNT%22%7D%5D&gmt_create=202
	//3-05-16+11%3A14%3A37&gmt_payment=2023-05-16+11%3A14%3A48&invoice_amount=0.01&notify_id=2023051601222111448014411420706088&notify_time=2023-05-16+11%3A14%3A48&notify_type=trade_status_sync&out_trade_no=202305161113450001&poin
	//t_amount=0.00&receipt_amount=0.01&seller_id=2088422339120873&sign=AOGgQPzmHf1aTY695Ey39sxAni7J5EvZybD%2BOvBDfWMUSWRDAJm72Ciy4Rz3cxXYsfZO1t61qKKGVAjNoVDxAZfZdbZrKhk%2BFDRqM7n%2FODPdgI8pelo1NT4Af%2BGcYIF9zkhcmqHcpCJCMeh8yYAPdk
	//WkcTKWaGRwFAIELI9vd8DusrNegDLYKnPCrrNF1U4MSXAbhDXAnu5%2FONWBbWeedyY6xR5R%2BKWDnyWptcZaT8dJAWz23V3dVsH8vLMcv2Dx7q3SL7mQCiA3gAZuI0zitrIKfd7AybKQZD6Vjl%2FOEeyffnaE6D4kEiWOBSfXxwKr9uxPkcaFucoTw0ctWH3B8g%3D%3D&sign_type=RSA2&subject=%E7%94%A8%E6%88%B7aaaaaa%E5%85%85%E5%80%BC&total_amount=0.01&trade_no=2023051622001414411454464620&trade_status=TRADE_SUCCESS&version=1.0
	if err != nil {
		局_boyd := c.Request.PostForm.Encode()

		Ser_Log.Log_写风控日志(0, Ser_Log.Log风控类型_Api异常调用, "WebApi", c.ClientIP(), "支付宝当面付异步回调被异常调用:"+c.Request.RequestURI+"|"+局_boyd)
		c.Status(http.StatusInternalServerError)
		return
	}

	fmt.Printf("订单号:%s;状态:%s\n,%v", noti.OutTradeNo, noti.TradeStatus, noti)
	if noti.TradeStatus == "TRADE_SUCCESS" {
		局_订单详细信息, ok := Ser_RMBPayOrder.Order取订单详细(noti.OutTradeNo)
		if ok && 局_订单详细信息.Status == 1 { //有订单且订单信息为未支付
			局_订单详细信息.PayOrder2 = noti.TradeNo
			Z支付成功_后处理(局_订单详细信息)
		}
	}

	c.String(http.StatusOK, "success")
	return
}

// 支付宝H5支付异步回调 Notify - 支付成功后会回调这里;我们可以用来修改订单状态等等
func PayAliNotify_H5(c *gin.Context) {
	var privateKey = global.GVA_CONFIG.Z在线支付.Z支付宝H5商户私钥 // 必须，上一步中使用 RSA签名验签工具 生成的私钥
	client, err := alipay.New(global.GVA_CONFIG.Z在线支付.Z支付宝H5商户ID, privateKey, true)
	if err != nil {
		go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, "系统PayAliNotifyH5", "系统内部", global.X系统信息.B版本号当前, "PayAliNotifyH5回调商户私钥载入失败:"+err.Error(), c.ClientIP())
		// 开始时间
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	err = client.LoadAliPayPublicKey(global.GVA_CONFIG.Z在线支付.Z支付宝H5公钥) // 加载支付宝H5公钥证书
	if err != nil {
		go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, "系统PayAliNotifyH5", "系统内部", global.X系统信息.B版本号当前, "PayAliNotifyH5回调商户公钥载入失败:"+err.Error(), c.ClientIP())
		// 开始时间
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	noti, err := client.GetTradeNotification(c.Request) //这里就会校验的
	//fmt.Println(c.Request.PostForm.Encode())
	//app_id=2021001159688744&auth_app_id=2021001159688744&buyer_id=2088022724614415&buyer_pay_amount=0.01&charset=utf-8&fund_bill_list=%5B%7B%22amount%22%3A%220.01%22%2C%22fundChannel%22%3A%22ALIPAYACCOUNT%22%7D%5D&gmt_create=202
	//3-05-16+11%3A14%3A37&gmt_payment=2023-05-16+11%3A14%3A48&invoice_amount=0.01&notify_id=2023051601222111448014411420706088&notify_time=2023-05-16+11%3A14%3A48&notify_type=trade_status_sync&out_trade_no=202305161113450001&poin
	//t_amount=0.00&receipt_amount=0.01&seller_id=2088422339120873&sign=AOGgQPzmHf1aTY695Ey39sxAni7J5EvZybD%2BOvBDfWMUSWRDAJm72Ciy4Rz3cxXYsfZO1t61qKKGVAjNoVDxAZfZdbZrKhk%2BFDRqM7n%2FODPdgI8pelo1NT4Af%2BGcYIF9zkhcmqHcpCJCMeh8yYAPdk
	//WkcTKWaGRwFAIELI9vd8DusrNegDLYKnPCrrNF1U4MSXAbhDXAnu5%2FONWBbWeedyY6xR5R%2BKWDnyWptcZaT8dJAWz23V3dVsH8vLMcv2Dx7q3SL7mQCiA3gAZuI0zitrIKfd7AybKQZD6Vjl%2FOEeyffnaE6D4kEiWOBSfXxwKr9uxPkcaFucoTw0ctWH3B8g%3D%3D&sign_type=RSA2&subject=%E7%94%A8%E6%88%B7aaaaaa%E5%85%85%E5%80%BC&total_amount=0.01&trade_no=2023051622001414411454464620&trade_status=TRADE_SUCCESS&version=1.0
	if err != nil {
		局_boyd := c.Request.PostForm.Encode()

		Ser_Log.Log_写风控日志(0, Ser_Log.Log风控类型_Api异常调用, "WebApi", c.ClientIP(), "支付宝H5异步回调被异常调用:"+c.Request.RequestURI+"|"+局_boyd)
		c.Status(http.StatusInternalServerError)
		return
	}

	fmt.Printf("订单号:%s;状态:%s\n,%v", noti.OutTradeNo, noti.TradeStatus, noti)
	if noti.TradeStatus == "TRADE_SUCCESS" {
		局_订单详细信息, ok := Ser_RMBPayOrder.Order取订单详细(noti.OutTradeNo)
		if ok && 局_订单详细信息.Status == 1 { //有订单且订单信息为未支付
			局_订单详细信息.PayOrder2 = noti.TradeNo
			Z支付成功_后处理(局_订单详细信息)
		}
	}

	c.String(http.StatusOK, "success")
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

// Pay小叮当Notify- 支付成功后会回调这里;我们可以用来修改订单状态等等
func Pay小叮当Notify(c *gin.Context) {
	//order_no=123456&subject=&pay_type=43&money=10.00&realmoney=10.00&result=success&xddpay_order=654321&app_id=10088&extra=abc

	局_sign := fmt.Sprintf("order_no=%s&subject=%s&pay_type=%s&money=%s&realmoney=%s&result=success&xddpay_order=%s&app_id=%s&extra=%s&",
		c.PostForm("order_no"),
		c.PostForm("subject"),
		c.PostForm("pay_type"),
		c.PostForm("money"),
		c.PostForm("realmoney"),
		c.PostForm("xddpay_order"),
		c.PostForm("app_id"),
		c.PostForm("extra"),
	)

	局_sign = utils2.Md5String(局_sign + global.GVA_CONFIG.Z在线支付.X小叮当接口密钥)

	if strings.ToUpper(局_sign) != strings.ToUpper(c.PostForm("sign")) {
		局_boyd := c.Request.PostForm.Encode()
		Ser_Log.Log_写风控日志(0, Ser_Log.Log风控类型_Api异常调用, "WebApi", c.ClientIP(), "小叮当异步回调被异常调用:"+c.Request.RequestURI+"|"+局_boyd)
		return
	}

	if c.PostForm("result") == "success" {
		//这里是支付成功的回调
		局_订单详细信息, ok := Ser_RMBPayOrder.Order取订单详细(c.PostForm("order_no"))

		if ok && 局_订单详细信息.Status == 1 { //有订单且订单信息为未支付

			if c.PostForm("money") != c.PostForm("realmoney") {
				真实金额, err := strconv.ParseFloat(c.PostForm("realmoney"), 64)
				if err == nil {
					Ser_RMBPayOrder.Order更新订单金额(局_订单详细信息.PayOrder, 真实金额)
					局_订单详细信息.Rmb = 真实金额
				}

			}
			局_订单详细信息.PayOrder2 = c.PostForm("xddpay_order")
			Z支付成功_后处理(局_订单详细信息)
		}
	}

	c.String(http.StatusOK, "success")
	return
}

// 微信支付支付异步回调 Notify - 支付成功后会回调这里;我们可以用来修改订单状态等等
func PayWxNotify(c *gin.Context) {
	var 局_微信响应 微信回调响应
	err := c.ShouldBindJSON(&局_微信响应)
	if err != nil {
		c.String(http.StatusInternalServerError, AckFail)
		return
	}

	plaintext, err := WXutils.DecryptAES256GCM(
		global.GVA_CONFIG.Z在线支付.W微信支付商户v3密钥,
		局_微信响应.Resource.AssociatedData,
		局_微信响应.Resource.Nonce, 局_微信响应.Resource.Ciphertext,
	)

	if err != nil {
		局_原文, _ := jsoniter.MarshalToString(&局_微信响应)
		go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, "系统", "系统", global.X系统信息.B版本号当前, "微信回调"+局_原文+"解密失败:"+err.Error(), c.ClientIP())
		return
	}

	//{"mchid":"1613740956","appid":"wxeb886f382a7a71be","out_trade_no":"202305171129350001","transaction_id":"4200001827202305179902405083","trade_type":"NATIVE","trade_state":"SUCCESS","trade_state_desc":"支付成功","bank_type":"OTHERS","attach":"用户:aaaaaa_充值余额","success_time":"2023-05-17T11:30:21+08:00","payer":{"openid":"o-qvM6nT1T6lvh5wY-BK08oAdHpI"},"amount":{"total":1,"payer_total":1,"currency":"CNY","payer_currency":"CNY"}}
	fmt.Printf("微信支付回调:  %v\n %s\n", 局_微信响应, plaintext)
	局_回调, err := fastjson.Parse(plaintext)

	if err != nil {
		go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, "系统", "系统", global.X系统信息.B版本号当前, "微信回调解析失败:"+plaintext, c.ClientIP())
		return
	}

	if string(局_回调.GetStringBytes("trade_state")) == "SUCCESS" {
		//这里是支付成功的回调
		局_订单详细信息, ok := Ser_RMBPayOrder.Order取订单详细(string(局_回调.GetStringBytes("out_trade_no")))

		if ok && 局_订单详细信息.Status == 1 { //有订单且订单信息为未支付
			局_订单详细信息.PayOrder2 = string(局_回调.GetStringBytes("transaction_id"))
			Z支付成功_后处理(局_订单详细信息)
		}
	}

	c.String(http.StatusOK, "success")
	return
}
func Z支付成功_后处理(局_订单详细信息 DB.DB_LogRMBPayOrder) {
	if 局_订单详细信息.Status != 1 {
		return
	}
	var err error
	if 局_订单详细信息.PayOrder2 != "" {
		Ser_RMBPayOrder.Order更新订单状态和第三方订单号(局_订单详细信息.PayOrder, Ser_RMBPayOrder.D订单状态_已付待处理, 局_订单详细信息.PayOrder2) //修改订单信息为已支付 充值
	} else {
		Ser_RMBPayOrder.Order更新订单状态(局_订单详细信息.PayOrder, Ser_RMBPayOrder.D订单状态_已付待处理) //修改订单信息为已支付 充值

	}

	switch 局_订单详细信息.ProcessingType {
	case 0: //余额充值
		_, err = Ser_User.Id余额增减(局_订单详细信息.Uid, 局_订单详细信息.Rmb, true)
		if err != nil {
			Ser_RMBPayOrder.Order更新订单备注(局_订单详细信息.PayOrder, 局_订单详细信息.Note+err.Error())
			go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, "系统", "系统", global.X系统信息.B版本号当前, "订单ID"+局_订单详细信息.PayOrder+"支付异步回调成功,充值余额失败:"+err.Error(), 局_订单详细信息.Ip)
			return
		}
	case 1:
		局_fastjson, err2 := fastjson.Parse(局_订单详细信息.Extra)
		if err2 != nil {
			return
		}

		卡类ID := 局_fastjson.GetInt("KaClassId")
		AppUserId := 局_fastjson.GetInt("AppUserId")
		err12 := Ser_Ka.K卡类直冲_事务(卡类ID, AppUserId, 局_订单详细信息.Ip)
		if err12 != nil {
			Ser_RMBPayOrder.Order更新订单备注(局_订单详细信息.PayOrder, 局_订单详细信息.Note+err12.Error())
			return
		}
		Ser_RMBPayOrder.Order更新订单备注(局_订单详细信息.PayOrder, 局_订单详细信息.Note+"充值卡类ID:"+strconv.Itoa(卡类ID))
	case 2: //积分充值
		局_fastjson, err2 := fastjson.Parse(局_订单详细信息.Extra)
		if err2 != nil {
			return
		}
		AppID := 局_fastjson.GetInt("AppID")
		AppUserId := 局_fastjson.GetInt("AppUserId")
		局_软件用户信息, _ := Ser_AppUser.Id取详情(AppID, AppUserId)

		局_应用信息 := Ser_AppInfo.App取App详情(AppID)
		局_增加积分 := utils.Float64乘int64(局_订单详细信息.Rmb, int64(局_应用信息.RmbToVipNumber))
		局_软件用户信息.VipNumber += 局_增加积分

		err = Ser_AppUser.Id积分增减(AppID, 局_软件用户信息.Id, 局_增加积分, true)
		if err != nil {
			Ser_RMBPayOrder.Order更新订单备注(局_订单详细信息.PayOrder, 局_订单详细信息.Note+err.Error())
			go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, "系统", "系统", global.X系统信息.B版本号当前, "订单ID"+局_订单详细信息.PayOrder+"支付异步回调成功,充值积分失败:"+err.Error(), 局_订单详细信息.Ip)
			return
		}
		Ser_Log.Log_写积分点数时间日志(局_订单详细信息.User, 局_订单详细信息.Ip, fmt.Sprintf("支付订单:%s充值积分|剩余%v", 局_订单详细信息.PayOrder, 局_软件用户信息.VipNumber), 局_软件用户信息.VipNumber, 局_应用信息.AppId, 1)
		Ser_RMBPayOrder.Order更新订单备注(局_订单详细信息.PayOrder, 局_订单详细信息.Note+"充值积分:"+utils.Float64到文本(局_增加积分, 2))
	case Ser_Pay.D订单_处理类型_支付购卡:
		//没有订单信息没有Uid,用户名,需要修改
		局_fastjson, err2 := fastjson.Parse(局_订单详细信息.Extra)
		if err2 != nil {
			Ser_RMBPayOrder.Order更新订单备注(局_订单详细信息.PayOrder, 局_订单详细信息.Note+"附加参数错误:"+局_订单详细信息.Extra)
			return
		}
		卡类ID := 局_fastjson.GetInt("KaClassId")
		局_Ip := string(局_fastjson.GetStringBytes("Ip"))

		局_卡信息, err2 := Ser_Ka.Ka单卡创建(卡类ID, "系统自动", "支付购卡订单ID:"+局_订单详细信息.PayOrder, "", 0)
		if err2 != nil {
			Ser_RMBPayOrder.Order更新订单备注(局_订单详细信息.PayOrder, 局_订单详细信息.Note+err2.Error())
			return
		}
		局_扩展信息 := fmt.Sprintf(`{"Name":"%s","Id":%d}`, 局_卡信息.Name, 局_卡信息.Id)
		Ser_RMBPayOrder.Order更新订单备注和扩展信息(局_订单详细信息.PayOrder, 局_订单详细信息.Note+"购卡:"+局_卡信息.Name, 局_扩展信息)
		局_文本 := fmt.Sprintf("支付购卡订单ID:%s,卡类:%d,消费:%.2f)", 局_订单详细信息.PayOrder, 局_卡信息.KaClassId, 局_订单详细信息.Rmb)

		go Ser_Log.Log_写卡号操作日志("支付购卡", 局_Ip, 局_文本, []string{局_卡信息.Name}, 1, 5)
	default:
		return
	}
	Ser_RMBPayOrder.Order更新订单状态(局_订单详细信息.PayOrder, Ser_RMBPayOrder.D订单状态_成功) //修改订单信息为充值成功

}

// 微信支付支付退款回调 Notify成功后会回调这里;我们可以用来修改订单状态等等
func PayWx退款Notify(c *gin.Context) {
	var 局_微信响应 微信回调响应
	err := c.ShouldBindJSON(&局_微信响应)
	if err != nil {
		c.String(http.StatusInternalServerError, AckFail)
		return
	}

	plaintext, err := WXutils.DecryptAES256GCM(
		global.GVA_CONFIG.Z在线支付.W微信支付商户v3密钥,
		局_微信响应.Resource.AssociatedData,
		局_微信响应.Resource.Nonce, 局_微信响应.Resource.Ciphertext,
	)

	if err != nil {
		局_原文, _ := jsoniter.MarshalToString(&局_微信响应)
		go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, "系统", "系统", global.X系统信息.B版本号当前, "微信回调"+局_原文+"解密失败:"+err.Error(), c.ClientIP())
		return
	}

	//{"mchid":"1613740956","out_trade_no":"202305171129350001","transaction_id":"4200001827202305179902405083","out_refund_no":"202305171129350001","refund_id":"50302406042023051734540094143","refund_status":"SUCCESS","success_time":"2023-05-17T14:15:26+08:00","amount":{"total":1,"refund":1,"payer_total":1,"payer_refund":1},"user_received_account":"支付用户零钱"}
	fmt.Printf("微信支付退款回调:  %v\n %s\n", 局_微信响应, plaintext)

	局_回调, err := fastjson.Parse(plaintext)
	if err != nil {
		go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, "系统", "系统", global.X系统信息.B版本号当前, "微信回调解析失败:"+plaintext, c.ClientIP())
		return
	}

	if string(局_回调.GetStringBytes("refund_status")) == "SUCCESS" {
		//这里是退款成功 的回调
		局_订单详细信息, ok := Ser_RMBPayOrder.Order取订单详细(string(局_回调.GetStringBytes("out_trade_no")))

		if ok && 局_订单详细信息.Status == 4 { //有订单且订单信息为退款中
			Ser_RMBPayOrder.Order更新订单状态(局_订单详细信息.PayOrder, Ser_RMBPayOrder.D订单状态_退款成功)                                                 //修改订单信息为已支付 充值
			Ser_RMBPayOrder.Order更新订单备注(局_订单详细信息.PayOrder, 局_订单详细信息.Note+",已退到:"+string(局_回调.GetStringBytes("user_received_account"))) //修改订单信息为已支付 充值
		}
	} else {
		//这里是退款失败的回调
		局_订单详细信息, ok := Ser_RMBPayOrder.Order取订单详细(string(局_回调.GetStringBytes("out_trade_no")))
		if ok && 局_订单详细信息.Status == 4 { //有订单且订单信息为退款中
			Ser_RMBPayOrder.Order更新订单状态(局_订单详细信息.PayOrder, Ser_RMBPayOrder.D订单状态_退款失败)   //修改订单信息为已支付 充值
			Ser_RMBPayOrder.Order更新订单备注(局_订单详细信息.PayOrder, 局_订单详细信息.Note+局_微信响应.Summary) //修改订单信息为已支付 充值
		}
	}

	c.Status(http.StatusOK)
	return
}
