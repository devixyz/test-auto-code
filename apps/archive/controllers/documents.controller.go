package controllers

import (
	"context"
	"github.com/Arxtect/Einstein/apps/archive/models"
	"github.com/Arxtect/Einstein/apps/archive/service"
	"github.com/Arxtect/Einstein/apps/archive/service/dto"
	"github.com/Arxtect/Einstein/common/constants"
	openai_config "github.com/Arxtect/Einstein/common/openai-config"
	"github.com/Arxtect/Einstein/common/xminio"
	"github.com/Arxtect/Einstein/config"
	"github.com/Arxtect/Einstein/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	gogpt "github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"sync"
	"time"
)

type DocumentsController struct {
	DB          *gorm.DB
	Logger      *log.Logger
	RedisDb     *redis.Client
	MeiliClient *meilisearch.Client
}

func NewDocumentController(DB *gorm.DB, logger *log.Logger, redisDb *redis.Client, meili *meilisearch.Client) DocumentsController {
	return DocumentsController{
		DB,
		logger,
		redisDb,
		meili,
	}
}

// UploadDocumentsByUser 来自用户的上传文档
func (dc *DocumentsController) UploadDocumentsByUser(c *gin.Context) {
	currentUser := c.MustGet("currentUser").(models.User)

	tagsStr := c.PostForm("tags")
	tagSlice, err := utils.ParseTags(tagsStr)
	if err != nil {
		dc.Logger.Printf("Error parsing tags: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Error parsing tags"})
		return
	}

	// 使用PostForm获取其他字段
	req := dto.AddTagToDocumentRequest{
		Title:      c.PostForm("title"),
		Tags:       tagSlice,
		Content:    c.PostForm("content"),
		UploadType: c.PostForm("upload_type"),
	}

	// 获取文件名称
	metaDataFile, err := c.FormFile("file")
	fileName := metaDataFile.Filename
	hash := currentUser.ID.String() + "-" + fileName

	metaDataZip, err := c.FormFile("zip")
	fileNameZip := metaDataZip.Filename
	hashZip := currentUser.ID.String() + "-" + fileNameZip

	metaDataFileCover, err := c.FormFile("cover")
	hashCover := currentUser.ID.String() + "-" + "cover" + "-" + metaDataFileCover.Filename

	if err != nil {
		dc.Logger.Printf("Error getting file from form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Error getting file from form"})
		return
	}

	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "The title  cannot be less"})
		return
	}

	// 都默认上传minio，出现在桶内相同的文件名，会将之前的替换掉（自动执行的）。
	var fileResponse dto.FileResponse
	var zipResponse dto.FileResponse
	var coverResponse dto.FileResponse
	switch req.UploadType {
	//// 多文件上传
	//case "2":
	//	var done bool
	//	multipartFile, done := dc.multipleFile(c)
	//	if !done {
	//		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "upload failed"})
	//		return
	//	}
	//	// currentUser.Email:文件名,存储用户的草稿箱
	//	for _, file := range multipartFile {
	//		_ = dc.RedisDb.SAdd(context.Background(), currentUser.Email, file.FileStorageID).Err()
	//	}
	//
	//	c.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"file": multipartFile}})
	// 默认执行单文件上传
	default:
		var wg sync.WaitGroup
		var doneFile, doneZip, doneCover bool

		wg.Add(3)
		go func() {
			defer wg.Done()
			fileResponse, doneFile = SingleFile(hash, metaDataFile)
		}()

		go func() {
			defer wg.Done()
			zipResponse, doneZip = SingleFile(hashZip, metaDataZip)
		}()

		go func() {
			defer wg.Done()
			coverResponse, doneCover = SingleFile(hashCover, metaDataFileCover)
		}()

		wg.Wait()
		if !doneFile || !doneZip || !doneCover {
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "upload failed"})
			return
		}
	}

	//  对于数据库,判断hash是否存在,存在可修改title,tags,content,不存在则创建
	docIs, errByStorageKey := service.GetDocumentByFileHash(hash)
	if errByStorageKey == nil && len(req.Tags) > 0 {
		// 存在,只修改title,tags
		tagsByName, errByName := service.GetTagsByName(req.Tags)
		if errByName != nil {
			dc.Logger.Printf("Error getting tags from DB: %v", errByName)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Error getting tags"})
			return
		}
		err = service.UpdateDocumentTitleAndTags(docIs, req.Title, tagsByName, req.Content)
		if err != nil {
			dc.Logger.Printf("Error updating document: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Error updating document"})
			return
		}

		go func() {
			index := dc.MeiliClient.Index(constants.MeiliIndexDocuments)
			documents := []map[string]interface{}{
				{"id": docIs.ID, "title": docIs.Title, "content": docIs.Content, "tags": req.Tags, "author": docIs.AuthorID, "cover": docIs.Cover, "poster": docIs.Cover, "file_size": docIs.StorageSize, "file_storage_bucket": docIs.StorageBucket, "file_storage_id": docIs.StorageKey, "file_hash": docIs.FileHash},
			}
			_, _ = index.UpdateDocuments(documents, "id")

		}()
		c.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"fileHash": hash}})
		return
	}

	//  打上标签入库,可以是现有标签,也可以是自定义标签
	tagsByName, errByName := service.GetTagsByName(req.Tags)
	if errByName != nil {
		dc.Logger.Printf("Error getting tags from DB: %v", errByName)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Error getting tags"})
		return
	}
	for i, tag := range tagsByName {
		tagsByName[i].ID = tag.ID
	}
	doc := models.Document{
		Title:         req.Title,
		Content:       req.Content,
		AuthorID:      currentUser.ID,
		StorageBucket: config.Env.MinioBucket,
		StorageKey:    fileResponse.FileStorageID,
		Tags:          tagsByName,
		FileHash:      hash,
		StorageSize:   zipResponse.FileSize,
		StorageZip:    zipResponse.FileStorageID,
		Cover:         coverResponse.FileStorageID,
		Base: models.Base{
			CreatedAt: time.Now(),
		},
	}
	err = dc.DB.Model(&doc).Create(&doc).Error
	if err != nil {
		dc.Logger.Printf("Error creating document: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Error creating document"})
		return
	}
	_ = dc.RedisDb.SAdd(context.Background(), constants.RedisKeyDocuments+currentUser.Email, fileResponse.FileStorageID).Err()

	// TODO 将数据备份到meili
	go func() {
		index := dc.MeiliClient.Index(constants.MeiliIndexDocuments)
		documents := []map[string]interface{}{
			{"id": doc.ID, "title": doc.Title, "content": doc.Content, "tags": req.Tags, "author": doc.AuthorID, "cover": doc.Cover, "poster": doc.Cover, "file_size": doc.StorageSize, "file_storage_bucket": doc.StorageBucket, "file_storage_id": doc.StorageKey, "file_hash": doc.FileHash},
		}
		_, _ = index.AddDocuments(documents, "id")
	}()

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"file": fileResponse}})

}

