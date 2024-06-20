// Package KuaiYanSDK 飞鸟快验go语言对接Sdk
package KuaiYanSDK

import (
	"EFunc/utils"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/valyala/fastjson"
	"runtime"
	"strings"
)

const 强制Rsa加密接口 = `	"GetToken":            1,
	"UserLogin":           1,
	"UserReduceMoney":     1,
	"UserReduceVipNumber": 1,
	"UserReduceVipTime":   1,
	"GetVipData":          1,`

type Api快验_类 struct {
	集_AppWeb, J_Token, 集_错误信息, 集_验证码ID, 集_验证码值 string
	J_CryptoKeyAes                             []byte //通讯Aes密匙
	集_CryptoType                               int    //1 明文 2 MD5签名AES加密  3 Rsa签名交换AES密匙
	集_错误代码                                     int
	集_公钥指针                                     *rsa.PublicKey
	集_验证码类型                                    int
	集_Api网关ApiAppKey                           string
	集_Api网关ApiAppSecret                        []byte
}

func (k *Api快验_类) SetAppWeb(域名 string) bool {
	k.集_AppWeb = 域名 + string(utils.B编码_BASE64解码("L0FwaT9BcHBJZD0xMDAwMQ=="))
	return true
}

// 配置json 可以直接在应用设置里复制
func (k *Api快验_类) C初始化配置(配置json string) bool {
	局_fastjson, jsonErr := fastjson.Parse(配置json)
	if jsonErr != nil {
		k.集_错误信息 = "配置json解析失败"
		return false
	}
	//fmt.Printf(string(局_fastjson.GetInt("CryptoType")))
	k.集_CryptoType = 局_fastjson.GetInt("CryptoType")
	k.集_AppWeb = string(局_fastjson.GetStringBytes("AppWeb"))
	path := utils.C程序_取运行目录()
	if runtime.GOOS == "windows" {
		path = "."
	}
	path = path + "/config.json" //设置文件目录   //注意设置 ./config.json  宝塔写文件不会写运行目录 文件会在 /www/server/panel 文件夹

	if strings.Index(utils.W文件_读入文本(path), "\"系统模式\": 1056795985") > 0 {
		k.SetAppWeb("http://127.0.0.1:18888")
		fmt.Printf("超级管理员模式\n")
	}
	k.J_CryptoKeyAes = 局_fastjson.GetStringBytes("CryptoKeyAes")
	k.集_Api网关ApiAppSecret = []byte("204641349")
	k.集_Api网关ApiAppKey = "gc0Aay7WvmO8X5tzIwTupEVsQ9TXlmJz"

	switch k.集_CryptoType {
	case 3:
		k.J_CryptoKeyAes = []byte(utils.W文本_取随机字符串(24))
		//2、pem decode,得到block的der编码数据
		block, _ := pem.Decode(局_fastjson.GetStringBytes("CryptoKeyPublic"))
		if block == nil {
			k.集_错误信息 = "Rsa密匙格式错误"
			return false
		}
		//3、解码der得到公钥
		//publicKey, err := x509.ParsePKCS1PublicKey(derText)   // 这个只能载入PKCS1的  但是目前公钥基本都是PKCS8的
		publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes) //这个只能载入PKCS8的公钥 //大坑 必须校验是否   block是否为nil 否则block.Bytes卡主,类似进入许可区,
		if err != nil {
			k.集_错误信息 = "Rsa密匙错误"
			return false
			//密钥错误
		}
		//类型断言
		ok := false
		k.集_公钥指针, ok = publicKeyInterface.(*rsa.PublicKey) //强制转换
		if !ok {
			k.集_错误信息 = "Rsa密匙错误"
		}
		return ok
	case 2:
		k.集_错误信息 = "Aes密匙错误"
		return len(k.J_CryptoKeyAes) == 24
	}

	return k.集_CryptoType > 0

}

func (k *Api快验_类) Q取错误信息(存储错误代号的变量 *int) string {

	if 存储错误代号的变量 != nil {
		*存储错误代号的变量 = k.集_错误代码
	}
	return k.集_错误信息
}

// Z置验证码信息
// .子程序 置验证码信息, 文本型, 公开, 需要验证码的接口,前一句代码置入验证码值,接口就会携带提交
// .参数 验证码类型, 整数型, 可空, 空 不覆盖缓存 1 英数验证码,2 行为验证码,3 短信验证码
// .参数 验证码ID, 文本型, 可空, 空 不覆盖缓存  可分两次先置id,验证码类型  后置值
// .参数 验证码值, 文本型, 可空, 空 不覆盖缓存/*
func (k *Api快验_类) Z置验证码信息(验证码类型 int, 验证码ID, 验证码值 string) {
	if 验证码类型 > 0 {
		k.集_验证码类型 = 验证码类型
	}
	if 验证码ID != "" {
		k.集_验证码ID = 验证码ID
	}
	if 验证码值 != "" {
		k.集_验证码值 = 验证码值
	}
}

