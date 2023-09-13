package UserApi

import (
	"github.com/gin-gonic/gin"
	DB "server/structs/db"
	"server/utils"
	"strconv"
	"strings"
)

type 版本号格式 struct {
	大版本号   int
	小版本号   int
	编译版本号 int
}

func 版本号_检测可用(当前版本号, 可用版本号 string) bool {

	var 当前版本号数组 []string = utils.W文本_分割文本(当前版本号, ".")
	var 可用版本号数组 []string = utils.W文本_分割文本(可用版本号, "\n")

	for _, 值 := range 可用版本号数组 {
		局_分解版本号 := utils.W文本_分割文本(值, ".")
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
	局_分解版本号 := utils.W文本_分割文本(文本, ".")

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

func 用户数据信息还原(c *gin.Context, AppInfo *DB.DB_AppInfo, 在线信息 *DB.DB_LinksToken) {
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
