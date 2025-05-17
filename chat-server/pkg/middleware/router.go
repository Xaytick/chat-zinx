package middleware

import (
	"encoding/json"
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
	"github.com/Xaytick/zinx/ziface"
)

// AuthRouter 认证路由包装器
// 包装已有的路由并添加认证功能
type AuthRouter struct {
	// 被包装的原始路由
	router ziface.IRouter
	// 认证中间件
	auth *AuthMiddleware
	// 是否跳过认证（用于公开接口如登录）
	skipAuth bool
	// 认证失败时的错误消息
	errorMsgID uint32
}

// NewAuthRouter 创建新的认证路由包装器
func NewAuthRouter(router ziface.IRouter, auth *AuthMiddleware, skipAuth bool) *AuthRouter {
	return &AuthRouter{
		router:     router,
		auth:       auth,
		skipAuth:   skipAuth,
		errorMsgID: protocol.MsgIDLoginResp, // 默认以登录响应消息ID返回错误
	}
}

// PreHandle 预处理，实现认证逻辑
func (ar *AuthRouter) PreHandle(request ziface.IRequest) {
	// 如果跳过认证，直接调用原始路由的PreHandle
	if ar.skipAuth {
		ar.router.PreHandle(request)
		return
	}

	// 执行认证
	passed, userInfo, err := ar.auth.Verify(request)

	if !passed {
		// 认证失败，发送错误响应
		errMsg := "认证失败"
		if err != nil {
			errMsg = err.Error()
		}
		fmt.Printf("[认证] 请求认证失败: %s\n", errMsg)

		// 构造错误响应
		errResp := map[string]interface{}{
			"code": 401,
			"msg":  errMsg,
		}

		// 序列化并发送错误消息
		respData, _ := json.Marshal(errResp)
		request.GetConnection().SendMsg(ar.errorMsgID, respData)

		// 认证失败，不调用原始路由的PreHandle
		return
	}

	// 认证成功，将用户信息存储到连接属性中
	if userInfo != nil {
		request.GetConnection().SetProperty("userID", userInfo.UserUUID)
		request.GetConnection().SetProperty("username", userInfo.Username)
	}

	// 调用原始路由的PreHandle
	ar.router.PreHandle(request)
}

// Handle 处理请求，如果认证通过则调用原始路由的Handle
func (ar *AuthRouter) Handle(request ziface.IRequest) {
	// 直接调用原始路由的Handle
	// 在PreHandle中已经判断认证是否通过，如果未通过则不会执行到这里
	ar.router.Handle(request)
}

// PostHandle 后处理，直接调用原始路由的PostHandle
func (ar *AuthRouter) PostHandle(request ziface.IRequest) {
	// 直接调用原始路由的PostHandle
	ar.router.PostHandle(request)
}

// WithErrorMsgID 设置错误消息ID
func (ar *AuthRouter) WithErrorMsgID(msgID uint32) *AuthRouter {
	ar.errorMsgID = msgID
	return ar
}
