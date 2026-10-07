ws-game-demo/
├── go.mod
├── main.go       # 入口：启动服务
├── model.go      # 数据结构：所有 struct 定义
├── hub.go        # 全局状态：玩家表、星星表、广播
├── handler.go    # 连接生命周期：WebSocket 升级 + 读循环
├── message.go    # 业务逻辑：join / move / collect 三个 handler
└── util.go       # 工具函数：generateID



handler 接收 
message 组织处理
hub 响应

model 存玩家状态


<img width="3840" height="1080" alt="image" src="https://github.com/user-attachments/assets/ae62785c-b152-4044-aaed-a75a0e55beaa" />


# 构建
go run .
go build -o server && ./server


