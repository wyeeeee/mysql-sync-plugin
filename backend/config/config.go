package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config 应用配置
type Config struct {
	// 服务器配置
	ServerPort string `json:"serverPort"`

	// 签名密钥
	SecretKey string `json:"secretKey"`

	// MySQL配置
	MySQL MySQLConfig `json:"mysql"`

	// 应用配置
	Debug bool `json:"debug"`
}

// MySQLConfig MySQL数据库配置
type MySQLConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Load 加载配置
func Load() *Config {
	configFile := "config.json"

	// 检查配置文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// 生成默认配置文件
		defaultConfig := &Config{
			ServerPort: "7138",
			SecretKey:  "your-secret-key-here",
			MySQL: MySQLConfig{
				Host:     "localhost",
				Port:     "3306",
				Database: "mysql_sync_plugin",
				Username: "root",
				Password: "your-mysql-password",
			},
			Debug: true,
		}

		// 写入默认配置
		data, _ := json.MarshalIndent(defaultConfig, "", "  ")
		if err := os.WriteFile(configFile, data, 0644); err != nil {
			fmt.Printf("生成默认配置文件失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("已生成默认配置文件: config.json")
		fmt.Println("请修改配置文件中的 MySQL 密码后重新启动服务")
		os.Exit(0)
	}

	// 读取配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("读取配置文件失败: %v\n", err)
		os.Exit(1)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Printf("解析配置文件失败: %v\n", err)
		os.Exit(1)
	}

	// 验证配置
	if cfg.MySQL.Password == "your-mysql-password" {
		fmt.Println("错误: 请先修改配置文件中的 MySQL 密码")
		os.Exit(1)
	}

	return &cfg
}

// GetDSN 获取MySQL连接字符串
func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.MySQL.Username,
		c.MySQL.Password,
		c.MySQL.Host,
		c.MySQL.Port,
		c.MySQL.Database,
	)
}
