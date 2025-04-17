package controller

import (
	. "EFunc/utils"
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
// @ModuleName Exe加验证
type ExeTools struct {
	Common.Common
}

func NewExeToolsController() *ExeTools {
	return &ExeTools{}
}

// 获取文件上传token
func (C *ExeTools) GetUploadToken(c *gin.Context) {
	var 请求 struct {
		Path string `json:"Path" binding:"required"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	var 局_返回 string
	var 局_json gin.H

	if !global.Q快验.Y云存储_取文件上传授权(&局_返回, global.X系统信息.H会员帐号+"/"+请求.Path) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	err := json.Unmarshal([]byte(局_返回), &局_json)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	//{
	//	"Path": "10001/1.exe",
	//		"Type": 2,
	//		"Url": "http://upload.qiniup.com",
	//		"UpToken": "6MSwiVmlwTnVtYmVyIjo....."
	//}
	response.OkWithDetailed(局_json, "获取成功", c)
}

// 创建exe加验证任务
func (C *ExeTools) CreateExeAddFNKYTask(c *gin.Context) {
	var 请求 struct {
		Path     string `json:"Path" binding:"required"`
		FileName string `json:"fileName" binding:"required"`
		AppId    int    `json:"AppId" binding:"required"`
		Ui       int    `json:"Ui" binding:"required"`
		VMP一机一码  bool   `json:"VMP一机一码" `
		J检测调试器   bool   `json:"检测调试器" `
		J内存保护    bool   `json:"内存保护" `
		J导入表保护   bool   `json:"导入表保护" `
		J资源保护    bool   `json:"资源保护" `
		J压缩文件    bool   `json:"压缩文件" `
		J寄存器保护   bool   `json:"寄存器保护" `
		J常量保护    bool   `json:"常量保护" `
		J虚拟化工具   bool   `json:"虚拟化工具" `
		J移除重定位信息 bool   `json:"移除重定位信息" `
		J移除调试信息  bool   `json:"移除调试信息" `
	}

	if !C.ToJSON(c, &请求) {
		return
	}

	var 响应任务Uuid string
	var aaa = make(gin.H, 6)
	//把请求结构体 json 转换为 gin.h

	aaa["Path"] = 请求.Path
	aaa["FileName"] = 请求.FileName
	aaa["AppId"] = 请求.AppId
	aaa["Ui"] = 请求.Ui
	aaa["VMP一机一码"] = 请求.VMP一机一码
	aaa["检测调试器"] = 请求.J检测调试器
	aaa["内存保护"] = 请求.J内存保护
	aaa["导入表保护"] = 请求.J导入表保护
	aaa["资源保护"] = 请求.J资源保护
	aaa["压缩文件"] = 请求.J压缩文件
	aaa["寄存器保护"] = 请求.J寄存器保护
	aaa["常量保护"] = 请求.J常量保护
	aaa["虚拟化工具"] = 请求.J虚拟化工具
	aaa["移除重定位信息"] = 请求.J移除重定位信息
	aaa["移除调试信息"] = 请求.J移除调试信息

	局_Appinfo := Ser_AppInfo.App取App详情(请求.AppId)
	局_系统地址 := setting.Q系统设置().X系统地址
	局_可用版本 := W文本_分割文本(局_Appinfo.AppVer, "\n")
	var appInfo = make(gin.H, 4)
	if len(局_可用版本) == 0 || 局_Appinfo.AppVer == "" {
		appInfo["VerSion"] = "1.0.0"
	} else {
		局_分解版本号最新 := W文本_分割文本(局_可用版本[0], ".")
		局_分解版本号最新[len(局_分解版本号最新)-1] = strconv.Itoa(D到整数(局_分解版本号最新[len(局_分解版本号最新)-1]) + 1)
		appInfo["VerSion"] = S数组_合并文本(局_分解版本号最新, ".")
	}

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
	if !global.Q快验.R任务池_任务创建(&响应任务Uuid, 10, string(aaa2)) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	response.OkWithDetailed(响应任务Uuid, "操作成功", c)
	return
}

func (C *ExeTools) GetList(c *gin.Context) {
	var 请求 struct {
		request.List
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	var 局_返回 string
	var 局_map struct {
		List  []exe_list_item `json:"List"`  // 列表
		Count int64           `json:"Count"` // 总数
	}
	if !global.Q快验.R任务池_取任务列表(&局_返回, 请求.Page, 请求.Order, 请求.Size, 10) {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	局_json := gjson.New(局_返回)
	局_map.Count = 局_json.Get("Count").Int64()

	for i := range 局_json.Len("List") {
		局_任务提交 := gjson.New(局_json.Get("List." + strconv.Itoa(i) + ".SubmitData").String())
		局_任务结果 := gjson.New(局_json.Get("List." + strconv.Itoa(i) + ".ReturnData").String())
		局_map.List = append(局_map.List, exe_list_item{
			Uuid:        局_json.Get("List." + strconv.Itoa(i) + ".uuid").String(),
			Status:      局_json.Get("List." + strconv.Itoa(i) + ".Status").Int(),
			TimeStart:   局_json.Get("List." + strconv.Itoa(i) + ".TimeStart").Int64(),
			TimeEnd:     局_json.Get("List." + strconv.Itoa(i) + ".TimeEnd").Int64(),
			FileName:    局_任务提交.Get("FileName").String(),
			Path:        局_任务提交.Get("Path").String(),
			Ui:          局_任务提交.Get("Ui").Int(),
			AppId:       局_任务提交.Get("AppId").Int(),
			AppName:     Ser_AppInfo.AppId取应用名称(局_任务提交.Get("AppId").Int()),
			DownloadUrl: 局_任务结果.Get("Url").String(),
			ExeMd5:      局_任务结果.Get("ExeMd5").String(),
			Err:         局_任务结果.Get("msg").String(),
		})
	}

	response.OkWithDetailed(局_map, "操作成功", c)
	return
}

func (C *ExeTools) GetUiList(c *gin.Context) {

	//{ "list":[
	//{ id: 1, url: 'http://cdnjson.com/images/2025/04/10/165224b1w0ww55607xzw5x.png',label: '蓝色清新' },
	//{ id: 2, url: 'https://www.fnkuaiyan.cn/images/logo4.png',label: '蓝色清新2'  },
	//{ id: 3, url: 'https://www.fnkuaiyan.cn/images/logo4.png',label: '蓝色清新3'  },
	//]
	//}

	var 局_返回 string
	if !global.Q快验.Q取应用专属变量(&局_返回, "exeUi列表") {
		response.FailWithMessage(global.Q快验.Q取错误信息(nil), c)
		return
	}
	var 局_map []struct {
		Id    int    `json:"id"`
		Url   string `json:"url"`
		Label string `json:"label"`
	}
	局_json := gjson.New(局_返回)
	局_json = gjson.New(局_json.Get("exeUi列表").String())
	err := 局_json.Scan(&局_map)

	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(局_map, "操作成功", c)
	return
}

type exe_list_item struct {
	Uuid        string `json:"Uuid"`
	Status      int    `json:"Status"`
	TimeStart   int64  `json:"TimeStart"`
	TimeEnd     int64  `json:"TimeEnd"`
	FileName    string `json:"FileName"`
	Path        string `json:"Path"`
	Ui          int    `json:"Ui"`
	AppId       int    `json:"AppId"`
	AppName     string `json:"AppName"`
	DownloadUrl string `json:"DownloadUrl"`
	ExeMd5      string `json:"ExeMd5"`
	Err         string `json:"Err"`
}

// 获取任务池状态
func (C *ExeTools) GetTaskIdStatus(c *gin.Context) {

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
