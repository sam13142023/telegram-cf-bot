package bot

import (
	"fmt"
	"strconv"
	"strings"

	"telegram-cf-bot/config"
	"telegram-cf-bot/logger"

	"gopkg.in/telebot.v3"
)

// IsUserAuthorized 检查用户是否被授权
func IsUserAuthorized(userID int64, cfg *config.Config) bool {
	// 管理员总是被授权
	if userID == cfg.AdminID {
		logger.WithUserID(userID).Debug("管理员用户通过授权检查")
		return true
	}

	// 检查用户是否在授权列表中
	for _, id := range cfg.AuthorizedUserIDs {
		if id == userID {
			logger.WithUserID(userID).Debug("用户在授权列表中，通过授权检查")
			return true
		}
	}

	logger.WithUserID(userID).Warn("用户未在授权列表中，拒绝访问")
	return false
}

// IsAdmin 检查用户是否为管理员
func IsAdmin(userID int64, cfg *config.Config) bool {
	isAdmin := userID == cfg.AdminID
	logger.WithUserID(userID).WithFields(map[string]interface{}{
		"is_admin": isAdmin,
	}).Debug("管理员权限检查")
	return isAdmin
}

// AddUserToAuthorized 添加用户到授权列表
func (b *Bot) AddUserToAuthorized(userID int64) error {
	// 检查用户是否已经在授权列表中
	for _, id := range b.cfg.AuthorizedUserIDs {
		if id == userID {
			logger.WithUserID(userID).Warn("尝试添加已存在的授权用户")
			return fmt.Errorf("用户已经被授权")
		}
	}

	// 添加用户到授权列表
	b.cfg.AuthorizedUserIDs = append(b.cfg.AuthorizedUserIDs, userID)

	logger.WithUserID(userID).Info("用户已添加到授权列表")

	// 保存配置到文件
	err := b.cfg.SaveConfig()
	if err != nil {
		return logger.LogError(err, "保存配置文件失败")
	}

	logger.WithUserID(userID).Info("授权用户配置已保存")
	return nil
}

// RemoveUserFromAuthorized 从授权列表中移除用户
func (b *Bot) RemoveUserFromAuthorized(userID int64) error {
	// 查找并移除用户
	found := false
	for i, id := range b.cfg.AuthorizedUserIDs {
		if id == userID {
			// 移除用户
			b.cfg.AuthorizedUserIDs = append(b.cfg.AuthorizedUserIDs[:i], b.cfg.AuthorizedUserIDs[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		logger.WithUserID(userID).Warn("尝试移除不存在的授权用户")
		return fmt.Errorf("用户不在授权列表中")
	}

	logger.WithUserID(userID).Info("用户已从授权列表中移除")

	// 保存配置到文件
	err := b.cfg.SaveConfig()
	if err != nil {
		return logger.LogError(err, "保存配置文件失败")
	}

	logger.WithUserID(userID).Info("取消授权用户配置已保存")
	return nil
}

// HandleAuthCommand 处理 /auth 命令（仅管理员）
func (b *Bot) HandleAuthCommand(c telebot.Context) error {
	userID := c.Sender().ID
	username := c.Sender().Username

	logger.LogUserAction(userID, username, "auth_command", nil)

	// 检查是否为管理员
	if !IsAdmin(userID, b.cfg) {
		logger.WithUserID(userID).Warn("非管理员尝试使用auth命令")
		return c.Send("抱歉，您没有权限执行此操作。")
	}

	// 解析命令参数
	args := strings.Fields(c.Text())
	if len(args) != 2 {
		return c.Send("用法: /auth <用户ID>")
	}

	targetUserID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		logger.WithUserID(userID).WithFields(map[string]interface{}{
			"invalid_user_id": args[1],
		}).Error("管理员输入了无效的用户ID")
		return c.Send("无效的用户ID，请输入数字。")
	}

	// 添加用户到授权列表
	err = b.AddUserToAuthorized(targetUserID)
	if err != nil {
		logger.WithUserID(userID).WithFields(map[string]interface{}{
			"target_user_id": targetUserID,
			"error":          err.Error(),
		}).Error("添加授权用户失败")
		return c.Send(fmt.Sprintf("添加用户失败: %s", err.Error()))
	}

	logger.WithUserID(userID).WithFields(map[string]interface{}{
		"target_user_id": targetUserID,
	}).Info("管理员成功添加授权用户")

	return c.Send(fmt.Sprintf("用户 %d 已被添加到授权列表。", targetUserID))
}

// HandleUnauthCommand 处理 /unauth 命令（仅管理员）
func (b *Bot) HandleUnauthCommand(c telebot.Context) error {
	userID := c.Sender().ID
	username := c.Sender().Username

	logger.LogUserAction(userID, username, "unauth_command", nil)

	// 检查是否为管理员
	if !IsAdmin(userID, b.cfg) {
		logger.WithUserID(userID).Warn("非管理员尝试使用unauth命令")
		return c.Send("抱歉，您没有权限执行此操作。")
	}

	// 解析命令参数
	args := strings.Fields(c.Text())
	if len(args) != 2 {
		return c.Send("用法: /unauth <用户ID>")
	}

	targetUserID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		logger.WithUserID(userID).WithFields(map[string]interface{}{
			"invalid_user_id": args[1],
		}).Error("管理员输入了无效的用户ID")
		return c.Send("无效的用户ID，请输入数字。")
	}

	// 从授权列表中移除用户
	err = b.RemoveUserFromAuthorized(targetUserID)
	if err != nil {
		logger.WithUserID(userID).WithFields(map[string]interface{}{
			"target_user_id": targetUserID,
			"error":          err.Error(),
		}).Error("移除授权用户失败")
		return c.Send(fmt.Sprintf("移除用户失败: %s", err.Error()))
	}

	logger.WithUserID(userID).WithFields(map[string]interface{}{
		"target_user_id": targetUserID,
	}).Info("管理员成功移除授权用户")

	return c.Send(fmt.Sprintf("用户 %d 已从授权列表中移除。", targetUserID))
}

// HandleOtherFeatures 处理其他需要授权的功能
func (b *Bot) HandleOtherFeatures(c telebot.Context) error {
	userID := c.Sender().ID

	// 权限检查：用户必须是管理员或在授权列表中
	if IsAdmin(userID, b.cfg) || IsUserAuthorized(userID, b.cfg) {
		// 在这里执行机器人核心功能 - 这个函数将被其他处理器调用
		return nil // 表示用户有权限，可以继续处理
	} else {
		// 如果没有权限，则礼貌地拒绝
		return c.Reply("抱歉，你没有权限使用此机器人。请联系管理员进行授权。")
	}
}
