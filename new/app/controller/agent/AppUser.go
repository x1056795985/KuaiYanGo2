package controller

import (
	. "EFunc/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/Service/Ser_Admin"
	"server/Service/Ser_AppInfo"
	"server/Service/Ser_AppUser"
	"server/Service/Ser_Log"
	"server/Service/Ser_UserClass"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/logic/agent/L_appUser"
	"server/new/app/logic/common/agent"
	"server/new/app/models/constant"
	"server/new/app/models/request"
	. "server/new/app/models/response"
	"server/new/app/service"
	"server/structs/Http/response"
	DB "server/structs/db"
	"strconv"
	"time"
)

type AppUser struct {
	Common.Common
}

func NewAppUserController() *AppUser {
	return &AppUser{}
}

// GetAppUserInfo
func (C *AppUser) GetAppUserInfo(c *gin.Context) {
	var 请求 struct {
		Id    int `json:"Id"`
		AppId int `json:"AppId" binding:"required,min=10000"`
	}
	//{"Id":2}
	if !C.ToJSON(c, &请求) {
		return
	}
	var err error
	var DB_AppUser struct {
		DB.DB_AppUser
		AppType int `json:"AppType"` //登录平台App名字
	}
	tx := *global.GVA_DB
	err = tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(请求.AppId)).Omit("app_type").Where("id = ?", 请求.Id).Where("AgentUid = ?", c.GetInt("Uid")).Find(&DB_AppUser).Error
	// 没查到数据

	if err != nil {
		response.FailWithMessage("查询软件用户详细信息失败", c)
		return
	}
	var app信息 DB.DB_AppInfo
	app信息, _ = service.NewAppInfo(c, &tx).Info(请求.AppId)
	DB_AppUser.AppType = app信息.AppType

	response.OkWithDetailed(DB_AppUser, "获取成功", c)
	return
}

