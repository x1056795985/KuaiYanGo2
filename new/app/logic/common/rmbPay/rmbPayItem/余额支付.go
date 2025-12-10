package rmbPayItem

import (
	. "EFunc/utils"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"server/Service/Ser_Log"
	"server/Service/Ser_User"
	"server/global"
	"server/new/app/logic/common/rmbPay"
	m "server/new/app/models/common"
	"server/new/app/models/constant"
	"server/new/app/service"
	DB "server/structs/db"
)

func init() {
	rmbPay.L_rmbPay.Z注册接口(pay_余额支付)
}

var pay_余额支付 余额支付

type 余额支付 struct {
}

func (j 余额支付) Q取通道名称() string {
	return "余额支付"
}

func (j 余额支付) Q取订单id(c *gin.Context, 参数 *m.PayParams) string {

	return c.GetString("余额支付payOrder")

}
func (j 余额支付) D订单创建(c *gin.Context, 参数 *m.PayParams) (response m.Request, err error) {
	var 局_支付配置 m.Z在线支付_余额支付
	err = json.Unmarshal(参数.Z支付配置, &局_支付配置)

	if err != nil || !局_支付配置.Y余额支付开关 {
		err = errors.New(局_支付配置.Y余额支付显示名称 + "支付方式已关闭")
		return
	}

	if 参数.ProcessingType != constant.D订单类型_购卡直冲 && 参数.ProcessingType != constant.D订单类型_支付购卡 {
		err = errors.New("余额支付不支持该订单类型,请更换支付方式")
		return
	}
	var info = struct {
		likeInfo DB.DB_LinksToken
		appInfo  DB.DB_AppInfo
		userInfo DB.DB_User
	}{}
	局_临时通用, _ := c.Get("DB_LinksToken")
	info.likeInfo = 局_临时通用.(DB.DB_LinksToken)
	db := *global.GVA_DB
	if info.likeInfo.LoginAppid == constant.APPID_Web用户中心 {
		局_临时通用, err = service.NewAppInfo(c, &db).Info(D到整数(info.likeInfo.Tab))
		info.appInfo = 局_临时通用.(DB.DB_AppInfo)
	} else {
		局_临时通用, err = service.NewAppInfo(c, &db).Info(info.likeInfo.LoginAppid)
		info.appInfo = 局_临时通用.(DB.DB_AppInfo)
	}
	if err != nil {
		err = errors.New("读取应用信息错误")
		return
	}
	if info.appInfo.AppType > 2 {
		err = errors.New("余额支付仅限账号模式应用调用")
		return
	}
	info.userInfo, err = service.NewUser(c, &db).Info(info.likeInfo.Uid)
	if err != nil {
		err = errors.New("读取用户信息错误")
		return
	}
	if info.userInfo.Rmb < 参数.Rmb {
		err = errors.New("余额不足")
		return
	}
	info.userInfo.Rmb, err = Ser_User.Id余额增减(info.likeInfo.Uid, 参数.Rmb, false)
	if err != nil {
		err = errors.New("余额支付仅限账号模式应用调用")
		return
	}
	go Ser_Log.Log_写余额日志(info.likeInfo.User, info.likeInfo.Ip, "余额支付订单:"+参数.PayOrder+"|新余额≈"+Float64到文本(info.userInfo.Rmb, 2), 参数.Rmb)

	response = m.Request{
		Status:  1,
		PayURL:  "",
		OrderId: 参数.PayOrder,
	}
	return
}
func (j 余额支付) D订单退款(c *gin.Context, 参数 *m.PayParams) (err error) {

	return errors.New("暂不支持退款")
}
func (j 余额支付) D订单支付回调(c *gin.Context, 参数 *m.PayParams) (响应信息 string, 响应代码 int, err error) {
	defer func() {
		if err == nil {
			响应信息 = "success"
			响应代码 = http.StatusOK
		} else {
			响应信息 = "err"
			响应代码 = http.StatusInternalServerError
		}
	}()

	return
}
func (j 余额支付) D订单退款回调(c *gin.Context, 参数 *m.PayParams) (响应信息 string, 响应代码 int, err error) {
	return
}
