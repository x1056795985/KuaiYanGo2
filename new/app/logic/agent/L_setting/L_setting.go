package L_setting

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_UserConfig"
	"server/config"
	"server/new/app/models/constant"
)

func Q取代理在线支付信息(c *gin.Context) (data config.Z在线支付, err error) {
	var 配置值 = config.Z在线支付{}
	配置值.Z支付宝单次最大金额 = 2000
	配置值.Z支付宝当面付单次最大金额 = 2000
	配置值.Z支付宝H5单次最大金额 = 2000
	配置值.Z支付宝商户ID = "20210088888888"
	配置值.Z支付宝同步回调url = "https://www.baidu.com/s?wd=%E8%AE%A2%E5%8D%95{OrderId}%E6%94%AF%E4%BB%98%E6%88%90%E5%8A%9F"
	配置值.W微信支付单次最大金额 = 500
	配置值.X小叮当单次最大金额 = 500
	配置值.X小叮当支付类型 = 43
	配置值.H虎皮椒单次最大金额 = 500
	data = 配置值
	局_uid := c.GetInt("Uid")
	if 局_uid == 0 {
		err = errors.New("uid错误")
		return
	}
	func取值并解析 := func(key string, data any) (err error) {
		局_临时文本 := Ser_UserConfig.Q取值(constant.APPID_代理平台, 局_uid, key)
		if 局_临时文本 != "" {
			err = json.Unmarshal([]byte(局_临时文本), &data)
		}
		return
	}

	err = func取值并解析("支付宝PC", &data.Z在线支付_支付宝pc)
	err = func取值并解析("支付宝H5", &data.Z在线支付_支付宝H5)
	err = func取值并解析("支付宝当面付", &data.Z在线支付_支付宝当面付)
	err = func取值并解析("支付宝H5", &data.Z在线支付_支付宝H5)
	err = func取值并解析("微信支付", &data.Z在线支付_微信支付)
	err = func取值并解析("小叮当", &data.Z在线支付_小叮当)
	err = func取值并解析("虎皮椒", &data.Z在线支付_虎皮椒)
	return
}

func Z置代理在线支付信息(c *gin.Context, 在线支付 config.Z在线支付) (err error) {

	局_uid := c.GetInt("Uid")
	if 局_uid == 0 {
		err = errors.New("uid错误")
		return
	}
	func序列化并置值 := func(key string, v any) error {
		marshal, err2 := json.Marshal(&v)
		if err2 != nil {
			marshal = []byte("{}")
		}
		err2 = Ser_UserConfig.Z置值(constant.APPID_代理平台, 局_uid, key, string(marshal))
		return err2
	}

	err = func序列化并置值("支付宝PC", &在线支付.Z在线支付_支付宝pc)
	err = func序列化并置值("支付宝H5", &在线支付.Z在线支付_支付宝H5)
	err = func序列化并置值("支付宝当面付", &在线支付.Z在线支付_支付宝当面付)
	err = func序列化并置值("支付宝H5", &在线支付.Z在线支付_支付宝H5)
	err = func序列化并置值("微信支付", &在线支付.Z在线支付_微信支付)
	err = func序列化并置值("小叮当", &在线支付.Z在线支付_小叮当)
	err = func序列化并置值("虎皮椒", &在线支付.Z在线支付_虎皮椒)
	return
}
