package appInfo

import (
	"EFunc/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/songzhibin97/gkit/tools/rand_string"
	"gorm.io/gorm"
	"server/global"
	"server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	utils2 "server/utils"
	"strconv"
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
	err = global.GVA_DB.Model(DB.DB_AppInfo{}).Where("AppId = ?", AppId).Count(&count).Error
	// 没查到数据
	if count != 0 {
		return errors.New("AppId已存在")
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
	NewApp.CryptoType = 3                              //默认Rsa交换Aes密匙
	NewApp.CryptoKeyAes = rand_string.RandomLetter(24) //aes cbc 192长度固定24

	错误, 公钥base64, 私钥base64 := utils2.GetRsaKey()
	if err != nil {
		err = errors.New("新建app创建Rsa密匙失败:" + 错误.Error())
	}
	NewApp.CryptoKeyPublic = 公钥base64
	NewApp.CryptoKeyPrivate = 私钥base64
	NewApp.MaxOnline = 1
	NewApp.ExceedMaxOnlineOut = 1 //超过在线最大数量处理方式 1踢掉最先登录的账号  2 提示登录数量超过限制
	NewApp.RmbToVipNumber = 1     //1 人民币换多少积分

	// 使用事务处理数据库操作
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		// 创建卡类（使用事务的tx）
		局_注册送卡类 := DB.DB_KaClass{
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
			AutoMigrate(&DB.DB_AppUser{}); err != nil {
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
