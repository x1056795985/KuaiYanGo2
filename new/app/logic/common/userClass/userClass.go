package userClass

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/global"
	DB "server/structs/db"
	"strconv"
)

var L_userClass userClass

func init() {
	L_userClass = userClass{}

}

type userClass struct {
}

func (j *userClass) UserClass取map列表String(Appid int) map[string]string {

	var DB_UserClass = []DB.DB_UserClass{}
	tx := *global.GVA_DB
	_ = tx.Model(DB.DB_UserClass{}).Select("Id", "Name").Where("Appid=?", Appid).Find(&DB_UserClass).Error
	var AppName = make(map[string]string, len(DB_UserClass))
	//吧 id 和 app名字 放入map
	for 索引 := range DB_UserClass {
		AppName[strconv.Itoa(int(DB_UserClass[索引].Id))] = DB_UserClass[索引].Name
	}
	return AppName
}

// 只计算,计点请传入点数, 计时请传入剩余时间(viptime-现行时间戳), 自动处理权重=0 也就是未分类
func (j *userClass) J计算权重值(c *gin.Context, 旧用户类型权重, 新用户类型权重, 剩余时间或点数 int64) (新剩余时间 int64, err error) {
	旧用户类型权重 = S三元(旧用户类型权重 == 0, 1, 旧用户类型权重)
	新用户类型权重 = S三元(旧用户类型权重 == 0, 1, 新用户类型权重)
	新剩余时间 = 剩余时间或点数 * 旧用户类型权重 / 新用户类型权重
	return
}
