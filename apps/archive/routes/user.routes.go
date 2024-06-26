package routes

import (
	"github.com/Arxtect/Einstein/apps/archive/controllers"
	"github.com/Arxtect/Einstein/common/middleware"
	"github.com/gin-gonic/gin"
)

type UserRouteController struct {
	userController controllers.UserController
}

func NewRouteUserController(userController controllers.UserController) UserRouteController {
	return UserRouteController{userController}
}

func (uc *UserRouteController) UserRoute(rg *gin.RouterGroup) {

	router := rg.Group("users").Use(middleware.DeserializeUser())
	router.GET("/me", uc.userController.GetMe)
	router.POST("/admin_update_balance", uc.userController.AdminUpdateBalance)

	//router.POST("/admin_bulk_create_recharging_cards", uc.userController.AdminBulkCreateRechargingCards)
	//router.GET("/admin_get_recharging_cards", uc.userController.AdminFindAllCards)
	//router.DELETE("/admin_deactivate_recharging_cards/:rechargingCardId", uc.userController.AdminDeactivateCard)
	//router.POST("/recharge_with_recharging_card", uc.userController.RechargeMyselfWithRechargingCard)

	routerWs := rg.Group("ws").Use(middleware.DeserializeUser())
	{
		routerWs.POST("/establishWs", uc.userController.WsEditingRoom) // 1.用户协同编辑socket
		routerWs.POST("/room/:fileId", uc.userController.CreateRoom)   // 2.创建文件房间
		//routerWs.POST("/room/subscribe", uc.userController.WsSubscribeEdit)     // 3.订阅加入文件房间中,协同编辑
		//routerWs.POST("/room/unsubscribe", uc.userController.WsUnsubscribeEdit) // 4.取消订阅加入文件房间中,协同编辑
	}

}
