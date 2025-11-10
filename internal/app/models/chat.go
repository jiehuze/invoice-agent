package models

// ChatRequest 聊天请求结构体
// ChatRequest 聊天请求结构体
type ChatRequest struct {
	SessionId string   `json:"session_id"`
	Input     string   `json:"input" binding:"required"` // 用户输入
	History   string   `json:"history,omitempty"`        // 对话历史
	Parse     bool     `json:"parse"`
	FileIds   []string `json:"file_ids"`
}
