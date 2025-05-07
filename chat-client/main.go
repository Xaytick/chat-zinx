package main

import (
    "encoding/binary"
    "encoding/json"
    "fmt"
    "net"
)

func main() {
    conn, err := net.Dial("tcp", "127.0.0.1:9000")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    // 构造登录请求
    req := map[string]string{
        "username": "test",
        "password": "123456",
    }
    body, _ := json.Marshal(req)
    msgID := uint32(1) // MsgIDLoginReq
    length := uint32(len(body))

    // 组包
    buf := make([]byte, 8+len(body))
    binary.LittleEndian.PutUint32(buf[0:4], length)
    binary.LittleEndian.PutUint32(buf[4:8], msgID)
    copy(buf[8:], body)

    // 发送
    conn.Write(buf)

    // 读取响应
    head := make([]byte, 8)
    conn.Read(head)
    respLen := binary.LittleEndian.Uint32(head[0:4])
    respBody := make([]byte, respLen)
    conn.Read(respBody)
    fmt.Println("服务器响应：", string(respBody))
}