package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	// Logger 全局日志实例
	Logger *logrus.Logger
)

// LogLevel 日志级别
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// InitLogger 初始化日志系统
func InitLogger(level LogLevel, logToFile bool, logFilePath string) error {
	Logger = logrus.New()

	// 设置日志级别
	switch level {
	case DebugLevel:
		Logger.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		Logger.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		Logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		Logger.SetLevel(logrus.ErrorLevel)
	case FatalLevel:
		Logger.SetLevel(logrus.FatalLevel)
	default:
		Logger.SetLevel(logrus.InfoLevel)
	}

	// 自定义日志格式
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			// 获取相对路径
			filename := filepath.Base(f.File)
			return "", fmt.Sprintf("[%s:%d]", filename, f.Line)
		},
	})

	// 启用调用者信息
	Logger.SetReportCaller(true)

	// 设置输出
	if logToFile && logFilePath != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(logFilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("创建日志目录失败: %v", err)
		}

		// 打开日志文件
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("打开日志文件失败: %v", err)
		}

		// 同时输出到文件和控制台
		multiWriter := io.MultiWriter(os.Stdout, file)
		Logger.SetOutput(multiWriter)
	} else {
		Logger.SetOutput(os.Stdout)
	}

	return nil
}

// GetLogger 获取日志实例
func GetLogger() *logrus.Logger {
	if Logger == nil {
		// 如果未初始化，使用默认配置
		_ = InitLogger(InfoLevel, false, "")
	}
	return Logger
}

// WithFields 创建带字段的日志条目
func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}

// WithComponent 创建带组件名的日志条目
func WithComponent(component string) *logrus.Entry {
	return GetLogger().WithField("component", component)
}

// WithUserID 创建带用户ID的日志条目
func WithUserID(userID int64) *logrus.Entry {
	return GetLogger().WithField("user_id", userID)
}

// WithRequestID 创建带请求ID的日志条目
func WithRequestID(requestID string) *logrus.Entry {
	return GetLogger().WithField("request_id", requestID)
}

// Debug 调试日志
func Debug(msg string, args ...interface{}) {
	GetLogger().Debugf(msg, args...)
}

// Info 信息日志
func Info(msg string, args ...interface{}) {
	GetLogger().Infof(msg, args...)
}

// Warn 警告日志
func Warn(msg string, args ...interface{}) {
	GetLogger().Warnf(msg, args...)
}

// Error 错误日志
func Error(msg string, args ...interface{}) {
	GetLogger().Errorf(msg, args...)
}

// Fatal 致命错误日志（会退出程序）
func Fatal(msg string, args ...interface{}) {
	GetLogger().Fatalf(msg, args...)
}

// LogError 记录错误并返回格式化的错误
func LogError(err error, msg string, args ...interface{}) error {
	formattedMsg := fmt.Sprintf(msg, args...)
	GetLogger().WithError(err).Error(formattedMsg)
	return fmt.Errorf("%s: %v", formattedMsg, err)
}

// LogAndReturnError 记录错误日志并返回新的错误
func LogAndReturnError(msg string, args ...interface{}) error {
	formattedMsg := fmt.Sprintf(msg, args...)
	GetLogger().Error(formattedMsg)
	return fmt.Errorf(formattedMsg)
}

// LogUserAction 记录用户操作日志
func LogUserAction(userID int64, username string, action string, details interface{}) {
	GetLogger().WithFields(logrus.Fields{
		"user_id":  userID,
		"username": username,
		"action":   action,
		"details":  details,
	}).Info("用户操作")
}

// LogUploadAction 记录文件上传日志
func LogUploadAction(userID int64, username string, filename string, fileSize int64, success bool, errorMsg string) {
	entry := GetLogger().WithFields(logrus.Fields{
		"user_id":   userID,
		"username":  username,
		"filename":  filename,
		"file_size": fileSize,
		"success":   success,
	})

	if success {
		entry.Info("文件上传成功")
	} else {
		entry.WithField("error", errorMsg).Error("文件上传失败")
	}
}

// LogAPICall 记录API调用日志
func LogAPICall(api string, method string, url string, statusCode int, duration int64, err error) {
	entry := GetLogger().WithFields(logrus.Fields{
		"api":         api,
		"method":      method,
		"url":         sanitizeURL(url), // 清理敏感信息
		"status_code": statusCode,
		"duration_ms": duration,
	})

	if err != nil {
		entry.WithError(err).Error("API调用失败")
	} else {
		entry.Info("API调用成功")
	}
}

// sanitizeURL 清理URL中的敏感信息
func sanitizeURL(url string) string {
	// 替换API密钥等敏感信息
	if strings.Contains(url, "?") {
		return strings.Split(url, "?")[0] + "?[QUERY_PARAMS_HIDDEN]"
	}
	return url
}

// LogSystemEvent 记录系统事件日志
func LogSystemEvent(event string, details interface{}) {
	GetLogger().WithFields(logrus.Fields{
		"event":   event,
		"details": details,
	}).Info("系统事件")
}
