package controller

import (
	"EFunc/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastjson"
	"server/Service/Ser_Admin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_Log"
	"server/Service/Ser_PublicJs"
	"server/api/middleware"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/logic/common/appInfo"
	"server/new/app/logic/common/publicData"
	"server/new/app/logic/common/setting"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/router/webApi2"
	"server/new/app/service"
	"server/structs/Http/response"
	DB "server/structs/db"
	"sort"
	"strconv"
	"strings"
	"time"
)

type App struct {
	Common.Common
}

func NewAppController() *App {
	return &App{}
}

type 结构响应_GetAppInfo struct {
	AppInfo   DB.DB_AppInfo  `json:"appInfo"`
	KaClass   map[int]string `json:"kaClass"`
	ServerUrl string         `json:"serverUrl"`
	Port      int            `json:"port"`
}

type 结构请求_单id struct {
	Id int `json:"id"`
}

type 结构请求_GetAppList struct {
	Page     int    `json:"page"`
	Size     int    `json:"size"`
	Status   int    `json:"status"`
	Type     int    `json:"type"`
	Keywords string `json:"keywords"`
	Order    int    `json:"order"`
}

type DB_AppInfo_简化 struct {
	AppId            int    `json:"appId" gorm:"column:AppId;primarykey"`
	AppName          string `json:"appName" gorm:"column:AppName"`
	Status           int    `json:"status" gorm:"column:Status"`
	AppStatusMessage string `json:"appStatusMessage" gorm:"column:AppStatusMessage"`
	AppVer           string `json:"appVer" gorm:"column:AppVer"`
	CryptoType       int    `json:"cryptoType" gorm:"column:CryptoType"`
	AppType          int    `json:"appType" gorm:"column:AppType"`
	Sort             int64  `json:"sort" gorm:"column:Sort"`
}

type 结构响应_GetAppList struct {
	List  interface{} `json:"list"`
	Count int64       `json:"count"`
}

type 结构请求_ID数组 struct {
	Id []int `json:"id"`
}

// GetList 获取应用列表
func (a *App) GetList(c *gin.Context) {
	var 请求 结构请求_GetAppList
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var DB_AppInfo_简化1 []DB_AppInfo_简化
	var 总数 int64
	局_DB := global.GVA_DB.Model(DB.DB_AppInfo{})

	if 请求.Order == 1 {
		局_DB.Order("Sort DESC, AppId ASC")
	} else {
		局_DB.Order("Sort DESC, AppId DESC")
	}

	if 请求.Status == 1 || 请求.Status == 2 || 请求.Status == 3 {
		局_DB.Where("Status = ?", 请求.Status)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1:
			局_DB.Where("AppId = ?", 请求.Keywords)
		case 2:
			局_DB.Where("LOCATE(?, AppName)>0 ", 请求.Keywords)
		}
	}

	err = 局_DB.Count(&总数).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&DB_AppInfo_简化1).Error

	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("GetAppList:" + err.Error())
		return
	}
	response.OkWithDetailed(结构响应_GetAppList{DB_AppInfo_简化1, 总数}, "获取成功", c)
	return
}

// GetInfo 获取应用详细信息
func (a *App) GetInfo(c *gin.Context) {
	var 请求 结构请求_单id
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var DB_AppInfo DB.DB_AppInfo
	err = global.GVA_DB.Model(DB.DB_AppInfo{}).Where("AppId = ?", 请求.Id).Find(&DB_AppInfo).Error

	if err != nil {
		response.FailWithMessage("查询APPID:"+strconv.Itoa(请求.Id)+"详细信息失败", c)
		return
	}

	response.OkWithDetailed(结构响应_GetAppInfo{
		AppInfo:   DB_AppInfo,
		KaClass:   Ser_KaClass.KaName取map列表Int(请求.Id),
		ServerUrl: setting.Q系统设置().X系统地址,
		Port:      global.GVA_CONFIG.Port,
	}, "获取成功", c)
	return
}

