package SetSystem

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"server/api/middleware"
	"server/config"
	"server/global"
	"server/structs/Http/response"
	"server/utils"
	"strings"
)

type Api struct{}

func (a *Api) GetInfoSystem(c *gin.Context) {
	if err := global.GVA_Viper.Unmarshal(&global.GVA_CONFIG); err != nil {
		//感觉有点问题,这个重新读取配置文件速度慢有延迟 还是自己置值吧
		//补充,是我错了,测试的是系统关闭,发现, 逻辑值没有取反,所以开启系统 用户那边还是关闭,我就以为反序列化不好使
		fmt.Println("配置文件反序列化失败1:", err)
	}
	//middleware.G更新哈希APi名称(global.GVA_CONFIG.X系统设置.Y用户API加密盐)
	//方便手动修改配置后重新读取
	response.OkWithDetailed(global.GVA_CONFIG.X系统设置, "获取成功", c)
	return
}

//go:embed  \..\..\..\SDK/易语言/飞鸟快验网络验证对接模块.e

var 快验网络验证对接易模块 []byte

type 请求_S生成API加密源码SDK struct {
	Y用户API加密盐 string `mapstructure:"用户API加密盐" json:"用户API加密盐" yaml:"用户API加密盐"`
	Type      string `mapstructure:"Type" json:"Type" yaml:"Type"` //"E"  易源码
}
type 响应_S生成API加密源码SDK struct {
	Name       string `mapstructure:"Name" json:"Name" yaml:"Name"`
	Base64Data string `mapstructure:"Base64Data" json:"Base64Data" yaml:"Base64Data"` //"E"  易源码
}

func (a *Api) S生成API加密源码SDK(c *gin.Context) {
	var 请求 请求_S生成API加密源码SDK
	response.FailWithMessage("模块更新维护中", c)
	return

	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if 请求.Y用户API加密盐 == "" {
		response.FailWithMessage("API加密盐值不能为空", c)
		return
	}

	// 遍历map获取所有key
	APi列表 := make([]string, 0, len(middleware.J集_UserAPi路由)+1)
	APi列表 = append(APi列表, "GetToken")
	for key := range middleware.J集_UserAPi路由 {
		APi列表 = append(APi列表, key)
	}
	var SDK []byte
	switch 请求.Type {
	case "E":
		SDK = utils.Y易源码替换APi接口并修复(快验网络验证对接易模块, APi列表, 请求.Y用户API加密盐)
	}
	if len(SDK) == 0 {
		response.FailWithMessage("暂不支持:"+请求.Type, c)
	} else {
		response.OkWithDetailed(
			响应_S生成API加密源码SDK{
				Name:       "飞鸟快验APi加密盐值" + 请求.Y用户API加密盐 + ".e",
				Base64Data: base64.StdEncoding.EncodeToString(SDK),
			}, "生成成功,记得保存配置使功能生效", c)
	}
	return
}

