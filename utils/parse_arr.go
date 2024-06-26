package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func ParseTags(tagsStr string) ([]string, error) {
	// 如果入参是空字符串，立即返回
	if tagsStr == "" {
		return []string{}, errors.New("Invalid tags: empty string")
	}
	var tags []string

	// 检查tagsStr是否是一个有效的JSON数组
	if strings.HasPrefix(tagsStr, "[") && strings.HasSuffix(tagsStr, "]") {
		// 如果是，使用json.Unmarshal来解析它
		err := json.Unmarshal([]byte(tagsStr), &tags)
		if err != nil {
			return nil, fmt.Errorf("Invalid JSON array: %v", err)
		}
	} else {
		// 如果不是，使用strings.Split来分割它
		tags = strings.Split(tagsStr, ",")
	}

	return tags, nil
}
