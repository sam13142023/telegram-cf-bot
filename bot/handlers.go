package bot

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"telegram-cf-bot/cloudflare"
	"telegram-cf-bot/config"
	"telegram-cf-bot/logger"
	"telegram-cf-bot/validator"

	"gopkg.in/telebot.v3"
)

type Bot struct {
	bot *telebot.Bot
	cfg *config.Config
	// 用于存储待处理的图片FileID
	pendingUploads map[int64]string
	uploadMutex    sync.RWMutex
}

// NewBot 创建一个新的机器人实例
func NewBot(cfg *config.Config) (*Bot, error) {
	pref := telebot.Settings{
		Token: cfg.TelegramBotToken,
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		return nil, logger.LogError(err, "创建Telegram机器人失败")
	}

	logger.WithComponent("bot").Info("Telegram机器人实例创建成功")

	return &Bot{
		bot:            b,
		cfg:            cfg,
		pendingUploads: make(map[int64]string),
	}, nil
}

// HandleStartCommand 处理 /start 命令
func (b *Bot) HandleStartCommand(c telebot.Context) error {
	userID := c.Sender().ID
	username := c.Sender().Username

	// 记录用户操作日志
	logger.LogUserAction(userID, username, "start_command", nil)

	// 1. 验证使用者是否在授权清单内
	if !IsUserAuthorized(c.Sender().ID, b.cfg) {
		logger.WithUserID(userID).Warn("未授权用户尝试使用机器人")
		return c.Send("抱歉，您没有使用此机器人的权限。")
	}

	reply := "欢迎使用Cloudflare图片上传机器人。请以文件形式发送图片。"
	logger.WithUserID(userID).Info("向用户发送欢迎消息")
	return c.Send(reply)
}

// HandlePhoto 处理以"照片"形式发送的图片
// 2. 验证图片发送形式，如果不是以文件发送，向用户确认是否上传
func (b *Bot) HandlePhoto(c telebot.Context) error {
	userID := c.Sender().ID
	username := c.Sender().Username

	// 记录用户操作日志
	logger.LogUserAction(userID, username, "send_photo", nil)

	// 1. 验证使用者是否在授权清单内
	if !IsUserAuthorized(c.Sender().ID, b.cfg) {
		logger.WithUserID(userID).Warn("未授权用户尝试发送照片")
		return c.Send("抱歉，您没有使用此机器人的权限。")
	}

	photo := c.Message().Photo
	if photo == nil {
		logger.WithUserID(userID).Error("未检测到图片")
		return c.Send("未检测到图片")
	}

	logger.WithUserID(userID).WithFields(map[string]interface{}{
		"file_id":   photo.FileID,
		"file_size": photo.FileSize,
	}).Info("用户发送了压缩图片")

	// 将图片FileID存储到会话中
	b.uploadMutex.Lock()
	b.pendingUploads[userID] = photo.FileID
	b.uploadMutex.Unlock()

	// 创建内联键盘
	selector := &telebot.ReplyMarkup{}
	btnConfirm := selector.Data("确认上传", "confirm_upload")
	btnCancel := selector.Data("取消", "cancel_upload")
	selector.Inline(
		selector.Row(btnConfirm, btnCancel),
	)

	reply := "您发送的是一张压缩图片，可能会损失质量和元数据。确定要上传吗？"
	return c.Send(reply, selector)
}

// HandleDocument 处理以"文件"形式发送的内容
func (b *Bot) HandleDocument(c telebot.Context) error {
	userID := c.Sender().ID
	username := c.Sender().Username

	// 记录用户操作日志
	logger.LogUserAction(userID, username, "send_document", nil)

	// 1. 验证使用者是否在授权清单内
	if !IsUserAuthorized(c.Sender().ID, b.cfg) {
		logger.WithUserID(userID).Warn("未授权用户尝试发送文档")
		return c.Send("抱歉，您没有使用此机器人的权限。")
	}

	doc := c.Message().Document
	if doc == nil {
		logger.WithUserID(userID).Error("未检测到文档")
		return c.Send("未检测到文档")
	}

	logger.WithUserID(userID).WithFields(map[string]interface{}{
		"file_id":   doc.FileID,
		"file_name": doc.FileName,
		"file_size": doc.FileSize,
		"mime_type": doc.MIME,
	}).Info("用户发送了文档")

	if !strings.HasPrefix(doc.MIME, "image/") {
		logger.WithUserID(userID).WithFields(map[string]interface{}{
			"mime_type": doc.MIME,
			"file_name": doc.FileName,
		}).Warn("用户发送了非图片文件")
		return c.Send("请发送图片文件，而不是其他类型的文件。")
	}

	// 直接调用处理流程
	return b.processImageUpload(c, doc.FileID)
}

