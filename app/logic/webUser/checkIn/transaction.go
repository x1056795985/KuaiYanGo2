package checkIn

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"server/app/logic/common/ka"
	dbm "server/app/models/db"
	"server/app/service"
)

// Q签到_执行 在事务中写入签到、积分变动和连续签到状态。
func Q签到_执行(c *gin.Context, 数据库 *gorm.DB, appId int, uid int, 签到用户 dbm.DB_CheckInUser, 连续签到天数 int, 增加签到分 int, 今天标记 string) (结果 dbm.DB_CheckInUser, err error) {
	err = 数据库.Transaction(func(tx *gorm.DB) error {
		if _, 局_错误 := service.NewCheckInLog(c, tx).Create(&dbm.DB_CheckInLog{
			AppId: appId, UserId: uid, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix(), Day: 今天标记,
		}); 局_错误 != nil {
			return 局_错误
		}
		if 局_错误 := tx.Model(dbm.DB_CheckInUser{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id = ?", 签到用户.Id).First(&结果).Error; 局_错误 != nil {
			return 局_错误
		}
		结果.ContinuousDay = 连续签到天数 + 1
		if _, 局_错误 := service.NewCheckInUser(c, tx).UpdateMap([]int{结果.Id}, map[string]interface{}{
			"checkInScore":  结果.CheckInScore + 增加签到分,
			"continuousDay": 结果.ContinuousDay,
		}); 局_错误 != nil {
			return 局_错误
		}
		_, 局_错误 := service.NewCheckInScoreLog(c, tx).Create(&dbm.DB_CheckInScoreLog{
			AppId: appId, UserId: uid, CreatedAt: time.Now().Unix(), Number: int64(增加签到分), Msg: "每日签到",
			NumberBefore: 结果.CheckInScore, NumberAfter: 结果.CheckInScore + 增加签到分,
		})
		return 局_错误
	})
	return
}

// D签到_兑换奖励 在事务中扣除签到分并执行卡类直充。
func D签到_兑换奖励(c *gin.Context, 数据库 *gorm.DB, appInfo dbm.DB_AppInfo, 在线信息 dbm.DB_LinksToken, 签到用户 dbm.DB_CheckInUser, 卡类 dbm.DB_KaClass, 扣除签到分 int) (余额日志 []dbm.DB_LogMoney, 积分日志 []dbm.DB_LogVipNumber, err error) {
	err = 数据库.Transaction(func(tx *gorm.DB) error {
		c.Set("tx", tx)
		defer delete(c.Keys, "tx")

		if 局_错误 := tx.Model(dbm.DB_CheckInUser{}).Clauses(clause.Locking{Strength: "UPDATE"}).Where("Id = ?", 签到用户.Id).First(&签到用户).Error; 局_错误 != nil {
			return 局_错误
		}
		if 签到用户.CheckInScore-扣除签到分 < 0 {
			return errors.New("用户签到分不足")
		}
		if _, 局_错误 := service.NewCheckInScoreLog(c, tx).Create(&dbm.DB_CheckInScoreLog{
			AppId: appInfo.AppId, UserId: 在线信息.Uid, CreatedAt: time.Now().Unix(), Number: int64(-扣除签到分),
			Msg: "兑换:" + 卡类.Name, NumberBefore: 签到用户.CheckInScore, NumberAfter: 签到用户.CheckInScore - 扣除签到分,
		}); 局_错误 != nil {
			return 局_错误
		}
		if _, 局_错误 := service.NewCheckInUser(c, tx).UpdateMap([]int{签到用户.Id}, map[string]interface{}{
			"checkInScore": 签到用户.CheckInScore - 扣除签到分,
		}); 局_错误 != nil {
			return 局_错误
		}
		if 局_错误 := ka.L_ka.K卡类直冲_事务(c, 卡类.Id, 在线信息.Uid); 局_错误 != nil {
			return 局_错误
		}
		if 局_临时数据, 局_存在 := c.Get("logMoney"); 局_存在 {
			局_日志 := 局_临时数据.(dbm.DB_LogMoney)
			局_日志.Note = "签到分兑换," + 局_日志.Note
			余额日志 = append(余额日志, 局_日志)
		}
		if 局_临时数据, 局_存在 := c.Get("logVipNumber"); 局_存在 {
			局_日志 := 局_临时数据.(dbm.DB_LogVipNumber)
			局_日志.Note = "签到分兑换," + 局_日志.Note
			积分日志 = append(积分日志, 局_日志)
		}
		return nil
	})
	return
}
