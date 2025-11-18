package models

// 所有事件的基类（用于统一处理）
type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// 1. BalanceEvent
type BalanceEvent struct {
	Data int64 `json:"data"`
}

// 2. HeartbeatEvent —— 无字段
type HeartbeatEvent struct{}

// 3. QueryEvent
type QueryEvent struct {
	Data      []string `json:"data"`
	SessionID int64    `json:"sessionId"`
}

// 4. AppendTextEvent
type AppendTextEvent struct {
	Text string `json:"text"`
}

// 5. SetReferenceEvent
type SetReferenceEvent struct {
	ResultID string          `json:"resultId"`
	List     []ReferenceItem `json:"list"`
}

type ReferenceItem struct {
	ArticleType string           `json:"article_type"`
	Display     ReferenceDisplay `json:"display"`
	FileMeta    FileMeta         `json:"file_meta"`
	Index       int              `json:"index"`
	Page        int              `json:"page"`
	Title       string           `json:"title"`
	TotalPage   int              `json:"total_page"`
}

type ReferenceDisplay struct {
	ReferID int `json:"refer_id"`
}

type FileMeta struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}
