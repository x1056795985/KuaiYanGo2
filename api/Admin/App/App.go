package App

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
	"server/api/WebApi"
	"server/api/middleware"
	"server/global"
	"server/new/app/logic/common/appInfo"
	"server/new/app/logic/common/publicData"
	"server/new/app/logic/common/setting"
	"server/new/app/router/webApi2"
	"server/structs/Http/response"
	DB "server/structs/db"
	"sort"
	"strconv"
	"strings"
	"time"
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
		ServerUrl: setting.Q系统设置().X系统地址,
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
		局_DB.Order("Sort DESC, AppId ASC")
	} else {
		局_DB.Order("Sort DESC, AppId DESC")
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
	Sort             int64  `json:"Sort" gorm:"column:Sort;default:0;comment:排序权重; "`
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

	局_旧AppInfo := Ser_AppInfo.App取App详情(请求.AppData.AppId)
	err = Ser_AppInfo.App修改信息(请求.AppData)

	if err != nil {
		response.FailWithMessage("保存失败", c)
		return
	}
	//私钥被改变 写入到用户消息,方便误操作找回,因为私钥丢失无法恢复,必须记录一下,不然客户全部需要换公钥
	if 局_旧AppInfo.CryptoKeyPrivate != 请求.AppData.CryptoKeyPrivate {
		Ser_Log.Log_写用户消息(Ser_Log.Log用户消息类型_其他, Ser_Admin.Id取User(1), 请求.AppData.AppName, "", "防误操作应用"+strconv.Itoa(局_旧AppInfo.AppId)+"更换私钥旧私钥:"+局_旧AppInfo.CryptoKeyPrivate, c.ClientIP())
	}

	//===========检查专属变量
	for _, 专属变量 := range 请求.PublicData {
		局_临时, err2 := publicData.L_publicData.Q取值2(c, 专属变量.AppId, 专属变量.Name)
		if err2 != nil {
			continue
		}
		if 局_临时.Value != 专属变量.Value || 局_临时.IsVip != 专属变量.IsVip || 局_临时.Note != 专属变量.Note || 局_临时.Type != 专属变量.Type || 局_临时.Sort != 专属变量.Sort {
			//只有值改变了才修改时间戳
			if 局_临时.Value != 专属变量.Value {
				专属变量.Time = int(time.Now().Unix())
			}

			_ = publicData.L_publicData.Z置值_原值(c, 专属变量)
		}
	}

	//================检查apihook函数
	JSON, err := fastjson.Parse(请求.AppData.ApiHook)
	if err == nil {
		if object, err2 := JSON.Object(); err2 == nil {
			//JSON取全部成员
			object.Visit(func(key []byte, v *fastjson.Value) {
				// 获取所有键名
				//{"UserLogin":{"Before":"hook登录前","After":"hook登录后"}}
				局_hook函数名 := strings.TrimSpace(string(v.GetStringBytes("Before")))
				if len(局_hook函数名) > 0 && !Ser_PublicJs.Name是否存在(Ser_PublicJs.Js类型_ApiHook函数, 局_hook函数名) {
					Ser_PublicJs.C创建(DB.DB_PublicJs{
						AppId: 3,
						Name:  局_hook函数名,
						Value: "function " + 局_hook函数名 + Api之前Hook函数模板,
						Type:  2,
						IsVip: 0,
						Note:  请求.AppData.AppName + "(" + strconv.Itoa(请求.AppData.AppId) + ")函数" + string(key) + "函数hook进入前自动创建",
					})
				}
				局_hook函数名 = strings.TrimSpace(string(v.GetStringBytes("After")))
				if len(局_hook函数名) > 0 && !Ser_PublicJs.Name是否存在(Ser_PublicJs.Js类型_ApiHook函数, 局_hook函数名) {
					Ser_PublicJs.C创建(DB.DB_PublicJs{
						AppId: 3,
						Name:  局_hook函数名,
						Value: "function " + 局_hook函数名 + Api之后Hook函数模板,
						Type:  2,
						IsVip: 0,
						Note:  请求.AppData.AppName + "(" + strconv.Itoa(请求.AppData.AppId) + ")函数" + string(key) + "函数hook退出后自动创建",
					})
				}
			})

		}
	}

	//================================

	response.OkWithMessage("保存成功", c)
	return
}

const Api之前Hook函数模板 = `(JSON请求明文) {
    //这里的错误无法拦截,所以,如果js错误,可能会导致,用户返回"Api不存在"
    //JSON.stringify($Request)  //在 $Request里可以获取到 请求的大部分信息
    //{"Method":"POST","Url":{"Scheme":"","Opaque":"","User":null,"Host":"","Path":"/Api","RawPath":"","OmitHost":false,"ForceQuery":false,"RawQuery":"AppId=10002","Fragment":"","RawFragment":""},"Header":["Connection: Keep-Alive","Referer: http://127.0.0.1:18888/Api?AppId=10002","Content-Length: 467","Content-Type: application/x-www-form-urlencoded; Charset=UTF-8","Accept: */*","Accept-Language: zh-cn","User-Agent: Mozilla/4.0 (compatible; MSIE 9.0; Windows NT 6.1)","Token: PNYDKXDHLORTNVGEEY99YYSPQGFLQF7L"],"Host":"127.0.0.1:18888","Body":[]}

    //局_url = "https://www.baidu.com/"
    //局_返回 = $api_网页访问_GET(局_url, 15, "")
    //局_返回 = $api_网页访问_POST(局_url, "api=123", "",15, "")
    //{"StatusCode":200,"Headers":"Date: Sun, 21 May 2023 10:26:32 GMT\r\nContent-Length: 0\r\nContent-Type: application/x-www-form-urlencoded,\r\n","Cookies":"","Body":""}

    if (局_返回.Body !== "") {
        //$拦截原因 = "百度可以访问,所以不能登录." 
    }
	//这里可以替换请求明文信息,可以实现很多功能,比如自写算法解密
    return JSON请求明文
}`

