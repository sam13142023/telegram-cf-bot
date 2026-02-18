// Package bot provides Telegram bot functionality.
package bot

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/telebot.v3"

	"telegram-cf-bot/internal/cloudflare"
	"telegram-cf-bot/internal/config"
	apperrors "telegram-cf-bot/internal/errors"
	"telegram-cf-bot/internal/logger"
	"telegram-cf-bot/internal/validator"
)

// Bot represents the Telegram bot instance.
type Bot struct {
	telebot        *telebot.Bot
	config         *config.Config
	cfClient       *cloudflare.Client
	httpClient     *http.Client
	pendingUploads map[int64]string
	uploadMutex    sync.RWMutex
	stopChan       chan struct{}
	wg             sync.WaitGroup
}

// New creates a new bot instance.
func New(cfg *config.Config) (*Bot, error) {
	settings := telebot.Settings{
		Token: cfg.Telegram.BotToken,
	}

	tb, err := telebot.NewBot(settings)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInvalidConfig, "failed to create telegram bot", err)
	}

	return &Bot{
		telebot:        tb,
		config:         cfg,
		cfClient:       cloudflare.NewClient(cfg),
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		pendingUploads: make(map[int64]string),
		stopChan:       make(chan struct{}),
	}, nil
}

// Start starts the bot and registers handlers.
func (b *Bot) Start() error {
	logger.WithFields(map[string]interface{}{"username": b.telebot.Me.Username}).Info("starting bot")

	// Register handlers
	b.telebot.Handle("/start", b.handleStart)
	b.telebot.Handle("/auth", b.handleAuth)
	b.telebot.Handle("/unauth", b.handleUnauth)
	b.telebot.Handle(telebot.OnPhoto, b.handlePhoto)
	b.telebot.Handle(telebot.OnDocument, b.handleDocument)
	b.telebot.Handle(telebot.OnCallback, b.handleCallback)

	// Start polling in a goroutine
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		b.telebot.Start()
	}()

	// Wait for stop signal
	<-b.stopChan

	return nil
}

// Stop gracefully stops the bot.
func (b *Bot) Stop() {
	logger.Info("stopping bot")
	close(b.stopChan)
	b.telebot.Stop()
	b.wg.Wait()
	logger.Info("bot stopped")
}

// handleStart handles the /start command.
func (b *Bot) handleStart(c telebot.Context) error {
	userID := c.Sender().ID
	username := c.Sender().Username

	logger.LogUserAction(userID, username, "command_start", nil)

	if !b.config.IsAuthorized(userID) {
		logger.WithUser(userID, username).Warn("unauthorized access attempt")
		return c.Send("抱歉，您没有使用此机器人的权限。")
	}

	return c.Send("欢迎使用 Cloudflare 图片上传机器人。请以文件形式发送图片以保持最佳质量。")
}

// handlePhoto handles photo messages (compressed images).
func (b *Bot) handlePhoto(c telebot.Context) error {
	userID := c.Sender().ID
	username := c.Sender().Username

	logger.LogUserAction(userID, username, "send_photo", nil)

	if !b.config.IsAuthorized(userID) {
		logger.WithUser(userID, username).Warn("unauthorized photo upload attempt")
		return c.Send("抱歉，您没有使用此机器人的权限。")
	}

	photo := c.Message().Photo
	if photo == nil {
		return c.Send("未检测到图片")
	}

	// Store file ID for later
	b.uploadMutex.Lock()
	b.pendingUploads[userID] = photo.FileID
	b.uploadMutex.Unlock()

	logger.WithUser(userID, username).Debug("stored photo for confirmation", "file_id", photo.FileID)

	// Create confirmation keyboard
	selector := &telebot.ReplyMarkup{}
	btnConfirm := selector.Data("确认上传", "confirm_upload")
	btnCancel := selector.Data("取消", "cancel_upload")
	selector.Inline(selector.Row(btnConfirm, btnCancel))

	return c.Send("您发送的是压缩图片，可能会损失质量。确定要上传吗？", selector)
}

// handleDocument handles document messages (files).
func (b *Bot) handleDocument(c telebot.Context) error {
	userID := c.Sender().ID
	username := c.Sender().Username

	logger.LogUserAction(userID, username, "send_document", nil)

	if !b.config.IsAuthorized(userID) {
		logger.WithUser(userID, username).Warn("unauthorized document upload attempt")
		return c.Send("抱歉，您没有使用此机器人的权限。")
	}

	doc := c.Message().Document
	if doc == nil {
		return c.Send("未检测到文件")
	}

	// Check if it's an image
	if !strings.HasPrefix(doc.MIME, "image/") {
		logger.WithUser(userID, username).Warn("non-image file received", "mime", doc.MIME)
		return c.Send("请发送图片文件，不支持其他类型的文件。")
	}

	return b.processImageUpload(c, doc.FileID)
}

