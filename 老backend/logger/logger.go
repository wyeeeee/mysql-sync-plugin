package logger

import (
	"fmt"
	"log"
	"time"
)

// Logger 日志记录器
type Logger struct {
	store  *Store
	module string
}

// New 创建日志记录器
func New(module string) *Logger {
	return &Logger{
		store:  GetStore(),
		module: module,
	}
}

// log 记录日志
func (l *Logger) log(level LogLevel, action, message string, detail string) {
	entry := &LogEntry{
		Level:     level,
		Module:    l.module,
		Action:    action,
		Message:   message,
		Detail:    detail,
		CreatedAt: time.Now(),
	}

	// 同时输出到控制台
	log.Printf("[%s] [%s] [%s] %s", level, l.module, action, message)

	// 异步写入数据库
	go func() {
		if err := l.store.Insert(entry); err != nil {
			log.Printf("写入日志失败: %v", err)
		}
	}()
}

// LogWithRequest 记录带请求信息的日志
func (l *Logger) LogWithRequest(level LogLevel, action, message, detail, ip, userAgent string, duration int64) {
	entry := &LogEntry{
		Level:     level,
		Module:    l.module,
		Action:    action,
		Message:   message,
		Detail:    detail,
		IP:        ip,
		UserAgent: userAgent,
		Duration:  duration,
		CreatedAt: time.Now(),
	}

	log.Printf("[%s] [%s] [%s] %s (IP: %s, 耗时: %dms)", level, l.module, action, message, ip, duration)

	go func() {
		if err := l.store.Insert(entry); err != nil {
			log.Printf("写入日志失败: %v", err)
		}
	}()
}

// Debug 调试日志
func (l *Logger) Debug(action, message string) {
	l.log(LevelDebug, action, message, "")
}

// Debugf 格式化调试日志
func (l *Logger) Debugf(action, format string, args ...interface{}) {
	l.log(LevelDebug, action, fmt.Sprintf(format, args...), "")
}

// Info 信息日志
func (l *Logger) Info(action, message string) {
	l.log(LevelInfo, action, message, "")
}

// Infof 格式化信息日志
func (l *Logger) Infof(action, format string, args ...interface{}) {
	l.log(LevelInfo, action, fmt.Sprintf(format, args...), "")
}

// InfoWithDetail 带详情的信息日志
func (l *Logger) InfoWithDetail(action, message, detail string) {
	l.log(LevelInfo, action, message, detail)
}

// Warn 警告日志
func (l *Logger) Warn(action, message string) {
	l.log(LevelWarn, action, message, "")
}

// Warnf 格式化警告日志
func (l *Logger) Warnf(action, format string, args ...interface{}) {
	l.log(LevelWarn, action, fmt.Sprintf(format, args...), "")
}

// Error 错误日志
func (l *Logger) Error(action, message string) {
	l.log(LevelError, action, message, "")
}

// Errorf 格式化错误日志
func (l *Logger) Errorf(action, format string, args ...interface{}) {
	l.log(LevelError, action, fmt.Sprintf(format, args...), "")
}

// ErrorWithDetail 带详情的错误日志
func (l *Logger) ErrorWithDetail(action, message, detail string) {
	l.log(LevelError, action, message, detail)
}
