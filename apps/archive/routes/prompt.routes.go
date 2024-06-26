package routes

import (
	"github.com/Arxtect/Einstein/apps/archive/controllers"
	"github.com/gin-gonic/gin"
)

type PromptController struct {
	promptController controllers.PromptController
}

func NewPromptRouteController(promptController controllers.PromptController) PromptController {
	return PromptController{promptController}
}

func (dc *PromptController) PromptRoute(rg *gin.RouterGroup) {
	// .Use(middleware.DeserializeUser())
	prompt := rg.Group("prompt")
	{
		prompt.GET("list", dc.promptController.GetPromptList)
		prompt.GET("/:id", dc.promptController.GetPrompt)
		prompt.POST("", dc.promptController.CreatePrompt)
		prompt.PUT("", dc.promptController.UpdatePrompt)
		prompt.DELETE("/:id", dc.promptController.DeletePrompt)
	}

}