// New 新建应用
func (a *App) New(c *gin.Context) {
	var 请求 请求_NewApp
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if 请求.CopyAppId == 0 {
		err = appInfo.L_appInfo.NewApp信息(c, 请求.AppId, 请求.AppType, 请求.AppName)
	} else {
		err = Ser_AppInfo.CopyApp信息(请求.AppId, 请求.AppType, 请求.AppName, 请求.CopyAppId)
	}

	if err != nil {
		response.FailWithMessage(err.Error(), c)
	} else {
		if 请求.CopyAppId == 0 {
			response.OkWithMessage("添加成功", c)
		} else {
			response.OkWithMessage("复制成功", c)
		}
	}
	return
}

type 请求_NewApp struct {
	AppId     int    `json:"appId"`
	AppName   string `json:"appName"`
	AppType   int    `json:"appType"`
	CopyAppId int    `json:"copyAppId"`
}

// SaveInfo 保存应用信息
func (a *App) SaveInfo(c *gin.Context) {
	var 请求 struct {
		AppData        DB.DB_AppInfo         `json:"appData"`
		PublicData     []DB.DB_PublicData    `json:"publicData"`
		AppInfoWebUser dbm.DB_AppInfoWebUser `json:"appInfoWebUser"`
	}
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	if 请求.AppData.AppId <= 0 {
		response.FailWithMessage("AppId错误"+strconv.Itoa(请求.AppData.AppId), c)
		return
	}

	if Ser_AppInfo.App存在数量(请求.AppData.AppId) == 0 {
		response.FailWithMessage("应用不存在", c)
		return
	}

	if 请求.AppData.Status < 0 || 请求.AppData.Status > 3 {
		response.FailWithMessage("状态Id错误", c)
		return
	}
	if 请求.AppData.CryptoType == 2 && len(请求.AppData.CryptoKeyAes) != 24 {
		response.FailWithMessage("Aes_cbc_192密匙长度为24位字符", c)
		return
	}
	err = 版本号_检测通配符是否合法(请求.AppData.AppVer)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	局_旧AppInfo := Ser_AppInfo.App取App详情(请求.AppData.AppId)
	err = Ser_AppInfo.App修改信息(请求.AppData)

	if err != nil {
		response.FailWithMessage("保存失败", c)
		return
	}
	if 局_旧AppInfo.CryptoKeyPrivate != 请求.AppData.CryptoKeyPrivate {
		Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_其他, constant.APPID_管理平台, Ser_Admin.Id取User(1), 请求.AppData.AppName, "", "防误操作应用"+strconv.Itoa(局_旧AppInfo.AppId)+"更换私钥旧私钥:"+局_旧AppInfo.CryptoKeyPrivate, c.ClientIP())
	}

	tx := *global.GVA_DB
	_, err = service.NewAppInfoWebUser(c, &tx).Info(局_旧AppInfo.AppId)
	if err != nil {
		_, err = service.NewAppInfoWebUser(c, &tx).Create(dbm.DB_AppInfoWebUser{
			Id:             局_旧AppInfo.AppId,
			Status:         请求.AppInfoWebUser.Status,
			CaptchaLogin:   请求.AppInfoWebUser.CaptchaLogin,
			UrlDownload:    请求.AppInfoWebUser.UrlDownload,
			CaptchaReg:     请求.AppInfoWebUser.CaptchaReg,
			CaptchaSendSms: 请求.AppInfoWebUser.CaptchaSendSms,
			WebUserDomain:  请求.AppInfoWebUser.WebUserDomain,
			AgentOnlyOrder: 请求.AppInfoWebUser.AgentOnlyOrder,
		})
	} else {
		_, err = service.NewAppInfoWebUser(c, &tx).Update(局_旧AppInfo.AppId, map[string]interface{}{
			"status":         请求.AppInfoWebUser.Status,
			"captchaLogin":   请求.AppInfoWebUser.CaptchaLogin,
			"urlDownload":    请求.AppInfoWebUser.UrlDownload,
			"captchaReg":     请求.AppInfoWebUser.CaptchaReg,
			"captchaSendSms": 请求.AppInfoWebUser.CaptchaSendSms,
			"webUserDomain":  请求.AppInfoWebUser.WebUserDomain,
			"agentOnlyOrder": 请求.AppInfoWebUser.AgentOnlyOrder,
		})
	}
	if err != nil {
		response.FailWithMessage("保存网页用户中心配置失败", c)
		return
	}

	for _, 专属变量 := range 请求.PublicData {
		局_临时, err2 := publicData.L_publicData.Q取值2(c, 专属变量.AppId, 专属变量.Name)
		if err2 != nil {
			continue
		}
		if 局_临时.Value != 专属变量.Value || 局_临时.IsVip != 专属变量.IsVip || 局_临时.Note != 专属变量.Note || 局_临时.Type != 专属变量.Type || 局_临时.Sort != 专属变量.Sort {
			if 局_临时.Value != 专属变量.Value {
				专属变量.Time = time.Now().Unix()
			}
			_ = publicData.L_publicData.Z置值_原值(c, 专属变量)
		}
	}

	JSON, err := fastjson.Parse(请求.AppData.ApiHook)
	if err == nil {
		if object, err2 := JSON.Object(); err2 == nil {
			object.Visit(func(key []byte, v *fastjson.Value) {
				局_hook函数名 := strings.TrimSpace(string(v.GetStringBytes("Before")))
				if len(局_hook函数名) > 0 && !Ser_PublicJs.Name是否存在(Ser_PublicJs.Js类型_ApiHook函数, 局_hook函数名) {
					Ser_PublicJs.C创建(DB.DB_PublicJs{
						AppId: 3, Name: 局_hook函数名,
						Value: "function " + 局_hook函数名 + Api之前Hook函数模板,
						Type: 2, IsVip: 0,
						Note: 请求.AppData.AppName + "(" + strconv.Itoa(请求.AppData.AppId) + ")函数" + string(key) + "函数hook进入前自动创建",
					})
				}
				局_hook函数名 = strings.TrimSpace(string(v.GetStringBytes("After")))
				if len(局_hook函数名) > 0 && !Ser_PublicJs.Name是否存在(Ser_PublicJs.Js类型_ApiHook函数, 局_hook函数名) {
					Ser_PublicJs.C创建(DB.DB_PublicJs{
						AppId: 3, Name: 局_hook函数名,
						Value: "function " + 局_hook函数名 + Api之后Hook函数模板,
						Type: 2, IsVip: 0,
						Note: 请求.AppData.AppName + "(" + strconv.Itoa(请求.AppData.AppId) + ")函数" + string(key) + "函数hook退出后自动创建",
					})
				}
			})
		}
	}

	response.OkWithMessage("保存成功", c)
	return
}

