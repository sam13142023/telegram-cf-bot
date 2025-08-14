package validator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/gif"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"telegram-cf-bot/logger"

	"github.com/rwcarlsen/goexif/exif"
)

// Cloudflare 的限制
const (
	maxDimension     = 12000
	maxArea          = 100 * 1000 * 1000 // 1亿像素
	maxSizeBytes     = 10 * 1024 * 1024  // 10 MB
	maxMetadataBytes = 1024
	maxAnimatedArea  = 50 * 1000 * 1000 // 5千万像素
)

// ValidationResult 包含验证结果
type ValidationResult struct {
	IsValid  bool
	Error    string
	Metadata map[string]interface{}
}

// ValidateImage 验证图片数据是否符合Cloudflare标准
// 3. 验证以文件形式发送的内容是否符合cf的标准
func ValidateImage(imageBytes []byte) (*ValidationResult, error) {
	logger.WithComponent("validator").WithFields(map[string]interface{}{
		"file_size": len(imageBytes),
	}).Debug("开始验证图片")

	// 检查文件大小
	if len(imageBytes) > maxSizeBytes {
		err := fmt.Errorf("图片大小超过 %d MB", maxSizeBytes/1024/1024)
		logger.WithComponent("validator").WithFields(map[string]interface{}{
			"file_size": len(imageBytes),
			"max_size":  maxSizeBytes,
		}).Warn("图片文件大小超出限制")
		return nil, err
	}

	// 尝试解码图片以获取格式和尺寸
	imgConfig, format, err := image.DecodeConfig(bytes.NewReader(imageBytes))
	if err != nil {
		logger.WithComponent("validator").WithError(err).Error("无法解析图片格式")
		return nil, errors.New("无法识别的图片格式或文件已损坏")
	}

	logger.WithComponent("validator").WithFields(map[string]interface{}{
		"format": format,
		"width":  imgConfig.Width,
		"height": imgConfig.Height,
	}).Debug("图片格式解析成功")

	// 验证尺寸和像素面积
	if imgConfig.Width > maxDimension || imgConfig.Height > maxDimension {
		err := fmt.Errorf("图片尺寸超过 %d 像素", maxDimension)
		logger.WithComponent("validator").WithFields(map[string]interface{}{
			"width":         imgConfig.Width,
			"height":        imgConfig.Height,
			"max_dimension": maxDimension,
		}).Warn("图片尺寸超出限制")
		return nil, err
	}

	area := imgConfig.Width * imgConfig.Height
	maxAreaForFormat := maxArea
	if format == "gif" {
		maxAreaForFormat = maxAnimatedArea
	}

	if area > maxAreaForFormat {
		err := fmt.Errorf("图片像素面积超过限制")
		logger.WithComponent("validator").WithFields(map[string]interface{}{
			"area":     area,
			"max_area": maxAreaForFormat,
			"format":   format,
		}).Warn("图片像素面积超出限制")
		return nil, err
	}

	// 提取EXIF元数据
	metadata := make(map[string]interface{})
	if format == "jpeg" {
		exifData, err := exif.Decode(bytes.NewReader(imageBytes))
		if err != nil {
			logger.WithComponent("validator").WithFields(map[string]interface{}{
				"format": format,
			}).Debug("无法解析EXIF数据，可能是正常情况")
		} else {
			// 提取一些基本的EXIF信息
			if make, err := exifData.Get(exif.Make); err == nil {
				if makeStr, err := make.StringVal(); err == nil {
					metadata["camera_make"] = makeStr
				}
			}

			if model, err := exifData.Get(exif.Model); err == nil {
				if modelStr, err := model.StringVal(); err == nil {
					metadata["camera_model"] = modelStr
				}
			}

			if dateTime, err := exifData.Get(exif.DateTime); err == nil {
				if dateTimeStr, err := dateTime.StringVal(); err == nil {
					metadata["date_time"] = dateTimeStr
				}
			}

			logger.WithComponent("validator").WithFields(map[string]interface{}{
				"metadata_count": len(metadata),
			}).Debug("EXIF元数据提取成功")
		}
	}

	// 添加基本信息到元数据
	metadata["format"] = format
	metadata["width"] = imgConfig.Width
	metadata["height"] = imgConfig.Height
	metadata["file_size"] = len(imageBytes)

	// 验证元数据大小
	metadataJSON, _ := json.Marshal(metadata)
	if len(metadataJSON) > maxMetadataBytes {
		logger.WithComponent("validator").WithFields(map[string]interface{}{
			"metadata_size":     len(metadataJSON),
			"max_metadata_size": maxMetadataBytes,
		}).Warn("元数据过大，将使用简化版本")

		// 如果元数据太大，只保留基本信息
		metadata = map[string]interface{}{
			"format":    format,
			"width":     imgConfig.Width,
			"height":    imgConfig.Height,
			"file_size": len(imageBytes),
		}
	}

	logger.WithComponent("validator").WithFields(map[string]interface{}{
		"format":         format,
		"width":          imgConfig.Width,
		"height":         imgConfig.Height,
		"area":           area,
		"file_size":      len(imageBytes),
		"metadata_count": len(metadata),
	}).Info("图片验证通过")

	return &ValidationResult{
		IsValid:  true,
		Error:    "",
		Metadata: metadata,
	}, nil
}

// validateGif 验证GIF动画的总像素面积
func validateGif(imageBytes []byte) error {
	gifData, err := gif.DecodeAll(bytes.NewReader(imageBytes))
	if err != nil {
		return errors.New("无效的GIF文件")
	}

	totalArea := int64(0)
	for _, frame := range gifData.Image {
		bounds := frame.Bounds()
		frameArea := int64(bounds.Dx()) * int64(bounds.Dy())
		totalArea += frameArea
	}

	if totalArea > maxAnimatedArea {
		return errors.New("GIF动画总像素面积超过限制")
	}

	return nil
}

// extractMetadata 从图片字节中提取EXIF元数据
func extractMetadata(imageBytes []byte) (map[string]interface{}, error) {
	x, err := exif.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, err // 没有EXIF数据是正常情况
	}

	metadata := make(map[string]interface{})

	// 获取一些常见的EXIF字段
	commonTags := []exif.FieldName{
		exif.ImageWidth,
		exif.ImageLength,
		exif.Make,
		exif.Model,
		exif.DateTime,
		exif.Orientation,
	}

	for _, tagName := range commonTags {
		tag, err := x.Get(tagName)
		if err == nil && tag != nil {
			val, err := tag.StringVal()
			if err == nil && val != "" {
				metadata[string(tagName)] = val
			}
		}
	}

	return metadata, nil
}
