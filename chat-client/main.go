package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Xaytick/chat-zinx/chat-client/pkg/client"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	serverProtocol "github.com/Xaytick/chat-zinx/chat-server/pkg/protocol"
)

var cli *client.ChatClient
var serverAddr = "127.0.0.1:9000"  // Default server address
var doneChan = make(chan struct{}) // Channel to signal graceful shutdown

// For handling console output without clashing with input prompt
var outputChan = make(chan string, 100)

var expectingMessageContentForRecipient string // If not empty, next input is message content for this recipient

func main() {
	go consoleOutputRoutine() // Start a goroutine to handle all console output

	reader := bufio.NewReader(os.Stdin)
	outputChan <- "简易聊天客户端"
	outputChan <- "---------------------"
	outputChan <- fmt.Sprintf("默认服务器地址: %s", serverAddr)
	outputChan <- "输入 /connect [host:port] 连接到不同服务器。"
	outputChan <- "输入 /help 获取命令列表。"
	outputChan <- "输入 /quit 退出。"

	// Handle Ctrl+C for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan // Wait for SIGINT or SIGTERM
		outputChan <- "\n检测到中断信号，正在关闭..."
		if expectingMessageContentForRecipient != "" { // If waiting for content, cancel it
			outputChan <- "当前消息输入已取消。"
			expectingMessageContentForRecipient = ""
		}
		if cli != nil {
			cli.Close()
		}
		close(doneChan) // Signal main loop to terminate
	}()

mainLoop:
	for {
		if expectingMessageContentForRecipient == "" {
			fmt.Print("> ")
		} else {
			// Prompt is handled by the command that sets expectingMessageContentForRecipient
			// or could be set here, e.g., fmt.Printf("To %s: ", expectingMessageContentForRecipient)
			// For now, handleSendMsg will print "请输入消息内容..."
			// And consoleOutputRoutine will reprint "> " if it runs after a message is printed.
			// This might need refinement for perfect prompt handling.
		}
		var input string
		inputChan := make(chan string)
		readErrChan := make(chan error)

		go func() {
			line, err := reader.ReadString('\n') // Use the reader defined in main
			if err != nil {
				readErrChan <- err
				return
			}
			inputChan <- line
		}()

		select {
		case line := <-inputChan:
			input = line
		case err := <-readErrChan:
			outputChan <- fmt.Sprintf("读取输入错误: %v，客户端退出。", err)
			break mainLoop
		case <-doneChan: // Shutdown signal from Ctrl+C or /quit
			break mainLoop
		}

		input = strings.TrimSpace(input)

		if expectingMessageContentForRecipient != "" {
			if input == "/cancel" {
				outputChan <- "消息发送已取消。"
				expectingMessageContentForRecipient = ""
				continue // Go back to expecting a command
			}
			if strings.TrimSpace(input) == "" {
				outputChan <- "不能发送空消息。操作已取消。"
				expectingMessageContentForRecipient = ""
				continue
			}
			cli.SendTextMessage(expectingMessageContentForRecipient, input)
			expectingMessageContentForRecipient = "" // Reset after sending
			continue                                 // Go back to expecting a command
		}

		// If not expecting message content, process as a command
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}
		command := parts[0]
		args := parts[1:]

		switch command {
		case "/connect":
			handleConnect(args)
		case "/register":
			handleRegister(args)
		case "/login":
			handleLogin(args)
		case "/msg":
			handleSendMsg(args) // This will now set expectingMessageContentForRecipient if needed
		case "/history":
			handleHistory(args)
		case "/creategroup":
			handleCreateGroup(args)
		case "/joingroup":
			handleJoinGroup(args)
		case "/leavegroup":
			handleLeaveGroup(args)
		case "/help":
			handleHelp()
		case "/quit", "/exit":
			if cli != nil {
				cli.Close()
			}
			outputChan <- "客户端退出。"
			close(doneChan)
		default:
			outputChan <- "未知命令。输入 /help 获取命令列表。"
		}
	}
	outputChan <- "主程序循环结束。等待输出完成..."
	time.Sleep(100 * time.Millisecond)
}