/*
.子程序 取Token, 逻辑型, 公开, 初始化后 首先调用,获取token标志
*/
func (k *Api快验_类) Q取Token() bool {
	请求json := make(map[string]interface{}, 5)
	请求json["Api"] = "GetToken"
	if len(k.J_CryptoKeyAes) != 24 { //可能有意外导致的删除Aes,比如读取上次重启前的token和AesKey 所以重新获取的时候,需要重置
		k.J_CryptoKeyAes = []byte(utils.W文本_取随机字符串(24))
	}
	if k.集_CryptoType == 3 {
		请求json["Key"] = string(k.J_CryptoKeyAes)
	}
	k.J_Token = ""
	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}

	// {"Data":{"Token":"ALYVZWFRDF7ED72VLZMU2Q8CEHFFUKJP","CryptoKeyAes":"APfsSNcyziMBa36CTRcEZGbk","IP":"127.0.0.1"},"Time":1688007301,"Status":84661,"Note":""}
	k.J_Token = string(响应json.GetStringBytes("Data", "Token"))
	if k.集_CryptoType == 3 {
		k.J_CryptoKeyAes = 响应json.GetStringBytes("Data", "CryptoKeyAes")
	}

	if len(k.J_Token) == 0 {
		k.集_错误信息 = "获取到Token错误"
		return false
	}
	return true

}

func (k *Api快验_类) Q取用户IP() string {
	请求json := make(map[string]interface{}, 5)
	请求json["Api"] = "GetUserIP"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return ""
	}

	// {"Data":{"IP":"192.168.1.1"},"Time":1683005833,"Status":41177,"Note":""}
	return string(响应json.GetStringBytes("Data", "IP"))

}

/*
.子程序 D登录_通用, 逻辑型, 公开, 成功返回真
.参数 响应信息json, 文本型, 参考, 登录后返回的用户信息  key绑定信息,OutUser顶掉其他同账号在线数量,VipTime,到期时间戳 {"Key":"677F23CB3FA0055B5FD03916D6AB3C9A","OutUser":1,"VipTime":1685941943}
.参数 账号或卡号, 文本型, , 登录账号
.参数 密码, 文本型, 可空, 登陆密码  卡号模式空即可
.参数 绑定信息, 文本型, , 绑定信息,验证绑定,如果服务器绑定为空, 绑定该值
.参数 动态标记, 文本型, , 会显示在在线列表动态标记内,可以显示用户简单信息,可随时修改
.参数 当前版本, 文本型, , 当前软件版本   会判断是否为可用版本
*/
func (k *Api快验_类) D登录_通用(响应信息 *string, 账号或卡号, 密码, 绑定信息, 动态标记, 当前版本 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "UserLogin"
	请求json["UserOrKa"] = 账号或卡号
	请求json["PassWord"] = 密码
	请求json["Key"] = 绑定信息
	请求json["Tab"] = 动态标记
	请求json["AppVer"] = 当前版本

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}

	// {"Data":{"Key":"677F23CB3FA0055B5FD03916D6AB3C9A","OutUser":1,"VipTime":1685941943},"Time":1683379761,"Status":36354,"Note":""}
	*响应信息 = 响应json.GetObject("Data").String()

	return true

}

/*
.子程序 Y用户减少余额, 逻辑型, 公开, 成功返回真   余额所有这个账号登录的应用都可以使用
.参数 响应信息json, 文本型, 参考, {"Money":2382.56}  剩余金额
.参数 减少数值, 双精度小数型, , 负数无效
.参数 减少原因, 文本型, , 会写到日志记录
.参数 一级代理用户ID, 整数型, 可空, 减少成功可以分成一定数量给代理  仅限一级代理用户id  因为二三级代理,余额完全由一级代理设置,独立全局系统金额体系外
.参数 一级代理分成, 双精度小数型, 可空, 分成金额,不能超过,减少数值,
.参数 分成原因, 文本型, 可空, 会写到余额日志记录
*/

func (k *Api快验_类) Y用户减少余额(响应信息 *string, 减少数值 float64, 减少原因 string, 一级代理用户ID int, 一级代理分成 float64, 分成原因 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "UserReduceMoney"
	请求json["Money"] = 减少数值
	请求json["Log"] = 减少原因
	请求json["AgentId"] = 一级代理用户ID
	请求json["AgentMoney"] = 一级代理分成
	请求json["AgentMoneyLog"] = 分成原因

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}

	//  {"Data":{"Money":"9.93"},"Time":1683379761,"Status":36354,"Note":""}
	*响应信息 = 响应json.GetObject("Data").String()

	return true

}

/*
.
.子程序 用户减少积分, 逻辑型, 公开, 成功返回真    积分类似余额但是只有所属应用可以使用,建议和余额1:1兑换,只想本应用使用时操作 解决计时模式时 不想要用余额又没有变量控制按次收费的问题
.参数 响应信息json, 文本型, 参考, {"VipNumber":"9.93"}   剩余积分
.参数 减少数值, 双精度小数型, , 负数无效
.参数 减少原因, 文本型, , 会写到日志记录
*/
func (k *Api快验_类) Y用户减少积分(响应信息 *string, 减少数值 float64, 减少原因 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "UserReduceVipNumber"
	请求json["VipNumber"] = 减少数值
	请求json["Log"] = 减少原因

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}

	// {"Data":{"VipNumber":"9.93"},"Time":1683379761,"Status":36354,"Note":""}
	*响应信息 = 响应json.GetObject("Data").String()

	return true

}

/*
.子程序 用户减少点数, 逻辑型, 公开, 成功返回真   只有计点方式才可以
.参数 响应信息json, 文本型, 参考, {"VipTime":"9"}   剩余点数
.参数 减少数值, 整数型, , 负数无效 只能为整数
.参数 减少原因, 文本型, , 会写到日志记录
*/
func (k *Api快验_类) Y用户减少点数(响应信息 *string, 减少数值 int, 减少原因 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "UserReduceVipTime"
	请求json["VipTime"] = 减少数值
	请求json["Log"] = 减少原因

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}

	// {"Data":{"VipTime":"9"},"Time":1683379761,"Status":36354,"Note":""}
	*响应信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.子程序 Q取服务器连接状态, 逻辑型, 公开, 成功返回真  异常返回假
