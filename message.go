// message.go 负责消息分发和每个 type 的具体业务逻辑
// 每个函数只做一件事：解析参数 → 改状态 → 广播
package main

import (
	"encoding/json"
	"log"
)

// dispatchMessage 根据 type 把消息派发到对应 handler
func dispatchMessage(p *Player, msg ClientMessage) {
	switch msg.Type {
	case "join":
		handleJoin(p, msg.Data)
	case "move":
		handleMove(p, msg.Data)
	case "collect":
		handleCollect(p, msg.Data)
	default:
		log.Printf("unknown message type: %s", msg.Type)
	}
}

// 二次 JSON 中转
// parseData 通用：把 interface{} 再 marshal/unmarshal 成具体结构体
// 因为 msg.Data 是从 JSON 解析出来的 map[string]interface{}，要转成强类型
func parseData(raw interface{}, dst interface{}) error {
	b, err := json.Marshal(raw) //map[string]interface{}
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

// handleJoin 玩家上线：登记名字、放进在线表、欢迎消息、广播新玩家
func handleJoin(p *Player, data interface{}) {
	var d JoinData
	if err := parseData(data, &d); err != nil {
		log.Printf("join parse error: %v", err)
		return
	}

	p.Mu.Lock()
	p.Name = d.Name
	p.X, p.Y, p.Z = 0, 0, 0
	p.Mu.Unlock()

	addPlayer(p)

	//初始化其他玩家
	snapshot := snapshotPlayers()

	// 给自己发 welcome，带上自己的 ID
	// （这里简化，实际可以把全量在线玩家列表一起推过来）
	p.Mu.Lock()
	msg := ServerMessage{
		Type: "playerInitiate",
		Data: InitiatePlayerData{
			SelfID:  p.ID,
			Players: snapshot,
		},
	}
	_ = p.Conn.WriteJSON(msg)
	p.Mu.Unlock()

	// 广播给其他人：有新人上线
	broadcast(ServerMessage{
		Type: "playerUpdate",
		Data: map[string]interface{}{
			"id":   p.ID,
			"name": p.Name,
		},
	}, p.ID)
}

// handleMove 更新玩家位置，并广播给其他人
func handleMove(p *Player, data interface{}) {
	var d MoveData
	if err := parseData(data, &d); err != nil {
		log.Printf("move parse error: %v", err)
		return
	}

	p.Mu.Lock()
	p.X, p.Y, p.Z = d.X, d.Y, d.Z
	p.Mu.Unlock()

	broadcast(ServerMessage{
		Type: "playerUpdate",
		Data: map[string]interface{}{
			"id":   p.ID,
			"name": p.Name,
			"x":    d.X,
			"y":    d.Y,
			"z":    d.Z,
		},
	}, p.ID) // 排除自己，位置自己最清楚
}

// handleCollect 星星收集的服务器端仲裁（防作弊/防重复）
func handleCollect(p *Player, data interface{}) {
	var d CollectData
	if err := parseData(data, &d); err != nil {
		log.Printf("collect parse error: %v", err)
		return
	}

	starMu.Lock()
	if starClaimed[d.StarID] {
		starMu.Unlock()
		return // 已被别人抢先收集，忽略
	}
	starClaimed[d.StarID] = true
	starMu.Unlock()

	// 广播给所有人（包括收集者自己，便于客户端确认）
	broadcast(ServerMessage{
		Type: "starCollected",
		Data: map[string]interface{}{
			"starId":   d.StarID,
			"playerId": p.ID,
		},
	}, "")
}
