package App

import (
	"errors"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_KaClass"
	"server/Service/Ser_PublicData"
	"server/global"
	"server/structs/Http/response"
	DB "server/structs/db"
	"server/utils"
	"strconv"
	"strings"
)

type Api struct{}

// GetAppInfo
func (a *Api) GetAppInfo(c *gin.Context) {
	var 请求 结构请求_单id
	//{"Id":2}
	err := c.ShouldBindJSON(&请求)
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var DB_AppInfo DB.DB_AppInfo
	err = global.GVA_DB.Model(DB.DB_AppInfo{}).Where("AppId = ?", 请求.Id).Find(&DB_AppInfo).Error
	// 没查到数据

	if err != nil {
		response.FailWithMessage("查询APPID:"+strconv.Itoa(请求.Id)+"详细信息失败", c)
		return
	}

	response.OkWithDetailed(结构响应_GetAppInfo{
		AppInfo:   DB_AppInfo,
		KaClass:   Ser_KaClass.KaName取map列表Int(请求.Id),
		ServerUrl: global.GVA_CONFIG.X系统设置.X系统地址,
		Port:      global.GVA_CONFIG.Port,
	}, "获取成功", c)
	return
}

type 结构响应_GetAppInfo struct {
	AppInfo   DB.DB_AppInfo  `json:"AppInfo"`
	KaClass   map[int]string `json:"KaClass"`
	ServerUrl string         `json:"ServerUrl"`
	Port      int            `json:"Port"`
}
type 结构请求_单id struct {
	Id int `json:"Id"`
}

type 结构请求_GetAppList struct {
	Page     int    `json:"Page"`     // 页
	Size     int    `json:"Size"`     // 页数量
	Status   int    `json:"Status"`   // 状态id
	Type     int    `json:"Type"`     // 关键字类型  1 id 2 用户名
	Keywords string `json:"Keywords"` // 关键字
	Order    int    `json:"Order"`    // 0 倒序 1 正序
}

// GetAppList
// 获取用户信息列表
func (a *Api) GetAppList(c *gin.Context) {
	var 请求 结构请求_GetAppList
	//{"Type":"2","Size":10,"Page":1,"Status":"1","keywords":"1"}
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("提交参数错误:"+err.Error(), c)
		return
	}

	var DB_AppInfo_简化1 []DB_AppInfo_简化
	var 总数 int64
	局_DB := global.GVA_DB.Model(DB.DB_AppInfo{})

	if 请求.Order == 1 {
		局_DB.Order("AppId ASC")
	} else {
		局_DB.Order("AppId DESC")
	}

	if 请求.Status == 1 || 请求.Status == 2 || 请求.Status == 3 {
		局_DB.Where("Status = ?", 请求.Status)
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			局_DB.Where("AppId = ?", 请求.Keywords)
		case 2: //应用名称
			局_DB.Where("LOCATE(?, AppName)>0 ", 请求.Keywords)
		}
	}

	//Count(&总数) 必须放在where 后面 不然值会被清0
	err = 局_DB.Count(&总数).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Find(&DB_AppInfo_简化1).Error

	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("GetAppList:" + err.Error())
		return
	}
	//mark 需要处理卡信息
	response.OkWithDetailed(结构响应_GetAppList{DB_AppInfo_简化1, 总数}, "获取成功", c)
	return

}

type 结构响应_GetAppList struct {
	List  interface{} `json:"List"`  // 列表
	Count int64       `json:"Count"` // 总数

}

type DB_AppInfo_简化 struct {
	AppId            int    `json:"AppId" gorm:"column:AppId;primarykey"` // id
	AppName          string `json:"AppName" gorm:"column:AppName;comment:应用名称"`
	Status           int    `json:"Status" gorm:"column:Status;default:3;comment:状态(1>停止运营,2>免费模式,3>收费模式)"`
	AppStatusMessage string `json:"AppStatusMessage" gorm:"column:AppStatusMessage;comment:状态原因"`
	AppVer           string `json:"AppVer"  gorm:"column:AppVer;default:1.0.0;comment:软件版本"`
	CryptoType       int    `json:"CryptoType"  gorm:"column:CryptoType;default:1;comment:加密类型"` //加密类型 0: 明文 1des加密)
	AppType          int    `json:"AppType"  gorm:"column:AppType;default:1;comment:软件类型"`
}

