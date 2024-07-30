package init

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/logic/common/mqttClient"
	"server/new/app/logic/common/setting"
)

func InitMqttClient() {

	if global.GVA_DB == nil {
		return
	}
	mqttConfig := setting.Q取MQTT配置()
	if mqttConfig.L连接状态 {
		_ = mqttClient.L_mqttClient.L连接(&gin.Context{}, mqttConfig.F服务器地址, mqttConfig.F服务器端口, mqttConfig.Y用户名, mqttConfig.M密码)
	}

}
