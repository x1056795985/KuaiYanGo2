package Ser_KaClass

import (
	"EFunc/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_User"
	"server/global"
	"server/new/app/logic/common/agent"
	"server/new/app/logic/common/kaClassUpPrice"
	dbm "server/new/app/models/db"
	DB "server/structs/db"
	"strconv"
)

func KaClassId是否存在(id int) bool {
	var Count int64
	result := global.GVA_DB.Model(dbm.DB_KaClass{}).Select("1").Where("Id=?", id).First(&Count)
	return result.Error == nil

}

func KaName取map列表String(Appid int) map[string]string {

	var DB_KaClass []dbm.DB_KaClass
	var 总数 int64
	_ = global.GVA_DB.Model(dbm.DB_KaClass{}).Select("Id", "Name").Where("Appid=?", Appid).Count(&总数).Find(&DB_KaClass).Error
	var AppName = make(map[string]string, 总数)

	//吧 id 和 app名字 放入map
	for 索引 := range DB_KaClass {
		AppName[strconv.Itoa(int(DB_KaClass[索引].Id))] = DB_KaClass[索引].Name
	}
	return AppName
}

func KaName取map列表Int(Appid int) map[int]string {

	var DB_KaClass []dbm.DB_KaClass
	var 总数 int64
	_ = global.GVA_DB.Model(dbm.DB_KaClass{}).Select("Id", "Name").Where("Appid=?", Appid).Count(&总数).Find(&DB_KaClass).Error
	var AppName = make(map[int]string, 总数)

	//吧 id 和 app名字 放入map
	for 索引 := range DB_KaClass {
		AppName[int(DB_KaClass[索引].Id)] = DB_KaClass[索引].Name
	}
	return AppName
}
func KaClass取可购买卡类列表(Appid int) []dbm.DB_KaClass {

	var DB_KaClass []dbm.DB_KaClass
	_ = global.GVA_DB.Model(dbm.DB_KaClass{}).Where("Appid=?", Appid).Where("Money>0").Find(&DB_KaClass).Error
	return DB_KaClass
}

func KaClass取详细信息(id int) (dbm.DB_KaClass, error) {

	var KaClass详细信息 dbm.DB_KaClass

	err := global.GVA_DB.Model(dbm.DB_KaClass{}).Where("Id=?", id).First(&KaClass详细信息).Error

	return KaClass详细信息, err
}
func Id取详细信息_数组(id []int) ([]dbm.DB_KaClass, error) {

	var KaClass详细信息 = make([]dbm.DB_KaClass, 0, len(id))

	err := global.GVA_DB.Model(dbm.DB_KaClass{}).Where("Id IN ?", id).Find(&KaClass详细信息).Error

	return KaClass详细信息, err
}

func KaClass取map列表Int(AppId int) map[int]string {

	var DB_KaClass []dbm.DB_KaClass
	var 总数 int64
	_ = global.GVA_DB.Model(dbm.DB_KaClass{}).Select("Id", "Name").Where("AppId=?", AppId).Count(&总数).Find(&DB_KaClass).Error
	var AppName = make(map[int]string, 总数)

	//吧 id 和 app名字 放入map
	for 索引 := range DB_KaClass {
		AppName[DB_KaClass[索引].Id] = DB_KaClass[索引].Name
	}
	return AppName
}

func KaClass取map列表String(AppId int) map[string]string {

	var DB_KaClass []dbm.DB_KaClass
	var 总数 int64
	_ = global.GVA_DB.Model(dbm.DB_KaClass{}).Select("Id", "Name").Where("AppId=?", AppId).Count(&总数).Find(&DB_KaClass).Error
	var AppName = make(map[string]string, 总数)

	//吧 id 和 app名字 放入map
	for 索引 := range DB_KaClass {
		AppName[strconv.Itoa(DB_KaClass[索引].Id)] = DB_KaClass[索引].Name
	}
	return AppName
}

