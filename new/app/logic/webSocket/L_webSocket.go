package webSocket

import (
	json2 "encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"runtime/debug"
	"server/Service/Ser_Js"
	"server/Service/Ser_PublicJs"
	"server/global"
	"server/new/app/models/common"
	"server/new/app/models/constant"
	"server/new/app/service"
	DB "server/structs/db"
	"sync"
	"time"
)

var L_webSocket webSocket

// 定义连接存储结构
type WSConnection struct {
	ws              *websocket.Conn //ws连接
	linkId          int             //在线id,数据库对应在线记录id
	lastTime        int64           //最后心跳时间  //可以降低数据库的读取次数,不用每次扫描都读库
	lastWriteDbTime int64           //最后更新在线信息时间
}

func init() {
	L_webSocket = webSocket{}
	L_webSocket.wsObj = sync.Map{} // 使用全局变量或更好的方式存储活跃连接
}

type webSocket struct {
	wsObj sync.Map // 并发安全的map
}

func (j *webSocket) F发送消息给所有连接用户(c *gin.Context, message []byte) {
	j.wsObj.Range(func(key, value interface{}) bool {
		conn, ok2 := value.(*WSConnection)
		if !ok2 {
			return true
		}
		if err := conn.ws.WriteMessage(websocket.TextMessage, message); err != nil {
			// 处理发送失败的情况，可能需要清理无效连接
			j.wsObj.Delete(key)
		}
		return true
	})
}

func (j *webSocket) F发送ping消息给所有连接用户(c *gin.Context) {
	time := time.Now().Unix()
	j.wsObj.Range(func(key, value interface{}) bool {
		conn, ok2 := value.(*WSConnection)
		if !ok2 {
			return true
		}

		if time-conn.lastTime > 180 { //超过180秒无响应,直接断开连接
			j.RemoveConnection(conn.linkId)
			return true
		}
		if time-conn.lastTime > 30 {
			if err := conn.ws.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				// 处理发送失败的情况，可能需要清理无效连接
				j.RemoveConnection(conn.linkId)
			}
		}
		return true
	})
}

// 添加
func (j *webSocket) Add(c *gin.Context, linkId int, ws *websocket.Conn) {
	j.wsObj.Store(linkId, &WSConnection{
		linkId:   linkId,
		ws:       ws,
		lastTime: time.Now().Unix(),
	})

}

// HandleConnection 处理单个WebSocket连接的消息循环
func (j *webSocket) HandleConnection(ws *websocket.Conn, linkId int) {
	defer func() {
		if err := recover(); err != nil {
			局_上报错误 := fmt.Sprintln("WebSocket捕获错误:\n", err, "\n堆栈信息:\n", string(debug.Stack()))
			debug.PrintStack()
			log.Println("发生致命错误:", 局_上报错误)
			j.RemoveConnection(linkId)
		}
	}()
	defer ws.Close()

	// 处理WebSocket消息
	for {
		messageType, p, err := ws.ReadMessage()
		if err != nil {
			//fmt.Println("读取消息错误:", err)
			j.RemoveConnection(linkId) // 移除连接
			return
		}

		// 更新最后心跳时间
		if conn, ok := j.GetConnection(linkId); ok {
			conn.lastTime = time.Now().Unix() //是指针,直接改就行
			db := *global.GVA_DB
			_, err = service.NewLinksToken(&gin.Context{}, &db).Update(linkId, map[string]interface{}{"lastTime": time.Now().Unix()})
			if err != nil {
				log.Println("更新在线信息失败:", err)
			}
		}

		switch messageType {
		case websocket.TextMessage:
			fmt.Printf("处理文本消息, %s\n", string(p))
			// 可以在这里调用业务逻辑处理函数
			j.ProcessTextMessage(ws, linkId, p)

		case websocket.BinaryMessage:
			fmt.Println("处理二进制消息")
			// 处理二进制消息 //不支持二进制消息 因为 js无法处理

			返回 := `{"code":200,"msg":"不支持二进制消息"}`

			_ = ws.WriteMessage(websocket.TextMessage, []byte(返回))

		case websocket.CloseMessage:
			//fmt.Println("关闭websocket连接")
			j.RemoveConnection(linkId)
			return

		case websocket.PingMessage:
			//fmt.Println("处理ping消息")
			ws.WriteMessage(websocket.PongMessage, []byte("pong"))

		case websocket.PongMessage:
			//fmt.Println("处理pong消息")

		default:
			fmt.Printf("未知消息类型: %d\n", messageType)
		}
	}
}

