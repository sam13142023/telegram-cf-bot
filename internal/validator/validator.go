// Package validator provides image validation functionality.
package validator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/rwcarlsen/goexif/exif"

	"telegram-cf-bot/internal/constants"
	apperrors "telegram-cf-bot/internal/errors"
	"telegram-cf-bot/internal/logger"
)

// Result contains validation result and metadata.
type Result struct {
	IsValid  bool
	Format   string
	Width    int
	Height   int
	Size     int
	Metadata map[string]interface{}
}

// Validate validates image bytes against Cloudflare limits.
func Validate(imageBytes []byte) (*Result, error) {
	log := logger.WithFields(logger.Fields{
		"file_size": len(imageBytes),
		"component": "validator",
	})

	log.Debug("validating image")

	// Check file size
	if len(imageBytes) > constants.MaxFileSizeBytes {
		log.Warn("image exceeds size limit")
		return nil, apperrors.New(apperrors.ErrImageTooLarge,
			fmt.Sprintf("image size %d exceeds limit %d bytes", len(imageBytes), constants.MaxFileSizeBytes))
	}

	// Decode image config to get dimensions and format
	config, format, err := image.DecodeConfig(bytes.NewReader(imageBytes))
	if err != nil {
		log.WithError(err).Error("failed to decode image")
		return nil, apperrors.Wrap(apperrors.ErrInvalidImage, "failed to decode image", err)
	}

	log.Debugf("image decoded: format=%s, width=%d, height=%d", format, config.Width, config.Height)

	// Validate dimensions
	if config.Width > constants.MaxImageDimension || config.Height > constants.MaxImageDimension {
		return nil, apperrors.New(apperrors.ErrImageTooBig,
			fmt.Sprintf("image dimensions %dx%d exceed limit %d", config.Width, config.Height, constants.MaxImageDimension))
	}

	// Validate pixel area
	area := config.Width * config.Height
	maxArea := constants.MaxImageArea
	if format == "gif" {
		maxArea = constants.MaxAnimatedArea
	}

	if area > maxArea {
		return nil, apperrors.New(apperrors.ErrImageTooBig,
			fmt.Sprintf("image area %d exceeds limit %d", area, maxArea))
	}

	// Extract metadata
	metadata := extractMetadata(imageBytes, format)

	// Validate metadata size
	if metadataJSON, _ := json.Marshal(metadata); len(metadataJSON) > constants.MaxMetadataSizeBytes {
		log.Warn("metadata too large, using simplified version")
		metadata = map[string]interface{}{
			"format": format,
			"width":  config.Width,
			"height": config.Height,
			"size":   len(imageBytes),
		}
	}

	log.Infof("image validation passed: format=%s, dimensions=%dx%d", format, config.Width, config.Height)

	return &Result{
		IsValid:  true,
		Format:   format,
		Width:    config.Width,
		Height:   config.Height,
		Size:     len(imageBytes),
		Metadata: metadata,
	}, nil
}

// extractMetadata extracts EXIF metadata from JPEG images.
func extractMetadata(imageBytes []byte, format string) map[string]interface{} {
	metadata := map[string]interface{}{
		"format": format,
	}

	if format != "jpeg" {
		return metadata
	}

	exifData, err := exif.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		// No EXIF data is normal for many images
		return metadata
	}

	// Extract common EXIF fields
	if make, err := exifData.Get(exif.Make); err == nil {
		if str, err := make.StringVal(); err == nil {
			metadata["camera_make"] = str
		}
	}

	if model, err := exifData.Get(exif.Model); err == nil {
		if str, err := model.StringVal(); err == nil {
			metadata["camera_model"] = str
		}
	}

	if dateTime, err := exifData.Get(exif.DateTime); err == nil {
		if str, err := dateTime.StringVal(); err == nil {
			metadata["date_time"] = str
		}
	}

	return metadata
}

// IsSupportedFormat checks if the image format is supported.
func IsSupportedFormat(format string) bool {
	return constants.SupportedImageFormats[format]
}
