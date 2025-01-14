package routes

import (
	"github.com/Arxtect/Einstein/apps/archive/controllers"
	"github.com/Arxtect/Einstein/common/middleware"
	"github.com/gin-gonic/gin"
)

type ChatRouteController struct {
	chatController controllers.ChatController
}

func NewChatRouteController(chatController controllers.ChatController) ChatRouteController {
	return ChatRouteController{chatController}
}

func (crc *ChatRouteController) ChatRoute(rg *gin.RouterGroup) {

	chat := rg.Group("chat").Use(middleware.DeserializeUser())
	{
		chat.POST("/completion_with_model_info", crc.chatController.CompletionWithModelInfo)
	}
}
