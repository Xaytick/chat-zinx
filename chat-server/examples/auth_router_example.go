package examples

import (
	"github.com/Xaytick/chat-zinx/chat-server/pkg/middleware"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
	"github.com/Xaytick/chat-zinx/chat-server/router"
	"github.com/Xaytick/zinx/znet"
)

// SetupAuthRouters 设置认证路由示例
func SetupAuthRouters(server *znet.Server) {
	// 创建认证中间件
	authMiddleware := middleware.NewAuthMiddleware(
		middleware.WithRedisCheck(true), // 启用Redis会话验证
	)

	// 注册不需要认证的路由（如登录、注册）
	// 登录路由不需要认证，但会在登录成功后生成Token
	loginRouter := router.NewLoginRouter()
	// 把原始登录路由包装成认证路由，但标记为skipAuth=true，跳过认证
	wrappedLoginRouter := middleware.NewAuthRouter(loginRouter, authMiddleware, true)
	server.AddRouter(protocol.MsgIDLoginReq, wrappedLoginRouter)

	// 注册路由也不需要认证
	registerRouter := router.NewRegisterRouter()
	wrappedRegisterRouter := middleware.NewAuthRouter(registerRouter, authMiddleware, true)
	server.AddRouter(protocol.MsgIDRegisterReq, wrappedRegisterRouter)

	// 注册需要认证的路由（如消息发送、获取历史记录等）
	// 注意：这里的两个路由函数在当前实现中尚未提供，实际使用时需要实现这些路由

	/* 暂时注释未实现的路由
	// 文本消息路由需要认证
	textMsgRouter := router.NewTextMsgRouter()
	// 包装原始路由，不跳过认证（skipAuth=false）
	wrappedTextMsgRouter := middleware.NewAuthRouter(textMsgRouter, authMiddleware, false)
	server.AddRouter(protocol.MsgIDTextMsg, wrappedTextMsgRouter)

	// 历史消息路由需要认证
	historyMsgRouter := router.NewHistoryMsgRouter()
	wrappedHistoryMsgRouter := middleware.NewAuthRouter(historyMsgRouter, authMiddleware, false)
	server.AddRouter(protocol.MsgIDHistoryMsgReq, wrappedHistoryMsgRouter)
	*/
}

// GetJWTToken 登录成功后生成JWT令牌示例
func GetJWTToken(userID, username string) (string, error) {
	// 创建认证中间件
	authMiddleware := middleware.NewAuthMiddleware()

	// 生成JWT令牌
	token, err := authMiddleware.GenerateToken(userID, username)
	if err != nil {
		return "", err
	}

	// 保存会话到Redis
	// 注意：这里总是保存会话，如果不需要可以根据配置判断
	if err := authMiddleware.SaveSession(userID, token); err != nil {
		return "", err
	}

	return token, nil
}

// 扩展现有的登录路由，添加JWT令牌生成功能
func ExtendLoginRouter() {
	/*
		// 这是一个示例，展示如何在原有的LoginRouter中添加JWT令牌生成功能
		// 在实际项目中，你需要修改LoginRouter的代码，如下所示：

		// 在登录成功后
		if user != nil {
			// 生成JWT令牌
			token, err := GetJWTToken(user.UserID, user.Username)
			if err != nil {
				// 处理错误
			}

			// 将令牌添加到响应数据中
			responseData := &model.UserLoginResponse{
				UserID:    userID,
				Username:  user.Username,
				Email:     user.Email,
				LastLogin: user.LastLogin.Format(time.DateTime),
				Token:     token, // 添加令牌
			}

			// 发送响应
			sendLoginResponse(request, 0, "登录成功", responseData)
		}
	*/
}
