package webSocket

import (
	. "EFunc/utils"
	json2 "encoding/json"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"runtime/debug"
	"server/Service/Ser_PublicJs"
	"server/global"
	"server/new/app/logic/common/cycleNot"
	"server/new/app/models/common"
	"server/new/app/models/constant"
	"server/new/app/service"
	db "server/structs/db"
	"sync"
	"sync/atomic"
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
	wsObj            sync.Map // 并发安全的map
	heartbeatRunning uint32   // 使用原子操作标志位替代互斥锁
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

func (j *webSocket) F发送消息(linkId int, message []byte) error {
	if value, ok := j.wsObj.Load(linkId); ok {
		conn, ok2 := value.(*WSConnection)
		if !ok2 {
			return errors.New("id链接异常")
		}
		return conn.ws.WriteMessage(websocket.TextMessage, message)

	}
	return errors.New("id链接不存在")
}
func (j *webSocket) F发送消息_批量(linkIds []int, message []byte) []error {
	局_结果 := make([]error, len(linkIds))
	for i, linkId := range linkIds {
		if value, ok := j.wsObj.Load(linkId); ok {
			conn, ok2 := value.(*WSConnection)
			if !ok2 {
				局_结果[i] = errors.New("id链接异常")
				continue
			}

			局_结果[i] = conn.ws.WriteMessage(websocket.TextMessage, message)
		} else {
			局_结果[i] = errors.New("id链接不存在")
		}

	}
	return 局_结果

}
func (j *webSocket) F发送ping消息给所有连接用户() (剩余数量 int) {
	局_time := time.Now().Unix()
	局_ids := make([]int, 0, 100)
	局_临时计数 := 0
	j.wsObj.Range(func(key, value interface{}) bool {
		conn, ok2 := value.(*WSConnection)
		if !ok2 {
			return true
		}
		局_临时计数 += 1
		if 局_time-conn.lastTime > 180 { //超过180秒无响应,直接断开连接
			j.RemoveConnection(conn.linkId)
			return true
		}
		// 超过30秒发送ping
		if 局_time-conn.lastTime > 30 {
			if err := conn.ws.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				// 处理发送失败的情况，可能需要清理无效连接
				j.RemoveConnection(conn.linkId)
			}
		}

		if 局_time-conn.lastWriteDbTime > 60 { //如果距离上次更新入库超过了 60 秒,则更新入库
			局_ids = append(局_ids, conn.linkId)
			conn.lastWriteDbTime = 局_time //指针,直接改就行 降低写库频率
		}

		return true
	})

	if len(局_ids) > 0 {
		// 批量更新数据库
		db := *global.GVA_DB
		_, err := service.NewLinksToken(&gin.Context{}, &db).Updates(局_ids, map[string]interface{}{"lastTime": time.Now().Unix()})
		if err != nil {
			log.Println("更新在线信息失败:", err)
		}
	}
	//fmt.Println("F发送ping消息给所有连接用户耗时:", time.Now().Unix()-局_time, "\n")

	return 局_临时计数
}

// 添加
func (j *webSocket) Add(c *gin.Context, linkId int, ws *websocket.Conn) {
	j.wsObj.Store(linkId, &WSConnection{
		linkId:   linkId,
		ws:       ws,
		lastTime: time.Now().Unix(),
	})
	// 原子操作确保只启动一个心跳协程
	if atomic.CompareAndSwapUint32(&j.heartbeatRunning, 0, 1) {
		go j.runHeartbeat()
	}

}

func (j *webSocket) runHeartbeat() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("心跳协程异常恢复: %v", err)
			// 确保标志位被重置，允许下次重启
			atomic.StoreUint32(&j.heartbeatRunning, 0)
		}
	}()
	defer atomic.StoreUint32(&j.heartbeatRunning, 0)

	for {
		局_数量 := j.F发送ping消息给所有连接用户()
		//fmt.Print("心跳协程已发送:", 局_数量)
		if 局_数量 == 0 {
			break // 没有连接时跳出
		}
		time.Sleep(25 * time.Second)
	}
	//fmt.Print("心跳协程已停止")
}

