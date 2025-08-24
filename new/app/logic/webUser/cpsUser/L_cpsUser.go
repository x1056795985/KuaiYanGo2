package cpsUser

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
)

var L_cpsUser cpsUser

func init() {
	L_cpsUser = cpsUser{}

}

type cpsUser struct {
}

func (j *cpsUser) Q取有效邀请数量(c *gin.Context, AppId, 邀请人id int) (有效数量 int) {
	var info struct {
		已邀请用户 []dbm.DB_CpsInvitingRelation
		所有Uid []int
	}
	var err error
	tx := *global.GVA_DB
	info.已邀请用户, err = service.NewCpsInvitingRelation(c, &tx).Q取所有被邀请人(AppId, 邀请人id, 0)
	if err != nil {
		return
	}
	info.所有Uid = make([]int, 0, len(info.已邀请用户))
	for _, v := range info.已邀请用户 {
		info.所有Uid = append(info.所有Uid, v.InviteeId)
	}
	info.所有Uid = S数组_去重复(info.所有Uid)
	//获取这些uid的分成订单数量
	err = tx.Model(&dbm.DB_CpsPayOrder{}).Select("uid").Where("appId = ? and uid in (?) and inviterId = ? and inviterStatus =?", AppId, info.所有Uid, 邀请人id, constant.D订单状态_成功).Scan(&info.所有Uid).Error
	//失败就是0
	info.所有Uid = S数组_去重复(info.所有Uid)
	有效数量 = len(info.所有Uid)
	return
}
