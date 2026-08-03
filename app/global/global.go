// Package global 全局变量  割赖哦抱儿
package global

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/spf13/viper"
	"github.com/valyala/fastjson"
	"gorm.io/gorm"
	"log"
	"net/http"
	"server/app/models/common"
	"server/app/sdk/KuaiYanSDK"
	"time"
)

// H缓存接口 隔离全局运行状态与具体缓存实现。
type H缓存接口 interface {
	Set(k string, v interface{}, d time.Duration)
	Get(k string) (interface{}, bool)
	Delete(key string)
	Increment(k string, n int64) error
	Add(k string, x interface{}, d time.Duration) error
}

var (
	//  全局 配置处理
	GVA_Viper *viper.Viper
	//  全局配置 结构存放地址  由GVa_VIper 读取数据 反序列化而成
	GVA_CONFIG common.Server
	//  全局 日志处理
	GVA_LOG = log.Default()
	//数据库操作工具 gorm
	GVA_DB *gorm.DB

	GVA_Gin *http.Server

	//缓存 用来缓存验证码key
	H缓存 H缓存接口

	Q快验 KuaiYanSDK.Api快验_类

	X系统信息 = K快验帐号信息{
		B版本号当前: "1.0.482",
	}
	// 定义一个全局翻译器T
	Trans ut.Translator
)

type K快验帐号信息 struct {
	B绑定信息      string
	Y用户类型      string
	Y用户类型代号    int
	D到期时间      int64
	Z注册时间      int
	D登录时间      int
	D登录IP      string
	Y余额        float64
	J积分        float64
	H会员帐号      string
	H会员密码      string
	Y用户备注      string
	Json_vip   fastjson.Value
	K开启验证码接口列表 string
	L连接方式      string
	B版本号当前     string
	B版本号最新     string
	G公告_文字     string
	G公告_图片     []byte
	Y应用名称      []byte
	Y邮箱        string
	S手机号       string
	Qq         string
}