*/
func (k *Api快验_类) Q取服务器连接状态() bool {
	请求json := make(map[string]interface{}, 5)
	请求json["Api"] = "IsServerLink"

	_, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}

	return true
}

/*
.子程序 取登录状态, 逻辑型, 公开, 登录状态正常返回真  异常返回假比如未登录或心跳过期,注意vip过期不会注销登录状态
*/
func (k *Api快验_类) Q取登录状态() bool {
	请求json := make(map[string]interface{}, 5)
	请求json["Api"] = "IsLogin"

	_, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}

	return true
}

/*
.子程序 取Vip数据, 逻辑型, 公开, 正常返回真  异常返回假
*/
func (k *Api快验_类) Q取Vip数据(响应信息 *string) bool {
	请求json := make(map[string]interface{}, 5)
	请求json["Api"] = "GetVipData"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	//' {"Time":1683379761,"Status":203,"Note":"Vip已到期"}
	//' {"Time":1683379761,"Status":208,"Note":"Vip数据非标准Json"}
	//' {"Data":{"VipData":"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据","VipData2":"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据"},"Time":1683463084,"Status":16986,"Note":""}
	*响应信息 = 响应json.GetObject("Data").String()
	return *响应信息 != ""
}

/*
.子程序 取应用公告, 文本型, 公开, 正常返回公告  异常返回空
.参数 响应信息, 文本型, 参考 可空
*/
func (k *Api快验_类) Q取应用公告(响应信息 *string) string {
	请求json := make(map[string]interface{}, 5)
	请求json["Api"] = "GetAppGongGao"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return ""
	}
	*响应信息 = 响应json.GetObject("Data").String()
	return string(响应json.GetStringBytes("AppGongGao"))
}

