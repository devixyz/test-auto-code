package utils

import "math/rand"

const chars = "abcdefghijklmnopqrstuvwxyz0123456789"

// RoomIdCreate  随机生成roomId
func RoomIdCreate(n int) string {
	id := ""
	for i := 0; i < n; i++ {
		id += string(chars[rand.Intn(len(chars))])
	}
	return id
}
