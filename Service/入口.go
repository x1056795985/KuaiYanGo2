package Service

import "server/Service/Admin"

type _Service struct {
	Admin.InitDBService // admin的服务都集中到这里
}

// 服务实例化 api内可以调用

var Admin服务 = new(_Service)
