package Ser_AppInfo

import (
	"EFunc/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/songzhibin97/gkit/tools/rand_string"
	"gorm.io/gorm"
	"regexp"
	"server/global"
	"server/new/app/logic/common/cloudStorage"

	"server/new/app/logic/common/publicData"
	dbm "server/new/app/models/db"
	DB "server/structs/db"
	utils2 "server/utils"
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
	//.Order("Sort DESC, AppId ASC")  排序没啥用, 后面也会被排序
	var AppName = make(map[string]string, 总数+2)
	AppName["1"] = "管理平台"
	AppName["2"] = "代理平台"
	AppName["3"] = "WebApi"

	//吧 id 和 app名字 放入map
	for 索引 := range DB_AppInfo {
		AppName[strconv.Itoa(DB_AppInfo[索引].AppId)] = DB_AppInfo[索引].AppName
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
	下载地址 = App下载更新地址变量处理(DB_AppInfo)
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
	db := *global.GVA_DB
	_ = db.Model(DB.DB_AppInfo{}).Select("AppType").Where("AppId=?", Appid).First(&AppType).Error
	if AppType == 3 || AppType == 4 {
		return true
	}
	return false
}

func App是否为计点(Appid int) bool {
	var AppType int = 0 //1=账号限时,2=账号计点,3卡号限时,4=卡号计点
	db := *global.GVA_DB
	_ = db.Model(DB.DB_AppInfo{}).Select("AppType").Where("AppId=?", Appid).First(&AppType).Error
	if AppType == 2 || AppType == 4 {
		return true
	}
	return false
}

func App存在数量(Appid int) int64 {
	var count int64 = 0
	db := *global.GVA_DB
	_ = db.Model(DB.DB_AppInfo{}).Where("AppId = ?", Appid).Count(&count).Error

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
		"FreeUpKeyTime",
		"FreeUpKeyInterval",
		"UpKeyTime",
		"UpKeyInterval",
	).Omit("AppType", "AppWeb", "Sort")

	err := db.Where("AppId= ?", AppInfo.AppId).Updates(AppInfo).Error
	if err == nil { //如果修改成功删除缓存
		global.H缓存.Delete("DB_AppInfo_" + strconv.Itoa(AppInfo.AppId)) //10分钟有效
	}
	return err
}

