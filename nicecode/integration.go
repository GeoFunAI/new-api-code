package nicecode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
)

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

// sendConsumeLog 发送消耗记录到 nicecode
func sendConsumeLog(baseURL, apiKey string, req *ConsumeLogRequest) error {
	// 构建请求URL
	url := baseURL + "/api/system/consume-log"

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
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送请求
	resp, err := client.Do(httpReq)
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

// RecordConsumeLog 记录积分消耗到 nicecode
// 此函数被 model 包调用，避免循环依赖
func RecordConsumeLog(userId, tokenId int, modelName string, quota, promptTokens, completionTokens int, other map[string]interface{}, requestId, clientIP string) {
	if common.NicecodeURL == "" || common.NicecodeAPIKey == "" {
		return
	}

	// 获取 other 中的 cache tokens（如果存在）
	var cacheReadTokens, cacheWriteTokens int64
	if other != nil {
		if val, ok := other["cache_read_input_tokens"].(float64); ok {
			cacheReadTokens = int64(val)
		}
		if val, ok := other["cache_creation_input_tokens"].(float64); ok {
			cacheWriteTokens = int64(val)
		}
	}

	// 构建请求（userId 就是 new-api-code 的 user_id，也就是 api_user_id）
	req := &ConsumeLogRequest{
		ApiUserID:        userId,
		ApiTokenID:       tokenId,
		ModelName:        modelName,
		Quota:            int64(quota),
		PromptTokens:     int64(promptTokens),
		CompletionTokens: int64(completionTokens),
		CacheReadTokens:  cacheReadTokens,
		CacheWriteTokens: cacheWriteTokens,
		RequestID:        requestId,
		IP:               clientIP,
	}

	// 发送请求
	if err := sendConsumeLog(common.NicecodeURL, common.NicecodeAPIKey, req); err != nil {
		common.SysLog(fmt.Sprintf("Failed to record consume log to nicecode: %v", err))
	}
}
