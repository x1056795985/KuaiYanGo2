package controller

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"net/url"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	"server/new/app/logic/common/setting"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
	"time"
)

type CpsShortUrl struct {
	Common.Common
}

func NewCpsShortUrlController() *CpsShortUrl {
	return &CpsShortUrl{}
}

// 处理短链302跳转
func (C *CpsShortUrl) Jump(c *gin.Context) {
	var info struct {
		CpsShortUrl dbm.DB_CpsShortUrl
	}
	var err error
	局_shortUrl := c.Param("shortUrl")
	tx := *global.GVA_DB
	info.CpsShortUrl, err = service.NewCpsShortUrl(c, &tx).InfoShortUrl(局_shortUrl)
	if err != nil {
		return
	}
	//http://localhost:9000/user/10001/

	//跳转到中间页 //http://localhost:9000/user/10001/ ,然后

	var 局_跳转url string

	局_跳转url = info.CpsShortUrl.BaseUrl + "#/pages/other/jump?type=1&cpsCode=" + strconv.Itoa(info.CpsShortUrl.Uid) + "&routerUrl=" + url.QueryEscape(info.CpsShortUrl.RouterUrl)
	//跳转到本地中间页写入本地推荐人数据然后跳转
	c.Redirect(302, 局_跳转url)
	// 跳转成功,写入数据库计数+1
	_ = service.NewCpsShortUrl(c, &tx).ClickCountUP(info.CpsShortUrl.Id, 1)

}

// 创建短链,短链信息
func (C *CpsShortUrl) Create(c *gin.Context) {
	var 请求 struct {
		BaseUrl   string `json:"baseUrl"  binding:"required" zh:"基础地址"`
		RouterUrl string `json:"routerUrl"  binding:"required" zh:"路由地址"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	var err error
	var info = struct {
		appInfo      DB.DB_AppInfo
		likeInfo     DB.DB_LinksToken
		CpsShortUrls []dbm.DB_CpsShortUrl
		CpsShortUrl  dbm.DB_CpsShortUrl
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	tx := *global.GVA_DB
	//先读取这个长链接和uid 是否已经有符合条件的短连接了,如果有直接返回,如果没有在创建,
	info.CpsShortUrls, err = service.NewCpsShortUrl(c, &tx).Infos(map[string]interface{}{
		"uid":       info.likeInfo.Uid,
		"baseUrl":   请求.BaseUrl,
		"routerUrl": 请求.RouterUrl,
	})
	if len(info.CpsShortUrls) > 0 {
		info.CpsShortUrl = info.CpsShortUrls[0]
	} else {
		info.CpsShortUrl.CreatedAt = time.Now().Unix()
		info.CpsShortUrl.UpdatedAt = info.CpsShortUrl.CreatedAt
		info.CpsShortUrl.ShortUrl = W文本_取随机字符串(6)
		info.CpsShortUrl.BaseUrl = 请求.BaseUrl
		info.CpsShortUrl.RouterUrl = 请求.RouterUrl
		info.CpsShortUrl.ClickCount = 0
		info.CpsShortUrl.Uid = info.likeInfo.Uid
		info.CpsShortUrl.Status = 1
		_, err = service.NewCpsShortUrl(c, &tx).Create(&info.CpsShortUrl)
		if err != nil {
			response.FailWithMessage(c, err.Error())
			return
		}
	}
	response.OkWithData(c, gin.H{
		"shortUrl": setting.Q系统设置().X系统地址 + "/c/" + info.CpsShortUrl.ShortUrl,
	})
	return
}

// 获取短链,短链信息
func (C *CpsShortUrl) Info(c *gin.Context) {
	var 请求 struct {
		ShortUrl string `json:"ShortUrl"  zh:"长链接地址"` // 用户名
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	var info = struct {
		appInfo     DB.DB_AppInfo
		likeInfo    DB.DB_LinksToken
		CpsShortUrl dbm.DB_CpsShortUrl
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	var err error
	tx := *global.GVA_DB

	info.CpsShortUrl, err = service.NewCpsShortUrl(c, &tx).InfoShortUrl(请求.ShortUrl)

	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}

	response.OkWithData(c, gin.H{
		"shortUrl": info.CpsShortUrl.ShortUrl,
	})
	return
}