func KaClass创建New(AppId int, Name, 卡前缀 string, VipTime int64, 邀请人赠送 int64, 余额, 积分, Money, AgentMoney float64, UserClassId, NoUserClass, KaLength, KaStringType, Num, KaType, MaxOnline int) (新卡类id int, 错误信息 error) {

	请求 := dbm.DB_KaClass{
		Id:           0,
		AppId:        AppId,
		Name:         Name,
		Prefix:       卡前缀,
		VipTime:      VipTime,
		InviteCount:  邀请人赠送,
		RMb:          余额,
		VipNumber:    积分,
		Money:        Money,
		AgentMoney:   AgentMoney,
		UserClassId:  UserClassId,
		NoUserClass:  NoUserClass,
		KaLength:     KaLength,
		KaStringType: KaStringType,
		Num:          Num,
		KaType:       KaType,
		MaxOnline:    MaxOnline,
	}

	if 请求.Id > 0 {
		return 0, errors.New("添加用户不能有id值")
	}
	if 请求.AppId < 10000 {
		return 0, errors.New("AppId错误")
	}
	if 请求.Name == "" {
		return 0, errors.New("卡类名称不能为空")
	}

	if 请求.KaLength-len(请求.Prefix) < 10 {
		return 0, errors.New(`制卡可随机字符长度小于10,请增加卡长度或减少前缀长度`)
	}

	if 请求.VipNumber < 0 || 请求.VipTime < 0 || 请求.InviteCount < 0 || 请求.Num < 0 {
		return 0, errors.New(`时间点数积分次数值不能为为负数`)
	}

	if 请求.Money < -1 || 请求.AgentMoney < -1 {
		return 0, errors.New(`售价值不能为小于-1`)
	}
	//app_id 没有这个字段排除掉
	err := global.GVA_DB.Model(dbm.DB_KaClass{}).Create(&请求).Error
	if err != nil {
		return 0, errors.New(`添加失败` + err.Error())
	}
	return 请求.Id, nil
}
func Id取Name(卡类id int) (Name string) {
	_ = global.GVA_DB.Model(dbm.DB_KaClass{}).Select("Name").Where("Id=?", 卡类id).First(&Name).Error
	return Name
}

type K可制卡类授权树形框结构 struct {
	AppId    int    `json:"id"`    //应用AppID
	Label    string `json:"label"` //应用名称
	Children []struct {
		Id    int    `json:"id"`    //卡类Id
		Label string `json:"label"` //卡类名称
	} `json:"children"`
}

func Q取全部可制卡类树形框列表(c *gin.Context, 上级代理ID int) []K可制卡类授权树形框结构 {
	var DB_AppInfo []DB.DB_AppInfo
	_ = global.GVA_DB.Model(DB.DB_AppInfo{}).Select("AppId", "AppName").Find(&DB_AppInfo).Error

	var DB_KaClass []dbm.DB_KaClass
	if 上级代理ID < 0 {
		//如果小于0说明是开发者,或管理员,可以获取全部卡类   代理价格-1为禁止代理购买的卡号
		_ = global.GVA_DB.Model(dbm.DB_KaClass{}).Where("AgentMoney>0").Find(&DB_KaClass).Error
	} else {
		var 上级代理可制卡类ID []int
		上级代理可制卡类ID, _ = agent.L_agent.Id取代理可制卡类和可用代理功能列表(c, 上级代理ID)
		//只可以获取上级代理允许的ID
		_ = global.GVA_DB.Model(dbm.DB_KaClass{}).Where("Id IN ?", 上级代理可制卡类ID).Where("AgentMoney>0").Find(&DB_KaClass).Error
	}

	var 局_数据 []K可制卡类授权树形框结构
	局_临时上级代理id := Ser_User.Id取上级代理ID(上级代理ID)
	for _, app值 := range DB_AppInfo {
		var 局_临时数据 K可制卡类授权树形框结构
		局_临时数据.AppId = 0
		局_临时数据.Label = app值.AppName
		for _, 卡类值 := range DB_KaClass {
			if 卡类值.AppId == app值.AppId {
				if 局_临时上级代理id > 0 {
					局_临时双精度, _, err := kaClassUpPrice.L_kaClassUpPrice.J计算代理调价(c, 卡类值.Id, 局_临时上级代理id)
					if err == nil && 局_临时双精度 > 0 {
						卡类值.AgentMoney = utils.Float64加float64(卡类值.AgentMoney, 局_临时双精度, 2)
					}
				}

				局_临时数据.Children = append(局_临时数据.Children, struct {
					Id    int    `json:"id"`    //卡类Id
					Label string `json:"label"` //卡类名称
				}{
					Id:    卡类值.Id,
					Label: fmt.Sprintf("Id:%v %v(¥%v)", 卡类值.Id, 卡类值.Name, 卡类值.AgentMoney),
				})

			}
		}
		if len(局_临时数据.Children) > 0 { //只有卡类大于1的应用,才添加进去
			局_数据 = append(局_数据, 局_临时数据)
		}

	}

	return 局_数据
}
