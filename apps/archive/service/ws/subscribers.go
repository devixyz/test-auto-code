package ws

import (
	"errors"
	"fmt"
	"github.com/Arxtect/Einstein/apps/archive/models"
	"golang.org/x/net/websocket"
	"sync"
	"time"
)

var ssList subscriptionSvc

type subscriptionSvc struct {
	Subscribers map[string][]subscription
	RoomOwners  map[string][]string
	mu          sync.RWMutex
}

type subscription struct {
	UserInfo     *models.User
	LastActivity time.Time // 添加最后活动时间字段
	Connection   *websocket.Conn
}

type Subscriber interface {
	All() map[string][]subscription
	Subscribe(roomID string, s subscription) error
	Unsubscribe(roomID string, s subscription) error
	GetSubscribers(roomID string) ([]subscription, error)
	IsOwner(roomID, userID string) bool
	AddUserAsOwner(roomID, userID string) error
	RemoveUserAsOwner(roomID, userID string) error
	// DeleteRoomSubscribers 删除指定房间的所有订阅者
	DeleteRoomSubscribers(roomID string) error

	// Forward
	BroadcastContentUpdates(roomID string, content string)
	StartSubscriberListener()
}

func GetSubscriber() Subscriber {
	return &ssList
}

func (subSvc *subscriptionSvc) All() map[string][]subscription {
	subSvc.mu.RLock()
	defer subSvc.mu.RUnlock()
	return subSvc.Subscribers
}

func (subSvc *subscriptionSvc) Subscribe(roomID string, s subscription) error {
	subSvc.mu.Lock()
	defer subSvc.mu.Unlock()
	if subSvc.Subscribers == nil {
		subSvc.Subscribers = make(map[string][]subscription)
	}
	existingSubs, ok := subSvc.Subscribers[roomID]
	if ok {
		// 附加到现有订阅者列表
		subSvc.Subscribers[roomID] = append(existingSubs, s)
	} else {
		// 初始化订阅者列表
		subSvc.Subscribers[roomID] = []subscription{s}
	}
	return nil
}

func (subSvc *subscriptionSvc) GetSubscribers(roomID string) ([]subscription, error) {
	subSvc.mu.RLock()
	defer subSvc.mu.RUnlock()
	if subSvc.Subscribers == nil {
		return []subscription{}, nil
	}
	existingSubs, ok := subSvc.Subscribers[roomID]
	if ok {
		return existingSubs, nil
	}
	// 这个房间没有订阅
	return []subscription{}, nil
}

func (subSvc *subscriptionSvc) Unsubscribe(roomID string, s subscription) error {
	subSvc.mu.Lock()
	defer subSvc.mu.Unlock()
	if subSvc.Subscribers == nil {
		return nil
	}
	existingSubs, ok := subSvc.Subscribers[roomID]
	if !ok {
		return errors.New("room not found")
	}
	for index, ex := range existingSubs {
		if s.Connection == ex.Connection && s.UserInfo.ID.String() == ex.UserInfo.ID.String() && s.UserInfo.Name == ex.UserInfo.Name {
			// 找到了应删除的订阅.复制原始订阅,但跳过当前元素
			subSvc.Subscribers[roomID] = append(existingSubs[:index], existingSubs[index+1:]...)
			return nil
		}
	}
	return errors.New("could not find the subscription")
}

func (subSvc *subscriptionSvc) IsOwner(roomID, userID string) bool {
	subSvc.mu.RLock()
	defer subSvc.mu.RUnlock()
	ownerIDs, ok := subSvc.RoomOwners[roomID]
	if !ok {
		return false
	}
	for _, id := range ownerIDs {
		if id == userID {
			return true
		}
	}
	return false
}

func (subSvc *subscriptionSvc) AddUserAsOwner(roomID, userID string) error {
	subSvc.mu.Lock()
	defer subSvc.mu.Unlock()
	if subSvc.RoomOwners == nil {
		subSvc.RoomOwners = make(map[string][]string)
	}
	ownerIDs, ok := subSvc.RoomOwners[roomID]
	if ok {
		subSvc.RoomOwners[roomID] = append(ownerIDs, userID)
	} else {
		subSvc.RoomOwners[roomID] = []string{userID}
	}
	return nil
}

func (subSvc *subscriptionSvc) RemoveUserAsOwner(roomID, userID string) error {
	subSvc.mu.Lock()
	defer subSvc.mu.Unlock()
	ownerIDs, ok := subSvc.RoomOwners[roomID]
	if !ok {
		return errors.New("room not found")
	}
	for i, id := range ownerIDs {
		if id == userID {
			subSvc.RoomOwners[roomID] = append(ownerIDs[:i], ownerIDs[i+1:]...)
			return nil
		}
	}
	return errors.New("could not find the user as owner")
}

func (subSvc *subscriptionSvc) DeleteRoomSubscribers(roomID string) error {
	subSvc.mu.Lock()
	defer subSvc.mu.Unlock()
	// 删除订阅者列表
	delete(subSvc.Subscribers, roomID)
	delete(subSvc.RoomOwners, roomID)

	return nil
}

// RoomContentUpdate ---------------------------
// RoomContentUpdate 表示房间内容更新的消息 (待商榷)
type RoomContentUpdate struct {
	RoomID  string
	Content string
}

// 初始化内容更新广播通道
var broadcast = make(chan RoomContentUpdate, 100)

// BroadcastContentUpdates 用于通知所有订阅者房间内容更新
func (subSvc *subscriptionSvc) BroadcastContentUpdates(roomID string, content string) {
	update := RoomContentUpdate{RoomID: roomID, Content: content}
	broadcast <- update
}

// StartSubscriberListener 启动订阅者监听
func (subSvc *subscriptionSvc) StartSubscriberListener() {
	for {
		select {
		case update := <-broadcast:
			subSvc.mu.RLock()
			subs, ok := subSvc.Subscribers[update.RoomID]
			subSvc.mu.RUnlock()
			if ok {
				// 广播消息给所有订阅者
				for _, sub := range subs {
					go func(s subscription) {
						// 发送更新内容到订阅者的WebSocket连接
						err := websocket.Message.Send(s.Connection, update.Content)
						if err != nil {
							// 处理发送错误
							fmt.Printf("Error sending update to user %s: %v\n", s.UserInfo.ID, err)
						}
					}(sub)
				}
			}
		}
	}
}