const Api之前Hook函数模板 = `(JSON请求明文) {
    if (局_返回.Body !== "") {
    }
    return JSON请求明文
}`

const Api之后Hook函数模板 = `(JSON响应明文) {
    return JSON响应明文
}`

// Delete 删除应用
func (a *App) Delete(c *gin.Context) {
	var 请求 结构请求_ID数组
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}
	for _, 局_id := range 请求.Id {
		if 局_id < 10000 {
			response.FailWithMessage("appid不能小于10000", c)
			return
		}
	}

	var 影响行数 int64
	var db = global.GVA_DB
	db.Model(DB.DB_AppInfo{}).Count(&影响行数)

	if int(影响行数)-len(请求.Id) <= 0 {
		response.FailWithMessage("不能删除全部应用至少保留一个应用", c)
		return
	}

	影响行数 = db.Model(DB.DB_AppInfo{}).Where("AppId IN ? ", 请求.Id).Delete("").RowsAffected

	if db.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	for _, id值 := range 请求.Id {
		global.GVA_DB.Migrator().DropTable("db_AppUser_" + strconv.Itoa(id值))
		global.GVA_DB.Model(DB.DB_UserClass{}).Where("AppId IN ? ", 请求.Id).Delete("")
		global.GVA_DB.Model(dbm.DB_KaClass{}).Where("AppId IN ? ", 请求.Id).Delete("")
		global.GVA_DB.Model(DB.DB_Ka{}).Where("AppId IN ? ", 请求.Id).Delete("")
	}

	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}

// GetAppIdMax 取最大AppId
func (a *App) GetAppIdMax(c *gin.Context) {
	var AppIdMax int64
	err := global.GVA_DB.Model(DB.DB_AppInfo{}).Select("Max(AppId)").Find(&AppIdMax).Error
	if err != nil {
		response.OkWithDetailed(gin.H{"appIdMax": 10000}, "获取成功", c)
		return
	}
	response.OkWithDetailed(gin.H{"appIdMax": AppIdMax}, "获取成功", c)
	return
}

