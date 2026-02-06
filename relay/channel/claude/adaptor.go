package claude

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/model_setting"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

type Adaptor struct {
}

func (a *Adaptor) ConvertGeminiRequest(*gin.Context, *relaycommon.RelayInfo, *dto.GeminiChatRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.ClaudeRequest) (any, error) {
	return request, nil
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	baseURL := fmt.Sprintf("%s/v1/messages", info.ChannelBaseUrl)
	// 如果客户端指定了 beta=true，或者配置启用了默认 beta
	if info.IsClaudeBetaQuery || model_setting.GetClaudeSettings().DefaultBetaEnabled {
		baseURL = baseURL + "?beta=true"
	}
	return baseURL, nil
}

func CommonClaudeHeadersOperation(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) {
	claudeSettings := model_setting.GetClaudeSettings()

	// 透传或设置默认 anthropic-beta header
	anthropicBeta := c.Request.Header.Get("anthropic-beta")
	if anthropicBeta != "" {
		req.Set("anthropic-beta", anthropicBeta)
	} else if claudeSettings.DefaultBetaEnabled && claudeSettings.DefaultBetaHeader != "" {
		// 使用配置的默认 beta header
		req.Set("anthropic-beta", claudeSettings.DefaultBetaHeader)
	}

	// 透传或设置默认 anthropic-dangerous-direct-browser-access header
	dangerousAccess := c.Request.Header.Get("anthropic-dangerous-direct-browser-access")
	if dangerousAccess != "" {
		req.Set("anthropic-dangerous-direct-browser-access", dangerousAccess)
	} else if claudeSettings.DefaultBetaEnabled {
		// 默认设置为 true（仅在启用 beta 时）
		req.Set("anthropic-dangerous-direct-browser-access", "true")
	}

	// 透传其他可能的 anthropic 相关 header
	for key, values := range c.Request.Header {
		lowerKey := strings.ToLower(key)
		// 透传所有 x-stainless-* 和其他 anthropic 相关 header
		if strings.HasPrefix(lowerKey, "x-stainless-") ||
			strings.HasPrefix(lowerKey, "x-app") {
			for _, value := range values {
				req.Set(key, value)
			}
		}
	}

	// 如果没有 X-Stainless-* header，设置默认值（模拟 Claude CLI）
	if c.Request.Header.Get("X-Stainless-Lang") == "" && claudeSettings.DefaultBetaEnabled {
		req.Set("X-App", "cli")
		req.Set("X-Stainless-Lang", "js")
		req.Set("X-Stainless-Runtime", "node")
		req.Set("X-Stainless-Runtime-Version", "v22.0.0")
		req.Set("X-Stainless-Arch", "x64")
		req.Set("X-Stainless-Os", "Linux")
		req.Set("X-Stainless-Package-Version", "0.70.0")
		req.Set("X-Stainless-Retry-Count", "0")
		req.Set("X-Stainless-Timeout", "600")
		// 模拟 Claude CLI 的 User-Agent
		req.Set("User-Agent", "claude-cli/2.1.12 (external, cli)")
	}

	claudeSettings.WriteHeaders(info.OriginModelName, req)
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)

	// 默认使用 x-api-key 认证（Claude 官方格式）
	req.Set("x-api-key", info.ApiKey)

	// 同时设置 Authorization: Bearer 格式，兼容某些第三方代理
	// 如果上游只需要其中一种，可以通过 HeaderOverride 覆盖
	req.Set("Authorization", "Bearer "+info.ApiKey)

	anthropicVersion := c.Request.Header.Get("anthropic-version")
	if anthropicVersion == "" {
		anthropicVersion = "2023-06-01"
	}
	req.Set("anthropic-version", anthropicVersion)
	CommonClaudeHeadersOperation(c, req, info)
	return nil
}

func (a *Adaptor) ConvertOpenAIRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return RequestOpenAI2ClaudeMessage(c, *request)
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertOpenAIResponsesRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.OpenAIResponsesRequest) (any, error) {
	// TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	if info.IsStream {
		return ClaudeStreamHandler(c, resp, info)
	} else {
		return ClaudeHandler(c, resp, info)
	}
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
