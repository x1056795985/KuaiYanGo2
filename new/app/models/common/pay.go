package common

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	DB "server/structs/db"
)

type Z在线支付 struct {
	J禁止退款        bool `mapstructure:"禁止退款" json:"禁止退款" `
	Z在线支付_支付宝PC       //id 1
	Z在线支付_支付宝当面付      //id 2
	Z在线支付_支付宝H5       //id 3
	Z在线支付_微信支付        //id 4
	Z在线支付_小叮当         //id 5
	Z在线支付_虎皮椒         //id 6
	Z在线支付_易支付         //id7
	Z在线支付_易支付2        //id8
}
type PayParams struct {
	DB.DB_LogRMBPayOrder
	S商品名称   string
	Y异步回调地址 string
	T同步回调地址 string
	E额外信息   *gjson.Json
	Z支付配置   []byte
	Z支付配置s  Z在线支付
}

// 通用通道返回数据
type Request struct {
	OrderId string `json:"OrderId"` //订单id
	Status  int    `json:"Status"`  //状态码
	//下边三个,视接口而定,有的接口可能没有,但是也就这三种类型
	PayURL       string                 `json:"PayURL"`       //付款地址,比如支付宝pc
	PayQRCode    string                 `json:"PayQRCode"`    //订单代码,比如微信的文本,
	PayQRCodePNG string                 `json:"PayQRCodePNG"` //代码的二维码图片 base64格式
	Other        map[string]interface{} `json:"Other"`        //其他参数
}

type Z在线支付_支付宝PC struct {
	Z支付宝开关      bool   `mapstructure:"支付宝开关" json:"支付宝开关" `
	Z支付宝显示名称    string `mapstructure:"支付宝显示名称" json:"支付宝显示名称"`
	Z支付宝商户ID    string `mapstructure:"支付宝商户ID" json:"支付宝商户ID" `
	Z支付宝商户私钥    string `mapstructure:"支付宝商户私钥" json:"支付宝商户私钥" `
	Z支付宝商户公钥    string `mapstructure:"支付宝商户公钥" json:"支付宝商户公钥" `
	Z支付宝公钥      string `mapstructure:"支付宝公钥" json:"支付宝公钥" `
	Z支付宝同步回调url string `mapstructure:"支付宝同步回调url" json:"支付宝同步回调url" `
	Z支付宝单次最大金额  int    `mapstructure:"支付宝单次最大金额" json:"支付宝单次最大金额" `
}

type Z在线支付_支付宝当面付 struct {
	Z支付宝当面付开关      bool   `mapstructure:"支付宝当面付开关" json:"支付宝当面付开关" `
	Z支付宝当面付显示名称    string `mapstructure:"支付宝当面付显示名称" json:"支付宝当面付显示名称"`
	Z支付宝当面付商户ID    string `mapstructure:"支付宝当面付商户ID" json:"支付宝当面付商户ID" `
	Z支付宝当面付商户私钥    string `mapstructure:"支付宝当面付商户私钥" json:"支付宝当面付商户私钥" `
	Z支付宝当面付商户公钥    string `mapstructure:"支付宝当面付商户公钥" json:"支付宝当面付商户公钥" `
	Z支付宝当面付公钥      string `mapstructure:"支付宝当面付公钥" json:"支付宝当面付公钥" `
	Z支付宝当面付同步回调url string `mapstructure:"支付宝当面付同步回调url" json:"支付宝当面付同步回调url" `
	Z支付宝当面付单次最大金额  int    `mapstructure:"支付宝当面付单次最大金额" json:"支付宝当面付单次最大金额" `
}
type Z在线支付_支付宝H5 struct {
	Z支付宝H5开关      bool   `mapstructure:"支付宝H5开关" json:"支付宝H5开关"`
	Z支付宝H5显示名称    string `mapstructure:"支付宝H5显示名称" json:"支付宝H5显示名称"`
	Z支付宝H5商户ID    string `mapstructure:"支付宝H5商户ID" json:"支付宝H5商户ID" `
	Z支付宝H5商户私钥    string `mapstructure:"支付宝H5商户私钥" json:"支付宝H5商户私钥" `
	Z支付宝H5商户公钥    string `mapstructure:"支付宝H5商户公钥" json:"支付宝H5商户公钥"`
	Z支付宝H5公钥      string `mapstructure:"支付宝H5公钥" json:"支付宝H5公钥" `
	Z支付宝H5同步回调url string `mapstructure:"支付宝H5同步回调url" json:"支付宝H5同步回调url"`
	Z支付宝H5单次最大金额  int    `mapstructure:"支付宝H5单次最大金额" json:"支付宝H5单次最大金额" `
}
type Z在线支付_微信支付 struct {
	W微信支付开关      bool   `mapstructure:"微信支付开关" json:"微信支付开关" `
	W微信支付显示名称    string `mapstructure:"微信支付显示名称" json:"微信支付显示名称"`
	W微信支付商户ID    string `mapstructure:"微信支付商户ID" json:"微信支付商户ID" `
	W微信支付AppId   string `mapstructure:"微信支付AppId" json:"微信支付AppId" `
	W微信支付商户v3密钥  string `mapstructure:"微信支付商户v3密钥" json:"微信支付商户v3密钥" `
	W微信支付商户证书串   string `mapstructure:"微信支付商户证书串" json:"微信支付商户证书串" `
	W微信支付商户证书序列号 string `mapstructure:"微信支付商户证书序列号" json:"微信支付商户证书序列号" `
	W微信支付异步回调Url string `mapstructure:"微信支付异步回调Url" json:"微信支付异步回调Url" `
	W微信支付单次最大金额  int    `mapstructure:"微信支付单次最大金额" json:"微信支付单次最大金额" `
}
type Z在线支付_小叮当 struct {
	X小叮当支付开关   bool   `mapstructure:"小叮当支付开关" json:"小叮当支付开关"`
	X小叮当支付显示名称 string `mapstructure:"小叮当支付显示名称" json:"小叮当支付显示名称"`
	X小叮当app_id string `mapstructure:"小叮当app_id" json:"小叮当app_id"`
	X小叮当接口密钥   string `mapstructure:"小叮当接口密钥" json:"小叮当接口密钥" `
	X小叮当支付类型   int    `mapstructure:"小叮当支付类型" json:"小叮当支付类型" `
	X小叮当单次最大金额 int    `mapstructure:"小叮当单次最大金额" json:"小叮当单次最大金额" `
}