const Api之后Hook函数模板 = `(JSON响应明文) {
    //{"Time":1697630688,"Status":200,"Msg":"百度可以访问,所以不能登录."}
    //{"Data":{"Key":"绑定信息","LoginIp":"127.0.0.1","LoginTime":1697630755,"OutUser":0,"RegisterTime":1696677905,"UserClassMark":2,"UserClassName":"vip2","VipNumber":0,"VipTime":1701300424},"Time":1697630755,"Status":73386,"Msg":""}

/*    let 局_返回信息 = JSON.parse(JSON响应明文) //把响应信息明文转换成对象,好操作
    if (局_返回信息.Status > 10000) {
        局_返回信息.Data.Key = "99999999" //返回的绑定信息被我修改了
    }
    JSON响应明文 = JSON.stringify(局_返回信息) //再把对象转换回明文字符串
*/

    //局_url = "https://www.baidu.com/"
    //局_返回 = $api_网页访问_GET(局_url, 15, "")
    //    协议头 = [
    //        "Accept: application/json, text/javascript, */*; q=0.01",
    //        "Content-Type: application/json",
    //        "User-Agent: Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36"
    //    ]
    //    返回对象 = $api_网页访问_POST(局_url, "api=123",协议头,"", 15, "")
    //{"StatusCode":200,"Headers":"Date: Sun, 21 May 2023 10:26:32 GMT\r\nContent-Length: 0\r\nContent-Type: application/x-www-form-urlencoded,\r\n","Cookies":"","Body":""}

    //这里可以替换响应的json信息文本, 如果想拦截直接替换为报错的json就可以了,注意状态码,和时间戳
    return JSON响应明文
}`

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
	AppId     int    `json:"AppId" gorm:"column:AppId;primarykey"` // id
	AppName   string `json:"AppName" gorm:"column:AppName;comment:应用名称"`
	AppType   int    `json:"AppType"  gorm:"column:AppType;default:1;comment:软件类型"` //1=账号限时,2=账号计点,3卡号限时,4=卡号计点
	CopyAppId int    `json:"CopyAppId"`                                             //要复制的appId
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
	// 对 Name 数组 按键值对.Id 进行升序排序
	sort.Slice(Name, func(i, j int) bool {
		return Name[i].AppId < Name[j].AppId
	})
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
	局_path数组 := make([][]string, 0, len(middleware.J集_UserAPi路由))
	//把下边这些常用接口放在前面
	局_path数组 = append(局_path数组, []string{"NewUserInfo", "用户注册"})
	局_path数组 = append(局_path数组, []string{"UserLogin", "用户登录"})
	局_path数组 = append(局_path数组, []string{"UseKa", "卡号充值"})
	局_path数组 = append(局_path数组, []string{"SetPassWord", "密码找回或修改"})
	局_path数组 = append(局_path数组, []string{"GetSMSCaptcha", "取短信验证码信息"})
	局_path数组 = append(局_path数组, []string{"GetPayOrderStatus", "订单_取状态"})
	局_path数组 = append(局_path数组, []string{"PayKaUsa", "订单_购卡直冲"})
	局_path数组 = append(局_path数组, []string{"PayUserMoney", "订单_余额充值"})
	局_path数组 = append(局_path数组, []string{"PayUserVipNumber", "订单_积分充值"})
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

func (a *Api) Get全部WebAPi(c *gin.Context) {
	/*	局_path数组 := [...][2]string{
		{"TaskPoolGetTask", "任务池_任务处理获取"},
		{"TaskPoolSetTask", "任务池_任务处理返回"},
		{"RunJs", "运行公共js函数"},
		{"GetAppUpDataJson", "取App最新下载地址"},
		{"NewKa", "新制卡号"},
		{"GetKaInfo", "取卡号详细信息"},
	}*/

	局_path数组 := make([][]string, 0, len(WebApi.J集_UserAPi路由)+len(webApi2.J集_UserAPi路由2))
	for 键名, 键值 := range WebApi.J集_UserAPi路由 {
		局_path数组 = append(局_path数组, []string{键名, 键值.Z中文名})
	}
	//追加新版接口
	for 键名, 键值 := range webApi2.J集_UserAPi路由2 {
		局_path数组 = append(局_path数组, []string{键名, 键值.Z中文名})
	}
	response.OkWithDetailed(局_path数组, "获取成功", c)
	return

}
