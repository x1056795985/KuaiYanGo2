package kuaiYan

import (
	"fmt"
	"runtime"

	"EFunc/utils"

	"server/app/global"
	"server/app/service"
	utils2 "server/app/utils"
)

var 集_心跳容错计数 int

// K快验_心跳 上报运行状态并在连续失败后清理失效授权状态。
func K快验_心跳() {
	if global.Q快验.J_Token == "" {
		return
	}
	var 局_响应信息 string
	var 局_当前状态 int
	if !global.Q快验.X心跳(&局_响应信息, &局_当前状态) {
		集_心跳容错计数++
		if 集_心跳容错计数 >= 3 {
			global.Q快验.J_Token = ""
			global.X系统信息.H会员帐号 = ""
			global.X系统信息.D到期时间 = 0
			global.X系统信息.Y用户类型代号 = 0
		}
		return
	}
	集_心跳容错计数 = 0

	局_设备信息, 局_错误 := 快验_取服务器信息()
	if 局_错误 != nil {
		return
	}
	局_动态标记 := fmt.Sprintf("%s %dH%.2fG %dG %d协程,用户数:%d,卡总数:%d,在线数:%d",
		utils.S三元(global.Q快验.J集_连接方式 == 0, "直连", "网关"),
		runtime.NumCPU(),
		utils.Float64除int64(utils.Int64到Float64(int64(局_设备信息.Ram.TotalMB)), 1024, 2),
		局_设备信息.Disk.TotalGB,
		runtime.NumGoroutine(),
		service.NewUser(nil, global.GVA_DB).Q取总数(),
		service.NewKa(nil, global.GVA_DB).Q取总数(),
		service.NewLinksToken(nil, global.GVA_DB).Get取在线总数(true, true),
	)
	if 局_设备信息.Os.GOOS != "linux" {
		局_动态标记 += " " + 局_设备信息.Os.GOOS
	}
	global.Q快验.Z置动态标记(局_动态标记)
}

func 快验_取服务器信息() (*utils2.Server, error) {
	局_服务器 := &utils2.Server{}
	局_服务器.Os = utils2.InitOS()
	var 局_错误 error
	if 局_服务器.Cpu, 局_错误 = utils2.InitCPU(); 局_错误 != nil {
		return 局_服务器, 局_错误
	}
	if 局_服务器.Ram, 局_错误 = utils2.InitRAM(); 局_错误 != nil {
		return 局_服务器, 局_错误
	}
	if 局_服务器.Disk, 局_错误 = utils2.InitDisk(); 局_错误 != nil {
		return 局_服务器, 局_错误
	}
	return 局_服务器, nil
}
