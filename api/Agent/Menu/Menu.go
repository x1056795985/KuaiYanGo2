package Menu

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/encoding/gjson"
	"server/Service/Ser_Agent"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_RMBPayOrder"
	"server/Service/Ser_User"
	"server/global"
	"server/new/app/logic/common/rmbPay"
	"server/new/app/models/common"
	"server/new/app/models/constant"
	"server/structs/Http/response"
	DB "server/structs/db"
	"server/utils"
	"strconv"
)

type Api struct{}

// GetInfo
func (a *Api) GetAgentInfo(c *gin.Context) {
	Uid := c.GetInt("Uid")
	var DB_user DB.DB_User
	err := global.GVA_DB.Model(DB.DB_User{}).Omit("Note", "PassWord", "SuperPassWord").Where("id = ?", Uid).First(&DB_user).Error

	// 没查到数据  或  取反(密码正确)
	if err != nil {
		response.FailWithMessage("查询失败", c)
		global.GVA_LOG.Error("Uid:" + strconv.Itoa(Uid) + "GetUserInfo错误:" + err.Error())
		return
	}

	_, 功能权限 := Ser_Agent.Id取代理可制卡类和可用代理功能列表(c.GetInt("Uid"))
	response.OkWithDetailed(结构响应_GetAdminInfo{
		AgentInfo:     DB_user,
		UserMsgNoRead: 0,
		G功能权限:         功能权限,
	}, "获取成功", c)
	return
}

type 结构响应_GetAdminInfo struct {
	AgentInfo     DB.DB_User `json:"AgentInfo"`
	UserMsgNoRead int64      `json:"UserMsgNoRead"`
	G功能权限         []int      `json:"功能权限"`
}

type 结构请求_单str struct {
	NewPassword string `json:"NewPassword"`
}

// NewPassword
// 修改代理Token密码
func (a *Api) NewPassword(c *gin.Context) {
	var 请求 结构请求_单str
	//{"NewPassword":"aaaaaa"}
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	var msg = ""
	if !utils.Z正则_校验密码(请求.NewPassword, &msg) {
		response.FailWithMessage("密码"+msg, c)
		return
	}

	Uid := c.GetInt("Uid")
	err = Ser_User.Id置新密码(Uid, 请求.NewPassword)
	if err != nil {
		response.FailWithMessage("修改失败", c)
		return
	}
	response.OkWithMessage("修改成功", c)
	return
}

// OutLogin
// 退出登录
func (a *Api) OutLogin(c *gin.Context) {
	err := Ser_LinkUser.Set批量注销Uid(c.GetInt("Uid"), Ser_LinkUser.Z注销_用户操作注销)
	if err != nil {
		response.FailWithMessage("注销失败", c)
		return
	}
	response.OkWithMessage("注销成功", c)
	return
}

type 结构请求_余额充值 struct {
	Type  string  `json:"Type"` //支付方式
	C充值金额 float64 `json:"RMB"`
	D订单ID string  `json:"OrderId"`
}

func (a *Api) Q取支付通道状态(c *gin.Context) {
	局map := rmbPay.L_rmbPay.Pay_取支付通道状态()

	response.OkWithData(局map, c)
	return
}
func (a *Api) Q取支付通道状态2(c *gin.Context) {
	局_arr := rmbPay.L_rmbPay.Pay_取支付通道基本信息()

	response.OkWithData(局_arr, c)
	return
}
func (a *Api) Y余额充值(c *gin.Context) {
	var 请求 结构请求_余额充值
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	if !Ser_Agent.Id功能权限检测(c.GetInt("Uid"), DB.D代理功能_余额充值) {
		response.FailWithMessage("无余额充值权限,请联系上级代理", c)
		return
	}
	//========订单状态查询=======================
	if 请求.D订单ID != "" {
		局_订单信息, ok := Ser_RMBPayOrder.Order取订单详细(请求.D订单ID)
		if !ok {
			response.FailWithMessage("订单不存在", c)
		} else {
			response.OkWithData(gin.H{"Status": 局_订单信息.Status}, c)
		}
		return
	}
	//=======================状态查询借宿
	局_Uid := c.GetInt("Uid")
	局_Uid类型 := 1 //代理一定都是账号
	局_支付方式 := 请求.Type
	//==============下边为支付数据
	var 参数 common.PayParams
	参数.Uid = 局_Uid
	参数.UidType = 局_Uid类型
	参数.Type = 局_支付方式
	参数.Rmb = 请求.C充值金额
	参数.ProcessingType = constant.D订单类型_余额充值
	参数.E额外信息 = gjson.New("{}")
	err = 参数.E额外信息.Set("AppId", constant.APPID_代理平台)

	响应数据, err := rmbPay.L_rmbPay.D订单创建(c, 参数)

	if err != nil {
		response.FailWithMessage("充值方式 ["+请求.Type+"] "+err.Error(), c)
		return
	}
	response.OkWithDetailed(响应数据, "获取成功", c)
	return
}

func (a *Api) Q取余额充值订单状态(c *gin.Context) {
	var 请求 结构请求_余额充值
	err := c.ShouldBindJSON(&请求)
	if err != nil || 请求.D订单ID == "" {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	局_订单详细信息, ok := Ser_RMBPayOrder.Order取订单详细(请求.D订单ID)
	if !ok {
		//如果失败了,在判断是不是上传的第三方订单号
		局_订单详细信息, ok = Ser_RMBPayOrder.Order取订单详细_第三方订单(请求.D订单ID)
	}

	if !ok || 局_订单详细信息.Uid != c.GetInt("Uid") {
		response.FailWithMessage("不可查询其他人支付订单状态", c)
	}
	response.OkWithData(gin.H{"Status": 局_订单详细信息.Status}, c)
	return

}
