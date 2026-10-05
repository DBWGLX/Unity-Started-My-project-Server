// handler.go 负责 WebSocket 连接的生命周期：升级、读消息、断线清理
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// 升级为WebSocket：upgrader 把普通 HTTP 连接升级成 WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Demo 阶段允许所有来源
}

// handleConnection 是 /ws 的入口，每个新连接开一个 goroutine 跑这里
func handleConnection(w http.ResponseWriter, r *http.Request) {
	// 1.升级协议
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}

	// 2.自己信息
	player := &Player{
		ID:   generateID(),
		Conn: conn,
	}

	// 断线清理：无论下面 ReadMessage 因为什么退出，都会执行
	defer func() {
		removePlayer(player.ID)
		_ = conn.Close()
		broadcast(ServerMessage{
			Type: "playerLeft",
			Data: map[string]string{"id": player.ID},
		}, "")
	}()

	// 3.读循环：一条一条读客户端消息
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			log.Printf("read message error: %v", err)
			break // 读失败 = 连接断了，跳出循环触发 defer
		}

		// 测试阶段：打印原始报文
		log.Printf("recv raw: %s", raw)

		var msg ClientMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("unmarshal message error: %v, raw=%s", err, raw)
			continue
		}

		//处理消息
		dispatchMessage(player, msg)
	}
}
