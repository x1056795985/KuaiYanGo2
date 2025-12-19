package webSocket

import (
	"github.com/songzhibin97/gkit/tools/rand_string"
	"net/http"
	"server/Service/Ser_Log"
	"server/global"
	"server/new/app/logic/webSocket"
	"server/new/app/models/constant"
	"server/new/app/service"
	DB "server/structs/db"
	"server/utils/Qqwry"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type AllRouter struct {
}

func (r *AllRouter) InitWebSocketRouter(router *gin.RouterGroup) {

	//打包静态VueAdmin文件============================
	Router根WS := router.Group("ws") //127.0.0.1:18080/  这个后面第一个不需要 / 符号

	Router根WS.GET("/:Token", WebSocketHandler)
	Router根WS.GET("/Token", WebSocketHandler)

}

func WebSocketHandler(c *gin.Context) {
	// 获取WebSocket连接
	wsUpgrader := websocket.Upgrader{
		HandshakeTimeout: time.Second * 10,
		ReadBufferSize:   4096,
		WriteBufferSize:  4096,
		CheckOrigin: func(r *http.Request) bool {
			// 允许所有来源，生产环境中应该更严格地限制
			return true
		},
	}
	ws, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		//fmt.Println("请使用ws/wss连接:" + err.Error())
		return
	}
	返回 := ""
	局_Token := c.Param("Token")
	if 局_Token == "" {
		返回 = `{"code":200,"msg":"Token不能为空"}`
		_ = ws.WriteMessage(websocket.TextMessage, []byte(返回))
		ws.Close()
		return
	}

	db := *global.GVA_DB
	var 局_在线信息_旧 DB.DB_LinksToken
	//这里如果报错  invalid memory address or nil pointer dereference   可能是配置文件数据库配置北山,global.GVA_DB 值为空

	局_在线信息_旧, err = service.NewLinksToken(c, &db).InfoToken(局_Token)
	// 没查到数据 或状态不正常
	if err != nil || 局_在线信息_旧.Status != 1 || 局_在线信息_旧.LoginAppid == constant.APPID_WebSocket {
		返回 = `{"code":200,"msg":"Token不存在"}`
		_ = ws.WriteMessage(websocket.TextMessage, []byte(返回))
		ws.Close()
		return
	}
	// 没查到数据 或状态不正常
	if 局_在线信息_旧.LoginAppid == constant.APPID_WebSocket {
		返回 = `{"code":200,"msg":"Token不能为webSocket应用id"}`
		_ = ws.WriteMessage(websocket.TextMessage, []byte(返回))
		ws.Close()
		return
	}

	var 局_在线信息_新 DB.DB_LinksToken
	局_在线信息_新.Uid = 局_在线信息_旧.Uid
	局_在线信息_新.User = 局_在线信息_旧.User
	局_在线信息_新.Tab = strconv.Itoa(局_在线信息_旧.LoginAppid)
	局_在线信息_新.Key = ""
	局_在线信息_新.Ip = c.ClientIP()
	省市, 运行商, err := Qqwry.Ip查信息(局_在线信息_新.Ip)
	if err == nil && 省市 != "" {
		局_在线信息_新.IPCity = 省市 + " " + 运行商
	}
	局_在线信息_新.Status = 1
	局_在线信息_新.LoginTime = time.Now().Unix()
	局_在线信息_新.OutTime = 36000 //退出时间
	局_在线信息_新.LastTime = 局_在线信息_新.LoginTime
	局_在线信息_新.Token = strings.ToUpper(rand_string.RandomLetter(32))
	局_在线信息_新.LoginAppid = constant.APPID_WebSocket
	err = global.GVA_DB.Create(&局_在线信息_新).Error
	if err != nil {
		返回 := `{"code":200,"msg":"Token创建失败"}`
		_ = ws.WriteMessage(websocket.TextMessage, []byte(返回))
		return
	}
	go Ser_Log.Log_写登录日志(局_在线信息_旧.User, c.ClientIP(), "["+strconv.Itoa(局_在线信息_旧.LoginAppid)+"]已链接wss", constant.APPID_Web用户中心)

	// 添加连接到管理器
	webSocket.L_webSocket.Add(c, 局_在线信息_新.Id, ws)
	// 启动新的goroutine处理连接，避免阻塞
	go webSocket.L_webSocket.HandleConnection(ws, 局_在线信息_新.Id)
	// 立即返回，不阻塞HTTP handler
	c.Abort()
}
