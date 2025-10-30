package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func getUploadPolicy(apiKey, modelName string) (map[string]interface{}, error) {
	url := "https://dashscope.aliyuncs.com/api/v1/uploads"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("action", "getPolicy")
	q.Add("model", modelName)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.New(fmt.Sprintf("Failed to get upload policy: %s", string(body)))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result["data"].(map[string]interface{}), nil
}

func uploadFileWithReader(policyData map[string]interface{}, reader io.Reader, fileName string) (string, error) {
	key := fmt.Sprintf("%s/%s", policyData["upload_dir"].(string), fileName)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	_ = writer.WriteField("OSSAccessKeyId", policyData["oss_access_key_id"].(string))
	_ = writer.WriteField("Signature", policyData["signature"].(string))
	_ = writer.WriteField("policy", policyData["policy"].(string))
	_ = writer.WriteField("x-oss-object-acl", policyData["x_oss_object_acl"].(string))
	_ = writer.WriteField("x-oss-forbid-overwrite", policyData["x_oss_forbid_overwrite"].(string))
	_ = writer.WriteField("key", key)
	_ = writer.WriteField("success_action_status", "200")

	// Add file
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, reader)
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", policyData["upload_host"].(string), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", errors.New(fmt.Sprintf("Failed to upload file: %s", string(body)))
	}

	return fmt.Sprintf("oss://%s", key), nil
}

func uploadFileToOSS(policyData map[string]interface{}, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileName := filepath.Base(filePath)
	key := fmt.Sprintf("%s/%s", policyData["upload_dir"].(string), fileName)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	_ = writer.WriteField("OSSAccessKeyId", policyData["oss_access_key_id"].(string))
	_ = writer.WriteField("Signature", policyData["signature"].(string))
	_ = writer.WriteField("policy", policyData["policy"].(string))
	_ = writer.WriteField("x-oss-object-acl", policyData["x_oss_object_acl"].(string))
	_ = writer.WriteField("x-oss-forbid-overwrite", policyData["x_oss_forbid_overwrite"].(string))
	_ = writer.WriteField("key", key)
	_ = writer.WriteField("success_action_status", "200")

	// Add file
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", policyData["upload_host"].(string), body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", errors.New(fmt.Sprintf("Failed to upload file: %s", string(body)))
	}

	return fmt.Sprintf("oss://%s", key), nil
}

func UpdateFileAndGetUrlWithReader(apiKey, modelName string, reader io.Reader, fileName string) (string, error) {
	policyData, err := getUploadPolicy(apiKey, modelName)
	if err != nil {
		return "", err
	}

	ossURL, err := uploadFileWithReader(policyData, reader, fileName)
	if err != nil {
		return "", err
	}

	return ossURL, nil
}

func UploadFileAndGetURL(apiKey, modelName, filePath string) (string, error) {
	// 1. 获取上传凭证
	policyData, err := getUploadPolicy(apiKey, modelName)
	if err != nil {
		return "", err
	}
	log.Infoln("policyData:", policyData)
	// 2. 上传文件到OSS
	ossURL, err := uploadFileToOSS(policyData, filePath)
	if err != nil {
		return "", err
	}

	return ossURL, nil
}
