package protocol

const (
	// 基础消息ID范围: 1-100
	MsgIDPing = iota + 1 // 1: 心跳检测
	MsgIDPong            // 2: 心跳响应

	// 用户相关消息ID范围: 101-200
	MsgIDRegisterReq  = iota + 101 // 101: 注册请求
	MsgIDRegisterResp              // 102: 注册响应
	MsgIDLoginReq                  // 103: 登录请求
	MsgIDLoginResp                 // 104: 登录响应
	MsgIDLogoutReq                 // 105: 登出请求
	MsgIDLogoutResp                // 106: 登出响应

	// 聊天相关消息ID范围: 201-300
	MsgIDTextMsg          = iota + 201 // 201: 文本消息
	MsgIDImageMsg                      // 202: 图片消息
	MsgIDFileMsg                       // 203: 文件消息
	MsgIDHistoryMsgReq                 // 204: 历史消息请求
	MsgIDHistoryMsgResp                // 205: 历史消息响应
	MsgIDChatRelationReq               // 206: 聊天关系请求
	MsgIDChatRelationResp              // 207: 聊天关系响应
)
