package appInfo

import (
	"EFunc/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/songzhibin97/gkit/tools/rand_string"
	"gorm.io/gorm"
	"regexp"
	"server/app/global"
	"server/app/logic/common/cloudStorage"
	"server/app/logic/common/publicData"
	"server/app/models/db"
	dbm "server/app/models/db"
	"server/app/service"
	utils2 "server/app/utils"
	"strconv"
	"strings"
	"unicode/utf8"
)

var L_appInfo appInfo

func init() {
	L_appInfo = appInfo{}

}

type appInfo struct {
}

// NewApp信息(AppId, AppType int, AppName string)
func (j *appInfo) NewApp信息(c *gin.Context, AppId, AppType int, AppName string) (err error) {
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
	if AppType > 4 || AppType < 1 {
		return errors.New("应用类型错误")
	}
	var count int64
	service.NewAppInfo(c, global.GVA_DB)
	err = global.GVA_DB.Model(dbm.DB_AppInfo{}).Where("AppId = ?", AppId).Count(&count).Error
	// 没查到数据
	if count != 0 {
		return errors.New("AppId已存在")
	}

	var NewApp dbm.DB_AppInfo
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
	NewApp.VerifyKey = 1     //绑定模式
	NewApp.IsUserKeySame = 1 //不同用户可否相同
	NewApp.UpKeyData = 10    //修改绑定key增减值
	NewApp.UrlHome = "https://www.fnkuaiyan.com/"
	NewApp.UrlDownload = `{
 "htmlurl": "https://www.fnkuaiyan.cn",
 "data": [
  {
     "WenJianMin": "飞鸟快验{{AppVer}}.bin",
     "md5": "E655BDD4DF35C94AA2A706E2E55C4FF5",
     "Lujing": "/",
     "size": "",
     "url": "{{云存储_取外链('10001/飞鸟快验{{AppVer}}.bin',0)}}",
     "YunXing": "1"
   }
 ]
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
	NewApp.CryptoType = 3                              //默认Rsa交换Aes密匙
	NewApp.CryptoKeyAes = rand_string.RandomLetter(24) //aes cbc 192长度固定24

	错误, 公钥base64, 私钥base64 := utils2.GetRsaKey()
	if err != nil {
		err = errors.New("新建app创建Rsa密匙失败:" + 错误.Error())
	}
	NewApp.CryptoKeyPublic = 公钥base64
	NewApp.CryptoKeyPrivate = 私钥base64

	NewApp.ExceedMaxOnlineOut = 1 //超过在线最大数量处理方式 1踢掉最先登录的账号  2 提示登录数量超过限制

	// 使用事务处理数据库操作
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		// 创建卡类（使用事务的tx）
		局_注册送卡类 := dbm.DB_KaClass{
			AppId:        NewApp.AppId,
			Name:         "注册送卡",
			Prefix:       "ZC",
			Money:        -1,
			AgentMoney:   -1,
			NoUserClass:  1,
			KaLength:     25,
			KaStringType: 1,
			Num:          1,
			KaType:       1,
			MaxOnline:    0,
		}
		_, err = service.NewKaClass(c, tx).Create(&局_注册送卡类)
		if err != nil || 局_注册送卡类.Id == 0 {
			return fmt.Errorf("创建注册送卡类失败: %w", err)
		}
		NewApp.RegisterGiveKaClassId = 局_注册送卡类.Id
		// 创建应用记录
		if err = tx.Create(&NewApp).Error; err != nil {
			return fmt.Errorf("添加应用失败: %w", err)
		}

		// 创建用户表
		if err = tx.Set("gorm:table_options", "ENGINE=InnoDB").
			Table("db_AppUser_" + strconv.Itoa(NewApp.AppId)).
			AutoMigrate(&dbm.DB_AppUser{}); err != nil {
			return fmt.Errorf("用户表创建失败: %w", err)
		}

		// 创建唯一积分记录表
		if err = tx.Set("gorm:table_options", "ENGINE=InnoDB").
			Table(db.DB_UniqueNumLog{}.TableName() + "_" + strconv.Itoa(NewApp.AppId)).
			AutoMigrate(&db.DB_UniqueNumLog{}); err != nil {
			return fmt.Errorf("积分记录表创建失败: %w", err)
		}

		return nil
	})

	return nil

}

// App修改信息 修改应用信息(排除AppType,AppWeb,Sort字段,并删除缓存)
func (j *appInfo) App修改信息(c *gin.Context, AppInfo dbm.DB_AppInfo) error {
	//高频率读取数据 写入缓存

	//直接排除AppType  AppWeb 禁止修改
	var db = global.GVA_DB.Model(dbm.DB_AppInfo{}).Select(
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
		"Captcha",
		"RegisterGiveKa",
		"ApiHook",
		"FreeUpKeyTime",
		"FreeUpKeyInterval",
		"UpKeyTime",
		"UpKeyInterval",
		"AgentGiftKaClassId",
		"AgentKaUseModel",
	).Omit("AppType", "AppWeb", "Sort")

	err := db.Where("AppId= ?", AppInfo.AppId).Updates(AppInfo).Error
	if err == nil { //如果修改成功删除缓存
		global.H缓存.Delete("DB_AppInfo_" + strconv.Itoa(AppInfo.AppId)) //10分钟有效
	}
	return err
}

// CopyApp信息 复制应用信息(多表事务操作: AppInfo+KaClass+UserClass+AppUser表)
func (j *appInfo) CopyApp信息(c *gin.Context, AppId, AppType int, AppName string, CopyAppId int) error {
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
	err := global.GVA_DB.Model(dbm.DB_AppInfo{}).Where("AppId = ?", AppId).Count(&count).Error
	// 没查到数据
	if count != 0 {
		return errors.New("AppId已存在")
	}

	if AppType > 4 || AppType < 1 {
		return errors.New("应用类型错误")
	}

	var NewApp dbm.DB_AppInfo
	var 数组_卡类列表 []dbm.DB_KaClass
	var 数组_用户类型列表 []dbm.DB_UserClass
	err = global.GVA_DB.Model(dbm.DB_AppInfo{}).Where("AppId = ?", CopyAppId).First(&NewApp).Error
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
		global.GVA_LOG.Println("新建app创建Rsa密匙失败:" + err.Error())
	}
	NewApp.CryptoKeyPublic = 公钥base64
	NewApp.CryptoKeyPrivate = 私钥base64

	err = global.GVA_DB.Model(dbm.DB_KaClass{}).Where("AppId = ?", CopyAppId).Find(&数组_卡类列表).Error
	err = global.GVA_DB.Model(dbm.DB_UserClass{}).Where("AppId = ?", CopyAppId).Find(&数组_用户类型列表).Error
	//数据准备完毕,开启事务进行复制应用
	db := *global.GVA_DB
	err = db.Transaction(func(tx *gorm.DB) (err error) {
		for i1, v := range 数组_用户类型列表 {
			v.Id = 0
			v.AppId = AppId
			err = tx.Model(dbm.DB_UserClass{}).Create(&v).Error
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

		err = tx.Model(dbm.DB_AppInfo{}).Create(&NewApp).Error
		if err != nil {
			return errors.Join(err, errors.New("app复制失败"))
		}

		//应用添加完毕 创建这个应用的用户表
		err = tx.Set("gorm:table_options", "ENGINE=InnoDB").Table("db_AppUser_" + strconv.Itoa(NewApp.AppId)).AutoMigrate(&dbm.DB_AppUser{})
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

// App下载更新地址变量处理 处理下载地址中的变量替换(涉及正则匹配和云存储调用)
func (j *appInfo) App下载更新地址变量处理(c *gin.Context, DB_AppInfo dbm.DB_AppInfo) string {
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