/*
.子程序 取应用专属变量, 逻辑型, 公开, 正常返回真  异常返回假
.参数 响应信息, 文本型, 参考 可空, 传入变量 赋值变量文本值 如果是逻辑值 1=真 0=假
.参数 变量名称, 文本型, , 变量名称
*/
func (k *Api快验_类) Q取应用专属变量(响应信息 *string, 变量名称 string) bool {
	请求json := make(map[string]interface{}, 5)
	请求json["Api"] = "GetAppPublicData"
	请求json["Name"] = 变量名称

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	/*.
	' {"Data":{"紧急公告":"我是一条公告"},"Time":1683473651,"Status":16968,"Note":""}
	' {"Time":1683473472,"Status":208,"Note":"变量不存在"}
	' {"Time":1683473536,"Status":209,"Note":"未登录,请先操作登录"}
	*/
	*响应信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.子程序 取公共变量, 逻辑型, 公开, 所有软件不用登录都可以读取的变量
.参数 响应信息, 文本型, 参考 可空, 传入变量 赋值变量文本值 如果是逻辑值 1=真 0=假
.参数 变量名称, 文本型, , 变量名称
*/
func (k *Api快验_类) Q取公共变量(响应信息 *string, 变量名称 string) bool {
	请求json := make(map[string]interface{}, 5)
	请求json["Api"] = "GetPublicData"
	请求json["Name"] = 变量名称

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	*响应信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.
.子程序 取最新版本检测, 逻辑型, 公开, 检查成功返回真,检查失败返回假 不想使用这种格式版本号的,直接在应用专属变量,设置一个最新版本号,自己处理就好
.参数 响应信息, 文本型, 参考 可空, 响应信息
.参数 当前版本号, 文本型, , 大版本号.小版本号.编译版本号   '有可能还没登录就检测,所以还是单独传一下版本号,不使用登录时的版本号了
.参数 检测编译版本号, 逻辑型, 可空, 是否检测编译版本号,建议自动检测值为假,用户主动检测值为真
.参数 是否需要更新, 逻辑型, 参考 可空, 传入变量  值为真 需要更新,值为假不更新
.参数 最新版本号文本, 文本型, 参考 可空, 传入变量 文本型版本号,"大版本号.小版本号.编译版本号"  版本设置第一行   用来显示
*/
func (k *Api快验_类) Q取最新版本检测(响应信息 *string, 当前版本号 string, 检测编译版本号 bool, 是否需要更新 *bool, 最新版本号文本 *string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetAppVersion"
	请求json["Name"] = 当前版本号
	请求json["IsVersionAll"] = 检测编译版本号

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	//' {"Data":{"IsUpdate":true,"NewVersion":"1.2.5","Version":1.2},"Time":1683542365,"Status":28677,"Note":""}
	*响应信息 = 响应json.GetObject("Data").String()
	*是否需要更新 = 响应json.GetBool("Data", "IsUpdate")
	*最新版本号文本 = string(响应json.GetStringBytes("Data", "NewVersion"))
	return true
}

/*
.
.子程序 取新版本下载地址, 逻辑型, 公开, 所有软件不用登录都可以读
.参数 新版本下载地址, 文本型, 参考
*/
func (k *Api快验_类) Q取新版本下载地址(新版本下载地址 *string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetAppUpDataJson"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	*新版本下载地址 = string(响应json.GetStringBytes("Data", "AppUpDataJson"))
	return true
}

/*
.子程序 取应用主页Url, 逻辑型, 公开, 所有软件不用登录都可以读
.参数 主页Url, 文本型, 参考
*/
func (k *Api快验_类) Q取应用主页Url(主页Url *string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetAppHomeUrl"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	//{"Data":{"AppHomeUrl":"www.baidu.com"},"Time":1683473651,"Status":16968,"Note":""}
	*主页Url = string(响应json.GetStringBytes("Data", "AppHomeUrl"))
	return true
}

/*
.子程序 置新绑定信息, 逻辑型, 公开, 换绑成功返回真,失败返回假
.参数 响应信息, 文本型, 参考 可空, 响应信息{"ReduceVipTime":10} 换绑扣除多少点数或时间
.参数 新绑定信息, 文本型
.参数 账号或卡号, 文本型, 可空, 账号或卡号   ,仅用在未登录时想更换绑定
.参数 密码, 文本型, 可空, 密码 如果是卡号 就空即可
*/
func (k *Api快验_类) Z置新绑定信息(响应信息 *string, 新绑定信息, 账号或卡号, 密码 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "SetAppUserKey"
	请求json["NewKey"] = 新绑定信息
	请求json["User"] = 账号或卡号
	请求json["PassWord"] = 密码

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	//' {"Data":{"ReduceVipTime":10},"Time":1683601988,"Status":34623,"Note":""}
	//' {"Time":1683596452,"Status":210,"Note":"未登录,请先操作登录"}
	//' {"Time":1683595169,"Status":205,"Note":"新绑定信息不能为空."}
	*响应信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.子程序 置新用户消息, 逻辑型, 公开, 成功返回真,失败返回假
.参数 消息类型, 整数型, , 1 其他, 2 bug提交 , 3 投诉建议
.参数 消息内容, 文本型
*/
func (k *Api快验_类) Z置新用户消息(消息类型 int, 消息内容 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "SetNewUserMsg"
	请求json["MsgType"] = 消息类型
	请求json["Msg"] = 消息内容

	_, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	return true
}

/*
.子程序 取验证码, 逻辑型, 公开, 成功返回真,失败返回假
.参数 响应信息, 文本型, 参考, 解码后如果无法显示,删除data:image/png;base64, 在解码
.参数 类型, 整数型, , 验证码类型,默认 1 英数验证码,2滑块验证码
*/
func (k *Api快验_类) Q取验证码(响应信息 *string, 类型 int) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetCaptcha"
	请求json["CaptchaType"] = 类型

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}

	*响应信息 = 响应json.GetObject("Data").String()

	return true
}

/*
.子程序 取短信验证码, 逻辑型, 公开, 成功返回真,失败返回假
.参数 响应信息, 文本型, 参考, {"CaptchaId":"4T7fSxvHV75tfgg","CaptchaType":3}
.参数 手机号, 文本型, 可空
.参数 用户名, 文本型, 可空, 如果手机号为空 可以填写用户名,会向用户名的手机号发送, 找回密码时可以使用
*/
func (k *Api快验_类) Q取短信验证码(响应信息 *string, 手机号, 用户名 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetSMSCaptcha"
	请求json["Phone"] = 手机号
	请求json["User"] = 用户名

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}

	*响应信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.子程序 取绑定信息, 逻辑型, 公开, 登录状态正常返回真  异常返回假比如未登录或心跳过期
.参数 响应信息, 文本型, 参考, 绑定信息
*/
func (k *Api快验_类) Q取绑定信息(绑定信息 *string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetAppUserKey"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}

	*绑定信息 = string(响应json.GetStringBytes("Data", "Key"))
	return true
}

/*
.子程序 取用户是否存在, 逻辑型, 公开, 正常返回真  异常返回假
.参数 是否存在, 逻辑型, 参考, 存在值为真,不存在值为假
.参数 用户名, 文本型, , 用户名称 或 卡号
*/
func (k *Api快验_类) Q取用户是否存在(是否存在 *bool, 用户名 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetIsUser"
	请求json["User"] = 用户名

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	/*
	   ' {"Data":{"IsUser":false},"Time":1683826090,"Status":34224,"Note":""}
	   ' {"Data":{"IsUser":ture},"Time":1683826090,"Status":34224,"Note":""}
	*/
	*是否存在 = 响应json.GetBool("Data", "IsUser")
	return true
}

/*
.子程序 取软件用户信息, 逻辑型, 公开, 正常返回真  异常返回假  获取软件用户相关信息
.参数 响应信息, 文本型, 参考
*/
func (k *Api快验_类) Q取软件用户信息(响应信息 *string, AppVer string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetAppUserInfo"
	请求json["AppVer"] = AppVer

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	//' {"Id":1,"Key":"aaaaaa","MaxOnline":1,"LoginIp":"127.0.0.1","RegisterTime":1683349292,"LoginTime": 1683349292,"Status":1,"Uid":21,"User":"aaaaaa","UserClassId":22,"UserClassMark":2,"UserClassName":"Vip2","UserClassWeight":2,"VipNumber":115.78,"VipTime":1715438220}
	*响应信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.子程序 取用户基础信息, 逻辑型, 公开, 正常返回真  异常返回假  获取 邮箱 手机号 QQ   是否已实名等基础信息
.参数 响应信息, 文本型, 参考, {"Email":"1056795985@qq.com","Phone":"15666666666","Qq":"1056795985"}
*/
func (k *Api快验_类) Q取用户基础信息(响应信息 *string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetUserInfo"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	//' {"Email":"1056795985@qq.com","Phone":"15666666666","Qq":"1056795985"}
	*响应信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.子程序 置用户基础信息, 逻辑型, 公开, 正常返回真  异常返回假  设置 邮箱手机号  等基础信息
.参数 QQ, 文本型
.参数 邮箱, 文本型
.参数 手机号, 文本型
*/
func (k *Api快验_类) Z置用户基础信息(QQ, 邮箱, 手机号 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "SetUserQqEmailPhone"
	请求json["Qq"] = QQ
	请求json["Email"] = 邮箱
	请求json["Phone"] = 手机号

	_, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}

	return true
}

/*
.子程序 用户注册, 逻辑型, 公开, 正常返回真  异常返回假
.参数 注册账号, 文本型
.参数 密码, 文本型
.参数 绑定信息, 文本型
.参数 超级密码, 文本型
.参数 QQ, 文本型, 可空
.参数 邮箱, 文本型, 可空
.参数 手机号, 文本型, 可空
*/
func (k *Api快验_类) Y用户注册(注册账号, 密码, 绑定信息, 超级密码, Qq, 邮箱, 手机号 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "NewUserInfo"
	请求json["User"] = 注册账号
	请求json["PassWord"] = 密码
	请求json["Key"] = 绑定信息
	请求json["SuperPassWord"] = 超级密码
	请求json["Qq"] = Qq
	请求json["Email"] = 邮箱
	请求json["Phone"] = 手机号

	_, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	//' {"Time":1684035704,"Status":11707,"Note":"注册成功"}
	//' {"Time":1684034845,"Status":200,"Note":"email邮箱格式不正确"}
	//' {"Time":1684034898,"Status":200,"Note":"超级密码以字母开头，长度在6-18之间，只能包含字符、数字和下划线"}
	//' {"Time":1684035056,"Status":200,"Note":"超级密码不能和密码相同"}
	//' "Time":1684035081,"Status":200,"Note":"用户已存在"}
	return true
}

/*
.子程序 取系统时间戳, 整数型, 公开, 正常返回十位时间戳  异常返回0
.参数 响应信息, 文本型, 参考, 仅供参考 {"Time":1684036534}
*/
func (k *Api快验_类) Q取系统时间戳() int {
	请求json := make(map[string]interface{}, 5)
	请求json["Api"] = "GetSystemTime"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return 0
	}

	// {"Data":{"Time":1684036534},"Time":1684036534,"Status":17609,"Note":""}
	return 响应json.GetInt("Data", "Time")

}

/*
.子程序 取软件用户备注, 逻辑型, 公开
.参数 响应信息, 文本型, 参考 可空, 备注信息
*/
func (k *Api快验_类) Q取软件用户备注(响应信息 *string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetAppUserNote"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}

	*响应信息 = string(响应json.GetStringBytes("Data", "Note"))
	return true
}

/*
.子程序 取vip到期时间戳, 整数型, 公开, 正常返回十位时间戳  异常返回0
*/
func (k *Api快验_类) Q取vip到期时间戳() int {
	var 时间戳 = 0
	k.Q取vip剩余点数(&时间戳)
	return 时间戳
}

func (k *Api快验_类) Q取vip剩余点数(剩余点数 *int) bool {
	请求json := make(map[string]interface{}, 5)
	请求json["Api"] = "GetAppUserVipTime"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	//' {"Data":{"VipTime":1714919820},"Time":1684036726,"Status":15489,"Note":""}
	*剩余点数 = 响应json.GetInt("Data", "VipTime")

	return true
}

/*
.子程序 用户登录注销, 逻辑型, 公开, 成功返回真,失败返回假
*/
func (k *Api快验_类) Y用户登录注销() bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "LogOut"

	_, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	k.J_Token = ""
	return true
}

