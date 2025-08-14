# Telegram to Cloudflare 图片上传bot

[![Go Version](https://img.shields.io/badge/Go-1.23.4-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## ✨ 功能特性

- 🔐 **用户授权管理** - 经授权用户才能使用
- ✅ **验证文件** - 验证图片是否符合 Cloudflare Images 的要求
- 📝 **日志** - 支持多级别日志记录和文件输出

## 🎯 支持的图片格式和限制

根据 Cloudflare Images 的要求：
- **格式支持**: JPEG、PNG、GIF
- **最大尺寸**: 12,000 × 12,000 像素
- **最大文件大小**: 10 MB
- **最大像素数**: 1亿像素（静态图片）/ 5千万像素（动画GIF）

## 🚀 快速开始

### 1. 环境要求

- Go 1.23.4 或更高版本
- Telegram Bot Token
- Cloudflare Account ID 和 API Token

### 2. 下载和安装

#### 方式一：下载预编译版本
从 [Releases](https://github.com/sam13142023/telegram-cf-bot/releases) 页面下载对应系统的可执行文件。

#### 方式二：从源码编译
```bash
git clone https://github.com/sam13142023/telegram-cf-bot.git
cd telegram-cf-bot
go mod tidy
go build -o telegram-cf-bot.exe .
```

### 3. 配置bot

复制配置模板：
```bash
cp config.yaml.example config.yaml
```

编辑 `config.yaml` 文件：
```yaml
# Telegram 机器人配置
telegram:
  bot_token: "YOUR_TELEGRAM_BOT_TOKEN"

# Cloudflare 配置
cloudflare:
  account_id: "YOUR_CLOUDFLARE_ACCOUNT_ID"
  api_token: "YOUR_CLOUDFLARE_API_TOKEN"

# 授权用户ID列表
authorized_users:
  - 123456789  # 替换为实际的用户ID
  - 987654321  # 可以添加多个用户

# 管理员用户ID（可选）
admin_id: 123456789

# 日志配置
logging:
  level: "info"              # 日志级别: debug, info, warn, error, fatal
  to_file: true              # 是否输出到文件
  file_path: "logs/bot.log"  # 日志文件路径
```

### 4. 获取的 Token 和 ID

#### Telegram Bot Token
1. 在 Telegram 中搜索 `@BotFather`
2. 发送 `/newbot` 命令创建新机器人
3. 按提示设置机器人名称和用户名
4. 获得 Bot Token，格式类似：`123456789:ABCdefGHIjklMNOpqrsTUVwxyz`

#### Cloudflare 配置
1. **Account ID**：
   - 登录 [Cloudflare Dashboard](https://dash.cloudflare.com/)
   - 在右侧边栏找到 Account ID

2. **API Token**：
   - 进入 `My Profile` > `API Tokens`
   - 点击 `Create Token`
   - 使用 `Custom token` 模板
   - 权限设置：`Cloudflare Images:Edit`
   - 账户资源：`Include - 你的账户`

#### 用户ID获取
1. 在 Telegram 中搜索 `@userinfobot`
2. 发送任意消息获取你的用户ID
3. 将用户ID添加到配置文件的 `authorized_users` 列表中

### 5. 运行机器人

```bash
# Windows
telegram-cf-bot.exe

# Linux/macOS
./telegram-cf-bot
```

## 📋 使用说明

### 上传图片：
   - **推荐方式**：以文件形式发送图片（保持原始质量）
   - **压缩方式**：直接发送图片（可能被压缩，需确认上传）

### 命令列表

- `/start` - 启动机器人并显示使用说明
-  /auth && /unauth - 授权或取消授权用户


## 🛠️ 技术栈

- **语言**: Go 1.23.4
- **Telegram库**: telebot v3.3.8
- **配置解析**: gopkg.in/yaml.v3
- **图片处理**: github.com/rwcarlsen/goexif
- **日志系统**: github.com/sirupsen/logrus

## 📝 日志系统

机器人内置完善的日志系统，支持：
- 多级别日志（debug、info、warn、error、fatal）
- 控制台和文件双重输出
- 用户操作追踪
- 错误详情记录

查看日志文件：
```bash
tail -f logs/bot.log
```

## 🔧 开发指南

### 本地开发环境

```bash
# 克隆项目
git clone https://github.com/sam13142023/telegram-cf-bot.git
cd telegram-cf-bot

# 安装依赖
go mod tidy

# 运行开发版本
go run main.go
```

### 构建生产版本

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o telegram-cf-bot.exe .

# Linux
GOOS=linux GOARCH=amd64 go build -o telegram-cf-bot .

# macOS
GOOS=darwin GOARCH=amd64 go build -o telegram-cf-bot .
```

## ❗ 常见问题

### Q: 机器人无法启动？
**A**: 检查以下项目：
- 配置文件 `config.yaml` 是否存在且格式正确
- Bot Token 是否有效
- Cloudflare API Token 是否有正确的权限

### Q: 提示"没有权限使用机器人"？
**A**: 确保你的用户ID已添加到 `authorized_users` 列表中。

### Q: 图片上传失败？
**A**: 检查：
- 图片是否超过10MB或12000×12000像素
- Cloudflare API Token 是否有 `Cloudflare Images:Edit` 权限
- 网络连接是否正常

### Q: 如何查看详细错误信息？
**A**: 
- 将日志级别设置为 `debug`
- 查看 `logs/bot.log` 文件
- 检查控制台输出

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📞 支持

如果你在使用过程中遇到任何问题，可以：
- 提交 [GitHub Issue](https://github.com/sam13142023/telegram-cf-bot/issues)
- 查看项目文档
- 检查日志文件获取详细错误信息

---

⭐ 如果这个项目对你有帮助，请给它一个Star！
