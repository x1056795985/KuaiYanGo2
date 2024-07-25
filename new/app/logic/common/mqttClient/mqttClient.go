package mqttClient

import (
	. "EFunc/utils"
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"log"
	"server/global"
)

var L_mqttClient mqttClient

func init() {
	L_mqttClient = mqttClient{}
	L_mqttClient.mQTTClient = mqtt.NewClient(mqtt.NewClientOptions().AddBroker("tcp://0.0.0.0:8080"))

}

type mqttClient struct {
	mQTTClient mqtt.Client
}

func (j *mqttClient) L连接(c *gin.Context, 地址 string, 端口 int, 用户名, 密码 string) (err error) {

	if j.mQTTClient.IsConnected() {
		return nil
	}

	opts := mqtt.NewClientOptions().AddBroker("tcp://" + 地址 + ":" + D到文本(端口))
	opts.SetClientID(J校验_取md5_文本(global.X系统信息.H会员帐号, true))
	opts.SetUsername(用户名)
	opts.SetPassword(密码)
	opts.SetDefaultPublishHandler(j.onMessageReceived)
	opts.SetAutoReconnect(true)
	opts.SetOrderMatters(false)

	j.mQTTClient = mqtt.NewClient(opts)
	if token := j.mQTTClient.Connect(); token.Wait() && token.Error() != nil {
		return errors.New("MQTT连接失败: " + token.Error().Error())
	}
	log.Println("MQTT连接成功")
	return nil
}

func (j *mqttClient) onMessageReceived(client mqtt.Client, message mqtt.Message) {
	fmt.Printf("接收topic: %s\nMessage: %s\n", message.Topic(), message.Payload())

	// 在这里将消息转发回业务平台，您可以根据需要修改此部分
}

func (j *mqttClient) F发送消息(c *gin.Context, 主题, 消息 string) (err error) {
	if !j.mQTTClient.IsConnected() {
		return errors.New("MQTT未连接")
	}
	if token := j.mQTTClient.Publish(主题, 1, false, 消息); token.Wait() && token.Error() != nil {
		return errors.New("MQTT发送失败: " + token.Error().Error())
	}
	return
}
func (j *mqttClient) D断开(c *gin.Context) {
	if !j.mQTTClient.IsConnected() {
		return
	}
	j.mQTTClient.Disconnect(250)
}
func (j *mqttClient) Q取连接状态(c *gin.Context) bool {

	return j.mQTTClient.IsConnected()
}
