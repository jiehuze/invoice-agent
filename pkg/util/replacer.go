package util

import "strings"

// StringReplacer 字符串替换接口
type StringReplacer interface {
	// Replace 执行字符串替换
	Replace(template string, placeholder string, replacement string) string
}

// SimpleStringReplacer 简单字符串替换器实现
type SimpleStringReplacer struct{}

// NewSimpleStringReplacer 创建新的简单字符串替换器
func NewSimpleStringReplacer() *SimpleStringReplacer {
	return &SimpleStringReplacer{}
}

// Replace 实现字符串替换
func (r *SimpleStringReplacer) Replace(template string, placeholder string, replacement string) string {
	return strings.Replace(template, placeholder, replacement, -1)
}
