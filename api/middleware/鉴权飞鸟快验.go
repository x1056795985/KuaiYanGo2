package middleware

import (
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/valyala/fastjson"
	"server/global"
	"server/new/app/logic/common/setting"
	"server/utils"
	"strconv"
	"sync"
)

func IsToken飞鸟快验() gin.HandlerFunc {
	return func(c *gin.Context) {
		D读取缓存Token()
		c.Next()
	}
}

// 定义互斥锁对象
var F飞鸟快验_互斥锁 sync.Mutex

func D读取缓存Token() bool {
	if global.Q快验.J_Token != "" {
		return false
	}
	F飞鸟快验_互斥锁.Lock()              //上锁
	defer F飞鸟快验_互斥锁.Unlock()      //解锁
	if global.Q快验.J_Token != "" { //进入后重新检测,防止前一个已经获取,后进入的,继续获取
		return false
	}
	//想缓存token,但是因为AES通讯密钥是动态的,也需要缓存,否则签名错误,但是保存可能导致不安全情况,暂时取消计划
	//缓存token已实现,加密AES通讯密钥保存到本地方式解决安全问题,还是用户体验更重要
	var (
		//防搜索中文关键字破解,字节集形式存放
		快验Token   = []byte("快验Token")
		ailiyunid = []byte("ailiyunid")
	)

	局_临时文本 := global.GVA_Viper.GetString(string(快验Token))
	if 局_临时文本 != "" {
		global.Q快验.J_Token = 局_临时文本
		global.Q快验.J_CryptoKeyAes, _ = hex.DecodeString(global.GVA_Viper.GetString(string(ailiyunid)))
		global.Q快验.J_CryptoKeyAes = utils.Aes解密_cbc192字节集2(global.Q快验.J_CryptoKeyAes, []byte(global.Q快验.J_Token[:24]))

		var 局_软件用户信息 string
		if global.Q快验.Q取软件用户信息(&局_软件用户信息, global.X系统信息.B版本号当前) { //旧Token还有效就继续用
			局_应用用户信息json, _ := fastjson.Parse(局_软件用户信息)
			var (
				//防搜索中文关键字破解,字节集形式存放
				User          = []byte("User")
				VipTime       = []byte("VipTime")
				Key           = []byte("Key")
				RegisterTime  = []byte("RegisterTime")
				UserClassName = []byte("UserClassName")
				VipNumber     = []byte("VipNumber")
				LoginTime     = []byte("LoginTime")
			)
			global.X系统信息.H会员帐号 = string(局_应用用户信息json.GetStringBytes(string(User)))
			global.X系统信息.D到期时间 = 局_应用用户信息json.GetInt64(string(VipTime))
			global.X系统信息.B绑定信息 = string(局_应用用户信息json.GetStringBytes(string(Key)))
			global.X系统信息.Z注册时间 = 局_应用用户信息json.GetInt(string(RegisterTime))
			global.X系统信息.Y用户类型 = string(局_应用用户信息json.GetStringBytes(string(UserClassName)))
			global.X系统信息.J积分 = 局_应用用户信息json.GetFloat64(string(VipNumber))
			global.X系统信息.D登录时间 = 局_应用用户信息json.GetInt(string(LoginTime))
			return true
		} else {
			//,如果已经无效了或没有,重新获取
			global.Q快验.J_Token = ""
			global.Q快验.J_CryptoKeyAes = []byte{}
			global.GVA_Viper.Set(string(快验Token), "")
			global.GVA_Viper.Set(string(ailiyunid), "")
			_ = global.GVA_Viper.WriteConfig()
		}

	}

	if !global.Q快验.Q取Token() {
		//fmt.Printf("快验Token获取失败:%v", global.Q快验.Q取错误信息(nil))
	} else {
		//fmt.Printf("快验Token获取成功:%v", global.Q快验.J_Token)
		局_系统设置 := setting.Q系统设置()
		if 局_系统设置.X系统地址 == "" {
			局_系统设置.X系统地址 = "http://" + global.Q快验.Q取用户IP() + ":" + strconv.Itoa(global.GVA_CONFIG.Port)
			_ = setting.Z系统设置(&局_系统设置)

		}
		global.GVA_Viper.Set(string(快验Token), global.Q快验.J_Token)                                                                             //写到配置重启备用
		global.GVA_Viper.Set(string(ailiyunid), hex.EncodeToString(utils.Aes加密_cbc192_2(global.Q快验.J_CryptoKeyAes, global.Q快验.J_Token[:24]))) //写到配置重启备用
		_ = global.GVA_Viper.WriteConfig()
	}

	return false
}