func consoleOutputRoutine() {
	for msg := range outputChan {
		fmt.Printf("\r\033[K%s\n", msg) // Clear line, print message
		// If not in mainLoop (e.g. after mainLoop exited), don't reprint prompt
		// A more sophisticated check would be needed if other goroutines could print after mainLoop finishes
		// but before consoleOutputRoutine itself closes. For now, this is okay.
		select {
		case <-doneChan: // If main loop is done, don't reprint prompt
		default:
			fmt.Print("> ") // Reprint prompt
		}
	}
}

func ensureConnected() bool {
	if cli == nil {
		outputChan <- "错误：未连接到服务器。请先使用 /connect [host:port] (默认 " + serverAddr + ")"
		return false
	}
	return true
}

func ensureLoggedIn() bool {
	if !ensureConnected() {
		return false
	}
	if !cli.IsLoggedIn() {
		outputChan <- "错误：请先登录。使用 /login <username> <password>"
		return false
	}
	return true
}

func handleConnect(args []string) {
	targetAddr := serverAddr
	if len(args) > 0 {
		targetAddr = args[0]
	}
	if cli != nil {
		outputChan <- "关闭现有连接..."
		cli.Close()
	}
	var err error
	newCli, err := client.NewChatClient(targetAddr)
	if err != nil {
		outputChan <- fmt.Sprintf("连接到 %s 失败: %v", targetAddr, err)
		return
	}
	cli = newCli // Assign to global cli only on success
	serverAddr = targetAddr
	outputChan <- fmt.Sprintf("成功连接到 %s。请使用 /login 或 /register。", targetAddr)
	cli.StartMsgListener(handleIncomingMessages)
}

