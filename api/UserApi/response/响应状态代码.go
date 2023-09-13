// 返回加密结果
package response

// 常量 回复状态码
const (
	//中间件鉴权使用的错误代码
	Status_系统已关闭    = 100
	Status_App不存在   = 101
	Status_Api不存在   = 102
	Status_签名错误     = 103
	Status_参数错误     = 104
	Status_加解密失败    = 105
	Status_Token无效  = 106
	Status_封包超时     = 107
	Status_状态码错误    = 108
	Status_Token已注销 = 109
	Status_已停止运营    = 110
	Status_验证码错误    = 111

	//Api使用的错误代码
	Status_操作失败         = 200
	Status_SQl错误        = 201
	Status_登录失败         = 202
	Status_版本不可用        = 203
	Status_Vip已到期       = 204
	Status_绑定信息验证失败     = 205
	Status_绑定信息已被其他用户使用 = 206
	Status_已冻结无法登录      = 207
	Status_同时在线超过最大值    = 208
	Status_未登录          = 210
)

var Status值键 = make(map[int]string)

func init() {
	//中间件鉴权使用的错误代码
	Status值键[Status_App不存在] = "App不存在"
	Status值键[Status_Api不存在] = "Api不存在"
	Status值键[Status_签名错误] = "请求签名错误"
	Status值键[Status_参数错误] = "参数错误"
	Status值键[Status_加解密失败] = "加解密失败请检查加密格式"
	Status值键[Status_Token无效] = "Token无效"
	Status值键[Status_封包超时] = "封包超时"
	Status值键[Status_状态码错误] = "状态码错误"
	Status值键[Status_Token已注销] = "Token已注销"
	Status值键[Status_验证码错误] = "验证码错误请重新输入"

	//Api使用的错误代码
	Status值键[Status_操作失败] = "操作失败"
	Status值键[Status_SQl错误] = "服务器内部错误,联系开发者查看日志"
	Status值键[Status_登录失败] = "登录失败"
	Status值键[Status_版本不可用] = "当前版本不可用,请更新最新版本"
	Status值键[Status_Vip已到期] = "Vip已到期"
	Status值键[Status_绑定信息验证失败] = "绑定信息验证失败"
	Status值键[Status_绑定信息已被其他用户使用] = "绑定信息已被其他用户使用"
	Status值键[Status_已冻结无法登录] = "已冻结无法登录"
	Status值键[Status_同时在线超过最大值] = "同时在线超过最大值,请退出其他登录."
	Status值键[Status_未登录] = "未登录,请先操作登录"

}
