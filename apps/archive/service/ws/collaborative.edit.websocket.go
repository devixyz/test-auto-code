package ws

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

type subscribeRequest struct {
	RoomID   string `json:"roomId"`
	UserID   string `json:"userId"`
	RoleID   string `json:"roleID"`
	UserName string `json:"userName"`
}

// HandlerWs 处理websocket请求
func HandlerWs(c *gin.Context) {
	ws := c.Writer
	req := c.Request

	websocket.Handler(func(conn *websocket.Conn) {
		defer conn.Close()

		for {
			var message string

			// 接收客户端发来的消息
			if err := websocket.Message.Receive(conn, &message); err != nil {
				// 如果读取出错，可能是连接关闭
				fmt.Printf("Error reading from WebSocket: %v\n", err)
				break
			}

			// 收到消息,做出反应(根据发送来的消息,做出不同的反应) 1.收到客户端发来的内容,需要转发给订阅的客户端

			var subReq subscribeRequest
			if err := json.Unmarshal([]byte(message), &subReq); err != nil {
				fmt.Printf("Error unmarshalling subscribe request: %v\n", err)
				continue
			}

			subscribers, err := ssList.GetSubscribers(subReq.RoomID)
			if err != nil {
				fmt.Printf("Error getting subscribers for room: %v\n", err)
				continue
			}

			for _, subscriber := range subscribers {
				if subscriber.Connection != conn {
					// 避免将消息发回给发送者
					if err := websocket.Message.Send(subscriber.Connection, message); err != nil {
						fmt.Printf("Error sending message to subscriber: %v\n", err)
						// 可以考虑在这里执行取消订阅等清理操作
						continue
					}
				}
			}
		}
	}).ServeHTTP(ws, req)
}
