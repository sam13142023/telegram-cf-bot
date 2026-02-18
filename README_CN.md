# Telegram 图床机器人

[![Go 版本](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org/)
[![许可证](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## ✨ 功能特点

- 🔐 **用户授权** - 只有授权用户才能上传图片
- ✅ **图片验证** - 验证图片是否符合 Cloudflare Images 的要求
- 📝 **结构化日志** - 使用 logrus 记录详细日志
- ⚡ **并发安全** - 线程安全

## 🎯 支持的图片格式和限制

遵循 Cloudflare Images API 规范：
- **格式**: JPEG, PNG, ~~GIF~~（由于telegram API限制，不支持）
- **最大尺寸**: 12,000 × 12,000 像素
- **最大文件大小**: 10 MB
- **最大像素面积**:
  - 静态图片：1 亿像素
  - 动图：5000 万像素

## 🚀 快速开始

### 前置要求

- Go 1.23 或更高版本
- Telegram Bot Token
- Cloudflare 账户 ID 和 API Token

### 安装

```bash
# 克隆仓库
git clone https://github.com/sam13142023/telegram-cf-bot.git
cd telegram-cf-bot

# 下载依赖
make deps

# 构建二进制文件
make build
```

### 配置

复制示例配置文件：

```bash
cp config.yaml.example config.yaml
```

编辑 `config.yaml`：

```yaml
# Telegram Bot 配置
telegram:
  bot_token: "YOUR_TELEGRAM_BOT_TOKEN"

# Cloudflare API 配置
cloudflare:
  account_id: "YOUR_CLOUDFLARE_ACCOUNT_ID"
  api_token: "YOUR_CLOUDFLARE_API_TOKEN"

# 授权用户（Telegram 用户 ID）
authorized_users:
  - 123456789

# 管理员用户 ID（用于用户管理）
admin_id: 123456789

# 日志配置
logging:
  level: "info"              # debug, info, warn, error, fatal
  to_file: true
  file_path: "logs/bot.log"
```

### 运行

```bash
# 运行二进制文件
./telegram-cf-bot

# 或使用 make
make run

# 开发模式
make dev
```

## 📖 使用说明

### 上传图片

1. **推荐方式**：以文件形式发送图片（保留原始质量）
2. **替代方式**：以照片形式发送（会提示确认）

### 命令

- `/start` - 启动机器人并查看欢迎信息
- `/auth <user_id>` - 添加用户到授权列表（仅管理员）
- `/unauth <user_id>` - 从授权列表移除用户（仅管理员）

### 获取必需的 ID

**Telegram Bot Token：**
1. 在 Telegram 中联系 [@BotFather](https://t.me/botfather)
2. 使用 `/newbot` 命令
3. 按照指示创建机器人并获取 Token

**Cloudflare 凭证：**
1. 登录 [Cloudflare 控制台](https://dash.cloudflare.com/)
2. 在右侧边栏找到账户 ID
3. 转到 "我的个人资料" → "API 令牌" → "创建令牌"
4. 使用自定义令牌，设置权限为 `Cloudflare Images:Edit`

**Telegram 用户 ID：**
1. 联系 [@userinfobot](https://t.me/userinfobot)
2. 将显示你的用户 ID

## 📝 日志

日志同时输出到控制台和文件（如果已配置）。日志级别：

- `debug` - 详细的调试信息
- `info` - 一般操作信息
- `warn` - 警告消息
- `error` - 错误情况
- `fatal` - 致命错误（退出应用程序）

查看日志：
```bash
tail -f logs/bot.log
```

## 🐛 故障排除

### 机器人无法启动
- 检查 `config.yaml` 是否存在且格式正确
- 验证 Bot Token 是否正确
- 确保 Cloudflare 凭证具有适当的权限

### "未授权"错误
- 将你的 Telegram 用户 ID 添加到 `authorized_users`
- 使用 @userinfobot 验证 ID

### 上传失败
- 检查图片大小和尺寸（参见上面的限制）
- 验证 Cloudflare API Token 是否具有 Images:Edit 权限
- 查看日志获取详细的错误信息

## 🔧 开发

### 🏗️ 项目结构

```
telegram-cf-bot/
├── cmd/
│   └── bot/
│       └── main.go              # 应用程序入口
├── internal/
│   ├── bot/
│   │   └── bot.go               # Telegram 机器人实现
│   ├── cloudflare/
│   │   └── client.go            # Cloudflare API 客户端
│   ├── config/
│   │   └── config.go            # 配置管理
│   ├── constants/
│   │   └── constants.go         # 应用常量
│   ├── errors/
│   │   └── errors.go            # 自定义错误类型
│   ├── logger/
│   │   └── logger.go            # 结构化日志
│   └── validator/
│       └── validator.go         # 图片验证
├── config.yaml.example          # 示例配置
├── Makefile                     # 构建自动化
├── go.mod                       # Go 模块定义
└── README.md                    # 本文件
```

## 📄 许可证

MIT 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件。

## 🤝 贡献

欢迎贡献！请遵循以下步骤：

1. Fork 仓库
2. 创建功能分支 (`git checkout -b feature/新功能`)
3. 提交更改 (`git commit -m '添加新功能'`)
4. 推送到分支 (`git push origin feature/新功能`)
5. 打开 Pull Request

## 📞 支持

- 提交 [Issue](https://github.com/sam13142023/telegram-cf-bot/issues)
- 查看日志获取错误详情
- 阅读本文档

---

⭐ 如果觉得这个项目有用，请给它点个星！