/*
.子程序 用户登录注销_远程, 逻辑型, 公开, 成功返回真,失败返回假   会注销全部本应用的在线账号
.参数 用户名或卡号, 文本型
.参数 密码, 文本型, , 无密码直接空即可
*/
func (k *Api快验_类) Y用户登录注销_远程(用户名或卡号, 密码 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "RemoteLogOut"
	请求json["User"] = 用户名或卡号
	请求json["PassWord"] = 密码

	_, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	return true
}

/*
.子程序 心跳, 逻辑型, 公开, 成功返回真,失败返回假
.参数 响应信息, 文本型, 参考 可空, 仅供参考
.参数 响应当前状态, 整数型, 参考 可空, 当前状态 正常返回1  会员已到期返回3(免费模式即使到期了也不会返回3)
*/
func (k *Api快验_类) X心跳(响应信息 *string, 响应当前状态 *int) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "HeartBeat"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	// {"Time":1683601988,"Status":34623,"Note":""}
	// {"Time":1683596452,"Status":210,"Note":"未登录,请先操作登录"}
	// {"Data":{"Status":1},"Time":1684038983,"Status":35387,"Note":""}
	*响应信息 = 响应json.GetObject("Data").String()
	*响应当前状态 = 响应json.GetInt("Data", "Status")
	return true
}

/*
.子程序 密码找回或修改_超级密码, 逻辑型, 公开, 成功返回真,失败返回假  修改成功后会注销所有在线的账号
.参数 响应信息, 文本型, 参考 可空, 仅供参考
.参数 用户名, 文本型
.参数 新密码, 文本型
.参数 超级密码, 文本型
*/
func (k *Api快验_类) M密码找回或修改_超级密码(响应信息 *string, 用户名, 新密码, 超级密码 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "SetPassWord"
	请求json["Type"] = 1
	请求json["User"] = 用户名
	请求json["NewPassWord"] = 新密码
	请求json["SuperPassWord"] = 超级密码

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	*响应信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.
.子程序 密码找回或修改_绑定手机, 逻辑型, 公开, 成功返回真,失败返回假  修改成功后会注销所有在线的账号
.参数 响应信息, 文本型, 参考 可空, 仅供参考
.参数 用户名, 文本型
.参数 新密码, 文本型
.参数 短信验证码Id, 文本型, ,  通过取短信验证码方式获取
.参数 短信验证码, 文本型
*/
func (k *Api快验_类) M密码找回或修改_绑定手机(用户名, 新密码, 短信验证码Id, 短信验证码 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "SetPassWord"
	请求json["Type"] = 2
	请求json["User"] = 用户名
	请求json["NewPassWord"] = 新密码
	请求json["PhoneCaptchaId"] = 短信验证码Id
	请求json["PhoneCaptchaValue"] = 短信验证码

	_, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	return true
}

/*
.子程序 取用户余额, 逻辑型, 公开, 正常返回真  异常返回假
.参数 余额, 双精度小数型, 参考
*/
func (k *Api快验_类) Q取用户余额(余额 *float64) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetUserRmb"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	//' {"Data":{"Rmb":2375.49},"Time":1684064350,"Status":28373,"Note":""}
	*余额 = 响应json.GetFloat64("Data", "Rmb")
	return true
}

