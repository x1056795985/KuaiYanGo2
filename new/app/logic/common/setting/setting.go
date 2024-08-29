package setting

import (
	"EFunc/utils"
	jsoniter "github.com/json-iterator/go"
	"server/config"
	"server/global"
	"server/new/app/service"
	"time"
)

func Z文本(配置名 string, 配置值 interface{}) error {

	db := service.S_Setting{}
	jsonStr, err := jsoniter.Marshal(配置值)
	if err != nil {
		return err
	}

	tx := *global.GVA_DB
	err = db.Update(&tx, 配置名, string(jsonStr))
	global.H缓存.Delete("config." + 配置名)

	return err

}

// T 为泛型  繁殖值t 必须为已经创建好的值,不能仅声明,否则可能出错空指针
func Q获取配置[T any](配置名 string) (T, error) {
	var 配置值 T
	if temp, ok := global.H缓存.Get("config." + 配置名); ok {
		if 配置值, ok = temp.(T); ok {
			return 配置值, nil
		}
	}

	计时 := utils.S时间_取现行时间戳13()

	tx := *global.GVA_DB
	db := service.S_Setting{}
	jsonStr, err := db.Info(&tx, 配置名)
	if err == nil {
		err = jsoniter.Unmarshal([]byte(jsonStr), &配置值)
	}
	计时 = utils.S时间_取现行时间戳13() - 计时

	if 计时 > 100 { //大于10毫秒 就缓存, 否则不用   本地数据库测试 2毫秒  这么快基本不用缓存
		global.H缓存.Set("config."+配置名, 配置值, time.Duration(计时)*time.Second) //最少缓存10秒
	}

	return 配置值, err
}

func Z系统设置(X系统设置 *config.X系统设置) error {
	return Z文本("系统设置", X系统设置)
}

func Q系统设置() config.X系统设置 {
	var 配置名 = "系统设置"
	//这里可以配置默认值,读取失败比如没有值会返回默认值
	var 配置值 = config.X系统设置{
		X系统名称:     "AI矩阵后台",
		X系统开关:     true,
		X系统关闭提示:   "系统已经关闭使用",
		D代理中心开关:   true,
		D代理中心关闭提示: "系统已经关闭使用",
		Y用户中心开关:   true,
		B备案号:      "粤ICP备88888888号-1",
	}
	局_临时配置值, err := Q获取配置[config.X系统设置](配置名)
	if err == nil {
		配置值 = 局_临时配置值
	}
	return 配置值

}

func Z行为验证码平台配置(配置值 *config.X行为验证码平台配置) error {
	return Z文本("行为验证码平台配置", 配置值)
}

func Q行为验证码平台配置() config.X行为验证码平台配置 {
	var 配置名 = "行为验证码平台配置"
	//这里可以配置默认值,读取失败比如没有值会返回默认值
	var 配置值 = config.X行为验证码平台配置{
		D当前选择: 1,
	}

	局_临时配置值, err := Q获取配置[config.X行为验证码平台配置](配置名)
	if err == nil {
		配置值 = 局_临时配置值
	}
	return 配置值
}

func Z在线支付配置(配置值 *config.Z在线支付) error {
	return Z文本("在线支付配置", 配置值)
}

func Q在线支付配置() config.Z在线支付 {
	var 配置名 = "在线支付配置"
	var 配置值 = config.Z在线支付{}
	配置值.Z支付宝单次最大金额 = 2000
	配置值.Z支付宝当面付单次最大金额 = 2000
	配置值.Z支付宝H5单次最大金额 = 2000
	配置值.Z支付宝商户ID = "20210088888888"
	配置值.Z支付宝同步回调url = "https://www.baidu.com/s?wd=%E8%AE%A2%E5%8D%95{OrderId}%E6%94%AF%E4%BB%98%E6%88%90%E5%8A%9F"
	配置值.W微信支付单次最大金额 = 500
	配置值.X小叮当单次最大金额 = 500
	配置值.X小叮当支付类型 = 43

	局_临时配置值, err := Q获取配置[config.Z在线支付](配置名)
	if err == nil {
		配置值 = 局_临时配置值
	}
	return 配置值
}

func Z短信平台配置(配置值 *config.D短信平台配置) error {
	return Z文本("短信平台配置", 配置值)
}

func Q短信平台配置() config.D短信平台配置 {
	var 配置名 = "短信平台配置"
	var 配置值 = config.D短信平台配置{
		D当前选择: 1,
	}

	局_临时配置值, err := Q获取配置[config.D短信平台配置](配置名)
	if err == nil {
		配置值 = 局_临时配置值
	}
	return 配置值
}
func Z例子写出记录(配置值 *config.Test) error {
	return Z文本("例子写出记录", 配置值)
}

func Q例子写出记录() config.Test {
	var 配置名 = "例子写出记录"
	var 配置值 = config.Test{}

	局_临时配置值, err := Q获取配置[config.Test](配置名)
	if err == nil {
		配置值 = 局_临时配置值
	}
	return 配置值
}

func Z置MQTT配置(配置值 *config.MQTT配置) error {
	return Z文本("MQTT配置", 配置值)
}

func Q取MQTT配置() config.MQTT配置 {
	var 配置名 = "MQTT配置"
	var 配置值 = config.MQTT配置{}
	配置值.L连接状态 = false
	配置值.F服务器地址 = "broker.emqx.io"
	配置值.F服务器端口 = 1883
	配置值.Y用户名 = ""
	配置值.M密码 = ""

	局_临时配置值, err := Q获取配置[config.MQTT配置](配置名)
	if err == nil {
		配置值 = 局_临时配置值
	}
	return 配置值
}