// HandleConnection 处理单个WebSocket连接的消息循环
func (j *webSocket) HandleConnection(ws *websocket.Conn, linkId int) {
	defer func() {
		if err := recover(); err != nil {
			局_上报错误 := fmt.Sprintln("WebSocket捕获错误:\n", err, "\n堆栈信息:\n", string(debug.Stack()))
			debug.PrintStack()
			log.Println("发生致命错误:", 局_上报错误)
		}
		// 确保连接总是被清理
		j.RemoveConnection(linkId)
	}()
	defer func() {
		if ws != nil {
			ws.Close()
		}
	}()

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
		}

		switch messageType {
		case websocket.TextMessage:
			if len(p) == 1 && string(p) == "1" { //响应心跳
				_ = ws.WriteMessage(websocket.TextMessage, []byte("2"))
			} else {
				//fmt.Printf("处理文本消息, %s\n", string(p))
				// 可以在这里调用业务逻辑处理函数
				j.ProcessTextMessage(ws, linkId, &p)
			}

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

// 处理文本消息的业务逻辑
func (j *webSocket) ProcessTextMessage(ws *websocket.Conn, linkId int, message *[]byte) {
	var 局_json common.WsMsgRequest
	err := json2.Unmarshal(*message, &局_json)
	if err != nil {
		//消息格式不对,断开链接
		j.RemoveConnection(linkId)
		return
	}

	var 局_PublicJs db.DB_PublicJs
	if W文本_是否为数字(局_json.Api) {
		局_PublicJs, err = Ser_PublicJs.Q取值2(D到整数(局_json.Api))
	} else {
		局_PublicJs, err = Ser_PublicJs.P取值2(constant.APPID_WebSocket, 局_json.Api)
	}

	if err != nil || 局_PublicJs.AppId != constant.APPID_WebSocket {
		return
	}

	var AppInfo db.DB_AppInfo
	var 局_在线信息 db.DB_LinksToken
	var c = gin.Context{}
	var response common.WsMsgResponse
	response.I = 局_json.I
	response.Code = constant.Status_操作失败

	// 检查是否需要登录
	db := *global.GVA_DB
	局_在线信息, err = service.NewLinksToken(&c, &db).Info(linkId)
	if err != nil {
		return
	}

	if 局_PublicJs.IsVip > 0 && 局_在线信息.Uid == 0 {
		response.Msg = "未登录,请使用登陆后Token链接"
		j.sendResponse(ws, &response)
		return
	}

	// 初始化JS引擎并执行代码
	vm := cycleNot.GlobalJsEngineInit(&gin.Context{}, &AppInfo, &局_在线信息, &局_PublicJs)
	_, err = vm.RunString(局_PublicJs.Value)
	if 局_详细错误, ok := err.(*goja.Exception); ok {
		response.Msg = "JS代码运行失败:" + 局_详细错误.String()
		j.sendResponse(ws, &response)
		return
	}

	// 获取并调用JS函数
	ret := vm.Get(局_PublicJs.Name)
	if ret == nil {
		response.Msg = "Js中没有[" + 局_PublicJs.Name + "()]函数"
		j.sendResponse(ws, &response)
		return
	}

	var 局_待执行js函数名 func(map[string]interface{}) string
	err = vm.ExportTo(ret, &局_待执行js函数名)
	if err != nil {
		response.Msg = "Js绑定函数到变量失败"
		j.sendResponse(ws, &response)
		return
	}

	// 执行JS函数并返回结果

	response.Code = constant.Status_操作成功
	局_返回 := 局_待执行js函数名(局_json.Data)
	var mapkv map[string]interface{}
	//判断字符串是否为json格式如果是json则解析

	if W文本_可能为json(局_返回) && json2.Unmarshal([]byte(局_返回), &mapkv) == nil {
		response.Data = mapkv
	} else {
		response.Data = 局_返回
	}
	j.sendResponse(ws, &response)
}

// 提取公共的响应发送逻辑
func (j *webSocket) sendResponse(ws *websocket.Conn, response *common.WsMsgResponse) {
	返回, err := json2.Marshal(response)
	if err != nil {
		return
	}
	_ = ws.WriteMessage(websocket.TextMessage, 返回)
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
		//fmt.Print("剩余连接数:", count, "\n")
		db := *global.GVA_DB
		_, err := service.NewLinksToken(&gin.Context{}, &db).Update(linkId, map[string]interface{}{"Status": 2})
		if err != nil {
			//fmt.Println("更新在线状态失败:", err.Error())
			return
		}
	}
}

func (j *webSocket) D断开所有连接() {
	j.wsObj.Range(func(key, value interface{}) bool {
		conn, ok2 := value.(*WSConnection)
		if !ok2 {
			return true
		}
		j.RemoveConnection(conn.linkId)
		return true
	})
	// 批量更新数据库
	db := *global.GVA_DB
	info, err := service.NewLinksToken(&gin.Context{}, &db).Infos(map[string]interface{}{"LoginAppid": constant.APPID_WebSocket, "Status": 1})
	if err != nil || len(info) == 0 {
		return
	}
	var 局_ids = make([]int, 0, len(info))
	for _, v := range info {
		局_ids = append(局_ids, v.Id)
	}

	fmt.Println("批量注销ws在线状态:", 局_ids)
	_, err = service.NewLinksToken(&gin.Context{}, &db).Updates(局_ids, map[string]interface{}{"Status": 2})

	return
}