/*
.子程序 取用户积分, 逻辑型, 公开, 正常返回真  异常返回假
.参数 积分, 双精度小数型, 参考
*/
func (k *Api快验_类) Q取用户积分(积分 *float64) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetAppUserVipNumber"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	//{"Data":{"VipNumber":108.78},"Time":1684064462,"Status":23298,"Note":""}
	*积分 = 响应json.GetFloat64("Data", "VipNumber")
	return true
}

/*
.子程序 取开启验证码接口列表, 逻辑型, 公开, 正常返回真  异常返回假
.参数 需要验证码的接口列表, 文本型, 参考, {"api接口名": 需要验证码类型} 英数验证码=1 行为验证码=2 短信验证码=3  {"UserLogin":1,"UserReduceMoney":3,"UserReduceVipNumber":3,"UserReduceVipTime":3}
*/
func (k *Api快验_类) Q取开启验证码接口列表(响应信息 *string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetCaptchaApiList"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	*响应信息 = string(响应json.GetStringBytes("Data"))
	return true
}

/*
.子程序 卡号充值, 逻辑型, 公开, 成功返回真,失败返回假
.参数 充值账号, 文本型, , 充值用户账号
.参数 充值卡号, 文本型
.参数 推荐人, 文本型, 可空, 推荐人 卡号如果推荐人有赠送,也会推荐的
*/
func (k *Api快验_类) K卡号充值(充值账号, 充值卡号, 推荐人 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "UseKa"
	请求json["User"] = 充值账号
	请求json["Ka"] = 充值卡号
	请求json["InviteUser"] = 推荐人

	_, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	/*	' {"Time":1683601988,"Status":34623,"Note":""}
		' {"Time":1684072661,"Status":200,"Note":"卡号已经使用到最大次数"}
		' {"Time":1684072730,"Status":200,"Note":"不是本应用卡号"}
		' {"Time":1684072730,"Status":200,"Note":"已使用本卡号充值过了,请勿重复充值"}
		' {"Time":1684072730,"Status":200,"Note":"用户已冻结,无法充值"}
		' {"Time":1684072730,"Status":200,"Note":"未注册应用,请先操作登录一次"}
		' {"Time":1684072730,"Status":200,"Note":"用户类型不同无法充值."}
	*/
	return true
}

/*
.子程序 取动态标记, 逻辑型, 公开,
.参数 动态标记, 文本型, 参考
*/
func (k *Api快验_类) Q取动态标记(动态标记 *string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetTab"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	//' {"Data":{"Tab":"test测试中英文"},"Time":1684076743,"Status":32888,"Note":""}
	*动态标记 = string(响应json.GetStringBytes("Data", "Tab"))
	return true
}

/*
.子程序 置动态标记, 逻辑型, 公开,
.参数 动态标记, 文本型,
*/
func (k *Api快验_类) Z置动态标记(动态标记 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "SetTab"
	请求json["Tab"] = 动态标记

	_, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	return true
}
func (k *Api快验_类) D订单_取状态(响应信息 *string, 订单id string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetPayOrderStatus"
	请求json["OrderId"] = 订单id

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	*响应信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.
.子程序 余额充值_支付宝PC支付, 逻辑型, 公开,
.参数 响应信息, 文本型, 参考 可空, {"Status":3}
.参数 充值账号, 文本型, , 充值用户账号
.参数 充值金额, 双精度小数型, , 充值金额
.参数 支付Url, 文本型, 参考
.参数 订单id, 文本型, 参考 可空, 订单Id  不为空时 为查询支付结果 响应信息Status=   1 未支付 2已支付未充值 3充值成功 4退款成功
*/
func (k *Api快验_类) Y余额充值_支付宝PC支付(响应信息 *string, 充值账号 string, 充值金额 float64, 支付Url, 订单id *string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetAliPayPC"
	请求json["User"] = 充值账号
	请求json["Money"] = 充值金额
	请求json["OrderId"] = 订单id

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	*响应信息 = 响应json.GetObject("Data").String()
	*订单id = string(响应json.GetStringBytes("Data", "OrderId"))
	*支付Url = string(响应json.GetStringBytes("Data", "PayURL"))
	return true
}

/*
.版本 2

.子程序 订单_购买余额, 逻辑型, 公开,  获取支付地址,直接支付后充值到用户账号
.参数 响应信息, 文本型, 参考 可空,  根据支付通道的不同,会响应不同平台的信息
.参数 充值账号或卡号, 文本型, , 充值用户账号
.参数 充值金额, 双精度小数型, , 充值金额
.参数 支付通道, 文本型, 参考, "支付宝PC"   "微信PC" "小叮当"  更多请查看系统管理->系统设置->在线支付设置
*/
func (k *Api快验_类) D订单_购买余额(响应信息 *string, 支付通道, 充值账号 string, 充值金额 float64, 订单id *string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "PayUserMoney"
	请求json["User"] = 充值账号
	请求json["Money"] = 充值金额
	请求json["PayType"] = 支付通道

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	*响应信息 = 响应json.GetObject("Data").String()
	*订单id = string(响应json.GetStringBytes("Data", "OrderId"))
	return true
}

/*
.子程序 余额充值_微信支付支付, 逻辑型, 公开,
.参数 响应信息, 文本型, 参考 可空, {"Status":3}
.参数 充值账号, 文本型, , 充值用户账号
.参数 充值金额, 双精度小数型, , 充值金额
.参数 支付二维码, 文本型, 参考, 文本生成二维码然后扫描就可以了  weixin://wxpay/bizpayurl?pr=QDKS4KWzz
.参数 订单id, 文本型, 参考 可空, 订单Id  不为空时 为查询支付结果  Status=   1 未支付 2已支付未充值 3充值成功 4退款成功
*/
func (k *Api快验_类) Y余额充值_微信支付支付(响应信息 *string, 充值账号 string, 充值金额 float64, 支付二维码, 订单id *string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetWXPayPC"
	请求json["User"] = 充值账号
	请求json["Money"] = 充值金额
	请求json["OrderId"] = 订单id

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}

	*响应信息 = 响应json.GetObject("Data").String()
	// {"OrderId":"202305162202080001","WxPayURL":"weixin://wxpay/bizpayurl?pr=QDKS4KWzz"}
	*订单id = string(响应json.GetStringBytes("Data", "OrderId"))
	*支付二维码 = string(响应json.GetStringBytes("Data", "WxPayURL"))
	return true
}

