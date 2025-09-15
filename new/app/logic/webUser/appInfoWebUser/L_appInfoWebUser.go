package appInfoWebUser

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/logic/common/setting"
	"server/new/app/service"
)

var L_appInfoWebUser appInfoWebUser

func init() {
	L_appInfoWebUser = appInfoWebUser{}

}

type appInfoWebUser struct {
}

func (j *appInfoWebUser) Q用户中心域名(c *gin.Context, AppId int) (WebUserDomain string) {
	db := *global.GVA_DB
	局_WebUserInfo, err := service.NewAppInfoWebUser(c, &db).Info(AppId)
	if err != nil {
		return setting.Q系统设置().X系统地址
	}
	if 局_WebUserInfo.WebUserDomain == "" {
		局_WebUserInfo.WebUserDomain = setting.Q系统设置().X系统地址
	}
	return 局_WebUserInfo.WebUserDomain
}