// ProcessTextMessage 处理文本消息的业务逻辑
func (j *webSocket) ProcessTextMessage(ws *websocket.Conn, linkId int, message []byte) {
	var 局_json common.WsMsgRequest
	err := json2.Unmarshal(message, &局_json)
	if err != nil {
		//消息格式不对,断开链接
		j.RemoveConnection(linkId)
		return
	}
	局_PublicJs, err := Ser_PublicJs.P取值2(Ser_PublicJs.Js类型_webSocket, 局_json.Api)
	if err != nil {
		return
	}
	var AppInfo DB.DB_AppInfo
	var 局_在线信息 DB.DB_LinksToken
	var c = gin.Context{}
	var response common.WsMsgResponse
	response.I = 局_json.I

	db := *global.GVA_DB
	//获取该应用是否已开启了cps 可能会有多个符合时间的配置信息 只获取第一个
	局_在线信息, err = service.NewLinksToken(&c, &db).Info(linkId)
	vm := Ser_Js.JS引擎初始化_用户(&gin.Context{}, &AppInfo, &局_在线信息, &局_PublicJs)
	_, err = vm.RunString(局_PublicJs.Value)
	if 局_详细错误, ok := err.(*goja.Exception); ok {
		response.Code = constant.Status_操作失败
		response.Msg = "JS代码运行失败:" + 局_详细错误.String()

	}
	var 局_待执行js函数名 func(common.WsMsgRequest) interface{}
	ret := vm.Get(局_PublicJs.Name)
	if ret == nil {
		response.Code = constant.Status_操作失败
		response.Msg = "Js中没有[" + 局_PublicJs.Name + "()]函数"
	}
	err = vm.ExportTo(ret, &局_待执行js函数名)
	if err != nil {
		response.Code = constant.Status_操作失败
		response.Msg = "Js绑定函数到变量失败"
	}

	局_return := 局_待执行js函数名(局_json)
	response.Code = constant.Status_操作成功
	response.Data = 局_return
	返回 := fmt.Sprintf("%v", response) //不管是什么类型,直接转文本
	err = ws.WriteMessage(websocket.TextMessage, []byte(返回))
	if err != nil {
		return
	}

}

// GetConnection 获取连接信息
func (j *webSocket) GetConnection(linkId int) (*WSConnection, bool) {
	if value, ok := j.wsObj.Load(linkId); ok {
		conn, ok2 := value.(*WSConnection)
		if !ok2 {
			return nil, false
		}
		return conn, true
	}
	return nil, false
}

// 更新连接信息
func (j *webSocket) UpdateConnection(linkId int, conn *WSConnection) {
	j.wsObj.Store(linkId, conn) // 直接存储指针
}

// RemoveConnection 移除连接
func (j *webSocket) RemoveConnection(linkId int) {
	if value, ok := j.wsObj.Load(linkId); ok {
		conn, ok2 := value.(*WSConnection)
		if !ok2 {
			return
		}
		// 安全关闭WebSocket连接
		if conn.ws != nil {
			conn.ws.Close()
		}
		// 从map中删除
		j.wsObj.Delete(linkId)
		count := 0
		j.wsObj.Range(func(key, value interface{}) bool {
			count++
			return true
		})
		fmt.Print("剩余连接数:", count)
		db := *global.GVA_DB
		_, err := service.NewLinksToken(&gin.Context{}, &db).Update(linkId, map[string]interface{}{"Status": 2})
		if err != nil {
			return
		}

	}
}
