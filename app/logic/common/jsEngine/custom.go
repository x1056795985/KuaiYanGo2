package jsEngine

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"

	"server/app/global"
	"server/app/logic/common/appUser"
	"server/app/logic/common/ka"
	"server/app/logic/common/user"
	dbm "server/app/models/db"
	"server/app/service"
	"server/app/utils"
)

type 脚本引擎_批量用户结果 struct {
	User    dbm.DB_User
	AppUser dbm.DB_AppUser
	UseKa   string
	Name    string
	IsOk    bool
	Msg     string
}

func 脚本引擎_定制批量注册(online dbm.DB_LinksToken, usernames []string, password string) 脚本引擎_Api结果 {
	局_上下文 := 脚本引擎_后台上下文()
	局_应用信息, 局_错误 := 脚本引擎_取账号模式应用(局_上下文, online.LoginAppid)
	if 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	局_结果数组 := make([]脚本引擎_批量用户结果, 0, len(usernames))
	局_软件用户服务 := service.NewAppUser(局_上下文, global.GVA_DB, 局_应用信息.AppId)
	for _, 局_用户名 := range usernames {
		局_用户信息, 局_创建错误 := user.L_user.New用户信息(局_上下文, 局_用户名, password, password, "", "", "", "127.0.0.1", "批量注册", 0, 0, 0, "")
		if 局_创建错误 != nil {
			局_结果数组 = append(局_结果数组, 脚本引擎_批量用户结果{Name: 局_用户名, Msg: 局_创建错误.Error()})
			continue
		}
		局_创建错误 = appUser.L_appUser.New用户信息(局_上下文, 局_应用信息.AppId, 局_用户信息.Id, "", 1, time.Now().Unix(), 0, 0, "")
		if 局_创建错误 != nil {
			局_结果数组 = append(局_结果数组, 脚本引擎_批量用户结果{User: 局_用户信息, Name: 局_用户名, Msg: 局_创建错误.Error()})
			continue
		}
		if 局_应用信息.RegisterGiveKaClassId > 0 {
			_ = ka.L_ka.K卡类直冲_事务(局_上下文, 局_应用信息.RegisterGiveKaClassId, 局_用户信息.Id)
		}
		局_软件用户信息, 局_详情错误 := 局_软件用户服务.InfoUid(局_用户信息.Id)
		if 局_详情错误 != nil {
			局_结果数组 = append(局_结果数组, 脚本引擎_批量用户结果{User: 局_用户信息, Name: 局_用户名, Msg: "无软件用户信息: " + 局_详情错误.Error()})
			continue
		}
		局_结果数组 = append(局_结果数组, 脚本引擎_批量用户结果{User: 局_用户信息, AppUser: 局_软件用户信息, Name: 局_用户名, IsOk: true, Msg: "成功"})
	}
	return 脚本引擎_成功("成功", 局_结果数组)
}

