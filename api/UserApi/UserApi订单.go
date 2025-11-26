package UserApi

import (
	. "EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/valyala/fastjson"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Ka"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_Log"
	"server/Service/Ser_RMBPayOrder"
	"server/Service/Ser_User"
	"server/Service/Ser_UserClass"
	"server/api/UserApi/response"
	"server/global"
	"server/new/app/logic/common/agent"
	"server/new/app/logic/common/kaClassUpPrice"
	"server/new/app/logic/common/rmbPay"
	"server/new/app/models/common"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	DB "server/structs/db"
	"strings"
)

func UserApi_订单_取状态(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	// {"Api":"GetPayOrderStatus","OrderId":"","Time":1684152719,"Status":15959}
	局_订单Id := string(请求json.GetStringBytes("OrderId"))
	if 局_订单Id == "" {
		response.X响应状态消息(c, response.Status_操作失败, "订单不存在")
		return
	}

	局_订单详细信息, ok := Ser_RMBPayOrder.Order取订单详细(局_订单Id)
	if !ok {
		// 如果失败了,在判断是不是上传的第三方订单号
		局_订单详细信息, ok = Ser_RMBPayOrder.Order取订单详细_第三方订单(局_订单Id)
	}

	// 可能存在未登录充值的情况,所以不检测在线了
	if !ok { //|| 局_订单详细信息.Uid != 局_在线信息.Uid
		response.X响应状态消息(c, response.Status_操作失败, "订单不存在")
	} else {
		局_响应 := gin.H{"Status": 局_订单详细信息.Status}
		if 局_卡号 := fastjson.GetString([]byte(局_订单详细信息.Extra), "卡号"); 局_卡号 != "" {
			局_响应["KaName"] = 局_卡号
		}
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), 局_响应)
	}
	return
}

func UserApi_订单_购卡直冲(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	if AppInfo.AppId < 10000 {
		response.X响应状态消息(c, response.Status_操作失败, "应用不存在")
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"GetAliPayPC","User":"aaaaaa","KaClassId":1,"PayType":"小叮当","Time":1684152719,"Status":15959}

	局_用户名 := strings.TrimSpace(string(请求json.GetStringBytes("User")))
	局_卡号 := Ser_AppInfo.App是否为卡号(AppInfo.AppId)
	var 局_Uid = 0
	var 局_Uid类型 = 0

	if 局_卡号 {
		局_Uid类型 = 2
		局_Uid = Ser_Ka.Ka卡号取id(AppInfo.AppId, 局_用户名)

	} else {
		局_Uid类型 = 1
		局_Uid = Ser_User.User用户名取id(局_用户名)
	}

	if 局_Uid == 0 {
		response.X响应状态消息(c, response.Status_操作失败, "要充值的用户不存在")
		return
	}
	局_AppUser, ok := Ser_AppUser.Uid取详情(AppInfo.AppId, 局_Uid)
	if !ok {
		response.X响应状态消息(c, response.Status_操作失败, "要充值的用户未登录过应用,请先操作登录一次")
		return
	}

	局_卡类信息, err := Ser_KaClass.KaClass取详细信息(请求json.GetInt("KaClassId"))
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "卡类不存在")
		return
	}
	if AppInfo.AppId != 局_卡类信息.AppId {
		response.X响应状态消息(c, response.Status_操作失败, "非本应用卡类")
		return
	}

	if 局_卡类信息.Money <= 0 {
		response.X响应状态消息(c, response.Status_操作失败, "该卡类用户价格小于0不可购买")
		return
	}

	if 局_AppUser.UserClassId != 0 && 局_卡类信息.NoUserClass == 2 && 局_AppUser.UserClassId != 局_卡类信息.UserClassId {
		response.X响应状态消息(c, response.Status_操作失败, "禁止购买，充值卡用户类型与当前用户类型不相同，请重新选择！")
		return
	}

	局_支付方式 := strings.TrimSpace(string(请求json.GetStringBytes("PayType")))
	// ==============下边为支付数据
	var 参数 common.PayParams
	参数.Uid = 局_AppUser.Uid
	参数.UidType = 局_Uid类型
	参数.Type = 局_支付方式
	参数.ReceivedUid = 局_AppUser.AgentUid
	参数.ProcessingType = constant.D订单类型_购卡直冲
	参数.E额外信息 = gjson.New("{}")
	err = 参数.E额外信息.Set("AppId", 局_在线信息.LoginAppid)
	err = 参数.E额外信息.Set("KaClassId", 局_卡类信息.Id)
	err = 参数.E额外信息.Set("KaClassName", 局_卡类信息.Name)
	err = 参数.E额外信息.Set("AppUserUid", 局_AppUser.Uid)
	err = 参数.E额外信息.Set("在线信息AgentUid", 局_在线信息.AgentUid)
	//开始处理调价信息
	总调价, 调价信息列表, err2 := kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, 局_卡类信息.Id, 局_在线信息.AgentUid)
	if err2 != nil && 总调价 > 0 {
		response.X响应状态消息(c, response.Status_操作失败, err2.Error())
		return
	}
	err = 参数.E额外信息.Set("卡类金额", 局_卡类信息.Money)
	err = 参数.E额外信息.Set("调价详情", 调价信息列表)
	err = 参数.E额外信息.Set("总调价", 总调价)
	参数.Rmb = Float64加float64(局_卡类信息.Money, 总调价, 2)

	var 响应数据 common.Request
	响应数据, err = rmbPay.L_rmbPay.D订单创建(c, 参数)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
	} else {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), 响应数据)
	}
	return
}

