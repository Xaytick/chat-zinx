package protocol

const (
	// 基础消息ID范围: 1-100
	MsgIDPing = iota + 1 // 1: 心跳检测
	MsgIDPong            // 2: 心跳响应

	// 用户相关消息ID范围: 101-200
	MsgIDRegisterReq  = iota + 100 // 101: 注册请求
	MsgIDRegisterResp              // 102: 注册响应
	MsgIDLoginReq                  // 103: 登录请求
	MsgIDLoginResp                 // 104: 登录响应
	MsgIDLogoutReq                 // 105: 登出请求
	MsgIDLogoutResp                // 106: 登出响应

	// 聊天相关消息ID范围: 201-300
	MsgIDTextMsg          = iota + 200 // 201: 文本消息
	MsgIDImageMsg                      // 202: 图片消息
	MsgIDFileMsg                       // 203: 文件消息
	MsgIDHistoryMsgReq                 // 204: 历史消息请求
	MsgIDHistoryMsgResp                // 205: 历史消息响应
	MsgIDChatRelationReq               // 206: 聊天关系请求
	MsgIDChatRelationResp              // 207: 聊天关系响应

	// 群组相关消息ID (Group related message IDs) - Starts after Chat related iota
	MsgIDCreateGroupReq      // 208
	MsgIDCreateGroupResp     // 209
	MsgIDJoinGroupReq        // 210
	MsgIDJoinGroupResp       // 211
	MsgIDLeaveGroupReq       // 212
	MsgIDLeaveGroupResp      // 213
	MsgIDGroupChatMsg        // 214: 群聊消息
	MsgIDGetGroupMembersReq  // 215: 获取群成员列表请求
	MsgIDGetGroupMembersResp // 216: 获取群成员列表响应
	MsgIDGetUserGroupsReq    // 217: 获取用户加入的群列表请求
	MsgIDGetUserGroupsResp   // 218: 获取用户加入的群列表响应
	// 后续可以添加更多群管理相关的消息ID，如踢人、授权等

	// 群组消息相关 310 - 319
	MsgIDGroupTextMsgReq  uint32 = 310 // C->S 发送群组文本消息请求
	MsgIDGroupTextMsgResp uint32 = 311 // S->C 发送群组文本消息响应
	MsgIDGroupTextMsgPush uint32 = 312 // S->C 推送群组文本消息
)

// 单独定义通用错误响应ID，避免破坏现有 iota 序列
const (
	MsgIDErrorResp = 99 // 通用错误响应ID (选择一个在基础范围且未被iota占用的值)
)
