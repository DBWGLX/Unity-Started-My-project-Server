// Package main 是整个多人小游戏 WebSocket 服务的入口
// 只负责：初始化模块变量、注册路由、启动 HTTP 服务
package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/ws", handleConnection)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