func 脚本引擎_定制批量充值(online dbm.DB_LinksToken, usernames, cardNumbers []string) 脚本引擎_Api结果 {
	局_上下文 := 脚本引擎_后台上下文()
	局_应用信息, 局_错误 := 脚本引擎_取账号模式应用(局_上下文, online.LoginAppid)
	if 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	if len(usernames) != len(cardNumbers) {
		return 脚本引擎_失败消息("账号和卡号数量必须一致")
	}
	局_已检查卡号 := make(map[string]struct{}, len(cardNumbers))
	局_卡号服务 := service.NewKa(局_上下文, global.GVA_DB)
	for _, 局_卡号 := range cardNumbers {
		if _, 局_存在 := 局_已检查卡号[局_卡号]; 局_存在 {
			return 脚本引擎_失败消息("卡号有重复不可充值")
		}
		局_已检查卡号[局_卡号] = struct{}{}
		局_卡号信息, 局_详情错误 := 局_卡号服务.Ka卡号取详情(局_卡号)
		if 局_详情错误 != nil {
			return 脚本引擎_失败消息("卡号[" + 局_卡号 + "]不存在")
		}
		if 局_卡号信息.NumMax <= 局_卡号信息.Num {
			return 脚本引擎_失败消息("卡号[" + 局_卡号 + "]已耗尽使用次数")
		}
	}
	局_软件用户服务 := service.NewAppUser(局_上下文, global.GVA_DB, 局_应用信息.AppId)
	for _, 局_用户名 := range usernames {
		if 局_软件用户服务.User或卡号取Id(局_应用信息.AppId, 局_用户名) == 0 {
			return 脚本引擎_失败消息("账号[" + 局_用户名 + "]不存在")
		}
	}

	局_结果数组 := make([]脚本引擎_批量用户结果, 0, len(usernames))
	局_用户服务 := service.NewUser(局_上下文, global.GVA_DB)
	for 局_索引, 局_用户名 := range usernames {
		局_卡号 := cardNumbers[局_索引]
		if 局_充值错误 := ka.L_ka.K卡号充值_事务(局_上下文, 局_应用信息.AppId, 局_卡号, 局_用户名, ""); 局_充值错误 != nil {
			局_结果数组 = append(局_结果数组, 脚本引擎_批量用户结果{UseKa: 局_卡号, Name: 局_用户名, Msg: 局_充值错误.Error()})
			continue
		}
		局_用户信息, 局_用户错误 := 局_用户服务.InfoName(局_用户名)
		if 局_用户错误 != nil {
			局_结果数组 = append(局_结果数组, 脚本引擎_批量用户结果{UseKa: 局_卡号, Name: 局_用户名, Msg: 局_用户错误.Error()})
			continue
		}
		局_软件用户信息, 局_详情错误 := 局_软件用户服务.InfoUid(局_用户信息.Id)
		if 局_详情错误 != nil {
			局_结果数组 = append(局_结果数组, 脚本引擎_批量用户结果{User: 局_用户信息, UseKa: 局_卡号, Name: 局_用户名, Msg: "无软件用户信息"})
			continue
		}
		局_结果数组 = append(局_结果数组, 脚本引擎_批量用户结果{User: 局_用户信息, AppUser: 局_软件用户信息, UseKa: 局_卡号, Name: 局_用户名, IsOk: true, Msg: "成功"})
	}
	return 脚本引擎_成功("成功", 局_结果数组)
}

func 脚本引擎_定制批量取账号信息(online dbm.DB_LinksToken, usernames []string, password string) 脚本引擎_Api结果 {
	局_上下文 := 脚本引擎_后台上下文()
	局_应用信息, 局_错误 := 脚本引擎_取账号模式应用(局_上下文, online.LoginAppid)
	if 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	局_结果数组 := make([]脚本引擎_批量用户结果, 0, len(usernames))
	局_用户服务 := service.NewUser(局_上下文, global.GVA_DB)
	局_软件用户服务 := service.NewAppUser(局_上下文, global.GVA_DB, 局_应用信息.AppId)
	for _, 局_用户名 := range usernames {
		局_用户信息, 局_详情错误 := 局_用户服务.InfoName(局_用户名)
		if 局_详情错误 != nil {
			局_结果数组 = append(局_结果数组, 脚本引擎_批量用户结果{Name: 局_用户名, Msg: 局_详情错误.Error()})
			continue
		}
		if !utils.BcryptCheck(password, 局_用户信息.PassWord) {
			局_结果数组 = append(局_结果数组, 脚本引擎_批量用户结果{Name: 局_用户名, Msg: "密码错误"})
			continue
		}
		局_软件用户信息, 局_软件用户错误 := 局_软件用户服务.InfoUid(局_用户信息.Id)
		if 局_软件用户错误 != nil {
			局_结果数组 = append(局_结果数组, 脚本引擎_批量用户结果{User: 局_用户信息, Name: 局_用户名, Msg: "无软件用户信息"})
			continue
		}
		局_结果数组 = append(局_结果数组, 脚本引擎_批量用户结果{User: 局_用户信息, AppUser: 局_软件用户信息, Name: 局_用户名, IsOk: true, Msg: "成功"})
	}
	return 脚本引擎_成功("成功", 局_结果数组)
}

func 脚本引擎_取账号模式应用(c *gin.Context, appID int) (dbm.DB_AppInfo, error) {
	if appID <= 10000 {
		return dbm.DB_AppInfo{}, errors.New("AppId必须大于10000")
	}
	局_应用信息, 局_错误 := service.NewAppInfo(c, global.GVA_DB).Info(appID)
	if 局_错误 != nil {
		return dbm.DB_AppInfo{}, errors.New("无该应用信息")
	}
	if 局_应用信息.AppType == 3 || 局_应用信息.AppType == 4 {
		return dbm.DB_AppInfo{}, errors.New("仅限账号模式应用调用")
	}
	return 局_应用信息, nil
}
