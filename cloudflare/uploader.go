package cloudflare

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"telegram-cf-bot/config"
	"telegram-cf-bot/logger"
)

// Cloudflare API 响应结构体
type CloudflareResponse struct {
	Result struct {
		ID       string   `json:"id"`
		Filename string   `json:"filename"`
		Uploaded string   `json:"uploaded"`
		Variants []string `json:"variants"` // 包含不同尺寸的图片URL
	} `json:"result"`
	Success bool `json:"success"`
	Errors  []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// UploadImage 将图片字节上传到 Cloudflare Images
func UploadImage(imageBytes []byte, userID int64, metadata map[string]interface{}, cfg *config.Config) (*CloudflareResponse, error) {
	startTime := time.Now()

	// 生成文件名: userID + unix时间戳 + 随机字符串
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, logger.LogError(err, "生成随机字符串失败")
	}
	randomStr := hex.EncodeToString(randomBytes)

	filename := fmt.Sprintf("%d_%d_%s.jpg", userID, timestamp, randomStr)

	logger.WithUserID(userID).WithFields(map[string]interface{}{
		"filename":  filename,
		"file_size": len(imageBytes),
	}).Info("开始上传图片到Cloudflare")

	// 创建一个 multipart/form-data 请求体
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// 创建文件部分，使用生成的文件名
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, logger.LogError(err, "创建multipart文件部分失败")
	}

	// 将图片数据写入文件部分
	_, err = part.Write(imageBytes)
	if err != nil {
		return nil, logger.LogError(err, "写入图片数据失败")
	}

	// 添加必要的元数据字段（Cloudflare Images API的标准字段）
	// 只添加Cloudflare API支持的字段，避免"Error parsing form fields"错误
	if metadata != nil {
		// 添加ID字段（可选）
		if id, ok := metadata["id"]; ok {
			if idStr := fmt.Sprintf("%v", id); idStr != "" {
				writer.WriteField("id", idStr)
			}
		}

		// 添加requireSignedURLs字段（可选）
		writer.WriteField("requireSignedURLs", "false")

		// 添加metadata作为JSON字符串（如果需要保存额外信息）
		filteredMetadata := make(map[string]interface{})
		for key, value := range metadata {
			// 只保留基本的元数据字段
			switch key {
			case "width", "height", "format", "file_size", "camera_make", "camera_model":
				filteredMetadata[key] = value
			}
		}

		if len(filteredMetadata) > 0 {
			metadataBytes, err := json.Marshal(filteredMetadata)
			if err == nil && len(metadataBytes) < 1024 { // 确保元数据不超过1KB
				writer.WriteField("metadata", string(metadataBytes))
			}
		}
	}

	// 关闭 writer
	err = writer.Close()
	if err != nil {
		return nil, logger.LogError(err, "关闭multipart writer失败")
	}

	// 构建 Cloudflare Images API URL
	apiURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/images/v1", cfg.CloudflareAccountID)

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", apiURL, &requestBody)
	if err != nil {
		return nil, logger.LogError(err, "创建HTTP请求失败")
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+cfg.CloudflareAPIToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	logger.WithUserID(userID).WithFields(map[string]interface{}{
		"api_url":      apiURL,
		"filename":     filename,
		"content_type": writer.FormDataContentType(),
	}).Debug("发送请求到Cloudflare API")

	resp, err := client.Do(req)
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		logger.LogAPICall("cloudflare", "POST", apiURL, 0, duration, err)
		return nil, logger.LogError(err, "发送HTTP请求失败")
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.WithUserID(userID).WithError(closeErr).Warn("关闭HTTP响应体失败")
		}
	}()

	duration := time.Since(startTime).Milliseconds()
	logger.LogAPICall("cloudflare", "POST", apiURL, resp.StatusCode, duration, nil)

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, logger.LogError(err, "读取响应失败")
	}

	logger.WithUserID(userID).WithFields(map[string]interface{}{
		"status_code":   resp.StatusCode,
		"response_size": len(responseBody),
	}).Debug("收到Cloudflare API响应")

	// 解析响应
	var cfResponse CloudflareResponse
	err = json.Unmarshal(responseBody, &cfResponse)
	if err != nil {
		logger.WithUserID(userID).WithFields(map[string]interface{}{
			"response_body": string(responseBody),
			"status_code":   resp.StatusCode,
		}).Error("解析Cloudflare响应失败")
		return nil, logger.LogError(err, "解析JSON响应失败")
	}

	// 检查是否成功
	if !cfResponse.Success {
		var errorMessages []string
		for _, cfErr := range cfResponse.Errors {
			errorMessages = append(errorMessages, cfErr.Message)
		}

		logger.WithUserID(userID).WithFields(map[string]interface{}{
			"errors":        errorMessages,
			"status_code":   resp.StatusCode,
			"response_body": string(responseBody),
		}).Error("Cloudflare API返回错误")

		logger.LogUploadAction(userID, "", filename, int64(len(imageBytes)), false, fmt.Sprintf("Cloudflare API错误: %v", errorMessages))
		return nil, fmt.Errorf("Cloudflare API 错误: %v", errorMessages)
	}

	logger.WithUserID(userID).WithFields(map[string]interface{}{
		"image_id":       cfResponse.Result.ID,
		"filename":       cfResponse.Result.Filename,
		"uploaded_time":  cfResponse.Result.Uploaded,
		"variants_count": len(cfResponse.Result.Variants),
		"duration_ms":    duration,
	}).Info("图片上传到Cloudflare成功")

	logger.LogUploadAction(userID, "", filename, int64(len(imageBytes)), true, "")

	return &cfResponse, nil
}

// GetImageURL 从 Cloudflare 响应中提取图片 URL
func GetImageURL(response *CloudflareResponse) (string, error) {
	if response == nil || !response.Success {
		return "", logger.LogAndReturnError("无效的Cloudflare响应")
	}

	if len(response.Result.Variants) == 0 {
		logger.WithFields(map[string]interface{}{
			"image_id": response.Result.ID,
		}).Error("Cloudflare响应中没有图片变体URL")
		return "", logger.LogAndReturnError("响应中没有图片URL")
	}

	// 返回第一个变体URL（通常是原图）
	imageURL := response.Result.Variants[0]

	logger.WithFields(map[string]interface{}{
		"image_id":       response.Result.ID,
		"image_url":      imageURL,
		"variants_count": len(response.Result.Variants),
	}).Debug("提取图片URL成功")

	return imageURL, nil
}
