package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/encoding/gjson"
	"server/Service/Ser_AppInfo"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/logic/common/setting"
	"server/new/app/models/request"

	"server/structs/Http/response"
	"strconv"
)

// Tools
// @MenuName 工具
// @ModuleName Apk加验证
type ApkTools struct {
	Common.Common
}

func NewApkToolsController() *ApkTools {
	return &ApkTools{}
}

// 获取文件上传token
func (C *ApkTools) GetUploadToken(c *gin.Context) {
	var 请求 struct {
		Path string `json:"Path" binding:"required"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	var 局_返回 string
	var 局_json gin.H

	if !global.Q快验.Y云存储_取文件上传授权(&局_返回, "ApkV0"+global.X系统信息.H会员帐号+"_"+请求.Path) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	err := json.Unmarshal([]byte(局_返回), &局_json)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	//{
	//	"Path": "10001/1.apk",
	//		"Type": 2,
	//		"Url": "http://upload.qiniup.com",
	//		"UpToken": "6MSwiVmlwTnVtYmVyIjo....."
	//}
	response.OkWithDetailed(局_json, "获取成功", c)
}

// 创建apk加验证任务
func (C *ApkTools) CreateApkAddFNKYTask(c *gin.Context) {
	var 请求 struct {
		Path     string `json:"Path" binding:"required"`
		FileName string `json:"fileName" binding:"required"`
		AppId    int    `json:"AppId" binding:"required"`
		Q签名方式    int    `json:"签名方式" binding:"required"`
		Activity string `json:"Activity"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	var 响应任务Uuid string
	var aaa = make(gin.H, 6)

	aaa["Path"] = 请求.Path
	aaa["FileName"] = 请求.FileName
	aaa["AppId"] = 请求.AppId
	aaa["签名方式"] = 请求.Q签名方式
	aaa["Activity"] = 请求.Activity

	局_Appinfo := Ser_AppInfo.App取App详情(请求.AppId)
	局_系统地址 := setting.Q系统设置().X系统地址

	var appInfo = make(gin.H, 3)
	appInfo["CryptoType"] = 局_Appinfo.CryptoType
	appInfo["AppWeb"] = 局_系统地址 + 局_Appinfo.AppWeb
	if 局_Appinfo.CryptoType == 2 {
		appInfo["CryptoKeyAes"] = 局_Appinfo.CryptoKeyAes
	} else if 局_Appinfo.CryptoType == 3 {
		appInfo["CryptoKeyPublic"] = 局_Appinfo.CryptoKeyPublic
	}
	aaa["AppInfo"] = appInfo

	aaa2, err := json.Marshal(aaa)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if !global.Q快验.R任务池_任务创建(&响应任务Uuid, 9, string(aaa2)) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	response.OkWithDetailed(响应任务Uuid, "操作成功", c)
	return
}

func (C *ApkTools) GetList(c *gin.Context) {
	var 请求 struct {
		request.List
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	var 局_返回 string
	var 局_map struct {
		List  []list_item `json:"List"`  // 列表
		Count int64       `json:"Count"` // 总数
	}
	if !global.Q快验.R任务池_取任务列表(&局_返回, 请求.Page, 请求.Order, 请求.Size, 9) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	局_json := gjson.New(局_返回)
	局_map.Count = 局_json.Get("Count").Int64()

	for i := range 局_json.Len("List") {
		局_任务提交 := gjson.New(局_json.Get("List." + strconv.Itoa(i) + ".SubmitData").String())
		局_任务结果 := gjson.New(局_json.Get("List." + strconv.Itoa(i) + ".ReturnData").String())
		局_map.List = append(局_map.List, list_item{
			Uuid:        局_json.Get("List." + strconv.Itoa(i) + ".uuid").String(),
			Status:      局_json.Get("List." + strconv.Itoa(i) + ".Status").Int(),
			TimeStart:   局_json.Get("List." + strconv.Itoa(i) + ".TimeStart").Int64(),
			TimeEnd:     局_json.Get("List." + strconv.Itoa(i) + ".TimeEnd").Int64(),
			FileName:    局_任务提交.Get("FileName").String(),
			Path:        局_任务提交.Get("Path").String(),
			Q签名方式:       局_任务提交.Get("签名方式").Int(),
			AppId:       局_任务提交.Get("AppId").Int(),
			AppName:     Ser_AppInfo.AppId取应用名称(局_任务提交.Get("AppId").Int()),
			DownloadUrl: 局_任务结果.Get("Url").String(),
			Err:         局_任务结果.Get("msg").String(),
		})
	}

	response.OkWithDetailed(局_map, "操作成功", c)
	return
}

type list_item struct {
	Uuid        string `json:"Uuid"`
	Status      int    `json:"Status"`
	TimeStart   int64  `json:"TimeStart"`
	TimeEnd     int64  `json:"TimeEnd"`
	FileName    string `json:"FileName"`
	Path        string `json:"Path"`
	Q签名方式       int    `json:"签名方式"`
	AppId       int    `json:"AppId"`
	AppName     string `json:"AppName"`
	DownloadUrl string `json:"DownloadUrl"`
	Err         string `json:"Err"`
}

// 获取任务池状态
func (C *ApkTools) GetTaskIdStatus(c *gin.Context) {

	var 局_返回 string
	var 局_json gin.H

	if !global.Q快验.R任务池_取任务状态(&局_返回) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}

	err := json.Unmarshal([]byte(局_返回), &局_json)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	//{"id1":1,"id2":1,"id3":1,"id4":1}

	response.OkWithDetailed(局_json, "获取成功", c)
}
