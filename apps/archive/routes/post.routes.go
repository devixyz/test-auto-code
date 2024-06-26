package routes

import (
	"github.com/Arxtect/Einstein/apps/archive/controllers"
	"github.com/Arxtect/Einstein/common/middleware"
	"github.com/gin-gonic/gin"
)

type PostRouteController struct {
	postController controllers.PostController
}

func NewRoutePostController(postController controllers.PostController) PostRouteController {
	return PostRouteController{postController}
}

func (pc *PostRouteController) PostRoute(rg *gin.RouterGroup) {

	router := rg.Group("posts").Use(middleware.DeserializeUser())
	router.POST("", pc.postController.CreatePost)
	router.GET("", pc.postController.FindPosts)
	router.PUT("/:postId", pc.postController.UpdatePost)
	router.GET("/getLatestPost", pc.postController.FindLatestPost)
	router.DELETE("/:postId", pc.postController.DeletePost)
}
