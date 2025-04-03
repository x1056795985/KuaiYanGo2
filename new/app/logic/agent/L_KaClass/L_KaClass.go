package L_KaClass

import (
	. "EFunc/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/logic/common/agent"
	"server/new/app/logic/common/kaClassUpPrice"
	dbm "server/new/app/models/db"
	"server/new/app/models/request"
	. "server/new/app/models/response"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
)

var L_KaClass = new(KaClass)

type KaClass struct {
}

// 四舍五入  索引越小,代理级别越靠下  代理专用
func (j *KaClass) GetList(c *gin.Context, 请求 request.List, AppId int) (总数 int64, 响应 []KaClassUp带调价信息, err error) {

	var info struct {
		AgentInfo    DB.DB_User
		已授权卡类Id      []int
		局_list卡类     []dbm.DB_KaClass
		map用户类型id_名称 map[int]string
	}

	info.已授权卡类Id, _ = agent.L_agent.Id取代理可制卡类和可用代理功能列表(c, c.GetInt("Uid"))
	if len(info.已授权卡类Id) == 0 {
		return
	}

	tx := *global.GVA_DB
	info.AgentInfo, err = service.NewUser(c, &tx).Info(c.GetInt("Uid"))

	总数, info.局_list卡类, err = service.NewKaClass(c, &tx).GetList(请求, AppId, info.已授权卡类Id)
	if err != nil {
		err = errors.Join(errors.New("卡类读取失败"), err)
		return
	}

	info.map用户类型id_名称 = make(map[int]string, len(info.局_list卡类))
	for _, v := range info.局_list卡类 {
		if _, ok := info.map用户类型id_名称[v.UserClassId]; !ok {
			if v.UserClassId == 0 {
				info.map用户类型id_名称[v.UserClassId] = "未分类"
				continue
			} else if 局_用户类型详情, err2 := service.NewUserClass(c, &tx).Info(v.UserClassId); err2 != nil {
				info.map用户类型id_名称[v.UserClassId] = "已删类型id" + strconv.Itoa(v.UserClassId)
			} else {
				info.map用户类型id_名称[v.UserClassId] = 局_用户类型详情.Name
			}

		}
	}
	//拼装数据
	响应 = make([]KaClassUp带调价信息, 0, len(info.局_list卡类))
	for _, v := range info.局_list卡类 {
		//这里如何高效的获取到代理成本价 和 局_自身调整价格   //不优化了,这个接口也没有高并发, 浪费我半天时间,
		//这里是计算上级的代理加价  自身加价不用计算,所以传上级代理id
		var 局_最终加价, 局_自身调整价格 float64
		局_最终加价, _, err = kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, v.Id, info.AgentInfo.UPAgentId)
		if err == nil && v.AgentMoney > 0 {
			v.Money = Float64加float64(v.Money, 局_最终加价, 2)
			v.AgentMoney = Float64加float64(v.AgentMoney, 局_最终加价, 2)
		}
		局_自身调整价格 = 0 // 默认值
		if 局_卡类自身调价信息, err2 := service.NewKaClassUpPrice(c, &tx).Info2(map[string]interface{}{"KaClassId": v.Id, "AgentId": info.AgentInfo.Id}); err2 == nil {
			局_自身调整价格 = 局_卡类自身调价信息.Markup
		}
		响应 = append(响应, KaClassUp带调价信息{
			DB_KaClass:    v,
			UserClassName: info.map用户类型id_名称[v.UserClassId],
			Markup:        局_自身调整价格, //自身设置的调整价格
		})
	}
	return
}
