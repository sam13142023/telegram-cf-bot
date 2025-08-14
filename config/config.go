package config

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 结构体用于存储所有配置信息
type Config struct {
	TelegramBotToken    string  `yaml:"-"`
	CloudflareAccountID string  `yaml:"-"`
	CloudflareAPIToken  string  `yaml:"-"`
	AuthorizedUserIDs   []int64 `yaml:"-"`
	AdminID             int64   `yaml:"-"`
	// 日志配置
	LogLevel    string `yaml:"-"`
	LogToFile   bool   `yaml:"-"`
	LogFilePath string `yaml:"-"`
	configPath  string // 存储配置文件路径，用于保存
}

// YamlConfig 用于解析 YAML 文件的结构体
type YamlConfig struct {
	Telegram struct {
		BotToken string `yaml:"bot_token"`
	} `yaml:"telegram"`
	Cloudflare struct {
		AccountID string `yaml:"account_id"`
		APIToken  string `yaml:"api_token"`
	} `yaml:"cloudflare"`
	AuthorizedUsers []int64 `yaml:"authorized_users"`
	AdminID         int64   `yaml:"admin_id"`
	Logging         struct {
		Level    string `yaml:"level"`
		ToFile   bool   `yaml:"to_file"`
		FilePath string `yaml:"file_path"`
	} `yaml:"logging"`
}

// LoadConfig 从config.yaml文件中加载配置
func LoadConfig() *Config {
	// 获取可执行文件所在目录
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("无法获取可执行文件路径: %v", err)
	}
	execDir := filepath.Dir(execPath)

	// 尝试不同的配置文件路径
	configPaths := []string{
		"config.yaml",                         // 当前工作目录
		filepath.Join(execDir, "config.yaml"), // 可执行文件目录
		filepath.Join(".", "config.yaml"),     // 相对路径
	}

	var yamlConfig YamlConfig
	var configFound bool
	var usedConfigPath string

	for _, configPath := range configPaths {
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err != nil {
				log.Printf("读取配置文件失败 %s: %v", configPath, err)
				continue
			}

			err = yaml.Unmarshal(data, &yamlConfig)
			if err != nil {
				log.Printf("解析配置文件失败 %s: %v", configPath, err)
				continue
			}

			log.Printf("成功加载配置文件: %s", configPath)
			configFound = true
			usedConfigPath = configPath
			break
		}
	}

	if !configFound {
		log.Fatalf("未找到配置文件 config.yaml，请确保配置文件存在于以下路径之一: %v", configPaths)
	}

	// 验证必需的配置项
	if yamlConfig.Telegram.BotToken == "" || yamlConfig.Telegram.BotToken == "YOUR_TELEGRAM_BOT_TOKEN" {
		log.Fatalf("请在 config.yaml 中设置有效的 telegram.bot_token")
	}

	if yamlConfig.Cloudflare.AccountID == "" || yamlConfig.Cloudflare.AccountID == "YOUR_CLOUDFLARE_ACCOUNT_ID" {
		log.Fatalf("请在 config.yaml 中设置有效的 cloudflare.account_id")
	}

	if yamlConfig.Cloudflare.APIToken == "" || yamlConfig.Cloudflare.APIToken == "YOUR_CLOUDFLARE_API_TOKEN" {
		log.Fatalf("请在 config.yaml 中设置有效的 cloudflare.api_token")
	}

	if len(yamlConfig.AuthorizedUsers) == 0 {
		log.Printf("警告: 未设置授权用户列表，机器人将无法使用")
	}

	if yamlConfig.AdminID == 0 {
		log.Printf("警告: 未设置管理员ID (admin_id)，某些管理功能将无法使用")
	}

	config := &Config{
		TelegramBotToken:    yamlConfig.Telegram.BotToken,
		CloudflareAccountID: yamlConfig.Cloudflare.AccountID,
		CloudflareAPIToken:  yamlConfig.Cloudflare.APIToken,
		AuthorizedUserIDs:   yamlConfig.AuthorizedUsers,
		AdminID:             yamlConfig.AdminID,
		configPath:          usedConfigPath,
		LogLevel:            yamlConfig.Logging.Level,
		LogToFile:           yamlConfig.Logging.ToFile,
		LogFilePath:         yamlConfig.Logging.FilePath,
	}

	return config
}

// SaveConfig 保存配置到文件
func (c *Config) SaveConfig() error {
	yamlConfig := YamlConfig{
		Telegram: struct {
			BotToken string `yaml:"bot_token"`
		}{
			BotToken: c.TelegramBotToken,
		},
		Cloudflare: struct {
			AccountID string `yaml:"account_id"`
			APIToken  string `yaml:"api_token"`
		}{
			AccountID: c.CloudflareAccountID,
			APIToken:  c.CloudflareAPIToken,
		},
		AuthorizedUsers: c.AuthorizedUserIDs,
		AdminID:         c.AdminID,
		Logging: struct {
			Level    string `yaml:"level"`
			ToFile   bool   `yaml:"to_file"`
			FilePath string `yaml:"file_path"`
		}{
			Level:    c.LogLevel,
			ToFile:   c.LogToFile,
			FilePath: c.LogFilePath,
		},
	}

	data, err := yaml.Marshal(&yamlConfig)
	if err != nil {
		return err
	}

	return os.WriteFile(c.configPath, data, 0644)
}

// SetConfigPath 设置配置文件路径
func (c *Config) SetConfigPath(path string) {
	c.configPath = path
}
