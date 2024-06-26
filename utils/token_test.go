package utils

import (
	"github.com/Arxtect/Einstein/config"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

// 测试token是否可用
func TestCreateAndValidateToken(t *testing.T) {
	// 设置测试数据
	payload := "user123" // 替换为实际的负载数据
	err := config.LoadEnv("../config/settings-dev.yml")
	if err != nil {
		log.Println("Error loading .env file", err)
		return
	}

	configCopy := config.Env

	// 测试 CreateToken 函数
	createdToken, err := CreateToken(configCopy.AccessTokenExpiresIn, payload, configCopy.AccessTokenPrivateKey)
	assert.NoError(t, err, "CreateToken should not return an error")

	// 测试 ValidateToken 函数
	validatedPayload, err := ValidateToken(createdToken, configCopy.AccessTokenPublicKey)
	assert.NoError(t, err, "ValidateToken should not return an error")

	// 断言负载数据是否相等
	assert.Equal(t, payload, validatedPayload, "Payloads should match")
	log.Println("createdToken correct: ", createdToken)
}