// Del批量删除App
func (a *Api) Del批量删除App(c *gin.Context) {
	var 请求 结构请求_ID数组
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}
	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}
	for _, 局_id := range 请求.Id {
		if 局_id <= 10000 {
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
		global.GVA_DB.Model(DB.DB_KaClass{}).Where("AppId IN ? ", 请求.Id).Delete("")
		global.GVA_DB.Model(DB.DB_Ka{}).Where("AppId IN ? ", 请求.Id).Delete("")
	}

	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(影响行数, 10), c)
	return
}

type 结构请求_ID数组 struct {
	Id []int `json:"Id"` //id数组
}

// save 保存
func (a *Api) SaveApp信息(c *gin.Context) {
	var 请求 结构响应_AppInfo
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	if 请求.AppData.AppId <= 0 {
		response.FailWithMessage("AppId错误"+strconv.Itoa(请求.AppData.AppId), c)
		return
	}

	// 没查到数据
	if Ser_AppInfo.App存在数量(请求.AppData.AppId) == 0 {
		response.FailWithMessage("应用不存在", c)
		return
	}
	msg := ""

	if 请求.AppData.Status < 0 || 请求.AppData.Status > 3 {
		response.FailWithMessage("状态Id错误"+msg, c)
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

	err = Ser_AppInfo.App修改信息(请求.AppData)

	if err != nil {
		response.FailWithMessage("保存失败", c)
		return
	}

	for _, 专属变量 := range 请求.PublicData {
		Ser_PublicData.P置值2(专属变量)
	}

	response.OkWithMessage("保存成功", c)
	return
}

type 结构响应_AppInfo struct {
	AppData    DB.DB_AppInfo      `json:"AppData"`    // 列表
	PublicData []DB.DB_PublicData `json:"PublicData"` // 列表
}

// NewApp信息
func (a *Api) NewApp信息(c *gin.Context) {
	var 请求 请求_NewApp
	err := c.ShouldBindJSON(&请求)
	//解析失败
	if err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	err = Ser_AppInfo.NewApp信息(请求.AppId, 请求.AppType, 请求.AppName)

	if err != nil {
		response.FailWithMessage(err.Error(), c)
	} else {
		response.OkWithMessage("添加成功", c)
	}

	return
}

type 请求_NewApp struct {
	AppId   int    `json:"AppId" gorm:"column:AppId;primarykey"` // id
	AppName string `json:"AppName" gorm:"column:AppName;comment:应用名称"`
	AppType int    `json:"AppType"  gorm:"column:AppType;default:1;comment:软件类型"` //1=账号限时,2=账号计点,3卡号限时,4=卡号计点
}

// GetAppIdMax 取最大appid值
func (a *Api) GetAppIdMax(c *gin.Context) {
	var AppIdMax int64

	err := global.GVA_DB.Model(DB.DB_AppInfo{}).Select("Max(AppId)").Find(&AppIdMax).Error
	// 没查到数据

	if err != nil {
		//出错可能是没有查到应用 直接返回10001
		response.OkWithDetailed(响应_AppIdMax{10000}, "获取成功", c)
		//response.FailWithMessage("查询APPID最大值失败", c)
		return
	}

	response.OkWithDetailed(响应_AppIdMax{AppIdMax}, "获取成功", c)
	return
}

type 响应_AppIdMax struct {
	AppIdMax int64 `json:"AppIdMax"`
}

// GetAppIdNameList 取appid和名字数组
func (a *Api) GetAppIdNameList(c *gin.Context) {

	AppIdName := Ser_AppInfo.App取map列表String()
	delete(AppIdName, "1")
	delete(AppIdName, "2")
	delete(AppIdName, "3")
	var 临时Int int
	var Name []键值对

	for Key := range AppIdName {
		临时Int, _ = strconv.Atoi(Key)
		Name = append(Name, 键值对{AppId: 临时Int, AppName: AppIdName[Key]})
	}

	response.OkWithDetailed(响应_AppIdNameList{AppIdName, Name}, "获取成功", c)
	return
}

type 响应_AppIdNameList struct {
	Map   map[string]string `json:"Map"`
	Array []键值对             `json:"Array"`
}

type 键值对 struct {
	AppId   int    `json:"Appid"`
	AppName string `json:"AppName"`
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
		//检测每一条版本号 * 是否在版本末尾
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

func (a *Api) Get全部用户APi(c *gin.Context) {
	局_path数组 := [...][2]string{
		{"NewUserInfo", "用户注册"},
		{"UserLogin", "用户登录"},
		{"UseKa", "卡号充值"},
		{"SetPassWord", "密码找回或修改"},
		{"GetSMSCaptcha", "取短信验证码信息"},
		{"UserReduceMoney", "用户减少余额"},
		{"UserReduceVipNumber", "用户减少积分"},
		{"UserReduceVipTime", "用户减少点数"},
		{"IsServerLink", "取服务器连接状态"},
		{"IsLogin", "取登录状态"},
		{"GetVipData", "取Vip数据"},
		{"GetAppGongGao", "取应用公告"},
		{"GetAppUpDataJson", "取新版本下载地址"},
		{"GetAppPublicData", "取应用专属变量"},
		{"GetPublicData", "取公共变量"},
		{"GetAppVersion", "取应用最新版本"},
		{"GetAppHomeUrl", "取应用主页Url"},
		{"SetAppUserKey", "置新绑定信息"},
		{"SetNewUserMsg", "置新用户消息"},
		{"GetCaptcha", "取验证码信息"},
		{"GetAppUserKey", "取用户绑定信息"},
		{"GetIsUser", "取用户是否存在"},
		{"GetAppUserInfo", "取软件用户信息"},
		{"GetUserInfo", "取用户基础信息"},
		{"SetUserQqEmailPhone", "置用户基础信息"},

		{"GetSystemTime", "取系统时间戳"},
		{"GetAppUserVipTime", "取Vip到期时间戳"},
		{"GetAppUserNote", "取软件用户备注"},
		{"LogOut", "用户登录注销"},
		{"RemoteLogOut", "用户登录远程注销"},
		{"HeartBeat", "心跳"},
		{"GetUserRmb", "取用户余额"},
		{"GetAppUserVipNumber", "取用户积分"},
		{"GetCaptchaApiList", "取开启验证码接口"},
		{"GetTab", "取动态标签"},
		{"SetTab", "置动态标签"},
		{"GetAliPayPC", "余额充值_支付宝PC支付"},
		{"GetWXPayPC", "余额充值_微信支付支付"},
		{"GetPayOrderStatus", "余额充值_支付订单状态查询"},
		{"GetPayStatus", "取支付通道状态"},
		{"GetPayKaList", "取可购买卡类列表"},
		{"GetPurchasedKaList", "取已购买充值卡列表"},
		{"PayMoneyToVipNumber", "余额购买积分"},
		{"PayMoneyToKa", "余额购买充值卡"},
		{"GetUserClassList", "取用户类型列表"},
		{"SetUserClass", "置用户类型"},
		{"RunJS", "云函数执行"},
		{"TaskPoolNewData", "任务池_任务创建"},
		{"TaskPoolGetData", "任务池_任务查询"},
		{"TaskPoolGetTask", "任务池_任务处理获取"},
		{"TaskPoolSetTask", "任务池_任务处理返回"},
	}
	response.OkWithDetailed(局_path数组, "获取成功", c)
	return

}

func (a *Api) Get全部WebAPi(c *gin.Context) {
	局_path数组 := [...][2]string{
		{"TaskPoolGetTask", "任务池_任务处理获取"},
		{"TaskPoolSetTask", "任务池_任务处理返回"},
		{"RunJs", "运行公共js函数"},
		{"GetAppUpDataJson", "取App最新下载地址"},
		{"NewKa", "新制卡号"},
		{"GetKaInfo", "取卡号详细信息"},
	}
	response.OkWithDetailed(局_path数组, "获取成功", c)
	return

}
