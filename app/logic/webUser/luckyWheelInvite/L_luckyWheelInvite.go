package luckyWheelInvite

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/app/global"
	"server/app/logic/webUser/user"
	"server/app/models/constant"
	dbm "server/app/models/db"
	"server/app/service"
	"time"
)

var L_luckyWheelInvite luckyWheelInvite

func init() {
	L_luckyWheelInvite = luckyWheelInvite{}
	// 注册拉新回调,与签到活动的拉新回调并列,互不影响
	user.L_user.Z邀请注册成功后处理(L_luckyWheelInvite.T邀请注册成功后处理)
}

type luckyWheelInvite struct {
}

// T邀请注册成功后处理 拉新注册成功后给邀请人增加抽奖次数
func (j *luckyWheelInvite) T邀请注册成功后处理(c *gin.Context, AppId, 邀请人, 被邀请人 int, Referer string) {
	var err error
	var info = struct {
		AppPromotionConfigs []dbm.DB_AppPromotionConfig
		AppPromotionConfig  dbm.DB_AppPromotionConfig
		LuckyWheelInfo      dbm.DB_LuckyWheelInfo
		luckyWheelUser      dbm.DB_LuckyWheelUser
	}{}

	db := *global.GVA_DB

	// 查询该App下进行中的大转盘活动
	info.AppPromotionConfigs, err = service.NewAppPromotionConfig(c, &db).Infos(
		map[string]interface{}{
			"appId":          AppId,
			"promotionType": constant.H活动类型_大转盘,
		})
	if err != nil && err.Error() != "record not found" {
		return
	}

	局_当前时间戳 := time.Now().Unix()
	for i := range info.AppPromotionConfigs {
		if info.AppPromotionConfigs[i].StartTime < 局_当前时间戳 && info.AppPromotionConfigs[i].EndTime > 局_当前时间戳 {
			info.AppPromotionConfig = info.AppPromotionConfigs[i]
			break
		}
	}
	if info.AppPromotionConfig.Id == 0 {
		// 没有开启的大转盘活动,不影响其他活动
		return
	}

	info.LuckyWheelInfo, err = service.NewLuckyWheelInfo(c, &db).Info(info.AppPromotionConfig.TypeAssociatedId)
	if err != nil {
		return
	}
	if info.LuckyWheelInfo.InviteGiveCount <= 0 {
		// 未开启拉新奖励
		return
	}

	// 查询或创建邀请人的用户记录
	info.luckyWheelUser, err = service.NewLuckyWheelUser(c, &db).Info(AppId, 邀请人)
	if err != nil && err.Error() != "record not found" {
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if info.luckyWheelUser.Id == 0 {
			// 创建用户记录
			info.luckyWheelUser = dbm.DB_LuckyWheelUser{
				AppId:            AppId,
				UserId:           邀请人,
				CreateTime:       局_当前时间戳,
				UpdateTime:       局_当前时间戳,
				RemainCount:      info.LuckyWheelInfo.InviteGiveCount,
				TotalInviteCount: info.LuckyWheelInfo.InviteGiveCount,
				DailyFreeDate:    time.Now().Format("20060102"),
			}
			_, e := service.NewLuckyWheelUser(c, tx).Create(&info.luckyWheelUser)
			return e
		}

		// 增加抽奖次数
		_, e := service.NewLuckyWheelUser(c, tx).UpdateMap([]int{info.luckyWheelUser.Id}, map[string]interface{}{
			"remainCount":       info.luckyWheelUser.RemainCount + info.LuckyWheelInfo.InviteGiveCount,
			"totalInviteCount":  info.luckyWheelUser.TotalInviteCount + info.LuckyWheelInfo.InviteGiveCount,
			"updateTime":        局_当前时间戳,
		})
		return e
	})

	if err != nil {
		global.GVA_LOG.Println("大转盘拉新奖励发放触发异常", err)
	}
	return
}
