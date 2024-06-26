package ws

import (
	"errors"
	"github.com/Arxtect/Einstein/common/constants"
)

var roomService roomSvc

type roomSvc struct {
	Rooms map[string]*room
}

type room struct {
	ID      string
	Content string
}

type EditFileRoomAccess interface {
	NewRoom() (room, error)
	Get(id string) (string, error)
	Update(id string, content string) bool
	Delete(id string) bool
}

func (rs *roomSvc) NewRoom(fileId string) (room, error) {
	if len(rs.Rooms) >= constants.MaxRooms {
		return room{}, errors.New("too many open rooms")
	}

	//newRoom := room{utils.RoomIdCreate(8), ""}

	//roomId 应该替换成文件名,文件名称在minio中是唯一标识
	newRoom := room{fileId, ""}
	if rs.Rooms == nil {
		rs.Rooms = make(map[string]*room, constants.MaxRooms)
	}
	rs.Rooms[newRoom.ID] = &newRoom
	return newRoom, nil
}

func (rs *roomSvc) Update(id string, content string) bool {
	r, ok := rs.Rooms[id]
	if !ok {
		return false
	}
	r.Content = content
	return true
}

func (rs *roomSvc) Get(id string) (string, error) {
	if rs.Rooms == nil {
		return "", errors.New("room is empty")
	}

	r, ok := rs.Rooms[id]
	if !ok {
		return "", errors.New("room not found")
	}
	return r.Content, nil
}

// Delete 每一个文件代表一个房间的连接,如果前端不管在编译前还是编译后删除,都该将房间释放
func (rs *roomSvc) Delete(id string) bool {
	if rs.Rooms == nil {
		return false
	}

	_, ok := rs.Rooms[id]
	if !ok {
		return false
	}

	delete(rs.Rooms, id)
	return true
}