// SingleFile 单文件上传
func SingleFile(hash string, files *multipart.FileHeader) (dto.FileResponse, bool) {
	f, _ := files.Open()
	defer f.Close()

	// 检查文件类型
	buffer := make([]byte, 512) // 512 bytes should be enough for most formats
	f.Read(buffer)
	f.Seek(0, 0) // Rewind the file after reading
	//contentType := http.DetectContentType(buffer)
	//if contentType != "application/pdf" && contentType != "application/zip" && contentType != "application/x-zip-compressed" {
	//	return dto.FileResponse{}, false
	//}

	content, err := ioutil.ReadAll(f)
	if err != nil {
		return dto.FileResponse{}, false
	}
	fileInfo := xminio.NewS3Manager(config.Env.MinioBucket, config.Env.MinioAccessKey, config.Env.MinioSecretKey, config.Env.MinioBucketUrl).
		UploadByteData(content, hash)

	fileResponse := dto.FileResponse{
		FileStorageID:     fileInfo.Key,
		FileName:          fileInfo.Key,
		FileSize:          fileInfo.Size,
		FileStorageBucket: fileInfo.Bucket,
	}

	return fileResponse, true
}

// multipleFile 多文件上传
func (dc *DocumentsController) multipleFile(c *gin.Context) ([]dto.FileResponse, bool) {
	form, _ := c.MultipartForm()
	files := form.File["file[]"]

	var fileArr []dto.FileResponse
	manager := xminio.NewS3Manager(config.Env.MinioBucket, config.Env.MinioAccessKey, config.Env.MinioSecretKey, config.Env.MinioBucketUrl)

	for _, file := range files {
		f, _ := file.Open()
		defer f.Close()

		content, _ := ioutil.ReadAll(f)
		data := manager.UploadByteData(content, file.Filename+"-"+uuid.New().String()[0:6])
		fileResponse := dto.FileResponse{
			FileStorageID:     data.Key,
			FileName:          data.Key,
			FileSize:          data.Size,
			FileStorageBucket: data.Bucket,
		}
		fileArr = append(fileArr, fileResponse)
	}
	return fileArr, true

}

// GetDocumentsAllTags 获取全部标签
func (dc *DocumentsController) GetDocumentsAllTags(ctx *gin.Context) {
	var tags []models.Tag
	err := dc.DB.Model(&tags).Find(&tags).Error
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Failed to get tags"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"tags": tags}})
}

