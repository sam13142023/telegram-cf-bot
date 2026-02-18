# Telegram Image Hosting Bot

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## ✨ Features

- 🔐 **User Authorization** - Only authorized users can upload images
- ✅ **Image Validation** - Validates images against Cloudflare Images requirements
- 📝 **Structured Logging** - Comprehensive logging with logrus
- ⚡ **Concurrent Safe** - Thread-safe operations

## 🎯 Supported Image Formats and Limits

Following Cloudflare Images API specifications:
- **Formats**: JPEG, PNG, ~~GIF~~ (Not supported due to Telegram API limitations)
- **Max Dimensions**: 12,000 × 12,000 pixels
- **Max File Size**: 10 MB
- **Max Pixel Area**:
  - Static images: 100 million pixels
  - Animated images: 50 million pixels

## 🚀 Quick Start

### Prerequisites

- Go 1.23 or higher
- Telegram Bot Token
- Cloudflare Account ID and API Token

### Installation

```bash
# Clone the repository
git clone https://github.com/sam13142023/telegram-cf-bot.git
cd telegram-cf-bot

# Download dependencies
make deps

# Build the binary
make build
```

### Configuration

Copy the example configuration:

```bash
cp config.yaml.example config.yaml
```

Edit `config.yaml`:

```yaml
# Telegram Bot Configuration
telegram:
  bot_token: "YOUR_TELEGRAM_BOT_TOKEN"

# Cloudflare API Configuration
cloudflare:
  account_id: "YOUR_CLOUDFLARE_ACCOUNT_ID"
  api_token: "YOUR_CLOUDFLARE_API_TOKEN"

# Authorized Users (Telegram user IDs)
authorized_users:
  - 123456789

# Admin User ID (for user management)
admin_id: 123456789

# Logging Configuration
logging:
  level: "info"              # debug, info, warn, error, fatal
  to_file: true
  file_path: "logs/bot.log"
```

### Running

```bash
# Run the binary
./telegram-cf-bot

# Or use make
make run

# Development mode
make dev
```

## 📖 Usage

### Uploading Images

1. **Recommended**: Send image as file (preserves original quality)
2. **Alternative**: Send as photo (will prompt for confirmation)

### Commands

- `/start` - Start the bot and see welcome message
- `/auth <user_id>` - Add user to authorized list (admin only)
- `/unauth <user_id>` - Remove user from authorized list (admin only)

### Getting Required IDs

**Telegram Bot Token:**
1. Message [@BotFather](https://t.me/botfather) on Telegram
2. Use `/newbot` command
3. Follow instructions to create bot and get token

**Cloudflare Credentials:**
1. Log in to [Cloudflare Dashboard](https://dash.cloudflare.com/)
2. Find Account ID in the right sidebar
3. Go to My Profile → API Tokens → Create Token
4. Use Custom token with `Cloudflare Images:Edit` permission

**Telegram User ID:**
1. Message [@userinfobot](https://t.me/userinfobot)
2. Your user ID will be displayed

## 📝 Logging

Logs are written to both console and file (if configured). Log levels:

- `debug` - Detailed debugging information
- `info` - General operational information
- `warn` - Warning messages
- `error` - Error conditions
- `fatal` - Fatal errors (exits application)

View logs:
```bash
tail -f logs/bot.log
```

## 🐛 Troubleshooting

### Bot won't start
- Check `config.yaml` exists and is valid
- Verify Bot Token is correct
- Ensure Cloudflare credentials have proper permissions

### "Unauthorized" errors
- Add your Telegram user ID to `authorized_users`
- Verify the ID with @userinfobot

### Upload failures
- Check image size and dimensions (see limits above)
- Verify Cloudflare API Token has Images:Edit permission
- Check logs for detailed error messages

## 🔧 Development

### 🏗️ Project Structure

```
telegram-cf-bot/
├── cmd/
│   └── bot/
│       └── main.go              # Application entry point
├── internal/
│   ├── bot/
│   │   └── bot.go               # Telegram bot implementation
│   ├── cloudflare/
│   │   └── client.go            # Cloudflare API client
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── constants/
│   │   └── constants.go         # Application constants
│   ├── errors/
│   │   └── errors.go            # Custom error types
│   ├── logger/
│   │   └── logger.go            # Structured logging
│   └── validator/
│       └── validator.go         # Image validation
├── config.yaml.example          # Example configuration
├── Makefile                     # Build automation
├── go.mod                       # Go module definition
└── README.md                    # This file
```

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/new-feature`)
3. Commit changes (`git commit -m 'Add new feature'`)
4. Push to branch (`git push origin feature/new-feature`)
5. Open a Pull Request

## 📞 Support

- Open an [Issue](https://github.com/sam13142023/telegram-cf-bot/issues)
- Check the logs for error details
- Review this documentation

---

⭐ Star this project if you find it useful!
