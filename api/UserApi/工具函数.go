package UserApi

import (
	. "EFunc/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_LinkUser"
	"server/global"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
	"strings"
	"time"
)

type 版本号格式 struct {
	大版本号  int
	小版本号  int
	编译版本号 int
}

func 版本号_检测可用(当前版本号, 可用版本号 string) bool {

	var 当前版本号数组 []string = W文本_分割文本(当前版本号, ".")
	var 可用版本号数组 []string = W文本_分割文本(可用版本号, "\n")

	for _, 值 := range 可用版本号数组 {
		局_分解版本号 := W文本_分割文本(值, ".")
		if len(局_分解版本号) != len(当前版本号数组) {
			//版本号不同,直接跳过 肯定不匹配
			continue
		}
		if len(当前版本号数组) == 1 && 版本号_单个匹配(当前版本号数组[0], 局_分解版本号[0]) {
			return true
		}
		if len(当前版本号数组) == 2 && 版本号_单个匹配(当前版本号数组[0], 局_分解版本号[0]) && 版本号_单个匹配(当前版本号数组[1], 局_分解版本号[1]) {
			return true
		}
		if len(当前版本号数组) == 3 && 版本号_单个匹配(当前版本号数组[0], 局_分解版本号[0]) && 版本号_单个匹配(当前版本号数组[1], 局_分解版本号[1]) && 版本号_单个匹配(当前版本号数组[2], 局_分解版本号[2]) {
			return true
		}
	}
	return false
}

func 版本号_单个匹配(当前, 匹配文本 string) bool {
	//"13","13"
	//"13","1"
	//"13","1*"
	//"13","*"
	当前数组 := strings.Split(当前, " ")
	匹配文本数组 := strings.Split(匹配文本, " ")
	for 索引, 值 := range 匹配文本数组 {
		if 当前数组[索引] != 值 && 值 != "*" {
			return false
		}
	}
	return true
}

func 版本号_检测更新(当前版本号, 最新版本号 string, 检测编译 bool) bool {

	局_当前版本号 := 版本号_分解(当前版本号)
	局_最新版本号 := 版本号_分解(最新版本号)
	if 局_当前版本号.大版本号 < 局_最新版本号.大版本号 {
		return true
	}
	if 局_当前版本号.小版本号 < 局_最新版本号.小版本号 {
		return true
	}
	if 检测编译 && 局_当前版本号.编译版本号 < 局_最新版本号.编译版本号 {
		return true
	}
	return false
}
func 版本号_分解(文本 string) (版本号 版本号格式) {
	局_分解版本号 := W文本_分割文本(文本, ".")

	for 索引, 值 := range 局_分解版本号 {
		switch 索引 {
		case 0:
			版本号.大版本号, _ = strconv.Atoi(值)
		case 1:
			版本号.小版本号, _ = strconv.Atoi(值)
		case 2:
			版本号.编译版本号, _ = strconv.Atoi(值)
		}
	}

	return 版本号
}

func Y用户数据信息还原(c *gin.Context, AppInfo *DB.DB_AppInfo, 在线信息 *DB.DB_LinksToken) {
	局_临时通用, _ := c.Get("AppInfo")
	*AppInfo = 局_临时通用.(DB.DB_AppInfo)
	局_临时通用, _ = c.Get("局_在线信息")
	*在线信息 = 局_临时通用.(DB.DB_LinksToken)
	return
}
func 检测用户登录在线正常(在线信息 *DB.DB_LinksToken) bool {
	if 在线信息.Uid > 0 && 在线信息.Status == 1 {
		return true
	}
	return false
}

func 更新上下文缓存在线信息(c *gin.Context) bool {
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	Y用户数据信息还原(c, &AppInfo, &局_在线信息)
	局_在线信息新, err := Ser_LinkUser.Token取User在线详情(局_在线信息.Token)
	if err == nil {
		c.Set("局_在线信息", 局_在线信息新) //修改在线信息缓存,因为hook里可能用到
	}
	return false
}

// 绑定信息更换规则校验 - 校验用户是否可以进行换绑操作
func 绑定信息更换规则校验(c *gin.Context, AppInfo DB.DB_AppInfo, Uid int) (err error, 扣费 int) {
	var info struct {
		logKey []dbm.DB_LogKey
	}
	db := *global.GVA_DB

	info.logKey, err = service.NewLogKey(c, &db).Infos(map[string]interface{}{
		"uid":   Uid,
		"appId": AppInfo.AppId,
	})

	if err != nil {
		return err, 0
	}

	// 获取当前时间戳
	局_现行时间戳 := time.Now().Unix()

	// 计算免费换绑限制
	免费时间内换绑次数 := 0
	for _, key := range info.logKey {
		if key.Type == constant.LogKey_绑定 { //绑定不算,只算解绑和换绑
			continue
		}
		if 局_现行时间戳-key.Time <= AppInfo.FreeUpKeyTime {
			免费时间内换绑次数++
		}
	}

	// 如果未超出免费换绑次数，则返回0表示无需扣除
	if 免费时间内换绑次数 < AppInfo.FreeUpKeyInterval {
		扣费 = 0
	} else {
		扣费 = AppInfo.UpKeyData
	}

	// 计算总换绑次数（包括付费）
	totalCount := 0
	for _, key := range info.logKey {
		if key.Type == constant.LogKey_绑定 { //绑定不算,只算解绑和换绑
			continue
		}
		if 局_现行时间戳-key.Time <= AppInfo.UpKeyTime {
			totalCount++
		}
	}

	// 如果超出总换绑次数，则不允许换绑
	if totalCount >= AppInfo.UpKeyInterval {
		//提示的更详细一些
		return errors.New(S时间_秒转时间文本(AppInfo.UpKeyTime) + "内最多换绑" + D到文本(AppInfo.UpKeyInterval) + "次"), 0
	}

	// 返回需要扣除的数据量
	return nil, 扣费
}
