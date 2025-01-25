package controller

import (
	"github.com/gin-gonic/gin"
	"server/new/app/controller/Common"
	"server/new/app/logic/common/cloudStorage"
	"server/structs/Http/response"
)

type CloudStorage struct {
	Common.Common
}

func NewCloudStorageController() *CloudStorage {
	return &CloudStorage{}
}

// 云存储_取文件上传授权
func (C *CloudStorage) GetUploadToken(c *gin.Context) {
	var 请求 struct {
		Path string `json:"Path"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	取文件上传授权, err := cloudStorage.L_云存储.Q取文件上传授权(c, 请求.Path)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(取文件上传授权, "操作成功", c)
	return
}

// 云存储_取外链
func (C *CloudStorage) GetDownloadUrl(c *gin.Context) {
	var 请求 struct {
		Path     string `json:"Path"`
		LongTime int64  `json:"LongTime"` //有效时间
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	下载地址, err := cloudStorage.L_云存储.Q取外链地址(c, 请求.Path, 请求.LongTime)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithDetailed(gin.H{"Url": 下载地址}, "操作成功", c)
	return
}
