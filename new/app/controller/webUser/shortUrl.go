package controller

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"net/url"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	"server/new/app/logic/common/setting"
	shortUr "server/new/app/logic/webUser/shortUrl"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
	"time"
)

type ShortUrl struct {
	Common.Common
}

func NewShortUrlController() *ShortUrl {
	return &ShortUrl{}
}

// 处理短链302跳转
func (C *ShortUrl) Jump(c *gin.Context) {
	var info struct {
		ShortUrl dbm.DB_ShortUrl
	}
	var err error
	局_shortUrl := c.Param("shortUrl")
	tx := *global.GVA_DB
	info.ShortUrl, err = service.NewShortUrl(c, &tx).InfoShortUrl(局_shortUrl)
	if err != nil {
		return
	}
	//http://localhost:9000/user/10001/

	//跳转到中间页 //http://localhost:9000/user/10001/ ,然后

	var 局_跳转url string

	局_跳转url = info.ShortUrl.BaseUrl + "#/pages/other/jump?type=1&cpsCode=" + strconv.Itoa(info.ShortUrl.Uid) + "&routerUrl=" + url.QueryEscape(info.ShortUrl.RouterUrl)
	//跳转到本地中间页写入本地推荐人数据然后跳转
	c.Redirect(302, 局_跳转url)
	//根据不同类型进行后处理
	shortUr.L_shortUr.D短链访问后处理(c, info.ShortUrl)

}

// 创建短链,短链信息
func (C *ShortUrl) Create(c *gin.Context) {
	var 请求 struct {
		BaseUrl   string `json:"baseUrl"  binding:"required" zh:"基础地址"`
		RouterUrl string `json:"routerUrl"  binding:"required" zh:"路由地址"`
		ShortType int    `json:"shortType"`
		Other     string `json:"other"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	var err error
	var info = struct {
		appInfo   DB.DB_AppInfo
		likeInfo  DB.DB_LinksToken
		ShortUrls []dbm.DB_ShortUrl
		ShortUrl  dbm.DB_ShortUrl
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	tx := *global.GVA_DB
	//先读取这个长链接和uid 是否已经有符合条件的短连接了,如果有直接返回,如果没有在创建,
	info.ShortUrls, err = service.NewShortUrl(c, &tx).Infos(map[string]interface{}{
		"uid":       info.likeInfo.Uid,
		"appId":     info.appInfo.AppId,
		"baseUrl":   请求.BaseUrl,
		"routerUrl": 请求.RouterUrl,
		"shortType": 请求.ShortType,
		"other":     请求.Other,
	})
	if len(info.ShortUrls) > 0 {
		info.ShortUrl = info.ShortUrls[0]
	} else {
		info.ShortUrl.CreatedAt = time.Now().Unix()
		info.ShortUrl.UpdatedAt = info.ShortUrl.CreatedAt
		info.ShortUrl.ShortUrl = W文本_取随机字符串(6)
		info.ShortUrl.BaseUrl = 请求.BaseUrl
		info.ShortUrl.RouterUrl = 请求.RouterUrl
		info.ShortUrl.ClickCount = 0
		info.ShortUrl.Uid = info.likeInfo.Uid
		info.ShortUrl.AppId = info.appInfo.AppId
		info.ShortUrl.Status = 1
		info.ShortUrl.ShortType = 请求.ShortType
		info.ShortUrl.Other = 请求.Other
		_, err = service.NewShortUrl(c, &tx).Create(&info.ShortUrl)
		if err != nil {
			response.FailWithMessage(c, err.Error())
			return
		}
	}
	response.OkWithData(c, gin.H{
		"shortUrl": setting.Q系统设置().X系统地址 + "/c/" + info.ShortUrl.ShortUrl,
	})
	return
}

// 获取短链,短链信息
func (C *ShortUrl) Info(c *gin.Context) {
	var 请求 struct {
		ShortUrl string `json:"ShortUrl"  zh:"长链接地址"` // 用户名
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	var info = struct {
		appInfo  DB.DB_AppInfo
		likeInfo DB.DB_LinksToken
		ShortUrl dbm.DB_ShortUrl
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	var err error
	tx := *global.GVA_DB

	info.ShortUrl, err = service.NewShortUrl(c, &tx).InfoShortUrl(请求.ShortUrl)

	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}

	response.OkWithData(c, gin.H{
		"shortUrl": info.ShortUrl.ShortUrl,
	})
	return
}