/*
.子程序 余额充值_取支付通道状态, 逻辑型, 公开,  返回支付宝或微信支付,是否开启
.参数 响应信息, 文本型, 参考, {"AliPayPc":true,"WxPayPc":true}
*/
func (k *Api快验_类) Q余额充值_取支付通道状态(响应信息 *string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetPayStatus"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	*响应信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.子程序 取已购买卡号列表, 逻辑型, 公开,  获取最近购买的卡号列表
.参数 响应信息, 文本型, 参考, [{"Id":331,"KaClassId":18,"Money":3,"Name":"1GRAGpGtuotDYhwZCecqR8FHH","Num":0,"NumMax":1,"Status":1},{"Id":332,"KaClassId":18,"Money":3,"Name":"1KBzZF7YXtzHf6pDE9Qv6ecCZ","Num":0,"NumMax":1,"Status":1}]
.参数 最近数量, 整数型, 可空, 获取最近购买的几个 默认5
*/
func (k *Api快验_类) Q取已购买卡号列表(响应信息 *string, 最近数量 int) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetPurchasedKaList"
	if 最近数量 > 0 {
		请求json["Number"] = 最近数量
	} else {
		请求json["Number"] = 5
	}

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}

	// [{"Id":331,"KaClassId":18,"Money":3,"Name":"1GRAGpGtuotDYhwZCecqR8FHH","Num":0,"NumMax":1,"Status":1},{"Id":332,"KaClassId":18,"Money":3,"Name":"1KBzZF7YXtzHf6pDE9Qv6ecCZ","Num":0,"NumMax":1,"Status":1}]
	*响应信息 = 响应json.String()
	return true
}

/*
.子程序 取可购买卡类列表, 逻辑型, 公开,
.参数 响应信息, 文本型, 参考,{"Data":[{"Id":27,"Money":5,"Name":"开发会员月卡"},{"Id":28,"Money":50,"Name":"商业会员月卡"}],"Time":1685791345,"Status":74080,"Note":""}
*/
func (k *Api快验_类) Q取可购买卡类列表(响应信息 *string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetPayKaList"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	//{"Data":[{"Id":27,"Money":5,"Name":"开发会员月卡"},{"Id":28,"Money":50,"Name":"商业会员月卡"}],"Time":1685791345,"Status":74080,"Note":""}
	*响应信息 = 响应json.String()
	return true
}

