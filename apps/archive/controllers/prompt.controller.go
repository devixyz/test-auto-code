package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Arxtect/Einstein/apps/archive/models"
	"github.com/Arxtect/Einstein/apps/archive/service/dto"
	"github.com/Arxtect/Einstein/common/constants"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"net/http"
)

type PromptController struct {
	DB      *gorm.DB
	RedisDb *redis.Client
	Logger  *log.Logger
}

func NewPromptController(DB *gorm.DB, logger *log.Logger, redisDb *redis.Client) PromptController {
	return PromptController{
		DB:      DB,
		RedisDb: redisDb,
		Logger:  logger,
	}
}

// GetPromptList Conditional query, paging, finding matching prompt
func (pc *PromptController) GetPromptList(c *gin.Context) {
	req := dto.GetPromptRequest{}
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid request"})
		return
	}

	var prompt []models.Prompt

	// 添加全文搜索条件
	query := pc.DB.Model(&models.Prompt{})
	if req.Keyword != "" {
		// 使用Postgres的全文搜索函数
		query = query.Where("to_tsvector('simple', content) @@ plainto_tsquery('simple', ?)", req.Keyword)
	}

	var count int64
	err = query.Scopes(
		dto.Paginate(req.GetPageSize(), req.GetPageIndex()),
	).Order("created_at desc").
		Count(&count).
		Find(&prompt).Error
	if err != nil {
		pc.Logger.Printf("Error getting  prompt from DB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Error getting  prompt"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"prompt": prompt, "total": count, "pageIndex": req.GetPageIndex(), "pageSize": req.GetPageSize()}})
}

// GetPrompt find one Prompt by ID
func (pc *PromptController) GetPrompt(c *gin.Context) {
	promptID := c.Param("id")
	if promptID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid request"})
		return
	}
	var prompt models.Prompt

	// 先从Redis中获取，如果没有，查询数据库
	val, err := pc.RedisDb.Get(context.Background(), constants.RedisKeyAiPrompt+promptID).Result()
	if err == redis.Nil {
		// Redis中没有找到，从数据库中查找
		err = pc.DB.First(&prompt, promptID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "Prompt not found"})
			} else {
				pc.Logger.Printf("Error getting prompt from DB: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Error getting prompt"})
			}
			return
		}
		// 找到后，将结果存储到Redis中
		jsonPrompt, _ := json.Marshal(prompt)
		pc.RedisDb.Set(context.Background(), promptID, jsonPrompt, 0)
	} else if err != nil {
		pc.Logger.Printf("Error getting prompt from Redis: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "Prompt not found"})
		return
	} else {
		// Redis中有缓存，反序列化
		json.Unmarshal([]byte(val), &prompt)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": prompt})
}

// CreatePrompt create Prompt
func (pc *PromptController) CreatePrompt(c *gin.Context) {
	var req dto.CreatePromptRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid request"})
		return
	}

	prompt := models.Prompt{
		Content:       req.Content,
		Keywords:      req.Keywords,
		Settings:      req.Settings,
		ReferenceFile: req.ReferenceFile,
	}

	err = pc.DB.Create(&prompt).Error
	if err != nil {
		pc.Logger.Printf("Error creating prompt in DB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Error creating prompt"})
		return
	}

	// 添加到数据库后，更新Redis缓存
	jsonPrompt, _ := json.Marshal(prompt)
	pc.RedisDb.Set(context.Background(), constants.RedisKeyAiPrompt+fmt.Sprintf("%v", prompt.ID), jsonPrompt, 0)

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": prompt})
}

// UpdatePrompt modify Prompt
func (pc *PromptController) UpdatePrompt(c *gin.Context) {
	var req dto.UpdatePromptRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid request"})
		return
	}
	if req.Base.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid request"})
		return
	}

	// 解析UUID
	uuidID, err := uuid.Parse(req.Base.ID)
	if err != nil {
		pc.Logger.Printf("Error parsing UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid UUID"})
		return
	}

	// 根据ID查询已存在的Prompt
	var prompt models.Prompt
	result := pc.DB.First(&prompt, "id = ?", uuidID)
	if result.Error != nil {
		pc.Logger.Printf("Error finding prompt in DB: %v", result.Error)
		c.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "Prompt not found"})
		return
	}

	// 更新prompt的字段
	prompt.Content = req.Content
	prompt.Keywords = req.Keywords
	prompt.Settings = req.Settings
	prompt.ReferenceFile = req.ReferenceFile

	// 使用Save方法保存更新
	result = pc.DB.Save(&prompt)
	if result.Error != nil {
		pc.Logger.Printf("Error updating prompt in DB: %v", result.Error)
		c.JSON(http.StatusOK, gin.H{"status": "fail", "message": "no prompt"})
		return
	}

	// 修改数据库后，更新Redis缓存
	jsonPrompt, _ := json.Marshal(prompt)
	pc.RedisDb.Set(context.Background(), constants.RedisKeyAiPrompt+req.Base.ID, jsonPrompt, 0)

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Prompt updated"})
}

// DeletePrompt delete Prompt
func (pc *PromptController) DeletePrompt(c *gin.Context) {
	promptID := c.Param("id")
	if promptID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid request"})
		return
	}

	// 解析字符串ID为uuid.UUID类型
	uuidID, err := uuid.Parse(promptID)
	if err != nil {
		pc.Logger.Printf("Error parsing UUID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid UUID"})
		return
	}

	result := pc.DB.Delete(&models.Prompt{}, "id = ?", uuidID)
	if result.Error != nil {
		pc.Logger.Printf("Error deleting prompt from DB: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Error deleting prompt"})
		return
	}
	// 删除数据库数据后，清理Redis缓存
	pc.RedisDb.Del(context.Background(), constants.RedisKeyAiPrompt+promptID)

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Prompt deleted"})
}