// GetAppIdNameList 取AppId和名称列表
func (a *App) GetAppIdNameList(c *gin.Context) {
	AppIdName := Ser_AppInfo.App取map列表String(false)
	var 临时Int int
	var Name []键值对
	for Key := range AppIdName {
		临时Int, _ = strconv.Atoi(Key)
		Name = append(Name, 键值对{AppId: 临时Int, AppName: AppIdName[Key]})
	}
	sort.Slice(Name, func(i, j int) bool {
		return Name[i].AppId < Name[j].AppId
	})
	response.OkWithDetailed(gin.H{"map": AppIdName, "array": Name}, "获取成功", c)
	return
}

type 键值对 struct {
	AppId   int    `json:"appId"`
	AppName string `json:"appName"`
}

// GetAllUserApi 获取全部用户API列表
func (a *App) GetAllUserApi(c *gin.Context) {
	局_path数组 := make([][]string, 0, len(middleware.J集_UserAPi路由))
	局_path数组 = append(局_path数组, []string{"NewUserInfo", "用户注册"})
	局_path数组 = append(局_path数组, []string{"UserLogin", "用户登录"})
	局_path数组 = append(局_path数组, []string{"UseKa", "卡号充值"})
	局_path数组 = append(局_path数组, []string{"SetPassWord", "密码找回或修改"})
	局_path数组 = append(局_path数组, []string{"GetSMSCaptcha", "取短信验证码信息"})
	局_path数组 = append(局_path数组, []string{"GetPayOrderStatus", "订单_取状态"})
	局_path数组 = append(局_path数组, []string{"PayKaUsa", "订单_购卡直冲"})
	局_path数组 = append(局_path数组, []string{"PayUserMoney", "订单_余额充值"})
	局_path数组 = append(局_path数组, []string{"PayGetKa", "订单_支付购卡"})

	for 键名, 键值 := range middleware.J集_UserAPi路由 {
		if 键名 == "NewUserInfo" || 键名 == "UserLogin" || 键名 == "UseKa" || 键名 == "SetPassWord" || 键名 == "GetSMSCaptcha" || 键名 == "GetPayOrderStatus" || 键名 == "PayKaUsa" || 键名 == "PayUserMoney" || 键名 == "PayUserVipNumber" || 键名 == "PayGetKa" {
			continue
		}
		if 键值.X显示 {
			局_path数组 = append(局_path数组, []string{键名, 键值.Z中文名})
		}
	}
	response.OkWithDetailed(局_path数组, "获取成功", c)
	return
}

// GetAllWebApi 获取全部WebApi列表
func (a *App) GetAllWebApi(c *gin.Context) {
	局_path数组 := make([][]string, 0, len(webApi2.J集_UserAPi路由2))
	for 键名, 键值 := range webApi2.J集_UserAPi路由2 {
		局_path数组 = append(局_path数组, []string{键名, 键值.Z中文名})
	}
	局_PublicJsName := Ser_PublicJs.P取全部公共函数名称(1)
	response.OkWithDetailed(gin.H{"api": 局_path数组, "publicJs": 局_PublicJsName}, "获取成功", c)
	return
}

func 版本号_检测通配符是否合法(可用版本号 string) error {
	var 可用版本号数组 []string = utils.W文本_分割文本(可用版本号, "\n")
	for 索引, 值 := range 可用版本号数组 {
		if 索引 == 0 {
			if strings.Contains(值, "*") {
				return errors.New("第一行为最新版本号,不能使用通配符*")
			} else {
				continue
			}
		}
		局_分解版本号 := utils.W文本_分割文本(值, ".")
		for _, 具体版本号 := range 局_分解版本号 {
			if strings.Contains(具体版本号, "*") {
				if strings.Index(具体版本号, "*") != len(具体版本号)-1 {
					return errors.New("第" + strconv.Itoa(索引+1) + "行版本号:" + 值 + ",*只能用在大,小,编译,版本号末尾")
				}
			}
		}
	}
	return nil
}
