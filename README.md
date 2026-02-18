# Telegram to Cloudflare Images Bot

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A high-quality Telegram bot for uploading images to Cloudflare Images with proper authorization, validation, and logging.

## ✨ Features

- 🔐 **User Authorization** - Only authorized users can upload images
- ✅ **Image Validation** - Validates images against Cloudflare Images requirements
- 📤 **Cloudflare Integration** - Direct upload to Cloudflare Images API
- 📝 **Structured Logging** - Comprehensive logging with logrus
- 🛡️ **Graceful Shutdown** - Proper signal handling and cleanup
- ⚡ **Concurrent Safe** - Thread-safe operations with proper mutex usage

## 🎯 Supported Image Formats and Limits

Following Cloudflare Images API specifications:
- **Formats**: JPEG, PNG, GIF (including animated)
- **Max Dimensions**: 12,000 × 12,000 pixels
- **Max File Size**: 10 MB
- **Max Pixel Area**: 
  - 100 million pixels (static images)
  - 50 million pixels (animated GIFs)

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

## 🏗️ Project Structure

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

## 🔧 Development

### Available Make Commands

```bash
make build       # Build for current platform
make build-all   # Build for all platforms (darwin/linux/windows)
make test        # Run tests
make run         # Build and run
make dev         # Run in development mode
make lint        # Run linters
make fmt         # Format code
make clean       # Clean build artifacts
make help        # Show all commands
```

### Code Quality Improvements

This rebuilt version includes several code quality improvements:

1. **Clean Architecture**: Proper separation of concerns with internal packages
2. **Custom Error Types**: Structured error handling with error wrapping
3. **Context Management**: Proper context propagation for cancellation
4. **Graceful Shutdown**: Signal handling with proper cleanup
5. **Structured Logging**: Consistent logging with fields and levels
6. **Configuration Validation**: Input validation with meaningful errors
7. **Concurrent Safety**: Thread-safe operations with sync primitives
8. **Constants**: Centralized constants for limits and timeouts
9. **Interface Segregation**: Clean interfaces for testability

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

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📞 Support

- Open an [Issue](https://github.com/sam13142023/telegram-cf-bot/issues)
- Check the logs for error details
- Review this documentation

---

⭐ Star this project if you find it useful!
