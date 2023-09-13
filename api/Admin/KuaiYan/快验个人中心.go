package KuaiYan

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastjson"
	"os"
	"os/exec"
	"path"
	"server/Service/KuaiYanUpdater"
	"server/Service/Ser_Ka"
	"server/Service/Ser_LinkUser"
	"server/Service/Ser_User"
	"server/global"
	"server/structs/Http/response"
	"server/utils"
	"syscall"
	"time"
)

type Api struct{}

func (a *Api) Q取开启验证码接口列表(c *gin.Context) {
	var 跳转过标记 = false
标记:
	var 响应信息 string
	if !global.Q快验.Q取开启验证码接口列表(&响应信息) {
		var 错误代码 = 0
		global.Q快验.Q取错误信息(&错误代码)
		if 错误代码 == 109 && !跳转过标记 {

			global.Q快验.J_Token = "" //后台被强行提出就会失败 需要重新获取token
			跳转过标记 = true
			goto 标记
		}

		response.FailWithMessage("取开启验证码接口列表失败:"+global.Q快验.Q取错误信息(nil), c)
		return
	}
	global.X系统信息.K开启验证码接口列表 = 响应信息
	response.OkWithDetailed(gin.H{"Map": 响应信息}, "获取成功", c)
	return
}
func (a *Api) Q注销(c *gin.Context) {
	var 响应信息 string
	if !global.Q快验.Y用户登录注销() {
		response.FailWithMessage("注销失败:"+global.Q快验.Q取错误信息(nil), c)
		return
	}

	global.X系统信息.K开启验证码接口列表 = 响应信息
	response.OkWithMessage("注销成功", c)
	return
}
func (a *Api) Q取英数验证码(c *gin.Context) {
	var 请求 结构请求_个人信息
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	var 跳转过标记 = false
标记:
	var 响应信息 string
	if !global.Q快验.Q取验证码(&响应信息, 1) {
		var 错误代码 = 0
		global.Q快验.Q取错误信息(&错误代码)
		if 错误代码 == 109 && !跳转过标记 {
			global.Q快验.J_Token = "" //后台被强行提出就会失败 需要重新获取token
			跳转过标记 = true
			goto 标记
		}
		response.FailWithMessage("获取英数验证码失败:"+global.Q快验.Q取错误信息(nil), c)
		return
	}
	局_json, err := fastjson.Parse(响应信息)

	response.OkWithDetailed(gin.H{"CaptchaId": string(局_json.GetStringBytes("CaptchaId")), "PicPath": string(局_json.GetStringBytes("CaptChaImg"))}, "获取成功", c)
	return
}

