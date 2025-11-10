package util

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
)

func GetJson(v interface{}) string {
	marshal, _ := json.Marshal(v)
	return string(marshal)
}

// calculateFileMD5 计算文件的MD5值
func CalculateFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