type Z在线支付_虎皮椒 struct {
	H虎皮椒支付开关      bool   `mapstructure:"虎皮椒支付开关" json:"虎皮椒支付开关"`
	H虎皮椒支付显示名称    string `mapstructure:"虎皮椒支付显示名称" json:"虎皮椒支付显示名称"`
	H虎皮椒appId     string `mapstructure:"虎皮椒appId" json:"虎皮椒appId"`
	H虎皮椒appSecret string `mapstructure:"虎皮椒appSecret" json:"虎皮椒appSecret" `
	H虎皮椒支付类型      int    `mapstructure:"虎皮椒支付类型" json:"虎皮椒支付类型" `
	H虎皮椒同步回调url   string `mapstructure:"虎皮椒同步回调url" json:"虎皮椒同步回调url"`
	H虎皮椒单次最大金额    int    `mapstructure:"虎皮椒单次最大金额" json:"虎皮椒单次最大金额" `
	H虎皮椒支付网关      string `mapstructure:"虎皮椒支付网关" json:"虎皮椒支付网关" `
}
type Z在线支付_易支付 struct {
	Y易支付开关      bool   `mapstructure:"易支付开关" json:"易支付开关" `
	Y易支付显示名称    string `mapstructure:"易支付显示名称" json:"易支付显示名称" `
	Y易支付网关      string `mapstructure:"易支付网关" json:"易支付网关" `
	Y易支付商户ID    string `mapstructure:"易支付商户ID" json:"易支付商户ID" `
	Y易支付支付方式    string `mapstructure:"易支付支付方式" json:"易支付支付方式" `
	Y易支付商户密钥KEY string `mapstructure:"易支付商户密钥KEY" json:"易支付商户密钥KEY" `
	Y易支付最大金额    int    `mapstructure:"易支付最大金额" json:"易支付最大金额" `
	Y易支付同步回调url string `mapstructure:"易支付同步回调url" json:"易支付同步回调url" `
	Y易支付设备类型    string `mapstructure:"易支付设备类型" json:"易支付设备类型" `
}

type Z在线支付_易支付2 struct {
	Y易支付2开关      bool   `mapstructure:"易支付2开关" json:"易支付2开关" `
	Y易支付2显示名称    string `mapstructure:"易支付2显示名称" json:"易支付2显示名称" `
	Y易支付2网关      string `mapstructure:"易支付2网关" json:"易支付2网关" `
	Y易支付2商户ID    string `mapstructure:"易支付2商户ID" json:"易支付2商户ID" `
	Y易支付2支付方式    string `mapstructure:"易支付2支付方式" json:"易支付2支付方式" `
	Y易支付2商户密钥KEY string `mapstructure:"易支付2商户密钥KEY" json:"易支付2商户密钥KEY" `
	Y易支付2最大金额    int    `mapstructure:"易支付2最大金额" json:"易支付2最大金额" `
	Y易支付2同步回调url string `mapstructure:"易支付2同步回调url" json:"易支付2同步回调url" `
	Y易支付2设备类型    string `mapstructure:"易支付2设备类型" json:"易支付2设备类型" `
}
