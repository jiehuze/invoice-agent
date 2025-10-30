package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenAiFileClient struct {
	client openai.Client
}

var FileClient *OpenAiFileClient

// NewOpenAIClient 创建新的OpenAI客户端
func NewOpenAIClient(apiKey string) *OpenAiFileClient {
	if FileClient == nil {
		FileClient = &OpenAiFileClient{
			client: openai.NewClient(
				option.WithAPIKey(apiKey),
				option.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1"),
			),
		}
	}
	return FileClient
}

// UploadFile 上传文件并返回文件ID
func (c *OpenAiFileClient) UploadFile(ctx context.Context, filePath, purpose string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	response, err := c.client.Files.New(ctx, openai.FileNewParams{
		Purpose: openai.FilePurpose(purpose),
		File:    io.Reader(file),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	return response.ID, nil
}

// GetFile 获取文件信息
func (c *OpenAiFileClient) GetFile(ctx context.Context, fileID string) (*openai.FileObject, error) {
	file, err := c.client.Files.Get(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	return file, nil
}

//// ListFiles 列出所有文件
//func (c *OpenAiFileClient) ListFiles(ctx context.Context) (*openai.FileListResponse, error) {
//	files, err := c.client.Files.List(ctx, openai.FileListParams{})
//	if err != nil {
//		return nil, fmt.Errorf("failed to list files: %w", err)
//	}
//	return files, nil
//}

// DeleteFile 删除文件
func (c *OpenAiFileClient) DeleteFile(ctx context.Context, fileID string) error {
	_, err := c.client.Files.Delete(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// WaitForFileProcessing 等待文件处理完成
func (c *OpenAiFileClient) WaitForFileProcessing(ctx context.Context, fileID string, timeout time.Duration) (*openai.FileObject, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for file processing")
		case <-ticker.C:
			file, err := c.GetFile(ctx, fileID)
			if err != nil {
				return nil, fmt.Errorf("error getting file status: %w", err)
			}

			switch file.Status {
			case "processed":
				return file, nil
			case "failed":
				return nil, fmt.Errorf("file processing failed")
			case "uploaded":
				fmt.Printf("File is still processing... (status: %s)\n", file.Status)
			}
		}
	}
}
