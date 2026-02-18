// Package constants provides application-wide constants.
package constants

import "time"

// Cloudflare Image API limits.
const (
	MaxImageDimension    = 12000             // Maximum width or height in pixels
	MaxImageArea         = 100 * 1000 * 1000 // 100 million pixels for static images
	MaxAnimatedArea      = 50 * 1000 * 1000  // 50 million pixels for animated GIFs
	MaxFileSizeBytes     = 10 * 1024 * 1024  // 10 MB
	MaxMetadataSizeBytes = 1024              // 1 KB
)

// Timeouts.
const (
	HTTPClientTimeout = 30 * time.Second
	ShutdownTimeout   = 5 * time.Second
	ContextTimeout    = 10 * time.Second
)

// File naming.
const (
	RandomStringLength = 8
	FilenameTemplate   = "%d_%d_%s.jpg" // userID_timestamp_random
)

// Telegram bot settings.
const (
	DefaultLogLevel    = "info"
	DefaultLogFilePath = "logs/bot.log"
	UpdateInterval     = 60 // seconds for polling interval
)

// HTTP status codes for logging.
const (
	StatusOK           = 200
	StatusBadRequest   = 400
	StatusUnauthorized = 401
	StatusNotFound     = 404
	StatusServerError  = 500
)

// Supported image formats.
var SupportedImageFormats = map[string]bool{
	"jpeg": true,
	"jpg":  true,
	"png":  true,
	"gif":  true,
}

// Metadata fields to extract from EXIF.
var ExifFields = []string{
	"Make",
	"Model",
	"DateTime",
	"ImageWidth",
	"ImageLength",
	"Orientation",
}
