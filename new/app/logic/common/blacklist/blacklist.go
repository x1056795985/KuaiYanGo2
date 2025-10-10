package blacklist

import (
	"server/global"
	"server/new/app/logic/common/log"
	"server/new/app/service"
)

func Is黑名单(配置名 string, AppId int) bool {
	db := service.S_Blacklist{}
	tx := *global.GVA_DB
	infos, err := db.InfoItemKey(&tx, 配置名)
	if err != nil {
		log.L_log.S上报异常("黑名单查询报错:" + err.Error())
		return false //直接放行
	}
	for i, _ := range infos {
		if infos[i].AppId == AppId || infos[i].AppId == 1 {
			_, _ = db.CountAdd1(&tx, infos[i].Id)
			return true //如果有本应用id,或全局,直接返回真
		}
	}
	return false
}