// save 保存
func (a *Api) Save信息System(c *gin.Context) {
	var 请求 config.X系统设置
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	global.GVA_Viper.Set("系统设置.系统名称", 请求.X系统名称)
	global.GVA_Viper.Set("系统设置.备案号", 请求.B备案号)
	//global.GVA_CONFIG.X系统设置.X系统名称 = 请求.X系统名称

	global.GVA_Viper.Set("系统设置.系统地址", 请求.X系统地址)
	global.GVA_Viper.Set("系统设置.用户API加密盐", 请求.Y用户API加密盐)

	global.GVA_Viper.Set("系统设置.管理员后台Host", strings.TrimSpace(请求.G管理员后台Host))
	global.GVA_Viper.Set("系统设置.WebApiHost", strings.TrimSpace(请求.WebApiHost))
	global.GVA_Viper.Set("系统设置.代理后台Host", strings.TrimSpace(请求.D代理后台Host))
	//global.GVA_CONFIG.X系统设置.X系统地址 = 请求.X系统地址

	global.GVA_Viper.Set("系统设置.系统开关", 请求.X系统开关)
	//global.GVA_CONFIG.X系统设置.X系统开关 = 请求.X系统开关

	global.GVA_Viper.Set("系统设置.系统关闭提示", 请求.X系统关闭提示)
	//global.GVA_CONFIG.X系统设置.X系统关闭提示 = 请求.X系统关闭提示

	global.GVA_Viper.Set("系统设置.代理中心开关", 请求.D代理中心开关)
	//global.GVA_CONFIG.X系统设置.D代理中心开关 = 请求.D代理中心开关

	global.GVA_Viper.Set("系统设置.代理中心关闭提示", 请求.D代理中心关闭提示)

	//global.GVA_CONFIG.X系统设置.D代理中心关闭提示 = 请求.D代理中心关闭提示

	global.GVA_Viper.Set("系统设置.用户中心开关", 请求.Y用户中心开关)
	//global.GVA_CONFIG.X系统设置.Y用户中心开关 = 请求.Y用户中心开关

	global.GVA_Viper.SetConfigFile(global.GVA_CONFIG.Q取运行目录 + "/config.json")
	global.GVA_Viper.SetConfigType("json")
	err = global.GVA_Viper.WriteConfig()
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}

	if err = global.GVA_Viper.Unmarshal(&global.GVA_CONFIG); err != nil {
		//感觉有点问题,这个重新读取配置文件速度慢有延迟 还是自己置值吧
		//补充,是我错了,测试的是系统关闭,发现, 逻辑值没有取反,所以开启系统 用户那边还是关闭,我就以为反序列化不好使
		fmt.Println("配置文件反序列化失败1:", err)
	}
	middleware.G更新哈希APi名称(global.GVA_CONFIG.X系统设置.Y用户API加密盐)

	/*filePath := "./var.txt"
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("open file error=%v\n", err)
		return
	}
	defer file.Close()
	str := "hello golang\n"
	writer := bufio.NewWriter(file)
	for i := 0; i < 5; i++ {
		writer.WriteString(str)
	}
	writer.Flush()*/

	response.OkWithMessage("保存成功", c)
	return
}

func (a *Api) GetInfo在线支付(c *gin.Context) {
	response.OkWithDetailed(global.GVA_CONFIG.Z在线支付, "获取成功", c)
	return
}

// save 保存
func (a *Api) Save信息在线支付(c *gin.Context) {
	var 请求 config.Z在线支付
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	global.GVA_Viper.Set("在线支付.支付宝开关", 请求.Z支付宝开关)
	global.GVA_Viper.Set("在线支付.支付宝显示名称", 请求.Z支付宝显示名称)
	global.GVA_Viper.Set("在线支付.支付宝商户ID", 请求.Z支付宝商户ID)
	global.GVA_Viper.Set("在线支付.支付宝商户私钥", 请求.Z支付宝商户私钥)
	global.GVA_Viper.Set("在线支付.支付宝商户公钥", 请求.Z支付宝商户公钥)
	global.GVA_Viper.Set("在线支付.支付宝公钥", 请求.Z支付宝公钥)
	global.GVA_Viper.Set("在线支付.支付宝同步回调url", 请求.Z支付宝同步回调url)
	global.GVA_Viper.Set("在线支付.支付宝单次最大金额", 请求.Z支付宝单次最大金额)

	global.GVA_Viper.Set("在线支付.支付宝当面付开关", 请求.Z支付宝当面付开关)
	global.GVA_Viper.Set("在线支付.支付宝当面付显示名称", 请求.Z支付宝当面付显示名称)
	global.GVA_Viper.Set("在线支付.支付宝当面付商户ID", 请求.Z支付宝当面付商户ID)
	global.GVA_Viper.Set("在线支付.支付宝当面付商户私钥", 请求.Z支付宝当面付商户私钥)
	global.GVA_Viper.Set("在线支付.支付宝当面付商户公钥", 请求.Z支付宝当面付商户公钥)
	global.GVA_Viper.Set("在线支付.支付宝当面付公钥", 请求.Z支付宝当面付公钥)
	global.GVA_Viper.Set("在线支付.支付宝当面付同步回调url", 请求.Z支付宝当面付同步回调url)
	global.GVA_Viper.Set("在线支付.支付宝当面付单次最大金额", 请求.Z支付宝当面付单次最大金额)

	global.GVA_Viper.Set("在线支付.微信支付开关", 请求.W微信支付开关)
	global.GVA_Viper.Set("在线支付.微信支付显示名称", 请求.W微信支付显示名称)
	global.GVA_Viper.Set("在线支付.微信支付商户ID", 请求.W微信支付商户ID)
	global.GVA_Viper.Set("在线支付.微信支付AppId", 请求.W微信支付AppId)
	global.GVA_Viper.Set("在线支付.微信支付商户v3密钥", 请求.W微信支付商户v3密钥)
	global.GVA_Viper.Set("在线支付.微信支付商户证书串", 请求.W微信支付商户证书串)
	global.GVA_Viper.Set("在线支付.微信支付商户证书序列号", 请求.W微信支付商户证书序列号)
	global.GVA_Viper.Set("在线支付.微信支付异步回调Url", 请求.W微信支付异步回调Url)
	global.GVA_Viper.Set("在线支付.微信支付单次最大金额", 请求.W微信支付单次最大金额)

	global.GVA_Viper.Set("在线支付.小叮当支付开关", 请求.X小叮当支付开关)
	global.GVA_Viper.Set("在线支付.小叮当支付显示名称", 请求.X小叮当支付显示名称)
	global.GVA_Viper.Set("在线支付.小叮当app_id", 请求.X小叮当app_id)
	global.GVA_Viper.Set("在线支付.小叮当接口密钥", 请求.X小叮当接口密钥)
	global.GVA_Viper.Set("在线支付.小叮当单次最大金额", 请求.X小叮当单次最大金额)
	global.GVA_Viper.Set("在线支付.小叮当支付类型", 请求.X小叮当支付类型)

	global.GVA_Viper.SetConfigFile(global.GVA_CONFIG.Q取运行目录 + "/config.json")
	global.GVA_Viper.SetConfigType("json")
	err = global.GVA_Viper.WriteConfig()
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}

	if err = global.GVA_Viper.Unmarshal(&global.GVA_CONFIG); err != nil {
		fmt.Println("配置文件反序列化失败2:", err)
	}
	response.OkWithMessage("保存成功", c)
	return
}