// HandleCallbackQuery 处理内联键盘的回调
func (b *Bot) HandleCallbackQuery(c telebot.Context) error {
	callback := c.Callback()
	if callback == nil {
		return nil
	}

	userID := c.Sender().ID
	username := c.Sender().Username

	// 回答回调请求
	err := c.Respond()
	if err != nil {
		logger.WithUserID(userID).WithError(err).Error("回答回调失败")
	}

	data := callback.Data
	// 清理回调数据中的空白字符
	data = strings.TrimSpace(data)

	logger.WithUserID(userID).WithFields(map[string]interface{}{
		"callback_data": data,
	}).Debug("收到用户回调")

	switch data {
	case "confirm_upload":
		logger.LogUserAction(userID, username, "confirm_upload", nil)

		// 从会话中获取图片FileID
		b.uploadMutex.RLock()
		fileID, exists := b.pendingUploads[userID]
		b.uploadMutex.RUnlock()

		if !exists {
			logger.WithUserID(userID).Error("未找到用户的待处理图片")
			return c.Edit("错误：未找到待处理的图片，请重新发送图片。")
		}

		logger.WithUserID(userID).WithFields(map[string]interface{}{
			"file_id": fileID,
		}).Info("找到待处理图片，开始上传")

		// 清理会话数据
		b.uploadMutex.Lock()
		delete(b.pendingUploads, userID)
		b.uploadMutex.Unlock()

		// 编辑确认消息
		err := c.Edit("好的，正在处理图片...")
		if err != nil {
			logger.WithUserID(userID).WithError(err).Error("编辑消息失败")
		}

		// 处理图片上传
		return b.processImageUpload(c, fileID)

	case "cancel_upload":
		logger.LogUserAction(userID, username, "cancel_upload", nil)

		// 清理会话数据
		b.uploadMutex.Lock()
		delete(b.pendingUploads, userID)
		b.uploadMutex.Unlock()

		return c.Edit("已取消上传。")

	default:
		logger.WithUserID(userID).WithFields(map[string]interface{}{
			"callback_data": data,
		}).Warn("未知的回调数据")
		return c.Edit("未知的操作。")
	}
}

// processImageUpload 是下载、验证和上传的统一处理函数
func (b *Bot) processImageUpload(c telebot.Context, fileID string) error {
	userID := c.Sender().ID

	// 发送处理中的消息
	msg, err := c.Bot().Send(c.Chat(), "图片已接收，正在下载...")
	if err != nil {
		logger.WithUserID(userID).WithError(err).Error("发送消息失败")
	}

	// 从Telegram下载文件
	file, err := c.Bot().FileByID(fileID)
	if err != nil {
		if msg != nil {
			_, _ = c.Bot().Edit(msg, "错误：无法获取文件信息。")
		}
		return logger.LogError(err, "获取文件信息失败")
	}

	// 下载文件内容
	fileURL := c.Bot().URL + "/file/bot" + c.Bot().Token + "/" + file.FilePath
	resp, err := http.Get(fileURL)
	if err != nil {
		if msg != nil {
			_, _ = c.Bot().Edit(msg, "错误：无法下载文件。")
		}
		return logger.LogError(err, "下载文件失败")
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.WithUserID(userID).WithError(closeErr).Warn("关闭HTTP响应体失败")
		}
	}()

	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		if msg != nil {
			_, _ = c.Bot().Edit(msg, "错误：读取文件失败。")
		}
		return logger.LogError(err, "读取文件失败")
	}

	// 验证图片
	if msg != nil {
		_, _ = c.Bot().Edit(msg, "图片已下载，正在验证...")
	}

	validationResult, err := validator.ValidateImage(imageBytes)
	if err != nil {
		if msg != nil {
			_, _ = c.Bot().Edit(msg, fmt.Sprintf("图片验证失败: %s", err.Error()))
		}
		return logger.LogError(err, "图片验证失败")
	}

	// 上传图片
	if msg != nil {
		_, _ = c.Bot().Edit(msg, "验证通过，正在上传到 Cloudflare...")
	}

	response, err := cloudflare.UploadImage(imageBytes, userID, validationResult.Metadata, b.cfg)
	if err != nil {
		if msg != nil {
			_, _ = c.Bot().Edit(msg, fmt.Sprintf("❌ 上传失败: %s", err.Error()))
		}
		return logger.LogError(err, "上传失败")
	}

	// 获取图片URL
	imageURL, err := cloudflare.GetImageURL(response)
	if err != nil {
		if msg != nil {
			_, _ = c.Bot().Edit(msg, fmt.Sprintf("❌ 获取图片URL失败: %s", err.Error()))
		}
		return logger.LogError(err, "获取图片URL失败")
	}

	// 返回成功信息和URL
	successText := fmt.Sprintf("✅ 上传成功！\n\n图片URL:\n%s", imageURL)
	if msg != nil {
		_, err := c.Bot().Edit(msg, successText)
		return err
	}
	return c.Send(successText)
}

// Start 启动机器人
func (b *Bot) Start() {
	// 注册处理器
	b.bot.Handle("/start", b.HandleStartCommand)
	b.bot.Handle("/auth", b.HandleAuthCommand)
	b.bot.Handle("/unauth", b.HandleUnauthCommand)
	b.bot.Handle(telebot.OnPhoto, b.HandlePhoto)
	b.bot.Handle(telebot.OnDocument, b.HandleDocument)
	b.bot.Handle(telebot.OnCallback, b.HandleCallbackQuery)

	logger.WithFields(map[string]interface{}{
		"username": b.bot.Me.Username,
	}).Info("机器人已启动")
	b.bot.Start()
}