func CopyApp信息(AppId, AppType int, AppName string, CopyAppId int) error {
	if AppId <= 10000 {
		return errors.New("AppId请输>10000的整数")
	}
	if utf8.RuneCountInString(AppName) < 2 || utf8.RuneCountInString(AppName) > 18 {
		return errors.New("应用名称长度必须大于2小于18")
	}
	msg := ""
	if !utils.Z正则_校验代理用户名(AppName, &msg) {
		return errors.New("应用名称" + msg)
	}

	var count int64
	err := global.GVA_DB.Model(DB.DB_AppInfo{}).Where("AppId = ?", AppId).Count(&count).Error
	// 没查到数据
	if count != 0 {
		return errors.New("AppId已存在")
	}

	if AppType > 4 || AppType < 1 {
		return errors.New("应用类型错误")
	}

	var NewApp DB.DB_AppInfo
	var 数组_卡类列表 []dbm.DB_KaClass
	var 数组_用户类型列表 []DB.DB_UserClass
	err = global.GVA_DB.Model(DB.DB_AppInfo{}).Where("AppId = ?", CopyAppId).First(&NewApp).Error
	if err != nil {
		return errors.New("复制应用不存在")
	}
	NewApp.AppId = AppId
	NewApp.AppType = AppType
	NewApp.AppName = AppName
	NewApp.AppWeb = `/Api?AppId=` + strconv.Itoa(AppId)
	NewApp.CryptoKeyAes = rand_string.RandomLetter(24) //aes cbc 192长度固定24
	err, 公钥base64, 私钥base64 := utils2.GetRsaKey()
	if err != nil {
		global.GVA_LOG.Error("新建app创建Rsa密匙失败:" + err.Error())
	}
	NewApp.CryptoKeyPublic = 公钥base64
	NewApp.CryptoKeyPrivate = 私钥base64

	err = global.GVA_DB.Model(dbm.DB_KaClass{}).Where("AppId = ?", CopyAppId).Find(&数组_卡类列表).Error
	err = global.GVA_DB.Model(DB.DB_UserClass{}).Where("AppId = ?", CopyAppId).Find(&数组_用户类型列表).Error
	//数据准备完毕,开启事务进行复制应用
	db := *global.GVA_DB
	err = db.Transaction(func(tx *gorm.DB) (err error) {
		for i1, v := range 数组_用户类型列表 {
			v.Id = 0
			v.AppId = AppId
			err = tx.Model(DB.DB_UserClass{}).Create(&v).Error
			if err != nil {
				return errors.Join(err, errors.New("用户类型复制失败"))
			}
			for 索引, _ := range 数组_卡类列表 {
				if 数组_用户类型列表[i1].Id == 数组_卡类列表[索引].UserClassId { //如果是旧的用户id==卡类用户id就修改为当前用户类型id
					数组_卡类列表[索引].UserClassId = v.Id
				}
			}
		}

		局_注册送卡id := 0
		for 索引, v := range 数组_卡类列表 {
			v.Id = 0
			v.AppId = AppId
			err = tx.Model(dbm.DB_KaClass{}).Create(&v).Error
			if err != nil {
				return err
			}
			if 数组_卡类列表[索引].Id == NewApp.RegisterGiveKaClassId {
				局_注册送卡id = v.Id
			}
		}
		NewApp.RegisterGiveKaClassId = 局_注册送卡id //注册赠送卡类的id 要重新设置

		err = tx.Model(DB.DB_AppInfo{}).Create(&NewApp).Error
		if err != nil {
			return errors.Join(err, errors.New("app复制失败"))
		}

		//应用添加完毕 创建这个应用的用户表
		err = tx.Set("gorm:table_options", "ENGINE=InnoDB").Table("db_AppUser_" + strconv.Itoa(NewApp.AppId)).AutoMigrate(&DB.DB_AppUser{})
		if err != nil {
			return errors.Join(err, errors.New("用户表创建失败,请删除该应用重新创建"))
		}

		// 创建唯一积分记录表
		if err = tx.Set("gorm:table_options", "ENGINE=InnoDB").
			Table(dbm.DB_UniqueNumLog{}.TableName() + "_" + strconv.Itoa(NewApp.AppId)).
			AutoMigrate(&dbm.DB_UniqueNumLog{}); err != nil {
			return fmt.Errorf("积分记录表创建失败: %w", err)
		}
		err = publicData.L_publicData.F复制app专属变量(CopyAppId, NewApp.AppId)
		if err != nil {
			return fmt.Errorf("复制app专属变量失败: %w", err)
		}

		return
	})

	return err
}

func App下载更新地址变量处理(DB_AppInfo DB.DB_AppInfo) string {
	局_新文本 := DB_AppInfo.UrlDownload

	局_新文本 = strings.Replace(局_新文本, "{{AppName}}", DB_AppInfo.AppName, -1)

	if strings.Index(局_新文本, "{{AppVer}}") != -1 && DB_AppInfo.AppVer != "" {
		局_可用版本 := utils.W文本_分割文本(DB_AppInfo.AppVer, "\n")
		if len(局_可用版本) > 0 {
			局_新文本 = strings.Replace(局_新文本, "{{AppVer}}", 局_可用版本[0], -1)
		}
	}

	//{{(.*?)\((.*?)\)}}  正则匹配指令,  子匹配1为指令名 子匹配2为参数
	if strings.Index(局_新文本, "{{") != -1 { //判断是否还有变量
		re := regexp.MustCompile(`{{(.*?)\((.*?)\)}}`)
		result := re.FindAllStringSubmatch(局_新文本, -1)
		for i, _ := range result {
			局_完整文本 := result[i][0]
			局_指令名 := result[i][1]
			局_参数 := utils.W文本_分割文本(result[i][2], ",")
			switch 局_指令名 {
			case "云存储_取外链":
				if len(局_参数) == 2 {
					下载地址, err := cloudStorage.L_云存储.Q取外链地址(&gin.Context{}, strings.Trim(局_参数[0], "'"), gconv.Int64(局_参数[1]))
					if err == nil {
						局_新文本 = strings.Replace(局_新文本, 局_完整文本, 下载地址, -1)
					}
				}
			}
		}
	}

	return 局_新文本
}
