package protocol

const (
	// 登录相关
	MsgIDLoginReq  uint32 = 1 // 登录请求
	MsgIDLoginResp uint32 = 2 // 登录响应

	// 注册相关
	MsgIDRegisterReq  uint32 = 3 // 注册请求
	MsgIDRegisterResp uint32 = 4 // 注册响应

	// 聊天相关
	MsgIDTextMsg uint32 = 10 // 文本消息

	// 系统相关
	MsgIDHeartbeat uint32 = 99 // 心跳包
)
