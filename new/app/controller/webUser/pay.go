package controller

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/valyala/fastjson"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_RMBPayOrder"
	"server/Service/Ser_UserClass"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	"server/new/app/logic/common/kaClassUpPrice"
	"server/new/app/logic/common/rmbPay"
	"server/new/app/models/common"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
)

type Pay struct {
	Common.Common
}

func NewPayController() *Pay {
	return &Pay{}
}

func (C *Pay) GetPayStatus(c *gin.Context) {
	局map := rmbPay.L_rmbPay.Pay_取支付通道状态()
	response.OkWithData(c, 局map)
	return
}
func (C *Pay) GetPayKaList(c *gin.Context) {
	var info = struct {
		ka       DB.DB_Ka
		likeInfo DB.DB_LinksToken
		appInfo  DB.DB_AppInfo
		appUser  DB.DB_AppUser
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	tx := *global.GVA_DB
	var err error
	info.appUser, err = service.NewAppUser(c, &tx, info.appInfo.AppId).InfoUid(info.likeInfo.Uid)
	if err != nil {
		response.FailWithMessage(c, "应用用户不存在")
		return
	}

	var DB_KaClass []dbm.DB_KaClass
	DB_KaClass = Ser_KaClass.KaClass取可购买卡类列表(info.appInfo.AppId)

	var 卡类列表_简化 = make([]gin.H, 0, len(DB_KaClass))
	var 局_用户类型 = DB.DB_UserClass{}
	var ok = true

	for 索引, _ := range DB_KaClass {
		局_用户类型, ok = Ser_UserClass.Id取详情(info.appInfo.AppId, DB_KaClass[索引].UserClassId)

		if !ok {
			局_用户类型.Name = ""
			局_用户类型.Mark = 0
			局_用户类型.Weight = 1
		}
		计算代理调价, _, err2 := kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, DB_KaClass[索引].Id, info.appUser.AgentUid)
		if err2 == nil {
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

	response.OkWithData(c, gin.H{
		"KaList": 卡类列表_简化,
	})
	return
}

func (C *Pay) GetPayOrderStatus(c *gin.Context) {
	var 请求 struct {
		OrderId string `json:"orderId" binding:"required" zh:"订单id"`
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}

	if 请求.OrderId == "" {
		response.FailWithMessage(c, "订单不存在")
		return
	}

	局_订单详细信息, ok := Ser_RMBPayOrder.Order取订单详细(请求.OrderId)
	if !ok {
		// 如果失败了,在判断是不是上传的第三方订单号
		局_订单详细信息, ok = Ser_RMBPayOrder.Order取订单详细_第三方订单(请求.OrderId)
	}

	// 可能存在未登录充值的情况,所以不检测在线了
	if !ok { //|| 局_订单详细信息.Uid != 局_在线信息.Uid
		response.FailWithMessage(c, "订单不存在")
	} else {
		局_响应 := gin.H{"Status": 局_订单详细信息.Status}
		if 局_卡号 := fastjson.GetString([]byte(局_订单详细信息.Extra), "卡号"); 局_卡号 != "" {
			局_响应["KaName"] = 局_卡号
		}
		response.OkWithData(c, 局_响应)
	}
	return
}

func (C *Pay) PayKaUsa(c *gin.Context) {
	var info = struct {
		likeInfo DB.DB_LinksToken
		appInfo  DB.DB_AppInfo
		appUser  DB.DB_AppUser
		KaClass  dbm.DB_KaClass
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	var 请求 struct {
		KaClassId int    `json:"kaClassId" binding:"required" zh:"卡类id"`
		PayType   string `json:"payType" binding:"required" zh:"支付类型"`
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	var err error
	tx := *global.GVA_DB
	info.KaClass, err = service.NewKaClass(c, &tx).Info(请求.KaClassId)

	if err != nil {
		response.FailWithMessage(c, "卡类不存在")
		return
	}

	if info.KaClass.Money <= 0 {
		response.FailWithMessage(c, "该卡类用户价格小于0不可购买")
		return
	}

	if info.appInfo.AppId != info.KaClass.AppId {
		response.FailWithMessage(c, "非本应用卡类")
		return
	}

	info.appUser, err = service.NewAppUser(c, &tx, info.appInfo.AppId).InfoUid(info.likeInfo.Uid)
	if err != nil { //理论不可能,因为webUser新用户登陆后也会写入用户信息表
		response.FailWithMessage(c, "要充值的用户未登录过应用,请先操作登录一次")
		return
	}

	if info.appUser.UserClassId != 0 && info.KaClass.NoUserClass == 2 && info.appUser.UserClassId != info.KaClass.UserClassId {
		response.FailWithMessage(c, "禁止购买，充值卡用户类型与当前用户类型不相同，请重新选择！")
		return
	}

	// ==============下边为支付数据
	var 参数 common.PayParams
	参数.Uid = info.appUser.Uid
	参数.UidType = 1 //默认账号类型
	if info.appInfo.AppType == 3 || info.appInfo.AppType == 4 {
		参数.UidType = 2
	}

	参数.Type = 请求.PayType
	参数.ReceivedUid = info.appUser.AgentUid
	参数.ProcessingType = constant.D订单类型_购卡直冲
	参数.E额外信息 = gjson.New("{}")
	err = 参数.E额外信息.Set("AppId", info.appInfo.AppId)
	err = 参数.E额外信息.Set("KaClassId", info.KaClass.Id)
	err = 参数.E额外信息.Set("KaClassName", info.KaClass.Name)
	err = 参数.E额外信息.Set("AppUserUid", info.appUser.Uid)
	err = 参数.E额外信息.Set("在线信息AgentUid", info.appUser.AgentUid)
	//开始处理调价信息
	总调价, 调价信息列表, err2 := kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, info.KaClass.Id, info.appUser.AgentUid)
	if err2 != nil && 总调价 > 0 {
		response.FailWithMessage(c, err2.Error())
		return
	}
	err = 参数.E额外信息.Set("卡类金额", info.KaClass.Money)
	err = 参数.E额外信息.Set("调价详情", 调价信息列表)
	err = 参数.E额外信息.Set("总调价", 总调价)
	参数.Rmb = Float64加float64(info.KaClass.Money, 总调价, 2)

	var 响应数据 common.Request
	响应数据, err = rmbPay.L_rmbPay.D订单创建(c, 参数)
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	response.OkWithData(c, 响应数据)
	return

}

//func (a *Pay) Y余额购卡直冲(c *gin.Context) {
//	var info = struct {
//		likeInfo DB.DB_LinksToken
//		appInfo  DB.DB_AppInfo
//		appUser  DB.DB_AppUser
//		KaClass  dbm.DB_KaClass
//	}{}
//	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
//
//	var 请求 struct {
//		KaClassId int `json:"KaClassId" binging:"required"` //  必填`
//	}
//	//解析失败
//	if !C.ToJSON(c, &请求) {
//		return
//	}
//
//	if info.appInfo.AppType > 2 {
//		response.FailWithMessage(c, "本接口仅限应用账密模式可用")
//		return
//	}
//
//	if info.likeInfo.Uid == 0 {
//		response.FailWithMessage(c, "本接口仅限登陆后可用")
//		return
//	}
//	var err error
//	db := *global.GVA_DB
//	info.KaClass, err = service.NewKaClass(c, &db).Info(请求.KaClassId)
//
//	if err != nil {
//		response.FailWithMessage(c, "要购买的充值卡类型ID不存在")
//		return
//	}
//	if info.appInfo.AppId != info.KaClass.AppId || info.KaClass.Money <= 0 {
//		response.FailWithMessage(c, "普通用户无法购买本类型充值卡")
//		return
//	}
//
//	//开始处理调价信息
//	var 局_价格组成 struct {
//		总调价  float64
//		调价详情 []dbm.DB_KaClassUpPrice
//		购买数量 int64
//
//		付款金额 float64
//	}
//
//	局_价格组成.总调价, 局_价格组成.调价详情, err = kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, info.KaClass.Id, info.likeInfo.AgentUid)
//	if err != nil {
//		response.FailWithMessage(c, err.Error())
//		return
//	}
//	局_价格组成.付款金额 = Float64加float64(info.KaClass.Money, 局_价格组成.总调价, 2)
//	局_价格组成.购买数量 = 1
//
//	新余额, err := Ser_User.Id余额增减(info.likeInfo.Uid, 局_价格组成.付款金额, false)
//	if err != nil {
//		response.FailWithMessage(c, "购买失败,"+err.Error())
//		return
//	}
//
//	局_卡信息, err2 := Ser_Ka.Ka单卡创建(info.KaClass.Id, info.likeInfo.Uid, info.likeInfo.User, "用户"+info.likeInfo.User+"自助通过Api购卡", "", 0)
//	if err2 != nil {
//		新余额, err = Ser_User.Id余额增减(info.likeInfo.Uid, info.KaClass.Money, true)
//		if err != nil {
//			go Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_系统执行错误, AppInfo.AppId, info.likeInfo.User, AppInfo.AppName, info.likeInfo.AppVer, "用户余额购卡,减余额成功,制卡失败,请手动处理,本次错误原因:"+err.Error(), c.ClientIP())
//			response.FailWithMessage(c, "购卡失败,费用退还失败,请联系开发者手动处理")
//		} else {
//			response.FailWithMessage(c, "购卡失败,请重试")
//		}
//		return
//	}
//	response.X响应状态带数据(c, c.GetInt("局_成功Status"), gin.H{"AppId": 局_卡信息.AppId, "KaClassId": 局_卡信息.KaClassId, "KaClassName": 局_卡类.Name, "KaName": 局_卡信息.Name})
//	//输出日志
//	str := fmt.Sprintf("自助购卡->:%s,->卡ID:%d,卡号:%s|新余额≈%s",
//		AppInfo.AppName,
//		局_卡信息.Id,
//		局_卡信息.Name,
//		Float64到文本(新余额, 2),
//	)
//	go Ser_Log.Log_写余额日志(info.likeInfo.User, c.ClientIP(), str, Float64取负值(局_价格组成.付款金额))
//	局_文本 := fmt.Sprintf("自助购卡应用:%s,卡类:%s,消费:%.2f)", AppInfo.AppName, info.KaClass.Name, 局_价格组成)
//	go Ser_Log.Log_写卡号操作日志(info.likeInfo.User, c.ClientIP(), 局_文本, []string{局_卡信息.Name}, 1, 0)
//	//代理分成 		//开始分利润 20240202 mark处理重构以后改事务
//	//先分成 代理调价信息的价格
//	if 局_价格组成.总调价 > 0 {
//		局_日志前缀 := fmt.Sprintf("用户:%s,余额制卡ID{%d}", info.likeInfo.User, 局_卡信息.Id)
//		err = agent.L_agent.Z执行调价信息分成(c, 局_价格组成.调价详情, 局_价格组成.购买数量, 局_日志前缀)
//		if err != nil {
//			global.GVA_LOG.Error(fmt.Sprintf("Z执行调价信息分成失败:", err.Error()))
//		}
//	}
//	if info.likeInfo.AgentUid > 0 && info.KaClass.AgentMoney > 0 {
//		//然后再计算百分比的价格
//		代理分成数据, err3 := agent.L_agent.D代理分成计算(c, info.likeInfo.AgentUid, info.KaClass.Money)
//		if err3 == nil {
//			局_日志前缀 := fmt.Sprintf("用户%s余额制卡ID:%d,", info.likeInfo.User, 局_卡信息.Id)
//			err = agent.L_agent.Z执行百分比代理分成(c, 代理分成数据, info.KaClass.Money, 局_日志前缀, 局_价格组成.总调价 == 0)
//			if err != nil {
//				global.GVA_LOG.Error(fmt.Sprintf("Z执行百分比代理分成:%s", err.Error()))
//			}
//		}
//	}
//	// 分成结束==============
//	return
//}
