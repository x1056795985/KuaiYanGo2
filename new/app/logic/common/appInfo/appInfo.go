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
	dbm "server/new/app/models/db"
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
	NewApp.MaxOnline = 1
	NewApp.ExceedMaxOnlineOut = 1 //超过在线最大数量处理方式 1踢掉最先登录的账号  2 提示登录数量超过限制
	NewApp.RmbToVipNumber = 1     //1 人民币换多少积分

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
