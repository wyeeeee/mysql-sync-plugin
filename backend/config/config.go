package config

import "os"

// Config 应用配置
type Config struct {
	// 服务器配置
	ServerPort string

	// 钉钉配置
	SecretKey string // 与钉钉约定的签名密钥

	// 数据库配置
	DBPath string // SQLite数据库路径

	// 应用配置
	Debug bool
}

// Load 加载配置
func Load() *Config {
	return &Config{
		ServerPort: getEnv("SERVER_PORT", "7139"),
		SecretKey:  getEnv("SECRET_KEY", "your-secret-key-here"),
		DBPath:     getEnv("DB_PATH", "./data/app.db"),
		Debug:      getEnv("DEBUG", "false") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
