package common

// websocket请求结构体
type WsMsgRequest struct {
	I    int                    `json:"I"` //请求计数,如果是同步响应的,需要填写
	Api  string                 `json:"api"`
	Data map[string]interface{} `json:"data"` //必须是json结构体
}

// websocket响应结构体
type WsMsgResponse struct {
	I    int         `json:"I"` // 请求计数,如果是同步响应的,需要填写
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}
