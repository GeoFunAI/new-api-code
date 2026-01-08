package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// NicecodeClient nicecode API 客户端
type NicecodeClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewNicecodeClient 创建 nicecode 客户端
func NewNicecodeClient(baseURL, apiKey string) *NicecodeClient {
	return &NicecodeClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ConsumeLogRequest 积分消耗记录请求
type ConsumeLogRequest struct {
	ApiUserID        int    `json:"api_user_id"`
	ApiTokenID       int    `json:"api_token_id"`
	ModelName        string `json:"model_name"`
	Quota            int64  `json:"quota"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	CacheReadTokens  int64  `json:"cache_read_tokens"`
	CacheWriteTokens int64  `json:"cache_write_tokens"`
	RequestID        string `json:"request_id"`
	IP               string `json:"ip"`
}

// RecordConsumeLog 记录积分消耗到 nicecode
func (c *NicecodeClient) RecordConsumeLog(req *ConsumeLogRequest) error {
	if c.baseURL == "" || c.apiKey == "" {
		return fmt.Errorf("nicecode not configured")
	}

	// 构建请求URL
	url := c.baseURL + "/api/system/consume-log"

	// 序列化请求体
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 发送请求
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("nicecode API returned status %d", resp.StatusCode)
	}

	// 解析响应
	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			TransactionID uint  `json:"transaction_id"`
			Balance       int64 `json:"balance"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("nicecode API error: %s", result.Message)
	}

	common.SysLog(fmt.Sprintf("Recorded consume log to nicecode: api_user_id=%d, transaction_id=%d, balance=%d",
		req.ApiUserID, result.Data.TransactionID, result.Data.Balance))

	return nil
}
