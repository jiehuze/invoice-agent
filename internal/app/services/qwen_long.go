package services

import (
	"context"
	"fmt"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"invoice-agent/pkg/config"
)

type QwenLongClient struct {
	client openai.Client
}

var ChatClient *QwenLongClient

func NewChantClient(apiKey string) *QwenLongClient {
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

func (p *QwenLongClient) Chat(ctx context.Context, fileIds []string) (*string, error) {
	msg := make([]openai.ChatCompletionMessageParamUnion, 0)
	msg = append(msg, openai.SystemMessage("You are a helpful assistant."))
	for _, s := range fileIds {
		msg = append(msg, openai.SystemMessage("fileid://"+s))
	}
	msg = append(msg, openai.UserMessage(config.GetOpenaiConf().Prompt))
	p.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{})
	chatCompletion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: msg,
		Model:    config.GetOpenaiConf().Model,
	})
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	fmt.Println(chatCompletion.Choices[0].Message.Content)
	return &chatCompletion.Choices[0].Message.Content, nil
	//splits := strings.Split(chatCompletion.Choices[0].Message.Content, "```")
	//err = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), target)
	//if nil != err {
	//	fmt.Println(err.Error())
	//	return err
	//}
}

func (p *QwenLongClient) ChatStream(ctx context.Context, fileIds []string) (<-chan string, <-chan error) {
	contentChan := make(chan string)
	errorChan := make(chan error, 1) // 缓冲通道，避免goroutine泄漏

	go func() {
		defer close(contentChan)
		defer close(errorChan)

		msg := make([]openai.ChatCompletionMessageParamUnion, 0)
		msg = append(msg, openai.SystemMessage("You are a helpful assistant."))
		for _, s := range fileIds {
			msg = append(msg, openai.SystemMessage("fileid://"+s))
		}
		msg = append(msg, openai.UserMessage(config.GetOpenaiConf().Prompt))

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
