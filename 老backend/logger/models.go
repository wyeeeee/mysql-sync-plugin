package logger

import "time"

// LogLevel 日志级别
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// LogEntry 日志条目
type LogEntry struct {
	ID        int64     `json:"id"`
	Level     LogLevel  `json:"level"`
	Module    string    `json:"module"`
	Action    string    `json:"action"`
	Message   string    `json:"message"`
	Detail    string    `json:"detail,omitempty"`
	IP        string    `json:"ip,omitempty"`
	UserAgent string    `json:"userAgent,omitempty"`
	Duration  int64     `json:"duration,omitempty"` // 毫秒
	CreatedAt time.Time `json:"createdAt"`
}

// LogQuery 日志查询参数
type LogQuery struct {
	Level     LogLevel `form:"level"`
	Module    string   `form:"module"`
	Action    string   `form:"action"`
	Keyword   string   `form:"keyword"`
	StartTime string   `form:"startTime"`
	EndTime   string   `form:"endTime"`
	Page      int      `form:"page"`
	PageSize  int      `form:"pageSize"`
}

// LogStats 日志统计
type LogStats struct {
	TotalCount  int64            `json:"totalCount"`
	TodayCount  int64            `json:"todayCount"`
	ErrorCount  int64            `json:"errorCount"`
	LevelCounts map[string]int64 `json:"levelCounts"`
}
