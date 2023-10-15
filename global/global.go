// Package global 全局变量  割赖哦抱儿
package global

import (
	"github.com/robfig/cron/v3"
	"github.com/songzhibin97/gkit/cache/local_cache"
	"github.com/spf13/viper"
	"github.com/valyala/fastjson"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"server/KuaiYanSDK"
	"server/config"
)

var (
	//  全局 配置处理
	GVA_Viper *viper.Viper
	//  全局配置 结构存放地址  由GVa_VIper 读取数据 反序列化而成
	GVA_CONFIG config.Server
	//  全局 日志处理
	GVA_LOG *zap.Logger
	//数据库操作工具 gorm
	GVA_DB *gorm.DB

	GVA_Gin *http.Server

	//缓存 用来缓存验证码key
	H缓存 local_cache.Cache

	Cron定时任务 定时任务

	Q快验 KuaiYanSDK.Api快验_类

	X系统信息 = K快验帐号信息{
		B版本号当前: "1.0.117",
	}
)

type K快验帐号信息 struct {
	B绑定信息           string
	Y用户类型           string
	Y用户类型代号       int
	D到期时间           int64
	Z注册时间           int
	D登录时间           int
	D登录IP             string
	Y余额               float64
	J积分               float64
	H会员帐号           string
	H会员密码           string
	Y用户备注           string
	Json_vip            fastjson.Value
	K开启验证码接口列表 string
	L连接方式           string
	B版本号当前         string
	B版本号最新         string
	G公告_文字          string
	G公告_图片          []byte
	Y应用名称           []byte
	Y邮箱               string
	S手机号             string
	Qq                  string
}
type 定时任务 struct {
	Cron        *cron.Cron
	Map任务列表 map[string]cron.EntryID
}

func (c *定时任务) T添加任务(任务名称, 表达式 string, cmd func()) error {
	if c.Map任务列表 == nil {
		c.Map任务列表 = map[string]cron.EntryID{}
	}
	EntryID, err := c.Cron.AddFunc(表达式, cmd)
	if err == nil {
		c.Map任务列表[任务名称] = EntryID
	} else {
		GVA_LOG.Error("定时任务添加失败:" + 任务名称 + err.Error())
		return err
	}
	return nil
}
