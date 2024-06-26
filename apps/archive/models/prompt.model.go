package models

type Prompt struct {
	Base
	Content       string `json:"content" gorm:"type:text;comment:用于存储上下文信息，可以是一段描述、历史对话记录或相关数据。"`
	Keywords      string `json:"keywords" gorm:"type:text;comment:用于存储与上下文相关的关键词，有助于AI理解上下文的重点。"`
	Settings      string `json:"settings" gorm:"type:text;comment:用于存储AI的行为设置信息，如特定的语言、专业领域或回答风格等。"`
	ReferenceFile string `json:"reference_file" gorm:"type:text;comment:引用的文件名,用逗号拼接。"`
}

//eg.     Context:     "关于提高工业生产效率的讨论。",
//        Keywords:    "工业, 生产效率, 自动化",
//        Settings:    "zh-CN, 工业领域, 专业回答",

func (s *Prompt) TableName() string {
	return "ai_prompts"
}
