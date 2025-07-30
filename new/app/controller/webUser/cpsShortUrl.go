package controller

import (
	"EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"time"
)

type CpsShortUrl struct {
	Common.Common
}

func NewCpsShortUrlController() *CpsShortUrl {
	return &CpsShortUrl{}
}

// 获取短链,短链信息
func (C *CpsShortUrl) Create(c *gin.Context) {
	var 请求 struct {
		Url string `json:"url" binding:"required,min=6,max=190" zh:"长链接地址"` // 用户名
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	var info = struct {
		appInfo  DB.DB_AppInfo
		likeInfo DB.DB_LinksToken
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	var err error
	tx := *global.GVA_DB

	result := utils.W文本_取随机字符串(8)

	_, err = service.NewCpsShortUrl(c, &tx).Create(&dbm.DB_CpsShortUrl{
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
		ShortUrl:   result,
		LongUrl:    请求.Url,
		ClickCount: 0,
		Uid:        info.likeInfo.Uid,
		Status:     1,
	})

	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}

	response.OkWithData(c, gin.H{
		"shortUrl": result,
	})
	return
}

// 获取短链,短链信息
func (C *CpsShortUrl) info(c *gin.Context) {
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