// handleCallback handles inline keyboard callbacks.
func (b *Bot) handleCallback(c telebot.Context) error {
	callback := c.Callback()
	if callback == nil {
		return nil
	}

	userID := c.Sender().ID
	username := c.Sender().Username

	// Answer callback to remove loading state
	c.Respond()

	data := strings.TrimSpace(callback.Data)
	logger.WithUser(userID, username).Debug("received callback", "data", data)

	switch data {
	case "confirm_upload":
		logger.LogUserAction(userID, username, "confirm_upload", nil)

		b.uploadMutex.RLock()
		fileID, exists := b.pendingUploads[userID]
		b.uploadMutex.RUnlock()

		if !exists {
			return c.Edit("错误：未找到待处理的图片，请重新发送。")
		}

		// Clear pending upload
		b.uploadMutex.Lock()
		delete(b.pendingUploads, userID)
		b.uploadMutex.Unlock()

		c.Edit("正在处理图片...")
		return b.processImageUpload(c, fileID)

	case "cancel_upload":
		logger.LogUserAction(userID, username, "cancel_upload", nil)

		b.uploadMutex.Lock()
		delete(b.pendingUploads, userID)
		b.uploadMutex.Unlock()

		return c.Edit("已取消上传。")

	default:
		logger.WithUser(userID, username).Warn("unknown callback", "data", data)
		return c.Edit("未知的操作。")
	}
}

// handleAuth handles the /auth command (admin only).
func (b *Bot) handleAuth(c telebot.Context) error {
	return b.handleUserCommand(c, "auth", b.config.AddAuthorizedUser)
}

// handleUnauth handles the /unauth command (admin only).
func (b *Bot) handleUnauth(c telebot.Context) error {
	return b.handleUserCommand(c, "unauth", b.config.RemoveAuthorizedUser)
}

// handleUserCommand handles admin user management commands.
func (b *Bot) handleUserCommand(c telebot.Context, action string, operation func(int64) error) error {
	userID := c.Sender().ID
	username := c.Sender().Username

	logger.LogUserAction(userID, username, "command_"+action, nil)

	if !b.config.IsAdmin(userID) {
		logger.WithUser(userID, username).Warn("non-admin attempted admin command")
		return c.Send("抱歉，只有管理员可以执行此操作。")
	}

	args := strings.Fields(c.Text())
	if len(args) != 2 {
		return c.Send(fmt.Sprintf("用法: /%s <用户ID>", action))
	}

	targetID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		logger.WithUser(userID, username).Error("invalid user ID format", "input", args[1])
		return c.Send("无效的用户ID，请输入数字。")
	}

	if err := operation(targetID); err != nil {
		logger.WithUser(userID, username).WithError(err).Error(action+" failed", "target", targetID)
		return c.Send(fmt.Sprintf("操作失败: %s", err.Error()))
	}

	actionText := map[string]string{
		"auth":   "添加",
		"unauth": "移除",
	}

	logger.WithUser(userID, username).Info(action+" successful", "target", targetID)
	return c.Send(fmt.Sprintf("用户 %d 已成功%s授权列表。", targetID, actionText[action]))
}

// processImageUpload handles the complete image upload flow.
func (b *Bot) processImageUpload(c telebot.Context, fileID string) error {
	userID := c.Sender().ID
	username := c.Sender().Username

	log := logger.WithUser(userID, username)

	// Send processing message
	msg, err := c.Bot().Send(c.Chat(), "正在下载图片...")
	if err != nil {
		log.WithError(err).Error("failed to send status message")
	}

	// Download file from Telegram
	file, err := c.Bot().FileByID(fileID)
	if err != nil {
		if msg != nil {
			c.Bot().Edit(msg, "错误：无法获取文件信息。")
		}
		return apperrors.Wrap(apperrors.ErrDownloadFailed, "failed to get file info", err)
	}

	// Download file content
	fileURL := fmt.Sprintf("%s/file/bot%s/%s", c.Bot().URL, c.Bot().Token, file.FilePath)
	resp, err := b.httpClient.Get(fileURL)
	if err != nil {
		if msg != nil {
			c.Bot().Edit(msg, "错误：无法下载文件。")
		}
		return apperrors.Wrap(apperrors.ErrDownloadFailed, "failed to download file", err)
	}
	defer resp.Body.Close()

	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		if msg != nil {
			c.Bot().Edit(msg, "错误：读取文件失败。")
		}
		return apperrors.Wrap(apperrors.ErrDownloadFailed, "failed to read file", err)
	}

	// Validate image
	if msg != nil {
		c.Bot().Edit(msg, "正在验证图片...")
	}

	validationResult, err := validator.Validate(imageBytes)
	if err != nil {
		if msg != nil {
			c.Bot().Edit(msg, fmt.Sprintf("❌ 验证失败: %s", err.Error()))
		}
		return err
	}

	// Upload to Cloudflare
	if msg != nil {
		c.Bot().Edit(msg, "正在上传到 Cloudflare...")
	}

	uploadResp, err := b.cfClient.Upload(imageBytes, userID, validationResult.Metadata)
	if err != nil {
		if msg != nil {
			c.Bot().Edit(msg, fmt.Sprintf("❌ 上传失败: %s", err.Error()))
		}
		return err
	}

	// Get image URL
	imageURL, err := cloudflare.GetImageURL(uploadResp)
	if err != nil {
		if msg != nil {
			c.Bot().Edit(msg, fmt.Sprintf("❌ 获取图片URL失败: %s", err.Error()))
		}
		return err
	}

	// Send success message
	successText := fmt.Sprintf("✅ 上传成功！\n\n图片URL:\n%s", imageURL)
	if msg != nil {
		_, err = c.Bot().Edit(msg, successText)
	} else {
		err = c.Send(successText)
	}

	return err
}
