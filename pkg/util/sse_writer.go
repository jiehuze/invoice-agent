package util

import (
	"encoding/json"
	"fmt"
	"invoice-agent/internal/app/models"
	"net/http"
)

// 注意：不再使用 SSEEvent 包装结构！

// WriteSSE 现在要求 data 是一个 map[string]interface{} 或可合并 type 的结构
// 但为了兼容原有调用，我们通过类型判断来“扁平化”输出
func WriteSSE(w http.ResponseWriter, eventType string, data interface{}) error {
	var event map[string]interface{}

	switch v := data.(type) {
	case models.BalanceEvent:
		event = map[string]interface{}{
			"type": eventType,
			"data": v.Data,
		}
	case models.HeartbeatEvent:
		event = map[string]interface{}{
			"type": eventType,
		}
	case models.QueryEvent:
		event = map[string]interface{}{
			"type":      eventType,
			"data":      v.Data,
			"sessionId": v.SessionID,
		}
	case models.AppendTextEvent:
		event = map[string]interface{}{
			"type": eventType,
			"text": v.Text,
		}
	case models.SetReferenceEvent:
		event = map[string]interface{}{
			"type":     eventType,
			"resultId": v.ResultID,
			"list":     v.List,
		}
	default:
		// 回退：尝试将 data 转为 map 并加入 type（不推荐）
		if m, ok := data.(map[string]interface{}); ok {
			m["type"] = eventType
			event = m
		} else {
			return fmt.Errorf("unsupported event data type: %T", data)
		}
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "data: %s\n\n", bytes)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}

// ===== 以下快捷函数保持完全不变（调用方无感知）=====

func WriteBalance(w http.ResponseWriter, balance int64) error {
	return WriteSSE(w, "balance", models.BalanceEvent{Data: balance})
}

func WriteHeartbeat(w http.ResponseWriter) error {
	return WriteSSE(w, "heartbeat", models.HeartbeatEvent{})
}

func WriteQuery(w http.ResponseWriter, data []string, sessionID int64) error {
	return WriteSSE(w, "query", models.QueryEvent{
		Data:      data,
		SessionID: sessionID,
	})
}

func WriteAppendText(w http.ResponseWriter, text string) error {
	return WriteSSE(w, "append-text", models.AppendTextEvent{Text: text})
}

func WriteSetReference(w http.ResponseWriter, resultID string, list []models.ReferenceItem) error {
	return WriteSSE(w, "set-reference", models.SetReferenceEvent{
		ResultID: resultID,
		List:     list,
	})
}

func WriteDone(w http.ResponseWriter) {
	_, _ = fmt.Fprintf(w, "data: [DONE]\n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
