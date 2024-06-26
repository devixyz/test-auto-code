package ws

import (
	"github.com/Arxtect/Einstein/common/xminio"
	"github.com/Arxtect/Einstein/config"
	"github.com/Arxtect/Einstein/utils"
	"io"
	"io/ioutil"
	"net/http"
)

func HandleGetStaticResource(filename string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(415)
			return
		}
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			utils.WriteJSON(w, 400, err.Error())
			return
		}
		w.WriteHeader(200)
		io.WriteString(w, string(b))
	}
}
func HandleCreateRoom(fileId string) (string, string, error) {
	// fileId 文件名 是否存在,如果不存在该返回400
	existence, errByMinio := xminio.NewS3Manager(config.Env.MinioBucket, config.Env.MinioAccessKey, config.Env.MinioSecretKey, config.Env.MinioBucketUrl).CheckFileExistence(fileId)
	if !existence || errByMinio != nil {
		return "", "", errByMinio
	}

	room, err := roomService.NewRoom(fileId)
	if err != nil {
		return "", "", err
	}
	// 邀请码 ， Invitation作为AES的密钥
	Invitation := utils.RoomIdCreate(6)
	// 返回加密后的邀请码
	return room.ID, Invitation, nil
}
