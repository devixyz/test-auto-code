package service

import (
	"github.com/Arxtect/Einstein/apps/archive/models"
	"github.com/Arxtect/Einstein/common/initializers"
)

// GetDocumentByFileHash 根据存储键获取文档
func GetDocumentByFileHash(hash string) (*models.Document, error) {
	var doc models.Document
	err := initializers.DB.Model(&doc).Where("file_hash = ?", hash).First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// GetTagsByName 根据标签名称获取标签
func GetTagsByName(names []string) ([]models.Tag, error) {
	var tags []models.Tag
	err := initializers.DB.Model(&models.Tag{}).Where("name IN (?)", names).Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// UpdateDocumentTitleAndTags 更新文档的标题和标签,还有摘要
func UpdateDocumentTitleAndTags(doc *models.Document, title string, tags []models.Tag, content string) error {
	update := models.Document{}
	if title != "" {
		update.Title = title
	}
	if tags != nil && len(tags) > 0 {
		update.Tags = tags
	}
	if content != "" {
		update.Content = content
	}

	return initializers.DB.Model(doc).Omit("User").Updates(update).Error
}

// GetDocumentByKey 根据文档Key查看单个细节
func GetDocumentByKey(key string) (*models.Document, error) {
	var doc models.Document
	err := initializers.DB.Model(&doc).
		Preload("User").Preload("Tags").
		Where("storage_key = ?", key).First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}
