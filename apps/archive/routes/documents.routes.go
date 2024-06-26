package routes

import (
	"github.com/Arxtect/Einstein/apps/archive/controllers"
	"github.com/Arxtect/Einstein/common/middleware"
	"github.com/gin-gonic/gin"
)

type DocumentsController struct {
	documentController controllers.DocumentsController
}

func NewDocumentRouteController(documentController controllers.DocumentsController) DocumentsController {
	return DocumentsController{documentController}
}

func (dc *DocumentsController) DocumentRoute(rg *gin.RouterGroup) {
	documents := rg.Group("documents").Use(middleware.DeserializeUser())
	{
		documents.POST("/upload", dc.documentController.UploadDocumentsByUser, middleware.Sentinel()) // 1.上传文档,并给文档打上标签
		documents.GET("/drafts", dc.documentController.GetDraftsByUser)                               // 1.1 获取用户的上传的全部文档ID
		documents.GET("/drafts/:key", dc.documentController.GetDocumentByKey)                         // 1.2 获取用户的上传的单个文档
		documents.POST("/gen/commitInfo", dc.documentController.GenCommitDocument)                    // 1.3 ai生成commit信息
	}

	// no auth
	documentsNoauth := rg.Group("documents")
	{
		documentsNoauth.GET("/tags/list", dc.documentController.GetDocumentsAllTags)           // 2.获取全部的标签
		documentsNoauth.GET("/list/search", dc.documentController.GetDocumentsListSearch)      // 4.获取文档列表(支持全文关键字搜索)
		documentsNoauth.GET("/list/search-v2", dc.documentController.GetDocumentsListSearchV2) // 4.1获取文档列表(meili全文关键字搜索)

		documentsNoauth.GET("/pre/download/:key", dc.documentController.GetDocumentDownloadUrl) // 5.1根据文件名称拿到预下载url
		documentsNoauth.GET("/pre/preview/:key", dc.documentController.PreViewFile)             // 5.2 浏览器预览文档
	}
}
