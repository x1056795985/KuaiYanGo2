package jsEngine

import (
	"strings"

	"github.com/google/uuid"

	"server/app/global"
	"server/app/logic/common/taskPool"
	"server/app/logic/webSocket"
	"server/app/models/constant"
	dbm "server/app/models/db"
	"server/app/service"
)

func 脚本引擎_任务池任务创建(online dbm.DB_LinksToken, taskTypeID int, taskData string) 脚本引擎_Api结果 {
	局_上下文 := 脚本引擎_后台上下文()
	局_任务类型, 局_错误 := service.NewTaskPoolType(局_上下文, global.GVA_DB).Task类型读取(taskTypeID)
	if 局_错误 != nil {
		return 脚本引擎_失败消息("任务类型ID不存在")
	}
	if 局_任务类型.Status != 1 {
		return 脚本引擎_失败消息("任务类型ID维护中")
	}
	局_应用信息, 局_错误 := service.NewAppInfo(局_上下文, global.GVA_DB).Info(online.LoginAppid)
	if 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	if 局_任务类型.HookSubmitDataStart != "" {
		taskData, _, 局_错误 = J脚本引擎_处理任务池Hook(局_上下文, &局_应用信息, &online, 局_任务类型.HookSubmitDataStart, taskData, 0)
		if 局_错误 != nil {
			return 脚本引擎_失败(局_错误)
		}
	}
	局_任务Id, 局_错误 := taskPool.L_taskPool.Task数据创建加入队列(局_上下文, 局_任务类型.Id, taskData, online.LoginAppid, online.Uid)
	if 局_错误 != nil {
		return 脚本引擎_失败消息("Task数据创建加入队列失败: " + 局_错误.Error())
	}
	if 局_任务类型.HookSubmitDataEnd != "" {
		if _, _, 局_错误 = J脚本引擎_处理任务池Hook(局_上下文, &局_应用信息, &online, 局_任务类型.HookSubmitDataEnd, taskData, 1); 局_错误 != nil {
			return 脚本引擎_失败(局_错误)
		}
	}
	return 脚本引擎_成功("", map[string]any{"TaskUuid": 局_任务Id})
}

func 脚本引擎_任务池任务查询(taskID string) 脚本引擎_Api结果 {
	if _, 局_错误 := uuid.Parse(taskID); 局_错误 != nil {
		return 脚本引擎_失败消息("任务Uuid错误")
	}
	局_任务, 局_错误 := service.NewTaskPoolData(脚本引擎_后台上下文(), global.GVA_DB).Task数据读取_单条(taskID)
	if 局_错误 != nil {
		return 脚本引擎_失败消息("任务Uuid错误")
	}
	return 脚本引擎_成功("", map[string]any{
		"Status": 局_任务.Status, "ReturnData": 局_任务.ReturnData,
		"TimeStart": 局_任务.TimeStart, "TimeEnd": 局_任务.TimeEnd,
	})
}

func 脚本引擎_任务池Uuid添加到队列(taskID string) 脚本引擎_Api结果 {
	if _, 局_错误 := uuid.Parse(taskID); 局_错误 != nil {
		return 脚本引擎_失败消息("任务Uuid错误")
	}
	if 局_错误 := taskPool.L_taskPool.Uuid_添加到队列(脚本引擎_后台上下文(), taskID); 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	return 脚本引擎_成功("成功", nil)
}

func 脚本引擎_任务池取队列长度() 脚本引擎_Api结果 {
	局_数据, 局_错误 := taskPool.L_taskPool.Task_取队列数量(脚本引擎_后台上下文())
	if 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	return 脚本引擎_成功("成功", 局_数据)
}

func 脚本引擎_WebSocket发送消息(id int, message string) 脚本引擎_Api结果 {
	if 局_错误 := webSocket.L_webSocket.F发送消息(id, []byte(message)); 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	return 脚本引擎_成功("ok", []string{})
}

func 脚本引擎_WebSocket批量发送消息(ids []int, message string) 脚本引擎_Api结果 {
	局_错误数组 := webSocket.L_webSocket.F发送消息_批量(ids, []byte(message))
	局_结果数组 := make([]string, len(局_错误数组))
	for 局_索引, 局_错误 := range 局_错误数组 {
		if 局_错误 == nil {
			局_结果数组[局_索引] = "成功"
		} else {
			局_结果数组[局_索引] = 局_错误.Error()
		}
	}
	return 脚本引擎_成功("ok", 局_结果数组)
}

func 脚本引擎_WebSocket筛选Id(appIDEx, uid int, tag string) 脚本引擎_Api结果 {
	局_查询 := global.GVA_DB.Model(dbm.DB_LinksToken{}).
		Select("Id").
		Where("LoginAppid = ?", constant.APPID_WebSocket).
		Where("Status = ?", 1)
	if uid != 0 {
		局_查询 = 局_查询.Where("Uid = ?", uid)
	}
	if appIDEx != 0 {
		局_查询 = 局_查询.Where("AppIdEx = ?", appIDEx)
	}
	if tag != "" {
		局_查询 = 局_查询.Where("Tab LIKE ?", "%"+脚本引擎_转义Like(tag)+"%")
	}
	var 局_Id数组 []int
	if 局_错误 := 局_查询.Find(&局_Id数组).Error; 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	return 脚本引擎_成功("ok", 局_Id数组)
}

func 脚本引擎_转义Like(value string) string {
	局_替换器 := strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_")
	return 局_替换器.Replace(value)
}
