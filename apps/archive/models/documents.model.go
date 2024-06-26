package models

import (
	"github.com/google/uuid"
)

type Document struct {
	Base
	Title         string    `gorm:"type:varchar(255);not null" json:"title"`
	Content       string    `gorm:"type:text" json:"content"`
	AuthorID      uuid.UUID `gorm:"type:uuid;not null" json:"author_id"`
	StorageBucket string    `gorm:"type:varchar(100);not null" json:"storage_bucket"`   // 存储桶
	StorageKey    string    `gorm:"type:varchar(100);not null" json:"storage_key"`      // 存储键
	StorageSize   int64     `gorm:"type:bigint" json:"storage_size"`                    // 存储大小
	FileHash      string    `gorm:"type:varchar(100);unique;not null" json:"file_hash"` // 文件hash 用来保证一致性
	StorageZip    string    `gorm:"type:varchar(100)" json:"storage_zip"`               // 存储zip
	Tags          []Tag     `gorm:"many2many:document_tags;" json:"tags"`               // 文档标签
	Cover         string    `gorm:"type:varchar(100)" json:"cover"`                     // 封面
	User          SafeUser  `gorm:"foreignKey:AuthorID" json:"user"`                    // 修改此处
}

type Tag struct {
	Base
	Name      string     `gorm:"type:varchar(100);unique;not null" json:"name"` // 标签名称
	Documents []Document `gorm:"many2many:document_tags;" json:"documents"`     // 关联的文档
}

func (s *Document) TableName() string {
	return "documents"
}

// DocumentMigrate 迁移表结构
type DocumentMigrate struct {
	Base
	Title         string    `gorm:"type:varchar(255);not null" json:"title"`
	Content       string    `gorm:"type:text" json:"content"`
	AuthorID      uuid.UUID `gorm:"type:uuid;not null" json:"author_id"`
	StorageBucket string    `gorm:"type:varchar(100);not null" json:"storage_bucket"`   // 存储桶
	StorageKey    string    `gorm:"type:varchar(100);not null" json:"storage_key"`      // 存储键
	StorageSize   int64     `gorm:"type:bigint" json:"storage_size"`                    // 存储大小
	FileHash      string    `gorm:"type:varchar(100);unique;not null" json:"file_hash"` // 文件hash 用来保证一致性
	StorageZip    string    `gorm:"type:varchar(100)" json:"storage_zip"`               // 存储zip
	Tags          []Tag     `gorm:"many2many:document_tags;" json:"tags"`               // 文档标签
	Cover         string    `gorm:"type:varchar(100)" json:"cover"`                     // 封面
}

func (s *DocumentMigrate) TableName() string {
	return "documents"
}