func handleIncomingMessages(msgID uint32, data []byte) {
	var output string
	switch msgID {
	case serverProtocol.MsgIDPong:
		// Pong 消息是心跳响应，通常不需要向用户显示
		// fmt.Println("[DEBUG] Received Pong") // 可以保留用于调试
		return // 直接返回，不设置 output，就不会打印到控制台
	case serverProtocol.MsgIDTextMsg:
		var msg model.TextMsg
		if err := json.Unmarshal(data, &msg); err == nil {
			// Try to get username if FromUserID is a UUID (requires server to send it, or a local cache)
			// For now, just using FromUserID (which client.go might populate with UserUUID)
			output = fmt.Sprintf("[消息] %s: %s", msg.FromUserID, msg.Content)
		} else {
			output = fmt.Sprintf("[错误] 解析文本消息失败: %v. 内容: %s", err, string(data))
		}
	case serverProtocol.MsgIDCreateGroupResp:
		var resp model.CreateGroupResp
		if err := json.Unmarshal(data, &resp); err == nil {
			output = fmt.Sprintf("[群组] 创建成功: ID=%d, 名称='%s', 创建者ID=%d, 成员数=%d, 创建于:%s",
				resp.ID, resp.Name, resp.OwnerUserID, resp.MemberCount, resp.CreatedAt)
		} else {
			var errResp map[string]string
			if json.Unmarshal(data, &errResp) == nil && errResp["error"] != "" {
				output = fmt.Sprintf("[错误] 创建群组失败: %s", errResp["error"])
			} else {
				output = fmt.Sprintf("[错误] 解析创建群组响应失败: %v. 内容: %s", err, string(data))
			}
		}
	case serverProtocol.MsgIDJoinGroupResp:
		var resp model.GenericMessageResp
		if err := json.Unmarshal(data, &resp); err == nil {
			if resp.Code == 0 {
				output = fmt.Sprintf("[群组] %s", resp.Message)
			} else {
				output = fmt.Sprintf("[错误] 加入群组操作: %s (code: %d)", resp.Message, resp.Code)
			}
		} else {
			output = fmt.Sprintf("[错误] 解析加入群组响应失败: %v. 内容: %s", err, string(data))
		}
	case serverProtocol.MsgIDLeaveGroupResp:
		var resp model.GenericMessageResp
		if err := json.Unmarshal(data, &resp); err == nil {
			if resp.Code == 0 {
				output = fmt.Sprintf("[群组] %s", resp.Message)
			} else {
				output = fmt.Sprintf("[错误] 离开群组操作: %s (code: %d)", resp.Message, resp.Code)
			}
		} else {
			output = fmt.Sprintf("[错误] 解析离开群组响应失败: %v. 内容: %s", err, string(data))
		}
	case serverProtocol.MsgIDHistoryMsgResp:
		var resp model.HistoryMsgResp
		if err := json.Unmarshal(data, &resp); err == nil {
			if resp.Code == 0 {
				var historyOutput strings.Builder
				historyOutput.WriteString("[历史消息]")
				if len(resp.Data) == 0 {
					historyOutput.WriteString("\n  (无历史消息)")
				}
				for i, itemMap := range resp.Data {
					from, _ := itemMap["from_user_id"].(string)
					content, _ := itemMap["content"].(string)
					timestampFloat, _ := itemMap["timestamp"].(float64)
					timestamp := time.Unix(int64(timestampFloat), 0).Format("2006-01-02 15:04:05")

					// Basic content decoding attempt (if it was base64 of simple string)
					// More complex content (e.g. JSON object within content) would need specific handling.
					// decodedContent, err := base64.StdEncoding.DecodeString(content)
					// if err == nil {
					// 	 content = string(decodedContent)
					// }
					historyOutput.WriteString(fmt.Sprintf("\n  %d. [%s] (%s): %s", i+1, from, timestamp, content))
				}
				output = historyOutput.String()
			} else {
				output = fmt.Sprintf("[错误] 获取历史消息失败: %s", resp.Message)
			}
		} else {
			output = fmt.Sprintf("[错误] 解析历史消息响应失败: %v. 内容: %s", err, string(data))
		}
	case serverProtocol.MsgIDErrorResp: // Generic error response from server
		var errResp model.GenericMessageResp
		if err := json.Unmarshal(data, &errResp); err == nil {
			output = fmt.Sprintf("[服务端错误] %s (code: %d)", errResp.Message, errResp.Code)
		} else {
			output = fmt.Sprintf("[服务端错误] 无法解析错误信息: %s", string(data))
		}
	default:
		output = fmt.Sprintf("[通知] 收到消息 ID=%d, 内容='%s'", msgID, string(data))
	}
	if output != "" {
		outputChan <- output
	}
}

func handleRegister(args []string) {
	if !ensureConnected() {
		return
	}
	if len(args) < 3 {
		outputChan <- "用法: /register <username> <password> <email>"
		return
	}
	resp, err := cli.Register(args[0], args[1], args[2])
	if err != nil {
		outputChan <- fmt.Sprintf("注册失败: %v", err)
		return
	}
	outputChan <- fmt.Sprintf("注册成功! 用户 UUID: %s. 请使用 /login 登录。", resp.UserUUID)
}

func handleLogin(args []string) {
	if !ensureConnected() {
		return
	}
	if len(args) < 2 {
		outputChan <- "用法: /login <username> <password>"
		return
	}
	loginResp, err := cli.Login(args[0], args[1])
	if err != nil {
		outputChan <- fmt.Sprintf("登录失败: %v", err)
		return
	}
	// Success message already printed by cli.Login(), or can be added here:
	outputChan <- fmt.Sprintf("登录成功: %s (UUID: %s)", loginResp.Username, loginResp.UserUUID)
}

