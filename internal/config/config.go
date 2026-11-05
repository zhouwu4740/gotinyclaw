// Package config 基于 github.com/joho/godotenv 提供 .env 文件的读取与写入能力。
//
// 读取策略：.env 文件中的变量不会覆盖已存在于系统环境变量中的同名变量，
// 即优先级为：真实环境变量 > .env 文件。
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const dotEnvFile = ".env"

// LoadDotEnv 从 .env 文件加载环境变量到当前进程。
// 文件不存在时不报错（视为无配置），返回 nil。
func LoadDotEnv() error {
	if err := godotenv.Load(dotEnvFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("load .env failed: %w", err)
	}
	return nil
}
