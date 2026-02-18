// Package cloudflare provides Cloudflare Images API integration.
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

	"telegram-cf-bot/internal/config"
	"telegram-cf-bot/internal/constants"
	apperrors "telegram-cf-bot/internal/errors"
	"telegram-cf-bot/internal/logger"
)

// Client provides Cloudflare API operations.
type Client struct {
	config     *config.Config
	httpClient *http.Client
}

// UploadResponse represents Cloudflare API upload response.
type UploadResponse struct {
	Success bool `json:"success"`
	Result  struct {
		ID       string   `json:"id"`
		Filename string   `json:"filename"`
		Uploaded string   `json:"uploaded"`
		Variants []string `json:"variants"`
	} `json:"result"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// NewClient creates a new Cloudflare API client.
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: constants.HTTPClientTimeout,
		},
	}
}

// Upload uploads an image to Cloudflare Images.
func (c *Client) Upload(imageBytes []byte, userID int64, metadata map[string]interface{}) (*UploadResponse, error) {
	start := time.Now()
	filename := generateFilename(userID)

	log := logger.WithUser(userID, "").WithFields(map[string]interface{}{
		"filename":  filename,
		"file_size": len(imageBytes),
	})

	log.Info("uploading image to cloudflare")

	// Build multipart request
	body, contentType, err := c.buildMultipartBody(imageBytes, filename, metadata)
	if err != nil {
		return nil, err
	}

	// Create request
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/images/v1",
		c.config.Cloudflare.AccountID)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrUploadFailed, "failed to create request", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.Cloudflare.APIToken)
	req.Header.Set("Content-Type", contentType)

	// Send request
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		logger.LogAPICall("cloudflare", "POST", url, 0, duration, err)
		return nil, apperrors.Wrap(apperrors.ErrUploadFailed, "request failed", err)
	}
	defer resp.Body.Close()

	logger.LogAPICall("cloudflare", "POST", url, resp.StatusCode, duration, nil)

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrUploadFailed, "failed to read response", err)
	}

	// Parse response
	var result UploadResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrUploadFailed, "failed to parse response", err)
	}

	// Check success
	if !result.Success {
		var msgs []string
		for _, e := range result.Errors {
			msgs = append(msgs, e.Message)
		}

		logger.LogUpload(userID, filename, int64(len(imageBytes)), false,
			fmt.Errorf("cloudflare errors: %v", msgs))

		return nil, apperrors.New(apperrors.ErrCloudflareAPI, fmt.Sprintf("API errors: %v", msgs))
	}

	logger.LogUpload(userID, filename, int64(len(imageBytes)), true, nil)
	log.WithFields(map[string]interface{}{
		"image_id": result.Result.ID,
		"duration": duration,
	}).Info("upload successful")

	return &result, nil
}

// GetImageURL extracts the image URL from upload response.
func GetImageURL(resp *UploadResponse) (string, error) {
	if resp == nil || !resp.Success {
		return "", apperrors.New(apperrors.ErrCloudflareAPI, "invalid response")
	}

	if len(resp.Result.Variants) == 0 {
		return "", apperrors.New(apperrors.ErrCloudflareAPI, "no image variants in response")
	}

	return resp.Result.Variants[0], nil
}

// buildMultipartBody creates multipart form data for upload.
func (c *Client) buildMultipartBody(imageBytes []byte, filename string, metadata map[string]interface{}) (*bytes.Buffer, string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, "", apperrors.Wrap(apperrors.ErrUploadFailed, "failed to create form file", err)
	}

	if _, err := part.Write(imageBytes); err != nil {
		return nil, "", apperrors.Wrap(apperrors.ErrUploadFailed, "failed to write image data", err)
	}

	// Add metadata if present
	if metadata != nil {
		filtered := filterMetadata(metadata)
		if len(filtered) > 0 {
			metaJSON, _ := json.Marshal(filtered)
			if len(metaJSON) < constants.MaxMetadataSizeBytes {
				writer.WriteField("metadata", string(metaJSON))
			}
		}
	}

	writer.WriteField("requireSignedURLs", "false")

	if err := writer.Close(); err != nil {
		return nil, "", apperrors.Wrap(apperrors.ErrUploadFailed, "failed to close writer", err)
	}

	return &body, writer.FormDataContentType(), nil
}

// generateFilename creates a unique filename for upload.
func generateFilename(userID int64) string {
	timestamp := time.Now().Unix()

	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomStr := hex.EncodeToString(randomBytes)

	return fmt.Sprintf(constants.FilenameTemplate, userID, timestamp, randomStr)
}

// filterMetadata keeps only allowed metadata fields.
func filterMetadata(metadata map[string]interface{}) map[string]interface{} {
	allowed := map[string]bool{
		"width": true, "height": true, "format": true,
		"file_size": true, "camera_make": true, "camera_model": true,
	}

	filtered := make(map[string]interface{})
	for k, v := range metadata {
		if allowed[k] {
			filtered[k] = v
		}
	}

	return filtered
}
