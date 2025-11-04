package models

// ChatRequest 聊天请求结构体
// ChatRequest 聊天请求结构体
type ChatRequest struct {
	Input   string `json:"input" binding:"required"` // 用户输入
	History string `json:"history,omitempty"`        // 对话历史
}