// GetDraftsByUser 根据标签获取草稿箱
func (dc *DocumentsController) GetDraftsByUser(c *gin.Context) {
	currentUser := c.MustGet("currentUser").(models.User)

	fileIDs, err := dc.RedisDb.SMembers(context.Background(), constants.RedisKeyDocuments+currentUser.Email).Result()
	if err != nil {
		log.Println("Error getting from Redis: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Error getting documents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"documents": fileIDs}})

}

// GetDocumentsListSearch 获取文档列表(支持全文关键字搜索)
func (dc *DocumentsController) GetDocumentsListSearch(c *gin.Context) {
	req := dto.GetDocumentsByUserRequest{}
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid request"})
		return
	}

	var documents []models.Document

	// 添加全文搜索条件
	query := dc.DB.Model(&models.Document{})
	if req.Keyword != "" {
		// 使用Postgres的全文搜索函数
		query = query.Where("to_tsvector('simple', content) @@ plainto_tsquery('simple', ?)", req.Keyword)
	}

	// 添加标签搜索条件
	if len(req.Tags) > 0 {
		query = query.Joins("JOIN document_tags ON document_tags.document_id = documents.id").
			Joins("JOIN tags ON tags.id = document_tags.tag_id").
			Where("tags.name IN (?)", req.Tags)
	}
	var count int64

	err = query.Preload("User").Preload("Tags").Scopes(
		dto.Paginate(req.GetPageSize(), req.GetPageIndex()),
	).Order("created_at desc").
		Count(&count).
		Find(&documents).Error
	if err != nil {
		dc.Logger.Printf("Error getting documents from DB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Error getting documents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"documents": documents, "total": count, "pageIndex": req.GetPageIndex(), "pageSize": req.GetPageSize()}})
}

// GetDocumentDownloadUrl 获取文档下载地址
func (dc *DocumentsController) GetDocumentDownloadUrl(c *gin.Context) {
	//req := dto.GetDocumentDownloadUrlRequest{}
	//err := c.ShouldBindJSON(&req)
	//if err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid request"})
	//	return
	//}

	// 从 URL 参数中获取文档的 key
	key := c.Param("key")

	manager := xminio.NewS3Manager(config.Env.MinioBucket, config.Env.MinioAccessKey, config.Env.MinioSecretKey, config.Env.MinioBucketUrl)
	manager.DownloadObject(c.Writer, key)
}

// GetDocumentByKey  获取用户的上传的单个文档
func (dc *DocumentsController) GetDocumentByKey(c *gin.Context) {
	// 从 URL 参数中获取文档的 key
	key := c.Param("key")

	// 从数据库中查找文档
	document, err := service.GetDocumentByKey(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "Document not found"})
		return
	}

	// 如果找到了文档，则返回文档
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": document})
}

// PreViewFile 浏览器预览文档
func (dc *DocumentsController) PreViewFile(c *gin.Context) {
	key := c.Param("key")

	manager := xminio.NewS3Manager(config.Env.MinioBucket, config.Env.MinioAccessKey, config.Env.MinioSecretKey, config.Env.MinioBucketUrl)

	manager.ServeObject(c.Writer, key)
}

// GetDocumentsListSearchV2 获取文档列表(支持全文关键字搜索)
func (dc *DocumentsController) GetDocumentsListSearchV2(c *gin.Context) {
	req := dto.GetDocumentsByUserRequest{}
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid request"})
		return
	}

	//var documents []models.Document

	searchResult, err := dc.MeiliClient.Index(constants.MeiliIndexDocuments).Search(req.Keyword, &meilisearch.SearchRequest{
		Page:        1,
		HitsPerPage: 10,
	})
	if err != nil {
		dc.Logger.Printf("Error getting documents from MeiliSearch: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Error getting documents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"documents": searchResult, "pageIndex": req.GetPageIndex(), "pageSize": req.GetPageSize()}})
}

// GenCommitDocument 生成commitInfo信息
func (dc *DocumentsController) GenCommitDocument(c *gin.Context) {
	//currentUser := c.MustGet("currentUser").(models.User)

	var req dto.GenCommitDocumentRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid request"})
		return
	}
	configCopy := config.Env
	cnf := openai_config.OpenAIConfiguration{
		ApiKey:        configCopy.ApiKey,
		ApiURL:        configCopy.ApiURL,
		Listen:        configCopy.Listen,
		Proxy:         configCopy.Proxy,
		AdminEmail:    configCopy.AdminEmail,
		AdminPassword: configCopy.AdminPassword,
	}

	promptContentPrefix := ""

	switch req.Language {
	case "en":
		// 如果 Language 是 "en"，则输出英文内容
		promptContentPrefix = "Please reply in English. This is a git diff content change. Please summarize it. If it is a new feature, please start with feat:. If it is a refactoring, please start with factor:. If it is formatted code, please start with format:. If it is a bug fix, please start with fix:. Finally, give a 30-50 character summary of the changes. The language should be concise and comprehensive, and try not to produce code output."
	default:
		// 默认输出中文内容
		promptContentPrefix = "请你中文回复.这是一段git diff的内容变化,请你总结,如果是新特性请开头格式为feat: ,如果是重构请开头factor:,如果是格式化代码请开头格式为format:,如果是bug修复请开头fix:.最后给出30-50的字符总结变化,言语言简意赅,尽量不要出现代码的输出"
	}

	gptConfig := gogpt.DefaultConfig(cnf.ApiKey)
	gptRequest := gogpt.ChatCompletionRequest{
		Model:  gogpt.GPT4,
		Stream: false,
		Messages: []gogpt.ChatCompletionMessage{
			{
				Role:    gogpt.ChatMessageRoleUser,
				Content: promptContentPrefix + req.SummaryDiff,
			},
		},
	}

	client := gogpt.NewClientWithConfig(gptConfig)
	resp, err := client.CreateChatCompletion(context.Background(), gptRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Error generating commit message"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": resp.Choices[0].Message.Content})
}