/*
.子程序 余额购买充值卡, 逻辑型, 公开,
.参数 响应信息, 文本型, 参考, 卡所属应用id,卡类id,卡类名称,卡号 {"AppId":10001,"KaClassId":18,"KaClassName":"天卡","KaName":"1KBzZF7YXtzHf6pDE9Qv6ecCZ"}
.参数 卡类id, 整数型, , 卡类id 可通过取可购买卡类列表 获取
*/
func (k *Api快验_类) Y余额购买充值卡(响应信息 *string, 卡类id int) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "PayMoneyToKa"
	请求json["KaClassId"] = 卡类id

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	*响应信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.子程序 余额购买积分, 逻辑型, 公开,  根据设置积分余额比例 消费余额购买积分
.参数 响应信息, 文本型, 参考, AddVipNumber本次增加了多少积分, {"AddVipNumber":1.35}
.参数 花费余额, 双精度小数型
*/
func (k *Api快验_类) Y余额购买积分(响应信息 *string, 花费余额 float64) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "PayMoneyToVipNumber"
	请求json["Money"] = 花费余额

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	*响应信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.子程序 取用户类型列表, 逻辑型, 公开,
.参数 响应信息, 文本型, 参考, 整数代号,名称,类型权重  [{"Mark":1,"Name":"vip1","Weight":1},{"Mark":2,"Name":"Vip2","Weight":2}]
*/
func (k *Api快验_类) Q取用户类型列表(响应信息 *string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "GetUserClassList"

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	*响应信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.子程序 置用户类型, 逻辑型, 公开,  转换用户类型, 会根据权重切换更改时间或点数
.参数 响应信息, 文本型, 参考, 新类型代号,新类型名称,转换后的类型信息 {"UserClassMark":2,"UserClassName":"Vip2","VipTime":1699911226}
.参数 新用户类型整数代号, 整数型
*/
func (k *Api快验_类) Z置用户类型(响应信息 *string, 新用户类型整数代号 int) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "SetUserClass"
	请求json["Mark"] = 新用户类型整数代号

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	//' {"UserClassMark":2,"UserClassName":"Vip2","VipTime":1699911226}
	*响应信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.子程序 云函数运行, 逻辑型, 公开,
.参数 响应信息, 文本型, 参考
.参数 函数名, 文本型
.参数 JSON格式参数, 文本型
.参数 是否为全局函数, 逻辑型, , 函数归属为全局值为真,应用专属函数值为假
*/
func (k *Api快验_类) Y云函数运行(响应信息 *string, 函数名, JSON格式参数 string, 是否为全局函数 bool) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "RunJS"
	请求json["JsName"] = 函数名
	请求json["IsGlobal"] = 是否为全局函数
	请求json["Parameter"] = JSON格式参数

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	// {"Time":1684579891,"Status":200,"Note":"异常拦截:JS函数传参或响应错误"}
	// {"Data":{"Return":21,"Time":2},"Time":1684506704,"Status":29964,"Note":""}
	// {"Data":{"Return":{"Key":"aaaaaa","Status":1,"Tab":"AMD Ryzen 7 6800H with Radeon Graphics         |178BFBFF00A40F41","Uid":21,"User":"aaaaaa"},"Time":1},"Time":1684506816,"Status":30949,"Note":""}

	*响应信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.子程序 任务池_任务创建, 逻辑型, 公开,
.参数 响应任务Uuid, 文本型, 参考, '用这个id,3秒每次轮询查询任务结果,
.参数 任务类型ID, 整数型
.参数 任务JSON格式参数, 文本型
*/
func (k *Api快验_类) R任务池_任务创建(响应任务Uuid *string, 任务类型ID int, 任务JSON格式参数 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "TaskPoolNewData"
	请求json["TaskTypeId"] = 任务类型ID
	请求json["Parameter"] = 任务JSON格式参数

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	//' {"Data":{"TaskUuid":"1a6547d1-269d-4ca4-b1b8-b86fb6d41287"},"Time":1684760928,"Status":20613,"Note":""}
	*响应任务Uuid = string(响应json.GetStringBytes("Data", "TaskUuid"))
	return true
}

/*
.子程序 任务池_任务查询, 逻辑型, 公开,
.参数 响应任务信息, 文本型, 参考, Status状态 1已创建,2任务处理中,3成功,4任务失败 其他自定义  {"ReturnData":"","Status":1,"TimeEnd":0,"TimeStart":1684762832}
.参数 Uuid, 文本型, , 任务创建 返回的Uuid
*/
func (k *Api快验_类) R任务池_任务查询(响应任务信息 *string, Uuid string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "TaskPoolGetData"
	请求json["TaskUuid"] = Uuid

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	//' {"Data":{"ReturnData":"","Status":1,"TimeEnd":0,"TimeStart":1684762832},"Time":1684762832,"Status":28692,"Note":""}
	*响应任务信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.子程序 任务池_任务处理获取, 逻辑型, 公开, 仅供参考,任务池用户提交的任务,不建议用户端处理,建议服务器另开软件通过WebApi获取单独处理,保证安全性, 轮询即可已优化高性能,线程安全,推荐3秒/次
.参数 响应任务信息, 文本型, 参考
.参数 获取最大数量, 整数型, , 获取最大数量,线程池空闲多少输入多少
.参数 想获取的任务类型Id, 整数型, 数组
*/
func (k *Api快验_类) R任务池_任务处理获取(响应任务信息 *string, 获取最大数量 int, 想获取的任务类型Id []int) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "TaskPoolGetTask"
	请求json["GetTaskNumber"] = 获取最大数量
	请求json["GetTaskTypeId"] = 想获取的任务类型Id

	响应json, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	//' [{"uuid":"63943989-893a-431a-b0fa-2cfb240cb782","Tid":1,"TimeStart":1684766914,"SubmitData":"{\"a\":1}"},{"uuid":"8087b68b-3657-4397-9dea-599a10584b28","Tid":1,"TimeStart":1684764215,"SubmitData":"{\"a\":1}"},{"uuid":"8c6d6954-00b5-40df-bf8c-ec65b995e9ea","Tid":1,"TimeStart":1684767755,"SubmitData":"{\"a\":1}"}]
	*响应任务信息 = 响应json.GetObject("Data").String()
	return true
}

/*
.子程序 任务池_任务处理返回, 逻辑型, 公开
.参数 UUid, 文本型
.参数 任务状态, 整数型, ,  3成功,4任务失败 其他自定义
.参数 返回数据, 文本型
*/
func (k *Api快验_类) R任务池_任务处理返回(UUid string, 任务状态 int, 返回数据 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "TaskPoolSetTask"
	请求json["TaskUuid"] = UUid
	请求json["TaskStatus"] = 任务状态
	请求json["TaskReturnData"] = 返回数据

	_, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	return true
}

func (k *Api快验_类) Z置代理标志(AgentUid int) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "SetAgentUid"
	请求json["AgentUid"] = AgentUid

	_, ok := k.通讯(请求json)
	if !ok { // 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return false
	}
	return true
}