func (a *Api) Q取短信验证码(c *gin.Context) {
	var 请求 结构请求_个人信息
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var 响应信息 string
	if utils.W文本_是否包含关键字(global.X系统信息.K开启验证码接口列表, `"GetSMSCaptcha"`) {
		if utils.W文本_是否包含关键字(global.X系统信息.K开启验证码接口列表, `"GetSMSCaptcha":2`) {
			global.Q快验.Z置验证码信息(2, 请求.CaptCha2.CaptchaId, 请求.CaptCha2.CaptChaValue)
		} else {
			global.Q快验.Z置验证码信息(1, 请求.CaptCha1.CaptchaId, 请求.CaptCha1.CaptChaValue)
		}
	}

	if !global.Q快验.Q取短信验证码(&响应信息, 请求.User, "") { //user就是手机号
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	局_json, err := fastjson.Parse(响应信息)
	response.OkWithDetailed(gin.H{"CaptchaId": string(局_json.GetStringBytes("CaptchaId"))}, "发送成功", c)
	return
}

func (a *Api) Z快验注册(c *gin.Context) {
	var 请求 结构请求_个人信息
	/*	局_临时字节集, _ := c.GetRawData()
		fmt.Printf(string(局_临时字节集))*/
	err := c.ShouldBindJSON(&请求)

	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	global.Q快验.Z置验证码信息(3, 请求.CaptCha3.CaptchaId, 请求.CaptCha3.CaptChaValue)
	if !global.Q快验.Y用户注册(请求.User, 请求.PassWord, c.ClientIP(), utils.W文本_取随机字符串(10), 请求.Qq, 请求.Qq+"@qq.com", 请求.User) { //user就是手机号
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	response.OkWithMessage("注册成功", c)
	return

}

func (a *Api) D登录(c *gin.Context) {
	var 请求 结构请求_个人信息
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var 响应信息 string
	if utils.W文本_是否包含关键字(global.X系统信息.K开启验证码接口列表, `"UserLogin"`) {
		//可能重启前段还没刷新,所以带滑动验证码参数,如果携带,大概率后台也是是滑动
		if utils.W文本_是否包含关键字(global.X系统信息.K开启验证码接口列表, `"UserLogin":2`) || 请求.CaptCha2.CaptChaValue != "" {
			global.Q快验.Z置验证码信息(2, 请求.CaptCha2.CaptchaId, 请求.CaptCha2.CaptChaValue)
		} else {
			global.Q快验.Z置验证码信息(1, 请求.CaptCha1.CaptchaId, 请求.CaptCha1.CaptChaValue)
		}
	}
	if !global.Q快验.D登录_通用(&响应信息, 请求.User, 请求.PassWord, c.ClientIP(), "", global.X系统信息.B版本号当前) { //user就是手机号
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	//{"Key":"677F23CB3FA0055B5FD03916D6AB3C9A","OutUser":1,"VipTime":1685941943}
	局_json, err := fastjson.Parse(响应信息)
	global.X系统信息.H会员帐号 = 请求.User
	global.X系统信息.H会员密码 = 请求.PassWord
	global.X系统信息.D到期时间 = 局_json.GetInt64("VipTime")
	global.X系统信息.B绑定信息 = string(局_json.GetStringBytes("Key"))
	global.X系统信息.Z注册时间 = 局_json.GetInt("RegisterTime")

	/*	User         string  `json:"User"`
		UserType     string  `json:"UserType"`
		VipNumber    float64 `json:"VipNumber"`
		RMB          float64 `json:"RMB"`
		VipTime      int     `json:"VipTime"`
		Email        string  `json:"Email"`
		LoginTime    int     `json:"loginTime"`
		LoginIp      string  `json:"loginIp"`
		RegisterTime int     `json:"RegisterTime"`*/
	局_是否需要更新 := false
	global.Q快验.Q取最新版本检测(&响应信息, global.X系统信息.B版本号当前, false, &局_是否需要更新, &global.X系统信息.B版本号最新)
	global.Q快验.Q取用户余额(&global.X系统信息.Y余额)
	局_基础信息 := ""
	global.Q快验.Q取用户基础信息(&局_基础信息)
	局_基础信息json, _ := fastjson.Parse(响应信息)
	//{"Email":"1056795985@qq.com","Phone":"15666666666","Qq":"1056795985"}

	global.X系统信息.Y用户类型 = string(局_json.GetStringBytes("UserClassName"))
	global.X系统信息.J积分 = 局_json.GetFloat64("VipNumber")
	global.X系统信息.Y邮箱 = string(局_基础信息json.GetStringBytes("Email"))
	global.X系统信息.D登录时间 = 局_json.GetInt("LoginTime")
	global.X系统信息.D登录IP = string(局_基础信息json.GetStringBytes("LoginIp"))
	global.X系统信息.Qq = string(局_基础信息json.GetStringBytes("Qq"))

	response.OkWithDetailed(gin.H{
		"User":          global.X系统信息.H会员帐号,
		"UserClassName": global.X系统信息.Y用户类型,
		"VipNumber":     global.X系统信息.J积分,
		"RMB":           global.X系统信息.Y余额,
		"VipTime":       global.X系统信息.D到期时间,
		"Email":         global.X系统信息.Y邮箱,
		"LoginTime":     global.X系统信息.D登录时间,
		"LoginIp":       global.X系统信息.D登录IP,
		"Qq":            global.X系统信息.Qq,
		"RegisterTime":  global.X系统信息.Z注册时间,
		"AppVer":        global.X系统信息.B版本号当前,
		"AppVerNew":     global.X系统信息.B版本号最新,
	}, "获取成功", c)
	return
}
func (a *Api) Q快验个人信息更新(c *gin.Context) {
	var 请求 结构请求_个人信息
	//耗时 := time.Now().UnixMilli()
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	var 文件修改日期 = ""
	if len(os.Args) > 0 {
		文件修改日期 = os.Args[0]
		文件信息, err1 := os.Stat(文件修改日期)

		if 文件信息 != nil {
			// linux环境下代码如下
			//*syscall.Stat_t 兼容windos 实际是linux的数据 防止ide报错 windos断言必定返回假
			linuxFileAttr, ok := 文件信息.Sys().(*syscall.Stat_t)
			if ok {
				//fmt.Println("文件创建时间", SecondToTime(linuxFileAttr.Ctim.Sec))
				//fmt.Println("最后访问时间", SecondToTime(linuxFileAttr.Atim.Sec))
				文件修改日期 = SecondToTime(linuxFileAttr.Atim.Sec).String()
				//fmt.Println("最后修改时间", SecondToTime(linuxFileAttr.Mtim.Sec))
			} else {
				文件修改日期 = 文件信息.ModTime().String()
			}

		} else {
			文件修改日期 = err1.Error()
		}
		if len(文件修改日期) > 19 {
			文件修改日期 = 文件修改日期[:19]
		}

	}
	var 局_软件用户信息 string
	if !global.Q快验.Q取软件用户信息(&局_软件用户信息, global.X系统信息.B版本号当前) { //失败直接返回空信息让前段重新登录
		global.Q快验.J_Token = "" //这个估计是被后台强行注销了
		response.OkWithDetailed(gin.H{
			"User":             "",
			"UserClassName":    "",
			"VipNumber":        0,
			"RMB":              0,
			"VipTime":          0,
			"Email":            "",
			"LoginTime":        0,
			"LoginIp":          "",
			"Qq":               "",
			"RegisterTime":     0,
			"AppVer":           global.X系统信息.B版本号当前,
			"AppVerUpdateTime": 文件修改日期,
			"AppVerNew":        global.X系统信息.B版本号最新,
			"linkTokenCount":   0,
		}, "获取成功", c)
		return
	}
	//fmt.Printf("取软件用户信息耗时:%d", time.Now().UnixMilli()-耗时)
	//耗时 = time.Now().UnixMilli()
	局_应用用户信息json, _ := fastjson.Parse(局_软件用户信息)
	global.X系统信息.D到期时间 = 局_应用用户信息json.GetInt64("VipTime")
	global.X系统信息.B绑定信息 = string(局_应用用户信息json.GetStringBytes("Key"))
	global.X系统信息.Z注册时间 = 局_应用用户信息json.GetInt("RegisterTime")
	global.X系统信息.Y用户类型 = string(局_应用用户信息json.GetStringBytes("UserClassName"))
	global.X系统信息.J积分 = 局_应用用户信息json.GetFloat64("VipNumber")
	global.X系统信息.D登录时间 = 局_应用用户信息json.GetInt("LoginTime")

	//{"Key":"677F23CB3FA0055B5FD03916D6AB3C9A","OutUser":1,"VipTime":1685941943}
	局_是否需要更新 := false
	响应信息 := ""
	global.Q快验.Q取最新版本检测(&响应信息, global.X系统信息.B版本号当前, false, &局_是否需要更新, &global.X系统信息.B版本号最新)
	//fmt.Printf("Q取最新版本检测:%d", time.Now().UnixMilli()-耗时)
	//耗时 = time.Now().UnixMilli()
	局_基础信息 := ""
	global.Q快验.Q取用户基础信息(&局_基础信息)
	//fmt.Printf("Q取用户基础信息:%d", time.Now().UnixMilli()-耗时)
	//耗时 = time.Now().UnixMilli()
	局_基础信息json, _ := fastjson.Parse(局_基础信息)
	//{"Email":"1056795985@qq.com","Phone":"15666666666","Qq":"1056795985"}

	global.X系统信息.Y余额 = 局_基础信息json.GetFloat64("RMB")
	global.X系统信息.Y邮箱 = string(局_基础信息json.GetStringBytes("Email"))
	global.X系统信息.D登录IP = string(局_基础信息json.GetStringBytes("LoginIp"))
	global.X系统信息.Qq = string(局_基础信息json.GetStringBytes("Qq"))
	局_在线计数 := 0
	if global.X系统信息.D到期时间 < time.Now().Unix() {
		局_在线计数 = int(Ser_LinkUser.Get取在线总数(true, true))
		global.H缓存.Set("在线数量", 局_在线计数, time.Minute*10) //10分钟有效
	}

	response.OkWithDetailed(gin.H{
		"User":             global.X系统信息.H会员帐号,
		"UserClassName":    global.X系统信息.Y用户类型,
		"VipNumber":        global.X系统信息.J积分,
		"RMB":              global.X系统信息.Y余额,
		"VipTime":          global.X系统信息.D到期时间,
		"Email":            global.X系统信息.Y邮箱,
		"LoginTime":        global.X系统信息.D登录时间,
		"LoginIp":          global.X系统信息.D登录IP,
		"Qq":               global.X系统信息.Qq,
		"RegisterTime":     global.X系统信息.Z注册时间,
		"AppVer":           global.X系统信息.B版本号当前,
		"AppVerUpdateTime": 文件修改日期,
		"AppVerNew":        global.X系统信息.B版本号最新,
		"linkTokenCount":   局_在线计数,
	}, "获取成功", c)
	return
}
func (a *Api) Z快验找回密码(c *gin.Context) {
	var 请求 结构请求_个人信息
	err := c.ShouldBindJSON(&请求)

	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	if !global.Q快验.M密码找回或修改_绑定手机(请求.User, 请求.PassWord, 请求.CaptCha3.CaptchaId, 请求.CaptCha3.CaptChaValue) { //user就是手机号
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	response.OkWithMessage("设置新密码["+请求.PassWord+"]成功", c)
	return
}

// 把秒级的时间戳转为time格式
func SecondToTime(sec int64) time.Time {
	return time.Unix(sec, 0)
}

type 结构请求_个人信息 struct {
	User     string `json:"User"`
	PassWord string `json:"PassWord"`
	Qq       string `json:"Qq"`
	CaptCha1 struct {
		CaptchaId    string `json:"captchaId"`
		CaptChaValue string `json:"CaptChaValue"`
	} `json:"CaptCha1"`
	CaptCha2 struct {
		CaptchaId    string `json:"captchaId"`
		CaptChaValue string `json:"CaptChaValue"`
	} `json:"CaptCha2"`
	CaptCha3 struct {
		CaptchaId    string `json:"captchaId"`
		CaptChaValue string `json:"CaptChaValue"`
	} `json:"CaptCha3"`
}

func (a *Api) Q取可购买充值卡列表(c *gin.Context) {

	var 响应信息 string
	if !global.Q快验.Q取可购买卡类列表(&响应信息) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}

	//{"Data":[{"Id":27,"Money":5,"Name":"开发会员月卡"},{"Id":28,"Money":50,"Name":"商业会员月卡"}],"Time":1685791345,"Status":74080,"msg":""}
	//
	局_json, _ := fastjson.Parse(响应信息)

	局_可买充值卡 := 局_json.GetArray("Data")
	var data = make([]gin.H, len(局_可买充值卡))
	for 索引, _ := range 局_可买充值卡 {
		data[索引] = gin.H{
			"Id":    局_可买充值卡[索引].GetInt("Id"),
			"Name":  string(局_可买充值卡[索引].GetStringBytes("Name")),
			"Money": 局_可买充值卡[索引].GetFloat64("Money"),
		}
	}

	response.OkWithDetailed(data, "获取成功", c)
	return
}

func (a *Api) Y余额购买充值卡(c *gin.Context) {
	var 请求 结构请求_KaCLassId
	err := c.ShouldBindJSON(&请求)

	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	var 响应信息 string
	if !global.Q快验.Y余额购买充值卡(&响应信息, 请求.KaCLassId) { //user就是手机号
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	//{"AppId":10001,"KaClassId":18,"KaClassName":"天卡","KaName":"1KBzZF7YXtzHf6pDE9Qv6ecCZ"}
	局_json, _ := fastjson.Parse(响应信息)

	response.OkWithDetailed(gin.H{"KaName": string(局_json.GetStringBytes("KaName"))}, "购买成功", c)
	return
}

type 结构请求_KaCLassId struct {
	KaCLassId int `json:"KaCLassId"`
}

func (a *Api) Q购买充值卡记录(c *gin.Context) {
	var 响应信息 string
	if !global.Q快验.Q取已购买卡号列表(&响应信息, 5) { //user就是手机号
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	// [{"Id":331,"KaClassId":18,"Money":3,"Name":"1GRAGpGtuotDYhwZCecqR8FHH","Num":0,"NumMax":1,"Status":1},{"Id":332,"KaClassId":18,"Money":3,"Name":"1KBzZF7YXtzHf6pDE9Qv6ecCZ","Num":0,"NumMax":1,"Status":1}]
	局_json, _ := fastjson.Parse(响应信息)

	局_可买充值卡 := 局_json.GetArray("Data")
	var data = make([]gin.H, len(局_可买充值卡))
	for 索引, _ := range 局_可买充值卡 {
		data[索引] = gin.H{
			"Id":           局_可买充值卡[索引].GetInt("Id"),
			"Name":         string(局_可买充值卡[索引].GetStringBytes("Name")),
			"Money":        局_可买充值卡[索引].GetFloat64("Money"),
			"Num":          局_可买充值卡[索引].GetInt("Num"),
			"NumMax":       局_可买充值卡[索引].GetInt("NumMax"),
			"KaClassId":    局_可买充值卡[索引].GetInt("KaClassId"),
			"KaClassName":  string(局_可买充值卡[索引].GetStringBytes("KaClassName")),
			"RegisterTime": 局_可买充值卡[索引].GetInt("RegisterTime"),
		}
	}

	response.OkWithDetailed(data, "获取成功", c)
	return
}

func (a *Api) K卡号充值(c *gin.Context) {
	var 请求 结构请求_充值
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	if !global.Q快验.K卡号充值(global.X系统信息.H会员帐号, 请求.Ka, 请求.InviteUser) { //user就是手机号
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}

	response.OkWithMessage("充值成功", c)
	return
}

type 结构请求_充值 struct {
	Ka         string `json:"Ka"`
	InviteUser string `json:"InviteUser"`
}

func (a *Api) Y余额充值(c *gin.Context) {
	var 请求 结构请求_余额充值
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}
	var 响应信息 string

	var 订单ID = 请求.D订单ID
	if 订单ID != "" {
		if !global.Q快验.D订单_取状态(&响应信息, 订单ID) {
			局_错误代号 := 0
			response.FailWithMessage(global.Q快验.Q取错误信息(&局_错误代号), c)
		} else {
			parse, _ := fastjson.Parse(响应信息)
			response.OkWithData(gin.H{"Status": parse.GetInt("Status")}, c)
		}
		return
	}
	if !global.Q快验.D订单_购买余额(&响应信息, 请求.Type, global.X系统信息.H会员帐号, 请求.C充值金额, &订单ID) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}

	var 局_MAP gin.H
	_ = json.Unmarshal([]byte(响应信息), &局_MAP)

	局_MAP["Status"] = 1
	response.OkWithData(局_MAP, c)
	return
}

type 结构请求_余额充值 struct {
	Type  string  `json:"Type"` //选择支付通道
	C充值金额 float64 `json:"RMB"`
	D订单ID string  `json:"OrderId"`
}

func (a *Api) Q取支付通道状态(c *gin.Context) {
	var 响应信息 string
	if !global.Q快验.Q余额充值_取支付通道状态(&响应信息) { //user就是手机号
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	response.OkWithData(响应信息, c)
	return
}

var 心跳容错计数 = 0

func K快验心跳() {
	//fmt.Printf("定时K快验心跳容错:%v\n", 心跳容错计数)
	if global.Q快验.J_Token == "" {
		return
	}

	var 响应信息 string
	var 当前状态 int
	if !global.Q快验.X心跳(&响应信息, &当前状态) {
		心跳容错计数++
		if 心跳容错计数 >= 3 {
			global.Q快验.J_Token = ""
			global.X系统信息.H会员帐号 = ""
			global.X系统信息.D到期时间 = 0
			global.X系统信息.Y用户类型代号 = 0
		}
		//fmt.Printf("定时K快验心跳失败:%v\n", 心跳容错计数)
		return
	}

	心跳容错计数 = 0

	if 当前状态 == 1 {
		//未过期
	}
	if 当前状态 == 3 {
		//因为是免费模式,这里不会返回3 所以自己通过  判断 到期时间和类型 仅限权限管理
		//已过期

	}
	局_动态标记 := fmt.Sprintf("用户数:%d,卡总数:%d,在线数:%d", Ser_User.Q取总数(), Ser_Ka.Q取总数(), Ser_LinkUser.Get取在线总数(true, true))
	global.Q快验.Z置动态标记(局_动态标记)

	//fmt.Printf("定时K快验心跳状态:%v\n", 当前状态)
}
func (a *Api) G更新程序(c *gin.Context) {

	if KuaiYanUpdater.J_系统更新状态 > 0 {
		response.FailWithMessage("已经在更新中当前进度:"+KuaiYanUpdater.J_系统更新提示, c)
		if KuaiYanUpdater.J_系统更新状态 == 2 { //只提示一次 重置更新状态
			KuaiYanUpdater.J_系统更新提示 = ""
			KuaiYanUpdater.J_系统更新状态 = 0
		}
		return
	}
	var 响应信息 string
	if !global.Q快验.Q取新版本下载地址(&响应信息) {
		response.FailWithMessage("获取新版本下载地址失败:"+global.Q快验.Q取错误信息(nil), c)
		return
	}
	if !utils.X系统_权限检测() {
		response.FailWithMessage("系统无写入文件权限,请检查权限是否777或755,运行用户是否为root", c)
		return
	}

	KuaiYanUpdater.J_系统更新提示 = "准备中"
	go KuaiYanUpdater.K快验系统开始更新(响应信息, aaA)

	var filename string
	filename = path.Base(os.Args[0])
	response.OkWithMessage("开始更新,当前:"+filename, c)
	return
}

func aaA(执行程序本地路径 string) {
	fmt.Printf("系统更新%v,马上启动新版本程序继承端口,退出旧版本\n\n", 执行程序本地路径)
	global.GVA_LOG.Info("系统更新,马上启动新版本程序继承端口,退出旧版本\n")
	err := os.Chmod(执行程序本地路径, 0755)
	if err != nil {
		KuaiYanUpdater.J_系统更新状态 = 2
		KuaiYanUpdater.J_系统更新提示 = fmt.Sprintf("启动修改下载文件权限失败:%v\n", err.Error())
		fmt.Printf(KuaiYanUpdater.J_系统更新提示)
		return
	} //通过chmod重新赋权限

	cmd := exec.Command(执行程序本地路径)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err = cmd.Start(); err != nil {
		KuaiYanUpdater.J_系统更新状态 = 2
		KuaiYanUpdater.J_系统更新提示 = fmt.Sprintf("启动新更新后程序失败:%v\n", err.Error())
		fmt.Printf(KuaiYanUpdater.J_系统更新提示)
		return
	}
	//先关闭端口 解除占用
	err = global.GVA_Gin.Shutdown(context.Background()) //这句话可以停止侦听关闭端口
	if err != nil {
		KuaiYanUpdater.J_系统更新状态 = 2
		KuaiYanUpdater.J_系统更新提示 = "下载文件成功,重启失败,请手动重启系统:" + err.Error()
		return
	}
	//不管如何都会复制失败,放弃统一文件名(方便宝塔重启)  2023-6-14 等待日后解决
	//_ = E.E删除文件("./飞鸟快验最新") //删除原来的
	//_ = E.E文件更名(执行程序本地路径, os.Args[0]) //重命名 为了防止重复下载所以保留原文件,改为直接复制
	/*	err = E.E复制文件(执行程序本地路径, "./飞鸟快验最新") //重命名
		if err != nil {
			fmt.Printf("新文件:%v,已复制为:%v,失败%s\n", 执行程序本地路径, os.Args[0], err.Error())
		} else {
			fmt.Printf("新文件:%v,已复制为:%v\n", 执行程序本地路径, os.Args[0])
		}*/

	// 退出当前进程
	os.Exit(0)

}