// GetList
// 获取用户信息列表
func (C *AppUser) GetList(c *gin.Context) {

	var 请求 struct {
		request.List
		Status        int `json:"Status"` // 1本软件正常,2本软件冻结
		AppId         int `json:"AppId" binding:"required,min=10000"`
		Sortable      int `json:"Sortable"`      //排序字段名id  0id 1=到期时间
		IsLogin       int `json:"IsLogin"`       //1 在线 2不在线
		VipTimeStatus int `json:"VipTimeStatus"` //vip剩余时间状态
		UserClassId   int `json:"UserClassId"`   //用户类型Id
		AgentUid      int `json:"AgentUid"`      //归属代理id
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	var err error
	var info struct {
		AppInfo   DB.DB_AppInfo
		AgentInfo DB.DB_User
	}

	tx := *global.GVA_DB
	info.AppInfo, err = service.NewAppInfo(c, &tx).Info(请求.AppId)
	if err != nil {
		response.FailWithMessage("读取应用详情失败:"+err.Error(), c)
		return
	}

	var DB_AppUser []DB_AppUser带User信息
	var 总数 int64
	var 表名_AppUser = "db_AppUser_" + strconv.Itoa(请求.AppId)

	局_DB := tx.Table(表名_AppUser).Where("AgentUid = ?", c.GetInt("Uid"))
	if info.AppInfo.AppType == 3 || info.AppInfo.AppType == 4 {
		局_DB = 局_DB.Select(表名_AppUser+".*", "db_Ka.Name", "(select count(db_links_Token.id)  FROM db_links_Token WHERE  "+表名_AppUser+".Uid=db_links_Token.Uid AND db_links_Token.Status=1 AND LoginAppid="+strconv.Itoa(请求.AppId)+" )as LinksCount").Joins("left join db_Ka on " + 表名_AppUser + ".Uid=db_Ka.Id")
	} else {
		//mark 现在只是 链接 user表,后期需要处理 链接卡号表 读取用户名
		局_DB = 局_DB.Select(表名_AppUser+".*", "db_User.User", "(select count(db_links_Token.id)  FROM db_links_Token WHERE  "+表名_AppUser+".Uid=db_links_Token.Uid AND db_links_Token.Status=1 AND LoginAppid="+strconv.Itoa(请求.AppId)+" )as LinksCount").Joins("left join db_User on " + 表名_AppUser + ".Uid=db_User.Id")
	}

	var 排序字段名 = "Id"
	switch 请求.Sortable {
	default:

	case 1: // VipTime
		排序字段名 = "VipTime"
	}

	if 请求.Order == 1 {
		局_DB.Order(表名_AppUser + "." + 排序字段名 + " ASC")
	} else {
		局_DB.Order(表名_AppUser + "." + 排序字段名 + " DESC")
	}

	switch 请求.VipTimeStatus {
	case 1:
		if info.AppInfo.AppType == 2 || info.AppInfo.AppType == 4 {
			局_DB.Where(表名_AppUser+".VipTime > ?", 0)
		} else {
			局_DB.Where(表名_AppUser+".VipTime > ?", time.Now().Unix())
		}
	case 2:
		if info.AppInfo.AppType == 2 || info.AppInfo.AppType == 4 {
			局_DB.Where(表名_AppUser+".VipTime <= ?", 0)
		} else {
			局_DB.Where(表名_AppUser+".VipTime <= ?", time.Now().Unix())
		}
	case 3: //计时模式 1日内到期
		if info.AppInfo.AppType == 1 || info.AppInfo.AppType == 3 {
			局_DB.Where(表名_AppUser+".VipTime > ?", time.Now().Unix())
			局_DB.Where(表名_AppUser+".VipTime <= ?", time.Now().Unix()+86400)
		}
	case 4: //计时模式 账号模式 3日内到期
		if info.AppInfo.AppType == 1 || info.AppInfo.AppType == 3 {
			局_DB.Where(表名_AppUser+".VipTime > ?", time.Now().Unix())
			局_DB.Where(表名_AppUser+".VipTime <= ?", time.Now().Unix()+(86400*3))
		}
	case 5: //计时模式 账号模式 7日内到期
		if info.AppInfo.AppType == 1 || info.AppInfo.AppType == 3 {
			局_DB.Where(表名_AppUser+".VipTime > ?", time.Now().Unix())
			局_DB.Where(表名_AppUser+".VipTime <= ?", time.Now().Unix()+(86400*7))
		}
	case 6: //计时模式 账号模式 30日内到期
		if info.AppInfo.AppType == 1 || info.AppInfo.AppType == 3 {
			局_DB.Where(表名_AppUser+".VipTime > ?", time.Now().Unix())
			局_DB.Where(表名_AppUser+".VipTime <= ?", time.Now().Unix()+(86400*30))
		}
	}

	if 请求.Keywords != "" {
		switch 请求.Type {
		case 1: //id
			局_DB.Where(表名_AppUser+".Id = ?", 请求.Keywords)
		case 2: //用户id
			局_DB.Where(表名_AppUser+".Uid = ?", 请求.Keywords)
		case 3: //用户名 '支持,号分割
			局_用户名数组 := Z正则_取全部匹配子文本(请求.Keywords, "([A-Za-z0-9]+)")
			if info.AppInfo.AppType == 3 || info.AppInfo.AppType == 4 {
				if len(局_用户名数组) == 1 {
					局_DB.Where(表名_AppUser+".Uid In ?", gorm.Expr("(Select Id from db_Ka where db_Ka.Name like ? )", "%"+请求.Keywords+"%"))
				} else {
					局_DB.Where(表名_AppUser+".Uid In ?", gorm.Expr("(Select Id from db_Ka where db_Ka.Name IN ? )", 局_用户名数组))
				}
			} else {
				if len(局_用户名数组) == 1 {
					局_DB.Where(表名_AppUser+".Uid In ?", gorm.Expr("(Select Id from db_User where db_User.User  LIKE ? )", "%"+请求.Keywords+"%"))
				} else {
					局_DB.Where(表名_AppUser+".Uid In ?", gorm.Expr("(Select Id from db_User where db_User.User IN ? )", 局_用户名数组))
				}

			}
		case 4: //绑定信息
			局_DB.Where("`Key` like ?", "%"+请求.Keywords+"%")
		case 5: //软件用户备注
			局_DB.Where("LOCATE( ?, "+表名_AppUser+".Note)>0 ", 请求.Keywords)
		case 6: //归属代理id
			info.AgentInfo, err = service.NewUser(c, &tx).Info2(map[string]interface{}{"User": 请求.Keywords})
			if err != nil {
				局_代理id, _ := strconv.Atoi(请求.Keywords)
				info.AgentInfo, err = service.NewUser(c, &tx).Info(局_代理id)
			}

			if info.AgentInfo.Id == 0 {
				response.FailWithMessage("代理用户不存在", c)
				return
			}
			局_DB.Where("AgentUid = ?", info.AgentInfo.Id)
		case 7: //归属代理id和子代理
			局_代理id含子级id := []int{}
			info.AgentInfo, err = service.NewUser(c, &tx).Info2(map[string]interface{}{"User": 请求.Keywords})
			if err != nil {
				局_代理id, _ := strconv.Atoi(请求.Keywords)
				info.AgentInfo, err = service.NewUser(c, &tx).Info(局_代理id)
			}

			if info.AgentInfo.Id == 0 {
				response.FailWithMessage("代理用户不存在", c)
				return
			}
			局_代理id含子级id = agent.L_agent.Q取下级代理数组含子级(c, []int{info.AgentInfo.Id})
			局_代理id含子级id = append(局_代理id含子级id, info.AgentInfo.Id)
			if len(局_代理id含子级id) == 0 {
				response.FailWithMessage("代理用户不存在", c)
				return
			}
			局_DB.Where("AgentUid IN ?", 局_代理id含子级id)
		}

	}

	switch 请求.IsLogin {
	case 1: //在线
		局_DB.Where("(select count(db_links_Token.id)  FROM db_links_Token WHERE  " + 表名_AppUser + ".Uid=db_links_Token.Uid AND db_links_Token.Status=1 AND LoginAppid=" + strconv.Itoa(请求.AppId) + " )>0 ")
	case 2: //不在线
		局_DB.Where("(select count(db_links_Token.id)  FROM db_links_Token WHERE  " + 表名_AppUser + ".Uid=db_links_Token.Uid AND db_links_Token.Status=1 AND LoginAppid=" + strconv.Itoa(请求.AppId) + " )=0 ")
	}
	if 请求.Status > 0 {
		局_DB.Where(表名_AppUser+".Status = ? ", 请求.Status)
	}
	if 请求.UserClassId >= 0 {
		局_DB.Where(表名_AppUser+".UserClassId = ? ", 请求.UserClassId)
	}

	//Count(&总数) 必须放在where 后面 不然值会被清0
	err = 局_DB.Count(&总数).Limit(请求.Size).Offset((请求.Page - 1) * 请求.Size).Scan(&DB_AppUser).Error
	//fmt.Println("用户总数%d", 总数, DB_LinksToken)
	if err != nil {
		response.FailWithMessage("查询失败,参数异常"+err.Error(), c)
		global.GVA_LOG.Error("GetAppUserList:" + err.Error())
		return
	}
	UserClass := Ser_UserClass.UserClass取map列表Int(请求.AppId)

	response.OkWithDetailed(结构响应_GetAppUserList{
		GetList:   GetList{DB_AppUser, 总数},
		AppType:   info.AppInfo.AppType,
		UserClass: UserClass,
	}, "获取成功", c)
	return
}

type 结构响应_GetAppUserList struct {
	GetList
	AppType   int            `json:"AppType"`   // //1=账号限时,2=账号计点,3卡号限时,4=卡号计点
	UserClass map[int]string `json:"UserClass"` //
}

type DB_AppUser带User信息 struct {
	DB.DB_AppUser
	User       string `json:"User" gorm:"column:User;index;comment:用户登录名"`                 // 用户登录名
	Name       string `json:"Name" gorm:"column:Name;index;comment:卡号"`                    // 用户登录名
	Status     int    `json:"Status" gorm:"column:Status;default:1;comment:用户是状态 1正常 2冻结"` // 1正常 2冻结
	LinksCount int    `json:"LinksCount" gorm:"column:LinksCount;index;comment:在线总数"`
}

// Del批量删除软件用户
func (C *AppUser) Del批量删除软件用户(c *gin.Context) {
	var 请求 struct {
		Id    []int `json:"Id"` //用户id数组
		AppId int   `json:"AppId" binding:"required,min=10000"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}

	tx := *global.GVA_DB
	var 软件用户Uid = service.NewUser(c, &tx).Id取Uid_批量(请求.AppId, 请求.Id)
	局_结果 := tx.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(请求.AppId)).Where("Id IN ? ", 请求.Id).Delete("")
	if 局_结果.Error != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	_ = tx.Model(DB.DB_UserConfig{}).Where("AppId = ? ", 请求.AppId).Where("Uid IN ? ", 软件用户Uid).Delete("").RowsAffected

	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(局_结果.RowsAffected, 10), c)
	return
}

// save 保存
func (C *AppUser) Save用户信息(c *gin.Context) {
	var 请求 struct {
		AppId  int    `json:"AppId" binging:"required,min=10000"`    // Appid 必填
		Id     int    `json:"Id" binging:"required,min=1"`           // Appid 必填
		Status int    `json:"Status" binging:"required,min=1,max=2"` //本应用用户状态 1正常 2冻结
		Key    string `json:"Key"`                                   // 绑定信息
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	var err error
	defer func() {
		if err != nil {
			response.FailWithMessage(err.Error(), c)
		}
	}()
	tx := *global.GVA_DB
	var info struct {
		局_旧用户信息 DB.DB_AppUser
		AppInfo DB.DB_AppInfo
	}
	info.局_旧用户信息, err = service.NewAppUser(c, &tx, 请求.AppId).Info(请求.Id)
	if err != nil {
		err = errors.New("用户Id不存在")
		return
	}
	info.AppInfo, err = service.NewAppInfo(c, &tx).Info(请求.AppId)
	if err != nil {
		err = errors.New("AppId不存在")
		return
	}

	if info.局_旧用户信息.AgentUid != c.GetInt("Uid") {
		err = errors.New("权限不足,只能修改自己的归属软件用户")
		return
	}

	if info.局_旧用户信息.Status != 请求.Status {
		if 请求.Status == 1 && !agent.L_agent.Id功能权限检测(c, c.GetInt("Uid"), DB.D代理功能_解冻软件用户) {
			err = errors.New("权限不足,请联系上级授权解冻软件用户")
			return
		}

		if 请求.Status == 2 && !agent.L_agent.Id功能权限检测(c, c.GetInt("Uid"), DB.D代理功能_冻结软件用户) {
			err = errors.New("权限不足,请联系上级授权冻结软件用户")
			return
		}
	}

	if info.局_旧用户信息.Key != 请求.Key {
		if !agent.L_agent.Id功能权限检测(c, c.GetInt("Uid"), DB.D代理功能_修改用户绑定) {
			response.FailWithMessage("权限不足,请联系上级授权修改用户绑定", c)
			return
		}
	}

	//卡号模式 软件用户和卡状态冻结解冻 关联,所以需要事务保证
	//开启事务执行
	err = tx.Transaction(func(tx3 *gorm.DB) error {
		_, err = service.NewAppUser(c, &tx, 请求.AppId).Update(请求.Id, map[string]interface{}{
			"Status": 请求.Status,
			"Key":    请求.Key,
		})
		if err != nil {
			return err
		}

		if info.AppInfo.AppType == 2 || info.AppInfo.AppType == 3 {
			_, err = service.NewKa(c, tx3).Update(请求.Id, map[string]interface{}{"Status": 请求.Status})
			if err != nil {
				return err //出错就返回并回滚
			}
		}
		return err //出错就返回并回滚
	})

	if err != nil {
		err = errors.New("保存失败")
		return
	}
	response.OkWithMessage("保存成功", c)

	//如果是冻结同时注销在线的uid
	if 请求.Status == 2 {
		_ = service.NewLinksToken(c, &tx).Set批量注销Uid数组([]int{info.局_旧用户信息.Uid}, 请求.AppId, constant.Z注销_管理员手动注销)
	}

	return
}

// New用户信息
func (C *AppUser) New用户信息(c *gin.Context) {
	var 请求 struct {
		AppId int `json:"AppId" binding:"required,min=10000"` // Appid 必填
		DB.DB_AppUser
	}
	if !C.ToJSON(c, &请求) {
		return
	}
	if 请求.Id > 0 {
		response.FailWithMessage("添加用户不能有id值", c)
		return
	}
	var err error
	defer func() {
		if err != nil {
			response.FailWithMessage(err.Error(), c)
		}
	}()
	var tx = *global.GVA_DB
	var info struct {
		AppInfo  DB.DB_AppInfo
		KaInfo   DB.DB_Ka
		UserInfo DB.DB_User
	}
	info.AppInfo, err = service.NewAppInfo(c, &tx).Info(请求.AppId)
	if err != nil {
		err = errors.New("AppId不存在")
		return
	}

	if info.AppInfo.AppType == 3 || info.AppInfo.AppType == 4 {
		info.KaInfo, err = service.NewKa(c, &tx).Info2(map[string]interface{}{"AppId": 请求.AppId, "Uid": 请求.Uid})
		if info.KaInfo.Id == 0 {
			err = errors.New(`卡号Uid不存在,
请先去[ 卡号列表 => 制新卡 ],
添加信息`)
			return
		}
	} else {
		info.UserInfo, err = service.NewUser(c, &tx).Info(请求.Uid)
		if info.UserInfo.Id == 0 {
			err = errors.New(`用户Uid不存在,
请先去[ 用户管理 => 用户账户 ],
添加该用户信息`)
			return
		}
	}

	_, err = service.NewAppUser(c, &tx, 请求.AppId).InfoUid(请求.Uid)
	if err == nil {
		err = errors.New("用户已存在")
		return
	}
	请求.RegisterTime = time.Now().Unix()
	//app_id 没有这个字段排除掉
	局_信息 := DB.DB_AppUser{
		Uid:          请求.Uid,
		Status:       请求.Status,
		Key:          请求.Key,
		VipTime:      请求.VipTime,
		VipNumber:    请求.VipNumber,
		Note:         请求.Note,
		MaxOnline:    请求.MaxOnline,
		UserClassId:  请求.UserClassId,
		RegisterTime: 请求.RegisterTime,
	}
	_, err = service.NewAppUser(c, &tx, 请求.AppId).Create(&局_信息)
	if err != nil {
		err = errors.Join(err, errors.New("添加失败"))
		return
	}
	response.OkWithMessage("添加成功", c)

	if 局_信息.VipNumber != 0 {
		go Ser_Log.Log_写积分点数时间日志(Ser_AppUser.Uid取User(请求.AppId, 请求.Uid), c.ClientIP(), fmt.Sprintf("管理员(%v),新增用户携带积分:%v", c.GetInt("Uid"), 局_信息.VipNumber), 局_信息.VipNumber, 请求.AppId, 1)
	}
	return
}

// 批量修改状态
func (C *AppUser) Set修改状态(c *gin.Context) {
	var 请求 struct {
		Id     []int `json:"Id"`
		AppId  int   `json:"AppId" binding:"required,min=1"`
		Status int   `json:"Status" binding:"required,min=1,max=2" zh:"状态"` //1 解冻 2冻结
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	if !agent.L_agent.Id功能权限检测(c, c.GetInt("Uid"), S三元(请求.Status == 1, DB.D代理功能_解冻软件用户, DB.D代理功能_冻结软件用户)) {
		response.FailWithMessage("权限不足,请联系上级授权", c)
		return
	}

	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}
	var err error
	defer func() {
		if err != nil {
			response.FailWithMessage(err.Error(), c)
		}
	}()
	err = L_appUser.L_appUser.Z置状态_同步卡号修改(c, 请求.AppId, 请求.Id, 请求.Status)
	if err != nil {
		return
	}

	//如果是冻结同时注销在线的uid
	if 请求.Status == 2 {
		局_uid数组 := make([]int, 0, len(请求.Id))
		tx := *global.GVA_DB
		for _, 值 := range 请求.Id {
			局_临时用户信息, _ := service.NewAppUser(c, &tx, 请求.AppId).Info(值)
			if 局_临时用户信息.Uid > 0 && 局_临时用户信息.AgentUid == c.GetInt("Uid") {
				局_uid数组 = append(局_uid数组, 局_临时用户信息.Uid)
			}
		}
		_ = service.NewLinksToken(c, &tx).Set批量注销Uid数组(局_uid数组, 请求.AppId, constant.Z注销_管理员手动注销)
	}

	response.OkWithMessage("修改成功", c)
	return
}

// 批量维护 增减时间点数
func (C *AppUser) Set批量维护_增减时间点数(c *gin.Context) {
	var 请求 struct {
		Id     []int `json:"Id" binding:"required,gt=0" zh:"Id数组"` //用户id数组
		AppId  int   `json:"AppId" binding:"required,min=1"`
		Status int   `json:"Status"`
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	if len(请求.Id) == 0 {
		response.FailWithMessage("Id数组为空", c)
		return
	}
	var err error
	defer func() {
		if err != nil {
			response.FailWithMessage(err.Error(), c)
		}
	}()
	var tx = *global.GVA_DB
	if 请求.Status > 0 {
		err = service.NewAppUser(c, &tx, 请求.AppId).Id点数增减_批量(请求.Id, int64(请求.Status), true)
	} else {
		err = service.NewAppUser(c, &tx, 请求.AppId).Id点数增减_批量(请求.Id, int64(-请求.Status), false)
	}

	if err != nil {
		return
	}

	response.OkWithMessage("修改成功", c)

	for _, 局_id := range 请求.Id {
		Ser_Log.Log_写积分点数时间日志(Ser_AppUser.Id取User(请求.AppId, 局_id), c.ClientIP(), "管理员"+Ser_Admin.Id取User(c.GetInt("Uid"))+"批量增减点数", float64(请求.Status), 请求.AppId, S三元(Ser_AppInfo.App是否为计点(请求.AppId), 2, 3))
	}
	return
}

// save 保存
func (C *AppUser) Set用户密码(c *gin.Context) {
	var 请求 struct {
		AppId       int    `json:"AppId" binging:"required,min=10000"` // Appid 必填
		Id          int    `json:"Id" binging:"required,min=1"`        // Appid 必填
		NewPassword string `json:"NewPassword"`                        // 新密码
	}
	if !C.ToJSON(c, &请求) {
		return
	}

	var err error
	defer func() {
		if err != nil {
			response.FailWithMessage(err.Error(), c)
		}
	}()
	tx := *global.GVA_DB
	var info struct {
		局_旧用户信息 DB.DB_AppUser
		AppInfo DB.DB_AppInfo
	}
	info.局_旧用户信息, err = service.NewAppUser(c, &tx, 请求.AppId).Info(请求.Id)
	if err != nil {
		err = errors.New("用户Id不存在")
		return
	}
	if info.局_旧用户信息.AgentUid != c.GetInt("Uid") {
		err = errors.New("权限不足,只能修改自己的归属软件用户")
		return
	}
	info.AppInfo, err = service.NewAppInfo(c, &tx).Info(请求.AppId)
	if err != nil {
		err = errors.New("AppId不存在")
		return
	}

	if info.AppInfo.AppType != 1 && info.AppInfo.AppType != 2 {
		err = errors.New("只有账号模式应用可修改用户密码")
		return
	}

	var msg = ""
	if !Z正则_校验密码(请求.NewPassword, &msg) {
		err = errors.New("密码" + msg)
		return
	}
	if _, err = service.NewUser(c, &tx).Update(info.局_旧用户信息.Uid, map[string]interface{}{"PassWord": J校验_取md5_文本(请求.NewPassword, false)}); err != nil {
		err = errors.Join(err, errors.New("修改失败"))
		return
	}
	response.OkWithMessage("修改成功", c)
	return
}
