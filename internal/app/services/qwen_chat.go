package services

import (
	"context"
	"fmt"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"invoice-agent/internal/app/models"
	"invoice-agent/pkg/config"
	"strings"
)

type QwenLongClient struct {
	client openai.Client
}

var ChatClient *QwenLongClient

func NewChatClient(apiKey string) *QwenLongClient {
	if ChatClient == nil {
		ChatClient = &QwenLongClient{
			client: openai.NewClient(
				option.WithAPIKey(apiKey),
				option.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1"),
			),
		}
	}
	return ChatClient
}

func (p *QwenLongClient) Chat(ctx context.Context, req models.ChatRequest) (*string, error) {
	msg := make([]openai.ChatCompletionMessageParamUnion, 0)
	msg = append(msg, openai.SystemMessage("You are a helpful assistant."))

	if req.Parse {
		prompt := strings.Replace(config.GetOpenaiConf().ParsePrompt, "{{input_question}}", req.Input, 1)
		msg = append(msg, openai.UserMessage(prompt))
	} else {
		history_prompt := strings.Replace(config.GetOpenaiConf().ChatPrompt, "{{history}}", req.History, 1)
		prompt := strings.Replace(history_prompt, "{{input_question}}", req.Input, 1)
		msg = append(msg, openai.UserMessage(prompt))
	}
	chatCompletion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages:    msg,
		Model:       config.GetOpenaiConf().Model,
		Temperature: openai.Float(0),
	})
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	fmt.Println(chatCompletion.Choices[0].Message.Content)
	return &chatCompletion.Choices[0].Message.Content, nil
}

func (p *QwenLongClient) ChatStream(ctx context.Context, fileIds []string) (<-chan string, <-chan error) {
	contentChan := make(chan string)
	errorChan := make(chan error, 1) // 缓冲通道，避免goroutine泄漏

	go func() {
		defer close(contentChan)
		defer close(errorChan)

		var fileids string
		msg := make([]openai.ChatCompletionMessageParamUnion, 0)
		msg = append(msg, openai.SystemMessage("You are a helpful assistant."))
		for _, s := range fileIds {
			msg = append(msg, openai.SystemMessage("fileid://"+s))
			fileids += s + ","
		}
		prompt := strings.Replace(config.GetOpenaiConf().Prompt, "{{file_ids}}", fileids, 1)
		msg = append(msg, openai.UserMessage(prompt))

		// 创建流式请求
		stream := p.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Messages: msg,
			Model:    config.GetOpenaiConf().Model,
		})

		// 处理流式响应
		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				if content != "" {
					contentChan <- content
				}
			}
		}

		if err := stream.Err(); err != nil {
			errorChan <- err
		}
	}()

	return contentChan, errorChan
}
