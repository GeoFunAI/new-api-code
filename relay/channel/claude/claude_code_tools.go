package claude

import (
	_ "embed"
	"encoding/json"
)

//go:embed claude_code_tools.json
var claudeCodeToolsJSON []byte

// GetClaudeCodeTools 返回 Claude Code 的标准 tools 定义
func GetClaudeCodeTools() ([]any, error) {
	var tools []any
	err := json.Unmarshal(claudeCodeToolsJSON, &tools)
	return tools, err
}
