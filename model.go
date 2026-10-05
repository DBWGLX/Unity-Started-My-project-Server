// model.go 集中定义所有数据结构（DTO / 领域模型）
// 不写任何业务逻辑，只声明字段和 JSON tag
package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Player 代表一个在线玩家
type Player struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	X, Y, Z float32         `json:"-"` // 位置不直接序列化，消息里单独打包
	Conn    *websocket.Conn `json:"-"`
	Mu      sync.RWMutex    `json:"-"`
}

// ---------- 客户端 → 服务器 ----------

// ClientMessage 外层信封：type 决定走哪个 handler，data 是具体业务参数
type ClientMessage struct {
	Type string      `json:"type"` // "join" | "move" | "collect"
	Data interface{} `json:"data"`
}

// JoinData 玩家上线
type JoinData struct {
	Name string `json:"name"`
}

// MoveData 玩家移动
type MoveData struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

// CollectData 收集星星
type CollectData struct {
	StarID int `json:"starId"`
}

// ---------- 服务器 → 客户端 ----------

// ServerMessage 服务器下发消息的统一信封
type ServerMessage struct {
	Type string      `json:"type"` // "welcome" | "playerUpdate" | "starCollected" | "playerLeft"
	Data interface{} `json:"data"`
}

type InitiatePlayerData struct {
	SelfID  string   `json:"selfId"`
	Players []Player `json:"players"`
}
