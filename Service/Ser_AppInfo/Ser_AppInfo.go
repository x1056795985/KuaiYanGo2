package Ser_AppInfo

import (
	"errors"
	"github.com/songzhibin97/gkit/tools/rand_string"
	"server/Service/Ser_KaClass"
	"server/global"
	DB "server/structs/db"
	"server/utils"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func AppInfo取map列表Int() map[int]string {

	var DB_AppInfo []DB.DB_AppInfo
	var 总数 int64
	_ = global.GVA_DB.Model(DB.DB_AppInfo{}).Select("AppId", "AppName").Count(&总数).Find(&DB_AppInfo).Error
	var AppName = make(map[int]string, 总数+2)
	AppName[1] = "管理平台"
	AppName[2] = "代理平台"
	AppName[3] = "WebApi"
	//吧 id 和 app名字 放入map
	for 索引 := range DB_AppInfo {
		AppName[DB_AppInfo[索引].AppId] = DB_AppInfo[索引].AppName
	}

	return AppName
}
func App取map列表String() map[string]string {

	var DB_AppInfo []DB.DB_AppInfo
	var 总数 int64
	_ = global.GVA_DB.Model(DB.DB_AppInfo{}).Select("AppId", "AppName").Count(&总数).Find(&DB_AppInfo).Error
	var AppName = make(map[string]string, 总数+2)
	AppName["1"] = "管理平台"
	AppName["2"] = "代理平台"
	AppName["3"] = "WebApi"

	//吧 id 和 app名字 放入map
	for 索引 := range DB_AppInfo {
		AppName[strconv.Itoa(int(DB_AppInfo[索引].AppId))] = DB_AppInfo[索引].AppName
	}

	return AppName
}

func App取AppName(Appid int) (AppName string) {
	_ = global.GVA_DB.Model(DB.DB_AppInfo{}).Select("AppName").Where("AppId=?", Appid).First(&AppName).Error
	return AppName
}

func App取App详情(Appid int) (AppName DB.DB_AppInfo) {
	Data缓存, ok := global.H缓存.Get("DB_AppInfo_" + strconv.Itoa(Appid)) //读取缓存
	if ok {
		return Data缓存.(DB.DB_AppInfo)
	}
	_ = global.GVA_DB.Model(DB.DB_AppInfo{}).Where("AppId=?", Appid).First(&AppName).Error

	//高频率读取数据 写入缓存
	global.H缓存.Set("DB_AppInfo_"+strconv.Itoa(Appid), AppName, time.Minute*10) //10分钟有效

	return AppName
}
func App取App最新下载地址Json(Appid int) (下载地址 string) {
	var DB_AppInfo DB.DB_AppInfo
	_ = global.GVA_DB.Model(DB.DB_AppInfo{}).Where("AppId=?", Appid).First(&DB_AppInfo).Error
	下载地址 = DB_AppInfo.UrlDownload
	if strings.Index(DB_AppInfo.UrlDownload, "{{AppVer}}") != -1 && DB_AppInfo.AppVer != "" {
		局_可用版本 := utils.W文本_分割文本(DB_AppInfo.AppVer, "\n")
		if len(局_可用版本) > 0 {
			下载地址 = strings.Replace(DB_AppInfo.UrlDownload, "{{AppVer}}", 局_可用版本[0], -1)
		}
	}

	return 下载地址
}
func AppId是否存在(AppId int) bool {
	var appInfo int
	result := global.GVA_DB.Model(DB.DB_AppInfo{}).Select("1").Where("AppId = ?", AppId).First(&appInfo)
	return result.Error == nil
}
func AppId取应用名称(AppId int) string {
	if AppId < 10000 {
		return ""
	}
	AppName := ""
	_ = global.GVA_DB.Model(DB.DB_AppInfo{}).Select("AppName").Where("AppId = ?", AppId).First(&AppName).Error
	return AppName
}
func App取AppType(Appid int) (AppType int) {
	_ = global.GVA_DB.Model(DB.DB_AppInfo{}).Select("AppType").Where("AppId=?", Appid).First(&AppType).Error
	return AppType
}

func App是否为卡号(Appid int) bool {
	var AppType int = 0 //1=账号限时,2=账号计点,3卡号限时,4=卡号计点
	_ = global.GVA_DB.Model(DB.DB_AppInfo{}).Select("AppType").Where("AppId=?", Appid).First(&AppType).Error
	if AppType == 3 || AppType == 4 {
		return true
	}
	return false
}

func App是否为计点(Appid int) bool {
	var AppType int = 0 //1=账号限时,2=账号计点,3卡号限时,4=卡号计点
	_ = global.GVA_DB.Model(DB.DB_AppInfo{}).Select("AppType").Where("AppId=?", Appid).First(&AppType).Error
	if AppType == 2 || AppType == 4 {
		return true
	}
	return false
}

func App存在数量(Appid int) int64 {
	var count int64 = 0
	_ = global.GVA_DB.Model(DB.DB_AppInfo{}).Where("AppId = ?", Appid).Count(&count).Error

	return count

}

func App修改信息(AppInfo DB.DB_AppInfo) error {
	//高频率读取数据 写入缓存

	//直接排除AppType  AppWeb 禁止修改
	var db = global.GVA_DB.Model(DB.DB_AppInfo{}).Select(
		"AppName",
		"Status",
		"AppStatusMessage",
		"AppVer",
		"RegisterGiveKaClassId",
		"VerifyKey",
		"IsUserKeySame",
		"UpKeyData",
		"PackTimeOut",
		"OutTime",
		"UrlHome",
		"UrlDownload",
		"AppGongGao",
		"VipData",
		"CryptoType",
		"CryptoKeyAes",
		"CryptoKeyPrivate",
		"CryptoKeyPublic",
		"MaxOnline",
		"ExceedMaxOnlineOut",
		"RmbToVipNumber",
		"Captcha",
		"RegisterGiveKa",
		"ApiHook",
	).Omit("AppType", "AppWeb")

	err := db.Where("AppId= ?", AppInfo.AppId).Updates(AppInfo).Error
	if err == nil { //如果修改成功删除缓存
		global.H缓存.Delete("DB_AppInfo_" + strconv.Itoa(AppInfo.AppId)) //10分钟有效
	}
	return err
}

// NewApp信息
func NewApp信息(AppId, AppType int, AppName string) error {
	if AppId < 10000 {
		return errors.New("AppId请输>=10000的整数")
	}
	if utf8.RuneCountInString(AppName) < 4 || utf8.RuneCountInString(AppName) > 18 {
		return errors.New("应用名称长度必须大于4小于18")
	}
	msg := ""
	if !utils.Z正则_校验代理用户名(AppName, &msg) {
		return errors.New("应用名称" + msg)
	}

	var count int64
	err := global.GVA_DB.Model(DB.DB_AppInfo{}).Where("AppId = ?", AppId).Count(&count).Error
	// 没查到数据
	if count != 0 {
		return errors.New("AppID已存在")
	}

	if AppType > 4 || AppType < 1 {
		return errors.New("应用类型错误")
	}

	var NewApp DB.DB_AppInfo
	NewApp.AppId = AppId
	NewApp.AppType = AppType
	NewApp.AppName = AppName

	NewApp.AppWeb = `/Api?AppId=` + strconv.Itoa(int(AppId))
	NewApp.Status = 3 //3>收费模式
	NewApp.AppStatusMessage = "正常运营中"
	NewApp.AppVer = `1.0.0
*.*.*
*.*
*`
	NewApp.RegisterGiveKaClassId = 0
	if NewApp.AppType <= 2 {
		新卡类ID, err2 := Ser_KaClass.KaClass创建New(int(NewApp.AppId), "注册送卡", "ZC", 0, 0, 0, 0, -1, -1, 0, 1, 25, 1, 1, 1, 1)
		if err2 != nil {
			global.GVA_LOG.Error("创建App时创建注册送卡类失败," + err.Error())
		}
		NewApp.RegisterGiveKaClassId = 新卡类ID //注册赠送卡类的id
	}

	NewApp.VerifyKey = 1     //绑定模式
	NewApp.IsUserKeySame = 1 //不同用户可否相同
	NewApp.UpKeyData = 10    //修改绑定key增减值

	NewApp.UrlHome = "https://www.baidu.com/"
	NewApp.UrlDownload = `{
    "htmlurl": "www.baidu.com(自动下载失败打开指定网址,手动更新地址",
    "data": [{
        "WenJianMin": "文件名{{AppName}}{{AppVer}}.exe(  {{AppName}}变量替换为应用名称{{AppVer}} 这个变量会替换为最新版本的版本号,省的每次更新都改版本号 )",
        "md5": "e10adc3949ba59abbe56e057f20f883e(小写文件md5可选,有就校验,空就只校验文件名)",
        "Lujing": "/(下载本地相对路径)",
        "size": "12345(可选,不填写也没问题)",
        "url": "https://www.baidu.com/文件名{{AppName}}{{AppVer}}.exe(下载路径)",
        "YunXing": "1(值为更新完成后会运行这个文件,只能有一个文件值为1)"

    }, {
        "WenJianMin": "文件名.dll",
        "md5": "e10adc3949ba59abbe56e057f20f883e(小写文件md5可选,有就校验,没有就文件名校验)",
        "Lujing": "/(下载本地相对路径)",
        "size": "12345",
        "url": "https://www.baidu.com/文件名.dll(下载路径)",
        "YunXing": "0"
    }]
}`
	NewApp.AppGongGao = "我是一条公告"
	if NewApp.AppType == 2 || NewApp.AppType == 4 {
		//1=账号限时,2=账号计点,3卡号限时,4=卡号计点
		NewApp.VipData = `{
"VipData":"这里的数据,只有登录成功并且账号有点数才会传输出去的数据",
"VipData2":"这里的数据,只有登录成功并且账号有点数才会传输出去的数据"
}`
	} else {
		NewApp.VipData = `{
"VipData":"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据",
"VipData2":"这里的数据,只有登录成功并且账号会员不过期才会传输出去的数据"
}`
	}
	NewApp.CryptoType = 3                                            //默认Rsa交换Aes密匙
	NewApp.CryptoKeyAes = rand_string.RandStringBytesMaskImprSrc(24) //aes cbc 192长度固定24

	错误, 公钥base64, 私钥base64 := utils.GetRsaKey()
	if err != nil {
		global.GVA_LOG.Error("新建app创建Rsa密匙失败:" + 错误.Error())
	}
	NewApp.CryptoKeyPublic = 公钥base64
	NewApp.CryptoKeyPrivate = 私钥base64
	NewApp.MaxOnline = 1
	NewApp.MaxOnline = 1
	NewApp.ExceedMaxOnlineOut = 1 //超过在线最大数量处理方式 1踢掉最先登录的账号  2 提示登录数量超过限制
	NewApp.RmbToVipNumber = 1     //1 人民币换多少积分

	err = global.GVA_DB.Model(DB.DB_AppInfo{}).Create(&NewApp).Error
	if err != nil {
		return errors.New("添加失败")
	}
	//应用添加完毕 创建这个应用的用户表
	//
	err = global.GVA_DB.Set("gorm:table_options", "ENGINE=InnoDB").Table("db_AppUser_" + strconv.Itoa(NewApp.AppId)).AutoMigrate(&DB.DB_AppUser{})
	if err != nil {
		return errors.New("用户表创建失败,请删除该应用重新创建")
	}

	return nil
}
