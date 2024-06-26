package model

import (
	"github.com/Arxtect/Einstein/apps/archive/models"
	"github.com/Arxtect/Einstein/common/initializers"
	"log"
)

func MigrateDocuments() {
	// 迁移表结构
	_ = initializers.DB.AutoMigrate(&models.Document{})

	// 执行原始 SQL 创建索引
	result := initializers.DB.Exec("CREATE INDEX documents_content_index ON documents USING gin(to_tsvector('simple_pdfs', content));")
	if result.Error != nil {
		log.Println("Error creating index:", result.Error)
		return
	} else {
		log.Println("Index created successfully")
		log.Println("🚗 Migration complete documents ...")
	}

}