func UserApi_订单_支付购卡(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	if AppInfo.AppId < 10000 {
		response.X响应状态消息(c, response.Status_操作失败, "应用不存在")
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"PayGetKa",,"KaClassId":1,"PayType":"小叮当","Time":1684152719,"Status":15959}

	var 局_Uid类型 = 0
	if Ser_AppInfo.App是否为卡号(AppInfo.AppId) {
		局_Uid类型 = 2
	} else {
		局_Uid类型 = 1
	}

	局_卡类信息, err := Ser_KaClass.KaClass取详细信息(请求json.GetInt("KaClassId"))
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "卡类不存在")
		return
	}
	if AppInfo.AppId != 局_卡类信息.AppId {
		response.X响应状态消息(c, response.Status_操作失败, "非本应用卡类")
		return
	}

	if 局_卡类信息.Money <= 0 {
		response.X响应状态消息(c, response.Status_操作失败, "该卡类用户价格小于0不可购买")
		return
	}

	局_支付方式 := strings.TrimSpace(string(请求json.GetStringBytes("PayType")))
	// ==============下边为支付数据
	var 参数 common.PayParams
	参数.Uid = 0
	参数.UidType = 局_Uid类型
	参数.Type = 局_支付方式
	参数.ReceivedUid = 局_在线信息.AgentUid
	参数.Rmb = 局_卡类信息.Money
	参数.ProcessingType = constant.D订单类型_支付购卡
	参数.E额外信息 = gjson.New("{}")
	err = 参数.E额外信息.Set("AppId", AppInfo.AppId)
	err = 参数.E额外信息.Set("KaClassId", 局_卡类信息.Id)
	err = 参数.E额外信息.Set("KaClassName", 局_卡类信息.Name)
	err = 参数.E额外信息.Set("在线信息AgentUid", 局_在线信息.AgentUid)
	//开始处理调价信息
	总调价, 调价信息列表, err2 := kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, 局_卡类信息.Id, 局_在线信息.AgentUid)
	if err2 != nil {
		response.X响应状态消息(c, response.Status_操作失败, err2.Error())
		return
	}
	err = 参数.E额外信息.Set("卡类金额", 局_卡类信息.Money)
	err = 参数.E额外信息.Set("调价详情", 调价信息列表)
	err = 参数.E额外信息.Set("总调价", 总调价)
	参数.Rmb = Float64加float64(局_卡类信息.Money, 总调价, 2)

	var 响应数据 common.Request
	响应数据, err = rmbPay.L_rmbPay.D订单创建(c, 参数)

	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
	} else {
		response.X响应状态带数据(c, c.GetInt("局_成功Status"), 响应数据)
	}
	return
}
func UserApi_取支付通道状态(c *gin.Context) {
	局map := rmbPay.L_rmbPay.Pay_取支付通道状态()
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), 局map)
	return
}
func UserApi_取可购买卡类列表(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	var DB_KaClass []dbm.DB_KaClass
	DB_KaClass = Ser_KaClass.KaClass取可购买卡类列表(AppInfo.AppId)

	var 卡类列表_简化 = make([]gin.H, 0, len(DB_KaClass))
	var 局_用户类型 = DB.DB_UserClass{}
	var ok = true

	for 索引, _ := range DB_KaClass {
		局_用户类型, ok = Ser_UserClass.Id取详情(AppInfo.AppId, DB_KaClass[索引].UserClassId)

		if !ok {
			局_用户类型.Name = ""
			局_用户类型.Mark = 0
			局_用户类型.Weight = 1
		}
		计算代理调价, _, err := kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, DB_KaClass[索引].Id, 局_在线信息.AgentUid)
		if err == nil {
			DB_KaClass[索引].Money = Float64加float64(DB_KaClass[索引].Money, 计算代理调价, 2)
		}

		卡类列表_简化 = append(卡类列表_简化, gin.H{
			"Id":              DB_KaClass[索引].Id,
			"Name":            DB_KaClass[索引].Name,
			"Money":           DB_KaClass[索引].Money,
			"NoUserClass":     DB_KaClass[索引].NoUserClass,
			"UserClassId":     DB_KaClass[索引].UserClassId,
			"UserClassName":   局_用户类型.Name,
			"UserClassMark":   局_用户类型.Mark,
			"UserClassWeight": 局_用户类型.Weight,
		})

	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), 卡类列表_简化)
	return
}
func UserApi_取已购买充值卡列表(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	局_数量 := 10
	if 请求json.GetInt("Number") > 0 {
		局_数量 = 请求json.GetInt("Number")
	}
	卡类名称map := Ser_KaClass.KaClass取map列表Int(AppInfo.AppId)
	var DB_Ka []DB.DB_Ka
	DB_Ka, _ = Ser_Ka.Ka取已购卡列表(局_在线信息.User, 1, 局_数量)

	var 卡列表_简化 = make([]gin.H, len(DB_Ka), len(DB_Ka)+1)
	for 索引, _ := range DB_Ka {
		卡列表_简化[索引] = gin.H{
			"Id":           DB_Ka[索引].Id,
			"AppId":        DB_Ka[索引].AppId,
			"Name":         DB_Ka[索引].Name,
			"Money":        DB_Ka[索引].Money,
			"KaClassId":    DB_Ka[索引].KaClassId,
			"KaClassName":  卡类名称map[DB_Ka[索引].KaClassId],
			"Status":       DB_Ka[索引].Status,
			"Num":          DB_Ka[索引].Num,
			"NumMax":       DB_Ka[索引].NumMax,
			"RegisterTime": DB_Ka[索引].RegisterTime,
		}
	}

	response.X响应状态带数据(c, c.GetInt("局_成功Status"), 卡列表_简化)
	return
}
func UserApi_余额购买充值卡(c *gin.Context) {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)

	if !检测_账密模式专用(c, AppInfo) {
		return
	}

	if !检测用户登录在线正常(&局_在线信息) {
		response.X响应状态(c, response.Status_未登录)
		return
	}

	请求json, _ := fastjson.Parse(c.GetString("局_json明文")) //必定是json 不然中间件就报错参数错误了
	//{"Api":"PayMoneyToKa","Money":1,"Time":1684550291,"Status":37674}
	var 局_卡类 dbm.DB_KaClass
	局_卡类.Id = 请求json.GetInt("KaClassId")
	局_卡类, err := Ser_KaClass.KaClass取详细信息(局_卡类.Id)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "要购买的充值卡类型ID不存在")
		return
	}
	if AppInfo.AppId != 局_卡类.AppId || 局_卡类.Money <= 0 {
		response.X响应状态消息(c, response.Status_操作失败, "普通用户无法购买本类型充值卡")
		return
	}

	//开始处理调价信息
	var 局_价格组成 struct {
		总调价  float64
		调价详情 []dbm.DB_KaClassUpPrice
		购买数量 int64

		付款金额 float64
	}

	局_价格组成.总调价, 局_价格组成.调价详情, err = kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, 局_卡类.Id, 局_在线信息.AgentUid)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, err.Error())
		return
	}
	局_价格组成.付款金额 = Float64加float64(局_卡类.Money, 局_价格组成.总调价, 2)
	局_价格组成.购买数量 = 1
	新余额, err := Ser_User.Id余额增减(局_在线信息.Uid, 局_价格组成.付款金额, false)
	if err != nil {
		response.X响应状态消息(c, response.Status_操作失败, "购买失败,"+err.Error())
		return
	}

	局_卡信息, err2 := Ser_Ka.Ka单卡创建(局_卡类.Id, 局_在线信息.Uid, 局_在线信息.User, "用户"+局_在线信息.User+"自助通过Api购卡", "", 0)
	if err2 != nil {
		新余额, err = Ser_User.Id余额增减(局_在线信息.Uid, 局_卡类.Money, true)
		if err != nil {
			go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, AppInfo.AppId, 局_在线信息.User, AppInfo.AppName, 局_在线信息.AppVer, "用户余额购卡,减余额成功,制卡失败,请手动处理,本次错误原因:"+err.Error(), c.ClientIP())
			response.X响应状态消息(c, response.Status_操作失败, "购卡失败,费用退还失败,请联系开发者手动处理")
		} else {
			response.X响应状态消息(c, response.Status_操作失败, "购卡失败,请重试")
		}
		return
	}
	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"AppId": 局_卡信息.AppId, "KaClassId": 局_卡信息.KaClassId, "KaClassName": 局_卡类.Name, "KaName": 局_卡信息.Name})
	//输出日志
	str := fmt.Sprintf("自助购卡->:%s,->卡ID:%d,卡号:%s|新余额≈%s",
		AppInfo.AppName,
		局_卡信息.Id,
		局_卡信息.Name,
		Float64到文本(新余额, 2),
	)
	go Ser_Log.Log_写余额日志(局_在线信息.User, c.ClientIP(), str, Float64取负值(局_价格组成.付款金额))
	局_文本 := fmt.Sprintf("自助购卡应用:%s,卡类:%s,消费:%.2f)", AppInfo.AppName, 局_卡类.Name, 局_价格组成)
	go Ser_Log.Log_写卡号操作日志(局_在线信息.User, c.ClientIP(), 局_文本, []string{局_卡信息.Name}, 1, 0)
	//代理分成 		//开始分利润 20240202 mark处理重构以后改事务
	//先分成 代理调价信息的价格
	if 局_价格组成.总调价 > 0 {
		局_日志前缀 := fmt.Sprintf("用户:%s,余额制卡ID{%d}", 局_在线信息.User, 局_卡信息.Id)
		err = agent.L_agent.Z执行调价信息分成(c, 局_价格组成.调价详情, 局_价格组成.购买数量, 局_日志前缀)
		if err != nil {
			global.GVA_LOG.Error(fmt.Sprintf("Z执行调价信息分成失败:", err.Error()))
		}
	}
	if 局_在线信息.AgentUid > 0 && 局_卡类.AgentMoney > 0 {
		//然后再计算百分比的价格
		代理分成数据, err3 := agent.L_agent.D代理分成计算(c, 局_在线信息.AgentUid, 局_卡类.Money)
		if err3 == nil {
			局_日志前缀 := fmt.Sprintf("用户%s余额制卡ID:%d,", 局_在线信息.User, 局_卡信息.Id)
			err = agent.L_agent.Z执行百分比代理分成(c, 代理分成数据, 局_卡类.Money, 局_日志前缀, 局_价格组成.总调价 == 0)
			if err != nil {
				global.GVA_LOG.Error(fmt.Sprintf("Z执行百分比代理分成:%s", err.Error()))
			}
		}
	}
	// 分成结束==============
	return
}