func (a *Api) GetInfo短信平台设置(c *gin.Context) {
	response.OkWithDetailed(global.GVA_CONFIG.D短信平台配置, "获取成功", c)
	return
}
func (a *Api) GetInfo行为验证码平台设置(c *gin.Context) {
	response.OkWithDetailed(global.GVA_CONFIG.X行为验证码平台配置, "获取成功", c)
	return
}

// save 保存
func (a *Api) Save短信平台设置(c *gin.Context) {
	var 请求 config.D短信平台配置
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	global.GVA_Viper.Set("短信平台配置.TX云Sms.SECRET_ID", 请求.TX云短信Sms.SECRET_ID)
	global.GVA_Viper.Set("短信平台配置.TX云Sms.SECRET_KEY", 请求.TX云短信Sms.SECRET_KEY)
	global.GVA_Viper.Set("短信平台配置.TX云Sms.短信应用ID", 请求.TX云短信Sms.D短信应用ID)
	global.GVA_Viper.Set("短信平台配置.TX云Sms.短信签名", 请求.TX云短信Sms.D短信签名)
	global.GVA_Viper.Set("短信平台配置.TX云Sms.正文模板ID", 请求.TX云短信Sms.Z正文模板ID)

	global.GVA_Viper.SetConfigFile(global.GVA_CONFIG.Q取运行目录 + "/config.json")
	global.GVA_Viper.SetConfigType("json")
	err = global.GVA_Viper.WriteConfig()
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}

	if err = global.GVA_Viper.Unmarshal(&global.GVA_CONFIG); err != nil {
		fmt.Println("配置文件反序列化失败2:", err)
	}
	response.OkWithMessage("保存成功", c)
	return
}

// save 保存
func (a *Api) Save行为验证码平台设置(c *gin.Context) {
	var 请求 config.X行为验证码平台配置
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	global.GVA_Viper.Set("行为验证码平台配置.极验行为验证4.验证_ID", 请求.J极验行为验证4.Y验证_ID)
	global.GVA_Viper.Set("行为验证码平台配置.极验行为验证4.验证_KEY", 请求.J极验行为验证4.Y验证_KEY)

	global.GVA_Viper.SetConfigFile(global.GVA_CONFIG.Q取运行目录 + "/config.json")
	global.GVA_Viper.SetConfigType("json")
	err = global.GVA_Viper.WriteConfig()
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}

	if err = global.GVA_Viper.Unmarshal(&global.GVA_CONFIG); err != nil {
		fmt.Println("配置文件反序列化失败2:", err)
	}
	response.OkWithMessage("保存成功", c)
	return
}