func handleSendMsg(args []string) {
	if !ensureLoggedIn() {
		return
	}
	if len(args) < 1 {
		outputChan <- "用法: /msg <接收者用户名/UserUUID> [消息内容...]"
		outputChan <- "如果未提供消息内容，将提示您输入。"
		return
	}
	recipient := args[0]

	if len(args) > 1 {
		content := strings.Join(args[1:], " ")
		if strings.TrimSpace(content) == "" {
			outputChan <- "不能发送空消息。"
			return
		}
		cli.SendTextMessage(recipient, content)
	} else {
		// No message content provided on the command line, set up to expect it on the next input
		expectingMessageContentForRecipient = recipient
		outputChan <- fmt.Sprintf("请输入消息内容给 %s (或输入 /cancel 取消):", recipient)
		// The main loop will now handle the next line of input as message content
	}
}

func handleHistory(args []string) {
	if !ensureLoggedIn() {
		return
	}
	if len(args) < 1 {
		outputChan <- "用法: /history <target_user_uuid> [limit]"
		return
	}
	targetUUID := args[0]
	limit := 20 // Default limit for history
	if len(args) > 1 {
		var err error
		limit, err = strconv.Atoi(args[1])
		if err != nil || limit <= 0 {
			outputChan <- "无效的 limit 参数，使用默认值 20。"
			limit = 20
		}
	}
	err := cli.SendHistoryMessageReq(targetUUID, limit)
	if err != nil {
		outputChan <- fmt.Sprintf("发送历史消息请求失败: %v", err)
	} else {
		outputChan <- "历史消息请求已发送。等待响应..."
	}
}

func handleCreateGroup(args []string) {
	if !ensureLoggedIn() {
		return
	}
	if len(args) < 1 {
		outputChan <- "用法: /creategroup <group_name> [description] [avatar_url]"
		return
	}
	name := args[0]
	description := ""
	avatar := ""
	if len(args) > 1 {
		description = args[1]
	}
	if len(args) > 2 {
		avatar = args[2]
	}
	err := cli.SendCreateGroupReq(name, description, avatar)
	if err != nil {
		outputChan <- fmt.Sprintf("创建群组请求发送失败: %v", err)
	} else {
		outputChan <- "创建群组请求已发送。等待响应..."
	}
}

func handleJoinGroup(args []string) {
	if !ensureLoggedIn() {
		return
	}
	if len(args) < 1 {
		outputChan <- "用法: /joingroup <group_id>"
		return
	}
	groupID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		outputChan <- "无效的群组ID。"
		return
	}
	err = cli.SendJoinGroupReq(uint(groupID))
	if err != nil {
		outputChan <- fmt.Sprintf("加入群组请求发送失败: %v", err)
	} else {
		outputChan <- "加入群组请求已发送。等待响应..."
	}
}

func handleLeaveGroup(args []string) {
	if !ensureLoggedIn() {
		return
	}
	if len(args) < 1 {
		outputChan <- "用法: /leavegroup <group_id>"
		return
	}
	groupID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		outputChan <- "无效的群组ID。"
		return
	}
	err = cli.SendLeaveGroupReq(uint(groupID))
	if err != nil {
		outputChan <- fmt.Sprintf("离开群组请求发送失败: %v", err)
	} else {
		outputChan <- "离开群组请求已发送。等待响应..."
	}
}

func handleHelp() {
	outputChan <- "可用命令:"
	outputChan <- "  /connect [host:port] - 连接到服务器 (默认 127.0.0.1:9000)"
	outputChan <- "  /register <username> <password> - 注册新用户"
	outputChan <- "  /login <username> <password> - 登录"
	outputChan <- "  /msg <接收者用户名/UserUUID> [消息内容...] - 发送消息"
	outputChan <- "  /history <对方UserUUID> [页码] [每页条数] - 获取与某人的历史消息"
	outputChan <- "  /creategroup <群名称> - 创建群组"
	outputChan <- "  /joingroup <群ID> - 加入群组"
	outputChan <- "  /leavegroup <群ID> - 离开群组"
	outputChan <- "  /cancel - 取消当前操作 (例如，在输入多行消息时)"
	outputChan <- "  /help - 显示此帮助信息"
	outputChan <- "  /quit 或 /exit - 退出客户端"
}
