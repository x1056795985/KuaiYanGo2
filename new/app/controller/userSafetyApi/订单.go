package userSafetyApi

import (
	. "EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/encoding/gjson"
	"server/new/app/controller/userSafetyApi/response"
	"server/new/app/global"
	"server/new/app/logic/common/agent"
	"server/new/app/logic/common/ka"
	"server/new/app/logic/common/kaClassUpPrice"
	"server/new/app/logic/common/log"
	"server/new/app/logic/common/rmbPay"
	"server/new/app/logic/common/user"
	"server/new/app/models/common"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	"strings"
)

func UserApi_订单_取状态(c *gin.Context) {
	局_ctx := 取上下文(c)
	// {"Api":"GetPayOrderStatus","OrderId":"","Time":1684152719,"Status":15959}
	局_订单Id := 局_ctx.Q请求明文.Get("OrderId").String()
	if 局_订单Id == "" {
		response.FailMsg(c, constant.Status_操作失败, "订单不存在")
		return
	}

	db := *global.GVA_DB
	// 可能存在未登录充值的情况,所以不检测在线了
	局_订单详细信息, err := service.NewRmbPayService(&db).Info2(map[string]interface{}{"PayOrder": 局_订单Id})
	if err != nil {
		// 如果失败了,在判断是不是上传的第三方订单号
		局_订单详细信息, err = service.NewRmbPayService(&db).Info2(map[string]interface{}{"PayOrder2": 局_订单Id})
	}

	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "订单不存在")
	} else {
		局_响应 := gin.H{"Status": 局_订单详细信息.Status}
		局_额外信息, _ := gjson.LoadJson(局_订单详细信息.Extra)
		if 局_卡号 := 局_额外信息.Get("卡号").String(); 局_卡号 != "" {
			局_响应["KaName"] = 局_卡号
		}
		response.OkData(c, 局_响应)
	}
	return
}

func UserApi_订单_购卡直冲(c *gin.Context) {
	局_ctx := 取上下文(c)

	if 局_ctx.AppInfo.AppId < 10000 {
		response.FailMsg(c, constant.Status_操作失败, "应用不存在")
		return
	}

	//{"Api":"GetAliPayPC","User":"aaaaaa","KaClassId":1,"PayType":"小叮当","Time":1684152719,"Status":15959}

	局_用户名 := strings.TrimSpace(局_ctx.Q请求明文.Get("User").String())
	局_卡号 := 局_ctx.AppInfo.AppType == 3 || 局_ctx.AppInfo.AppType == 4
	var 局_Uid = 0
	var 局_Uid类型 = 0

	db := *global.GVA_DB

	if 局_卡号 {
		局_Uid类型 = 2
		局_卡信息, err := service.NewKa(c, &db).Info2(map[string]interface{}{"Name": 局_用户名, "AppId": 局_ctx.AppInfo.AppId})
		if err == nil {
			局_Uid = 局_卡信息.Id
		}
	} else {
		局_Uid类型 = 1
		局_userInfo, err := service.NewUser(c, &db).InfoName(局_用户名)
		if err == nil {
			局_Uid = 局_userInfo.Id
		}
	}

	if 局_Uid == 0 {
		response.FailMsg(c, constant.Status_操作失败, "要充值的用户不存在")
		return
	}
	局_AppUser, err := service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoUid(局_Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "要充值的用户未登录过应用,请先操作登录一次")
		return
	}

	局_卡类信息, err := service.NewKaClass(c, &db).Info(局_ctx.Q请求明文.Get("KaClassId").Int())
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "卡类不存在")
		return
	}
	if 局_ctx.AppInfo.AppId != 局_卡类信息.AppId {
		response.FailMsg(c, constant.Status_操作失败, "非本应用卡类")
		return
	}

	if 局_卡类信息.Money <= 0 {
		response.FailMsg(c, constant.Status_操作失败, "该卡类用户价格小于0不可购买")
		return
	}

	if 局_AppUser.UserClassId != 0 && 局_卡类信息.NoUserClass == 2 && 局_AppUser.UserClassId != 局_卡类信息.UserClassId {
		response.FailMsg(c, constant.Status_操作失败, "禁止购买，充值卡用户类型与当前用户类型不相同，请重新选择！")
		return
	}

	局_支付方式 := strings.TrimSpace(局_ctx.Q请求明文.Get("PayType").String())
	// ==============下边为支付数据
	var 参数 common.PayParams
	参数.Uid = 局_AppUser.Uid
	参数.UidType = 局_Uid类型
	参数.Type = 局_支付方式
	参数.ReceivedUid = 局_AppUser.AgentUid
	参数.ProcessingType = constant.D订单类型_购卡直冲
	参数.E额外信息 = gjson.New("{}")
	_ = 参数.E额外信息.Set("AppId", 局_ctx.Z在线信息.LoginAppid)
	_ = 参数.E额外信息.Set("KaClassId", 局_卡类信息.Id)
	_ = 参数.E额外信息.Set("KaClassName", 局_卡类信息.Name)
	_ = 参数.E额外信息.Set("AppUserUid", 局_AppUser.Uid)
	_ = 参数.E额外信息.Set("在线信息AgentUid", 局_ctx.Z在线信息.AgentUid)
	//开始处理调价信息
	总调价, 调价信息列表, err2 := kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, 局_卡类信息.Id, 局_ctx.Z在线信息.AgentUid)
	if err2 != nil && 总调价 > 0 {
		response.FailMsg(c, constant.Status_操作失败, err2.Error())
		return
	}
	_ = 参数.E额外信息.Set("卡类金额", 局_卡类信息.Money)
	_ = 参数.E额外信息.Set("调价详情", 调价信息列表)
	_ = 参数.E额外信息.Set("总调价", 总调价)
	参数.Rmb = Float64加float64(局_卡类信息.Money, 总调价, 2)

	响应数据, err := rmbPay.L_rmbPay.D订单创建(c, 参数)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error())
	} else {
		response.OkData(c, 响应数据)
	}
	return
}

