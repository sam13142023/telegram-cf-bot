package main

import (
	"telegram-cf-bot/bot"
	"telegram-cf-bot/config"
	"telegram-cf-bot/logger"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化日志系统
	logLevel := cfg.LogLevel
	if logLevel == "" {
		logLevel = "info" // 默认日志级别
	}

	logFilePath := cfg.LogFilePath
	if logFilePath == "" {
		logFilePath = "logs/bot.log" // 默认日志文件路径
	}

	err := logger.InitLogger(logger.LogLevel(logLevel), cfg.LogToFile, logFilePath)
	if err != nil {
		logger.Fatal("初始化日志系统失败: %v", err)
	}

	logger.Info("机器人启动中...")
	logger.WithFields(map[string]interface{}{
		"log_level":   logLevel,
		"log_to_file": cfg.LogToFile,
		"log_path":    logFilePath,
	}).Info("日志系统初始化完成")

	// 创建机器人实例
	b, err := bot.NewBot(cfg)
	if err != nil {
		logger.Fatal("创建机器人失败: %v", err)
	}

	logger.Info("机器人创建成功，开始启动...")

	// 启动机器人
	b.Start()
}
