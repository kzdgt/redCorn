package redCorn

import (
	"fmt"
	"log"
	"os"
)

// Logger 日志接口
type Logger interface {
	// Debug logs a message at Debug level.
	Debug(args ...interface{})

	// Info logs a message at Info level.
	Info(args ...interface{})

	// Warn logs a message at Warning level.
	Warn(args ...interface{})

	// Error logs a message at Error level.
	Error(args ...interface{})

	// Fatal logs a message at Fatal level
	// and process will exit with status set to 1.
	Fatal(args ...interface{})
}

// DefaultLogger 默认日志实现
type DefaultLogger struct {
	logger *log.Logger
}

// NewDefaultLogger 创建默认日志器
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		logger: log.New(os.Stdout, "[RedCorn] ", log.LstdFlags|log.Lshortfile),
	}
}

// Debug 调试日志
func (d *DefaultLogger) Debug(args ...interface{}) {
	d.logger.Println("[DEBUG]", fmt.Sprint(args...))
}

// Info 信息日志
func (d *DefaultLogger) Info(args ...interface{}) {
	d.logger.Println("[INFO]", fmt.Sprint(args...))
}

// Warn 警告日志
func (d *DefaultLogger) Warn(args ...interface{}) {
	d.logger.Println("[WARN]", fmt.Sprint(args...))
}

// Error 错误日志
func (d *DefaultLogger) Error(args ...interface{}) {
	d.logger.Println("[ERROR]", fmt.Sprint(args...))
}

// Fatal 致命错误日志
func (d *DefaultLogger) Fatal(args ...interface{}) {
	d.logger.Println("[FATAL]", fmt.Sprint(args...))
	os.Exit(1)
}