func UserApi_订单_支付购卡(c *gin.Context) {
	局_ctx := 取上下文(c)

	if 局_ctx.AppInfo.AppId < 10000 {
		response.FailMsg(c, constant.Status_操作失败, "应用不存在")
		return
	}

	//{"Api":"PayGetKa",,"KaClassId":1,"PayType":"小叮当","Time":1684152719,"Status":15959}

	var 局_Uid类型 = 0
	if 局_ctx.AppInfo.AppType == 3 || 局_ctx.AppInfo.AppType == 4 {
		局_Uid类型 = 2
	} else {
		局_Uid类型 = 1
	}

	db := *global.GVA_DB
	局_卡类信息, err := service.NewKaClass(c, &db).Info(局_ctx.Q请求明文.Get("KaClassId").Int())
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "卡类不存在")
		return
	}
	if 局_ctx.AppInfo.AppId != 局_卡类信息.AppId {
		response.FailMsg(c, constant.Status_操作失败, "非本应用卡类")
		return
	}

	if 局_卡类信息.Money <= 0 {
		response.FailMsg(c, constant.Status_操作失败, "该卡类用户价格小于0不可购买")
		return
	}

	局_支付方式 := strings.TrimSpace(局_ctx.Q请求明文.Get("PayType").String())
	// ==============下边为支付数据
	var 参数 common.PayParams
	参数.Uid = 0
	参数.UidType = 局_Uid类型
	参数.Type = 局_支付方式
	参数.ReceivedUid = 局_ctx.Z在线信息.AgentUid
	参数.Rmb = 局_卡类信息.Money
	参数.ProcessingType = constant.D订单类型_支付购卡
	参数.E额外信息 = gjson.New("{}")
	_ = 参数.E额外信息.Set("AppId", 局_ctx.AppInfo.AppId)
	_ = 参数.E额外信息.Set("KaClassId", 局_卡类信息.Id)
	_ = 参数.E额外信息.Set("KaClassName", 局_卡类信息.Name)
	_ = 参数.E额外信息.Set("在线信息AgentUid", 局_ctx.Z在线信息.AgentUid)
	//开始处理调价信息
	总调价, 调价信息列表, err2 := kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, 局_卡类信息.Id, 局_ctx.Z在线信息.AgentUid)
	if err2 != nil {
		response.FailMsg(c, constant.Status_操作失败, err2.Error())
		return
	}
	_ = 参数.E额外信息.Set("卡类金额", 局_卡类信息.Money)
	_ = 参数.E额外信息.Set("调价详情", 调价信息列表)
	_ = 参数.E额外信息.Set("总调价", 总调价)
	参数.Rmb = Float64加float64(局_卡类信息.Money, 总调价, 2)

	响应数据, err := rmbPay.L_rmbPay.D订单创建(c, 参数)

	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error())
	} else {
		response.OkData(c, 响应数据)
	}
	return
}
func UserApi_取支付通道状态(c *gin.Context) {
	局map := rmbPay.L_rmbPay.Pay_取支付通道状态()
	response.OkData(c, 局map)
	return
}
func UserApi_取可购买卡类列表(c *gin.Context) {
	局_ctx := 取上下文(c)

	db := *global.GVA_DB
	//KaClass取可购买卡类列表: Where AppId=? and Money>0
	DB_KaClass, _ := service.NewKaClass(c, &db).Infos(map[string]interface{}{"AppId": 局_ctx.AppInfo.AppId})

	var 卡类列表_简化 = make([]gin.H, 0, len(DB_KaClass))
	var 局_用户类型 = dbm.DB_UserClass{}

	for 索引, _ := range DB_KaClass {
		if DB_KaClass[索引].Money <= 0 {
			continue //只显示价格大于0的
		}
		if DB_KaClass[索引].UserClassId == 0 {
			//未分类
			局_用户类型.Name = "未分类"
			局_用户类型.Mark = 0
			局_用户类型.Weight = 1
		} else {
			var err2 error
			局_用户类型, err2 = service.NewUserClass(c, &db).Info(DB_KaClass[索引].UserClassId)
			if err2 != nil {
				局_用户类型.Name = ""
				局_用户类型.Mark = 0
				局_用户类型.Weight = 1
			}
		}
		计算代理调价, _, err := kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, DB_KaClass[索引].Id, 局_ctx.Z在线信息.AgentUid)
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

	response.OkData(c, 卡类列表_简化)
	return
}
func UserApi_取已购买充值卡列表(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}

	局_数量 := 10
	if 局_ctx.Q请求明文.Get("Number").Int() > 0 {
		局_数量 = 局_ctx.Q请求明文.Get("Number").Int()
	}
	db := *global.GVA_DB
	//KaClass取map列表Int: 按AppId取Id和Name的map
	局_卡类列表, _ := service.NewKaClass(c, &db).GetListAll(局_ctx.AppInfo.AppId)
	卡类名称map := make(map[int]string, len(局_卡类列表))
	for _, v := range 局_卡类列表 {
		卡类名称map[v.Id] = v.Name
	}

	var DB_Ka []dbm.DB_Ka
	//Ka取已购卡列表: Where RegisterUser=? Order Id DESC Limit 数量 Offset 0
	db.Model(dbm.DB_Ka{}).Order("Id DESC").Where("RegisterUser=?", 局_ctx.Z在线信息.User).Limit(局_数量).Find(&DB_Ka)

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

	response.OkData(c, 卡列表_简化)
	return
}
func UserApi_余额购买充值卡(c *gin.Context) {
	局_ctx := 取上下文(c)

	if !检测_账密模式专用(c) {
		return
	}

	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}

	//{"Api":"PayMoneyToKa","Money":1,"Time":1684550291,"Status":37674}
	db := *global.GVA_DB
	局_卡类, err := service.NewKaClass(c, &db).Info(局_ctx.Q请求明文.Get("KaClassId").Int())
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "要购买的充值卡类型ID不存在")
		return
	}
	if 局_ctx.AppInfo.AppId != 局_卡类.AppId || 局_卡类.Money <= 0 {
		response.FailMsg(c, constant.Status_操作失败, "普通用户无法购买本类型充值卡")
		return
	}

	//开始处理调价信息
	var 局_价格组成 struct {
		总调价  float64
		调价详情 []dbm.DB_KaClassUpPrice
		购买数量 int64

		付款金额 float64
	}

	局_价格组成.总调价, 局_价格组成.调价详情, err = kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, 局_卡类.Id, 局_ctx.Z在线信息.AgentUid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error())
		return
	}
	局_价格组成.付款金额 = Float64加float64(局_卡类.Money, 局_价格组成.总调价, 2)
	局_价格组成.购买数量 = 1
	新余额, err := user.L_user.Id余额增减(c, 局_ctx.Z在线信息.Uid, 局_价格组成.付款金额, false)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "购买失败,"+err.Error())
		return
	}

	局_卡信息, err2 := ka.L_ka.Ka单卡创建(c, 局_卡类.Id, 局_ctx.Z在线信息.Uid, 局_ctx.Z在线信息.User, "用户"+局_ctx.Z在线信息.User+"自助通过Api购卡", "", 0)
	if err2 != nil {
		新余额, err = user.L_user.Id余额增减(c, 局_ctx.Z在线信息.Uid, 局_卡类.Money, true)
		if err != nil {
			//用户余额购卡,减余额成功,制卡失败,请手动处理,本次错误原因
			局_日志 := dbm.DB_LogUserMsg{
				User:    "系统",
				App:     局_ctx.AppInfo.AppName,
				AppId:   局_ctx.AppInfo.AppId,
				AppVer:  局_ctx.Z在线信息.AppVer,
				MsgType: 4, //Ser_Log.Log用户消息类型_系统执行错误
				Note:    "用户余额购卡,减余额成功,制卡失败,请手动处理,本次错误原因:" + err.Error(),
				Ip:      c.ClientIP(),
			}
			_ = log.L_log.S输出日志(c, 局_日志)
			response.FailMsg(c, constant.Status_操作失败, "购卡失败,费用退还失败,请联系开发者手动处理")
		} else {
			response.FailMsg(c, constant.Status_操作失败, "购卡失败,请重试")
		}
		return
	}
	response.OkData(c, gin.H{"AppId": 局_卡信息.AppId, "KaClassId": 局_卡信息.KaClassId, "KaClassName": 局_卡类.Name, "KaName": 局_卡信息.Name})
	//输出日志
	str := fmt.Sprintf("自助购卡->:%s,->卡ID:%d,卡号:%s|新余额≈%s",
		局_ctx.AppInfo.AppName,
		局_卡信息.Id,
		局_卡信息.Name,
		Float64到文本(新余额, 2),
	)
	_ = log.L_log.S输出日志(c, dbm.DB_LogMoney{
		User:  局_ctx.Z在线信息.User,
		Ip:    c.ClientIP(),
		Count: Float64取负值(局_价格组成.付款金额),
		Note:  str,
	})
	局_文本 := fmt.Sprintf("自助购卡应用:%s,卡类:%s,消费:%.2f)", 局_ctx.AppInfo.AppName, 局_卡类.Name, 局_价格组成.付款金额)
	_ = log.L_log.S输出日志(c, dbm.DB_LogKa{
		User:     局_ctx.Z在线信息.User,
		Ip:       c.ClientIP(),
		KaType:   constant.Log_卡操作_增,
		Note:     局_文本,
		Ka:       局_卡信息.Name,
		UserType: constant.Log_卡操作用户_普通用户,
	})
	//代理分成 		//开始分利润 20240202 mark处理重构以后改事务
	//先分成 代理调价信息的价格
	if 局_价格组成.总调价 > 0 {
		局_日志前缀 := fmt.Sprintf("用户:%s,余额制卡ID{%d}", 局_ctx.Z在线信息.User, 局_卡信息.Id)
		err = agent.L_agent.Z执行调价信息分成(c, 局_价格组成.调价详情, 局_价格组成.购买数量, 局_日志前缀)
		if err != nil {
			global.GVA_LOG.Println(fmt.Sprintf("Z执行调价信息分成失败:%s", err.Error()))
		}
	}
	if 局_ctx.Z在线信息.AgentUid > 0 && 局_卡类.AgentMoney > 0 {
		//然后再计算百分比的价格
		代理分成数据, err3 := agent.L_agent.D代理分成计算(c, 局_ctx.Z在线信息.AgentUid, 局_卡类.Money)
		if err3 == nil {
			局_日志前缀 := fmt.Sprintf("用户%s余额制卡ID:%d,", 局_ctx.Z在线信息.User, 局_卡信息.Id)
			err = agent.L_agent.Z执行百分比代理分成(c, 代理分成数据, 局_卡类.Money, 局_日志前缀, 局_价格组成.总调价 == 0)
			if err != nil {
				global.GVA_LOG.Println(fmt.Sprintf("Z执行百分比代理分成:%s", err.Error()))
			}
		}
	}
	// 分成结束==============
	return
}
