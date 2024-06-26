package dto

import (
	"github.com/google/uuid"
)

// FileResponse 文件上传响应
type FileResponse struct {
	FileStorageID     string `json:"file_storage_id"`
	FileName          string `json:"file_name"`
	FileSize          int64  `json:"file_size"`
	FileStorageBucket string `json:"file_storage_bucket"`
}

// AddTagToDocumentRequest 为文档添加标签请求
type AddTagToDocumentRequest struct {
	Title      string   `json:"title" binding:"required"`
	Tags       []string `json:"tags"`
	Content    string   `json:"content"`
	UploadType string   `json:"upload_type"`
}

// GetDocumentsByUserRequest 获取文档列表请求
type GetDocumentsByUserRequest struct {
	Pagination `search:"-"`
	Keyword    string   `form:"keyword"`
	Tags       []string `form:"tags"`
}

// GetDocumentDownloadUrlRequest 获取文档下载地址请求
type GetDocumentDownloadUrlRequest struct {
	FileStorageID string `json:"file_storage_id" binding:"required"`
}

// SearchListDocumentsResp 搜索文档列表响应
type SearchListDocumentsResp struct {
	ID    uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key" json:"id"`
	Name  string    `gorm:"type:varchar(255);not null" json:"name"`
	Email string    `gorm:"uniqueIndex;not null" json:"email"`
}

// GenCommitDocumentRequest  生成文档提交请求
type GenCommitDocumentRequest struct {
	SummaryDiff string `json:"summary_diff" binding:"required"`
	Language    string `json:"language"`
}
