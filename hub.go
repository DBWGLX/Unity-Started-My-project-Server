// hub.go 维护全局运行时状态：在线玩家表、星星已收集表、广播函数
// 所有对共享状态的读写都必须通过这里，避免到处加锁
package main

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// 全局状态（包级变量，整个进程一份）
var (
	players   = make(map[string]*Player) // playerID -> *Player
	playersMu sync.RWMutex

	starClaimed = make(map[int]bool) // starID -> 是否已被收集
	starMu      sync.Mutex
)

// broadcast 把一条消息发给除 excludeID 以外的所有在线玩家
func broadcast(msg ServerMessage, excludeID string) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("broadcast marshal error: %v", err)
		return
	}

	playersMu.RLock()
	defer playersMu.RUnlock()

	for id, p := range players {
		if id == excludeID {
			continue
		}
		p.Mu.Lock()
		err := p.Conn.WriteMessage(websocket.TextMessage, data)
		p.Mu.Unlock()
		if err != nil {
			// 写失败说明连接已断，留给 ReadMessage 的下次循环触发 defer 清理
			log.Printf("write to %s failed: %v", id, err)
		}
	}
}

// addPlayer 把玩家放进在线表（加写锁）
func addPlayer(p *Player) {
	playersMu.Lock()
	players[p.ID] = p
	playersMu.Unlock()

	dumpPlayers()
}

// removePlayer 把玩家从在线表删掉（加写锁）
func removePlayer(playerID string) {
	playersMu.Lock()
	delete(players, playerID)
	playersMu.Unlock()

	dumpPlayers()
}

// snapshotPlayers 返回当前所有在线玩家的快照
// ⚠️ 调用方不能已经持有 playersMu 或任意 p.Mu，否则会死锁
func snapshotPlayers() []Player {
	playersMu.RLock()
	defer playersMu.RUnlock()

	result := make([]Player, 0, len(players))
	for id, p := range players {
		p.Mu.Lock()
		result = append(result, Player{
			ID:   id,
			Name: p.Name,
			X:    p.X,
			Y:    p.Y,
			Z:    p.Z,
		})
		p.Mu.Unlock()
	}
	return result
}

// dumpPlayers 打印当前所有在线玩家的信息
func dumpPlayers() {
	curPlayers := snapshotPlayers()

	if len(curPlayers) == 0 {
		log.Println("[dumpPlayers] 当前无在线玩家")
		return
	}

	log.Printf("[dumpPlayers] 当前在线玩家数量: %d", len(curPlayers))
	for id, p := range curPlayers {
		p.Mu.RLock()
		log.Printf("id=%s name=%s",
			id, p.Name)

		p.Mu.RUnlock()
	}
}
