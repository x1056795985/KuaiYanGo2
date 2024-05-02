package agent

import (
	. "EFunc/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"server/global"
	m "server/new/app/models/common"
	"server/new/app/service"
	DB "server/structs/db"
)

var L_agent agent

func init() {
	L_agent = agent{}

}

type agent struct {
}

// 四舍五入  索引越小,代理级别越靠下
func (j *agent) D代理分成计算(c *gin.Context, 代理id int, 局_总计金额 float64) (局_返回 []m.D代理分成结构, err error) {

	局_返回 = make([]m.D代理分成结构, 0, 3)
	//开始分利润
	var 下级信息 DB.DB_User
	db := *global.GVA_DB
	s_user := service.NewUser(c, &db)
	if 下级信息, err = s_user.Info(代理id); err != nil {
		return 局_返回, fmt.Errorf("代理id:%d,不存在", 代理id)
	}
	局_下级分成百分比 := 0
	for {
		局_临时 := m.D代理分成结构{}
		局_临时.Uid = 下级信息.Id
		局_临时.User = 下级信息.User
		局_临时.F分成百分比 = 下级信息.AgentDiscount
		局_临时.F分给下级百分比 = 局_下级分成百分比
		局_临时.S实际自身百分比 = 下级信息.AgentDiscount - 局_下级分成百分比
		if 局_临时.S实际自身百分比 == 0 {
			局_临时.S实际分成金额 = 0
		} else {
			局_百分比小数 := Float64除int64(D到数值(局_临时.S实际自身百分比), 100, 2) //转换成小数百分比
			局_临时.S实际分成金额 = Float64乘Float64(局_总计金额, 局_百分比小数)
		}

		局_返回 = append(局_返回, 局_临时) //加入到返回数组
		if 下级信息.UPAgentId <= 0 {
			//上级是管理员了 跳出循环
			break
		}

		局_下级分成百分比 = 局_临时.F分成百分比

		if 下级信息, err = s_user.Info(下级信息.UPAgentId); err != nil {
			//代理不存在代理被删了, 结束,返回
			break
		}
		//继续往上找代理
	}
	return 局_返回, nil
}

func (j *agent) Id功能权限检测(c *gin.Context, 代理ID, 权限代号 int) bool {
	var 临时 int
	db := *global.GVA_DB

	db.Model(DB.Db_Agent_卡类授权{}).Select("1").Where("KId=?", 权限代号).Where("Uid=?", 代理ID).Take(&临时)
	return 临时 > 0
}
