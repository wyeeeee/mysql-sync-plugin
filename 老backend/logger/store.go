package logger

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Store 日志存储
type Store struct {
	db *sql.DB
	mu sync.RWMutex
}

var (
	instance *Store
	once     sync.Once
)

// GetStore 获取日志存储单例
func GetStore() *Store {
	once.Do(func() {
		instance = &Store{}
	})
	return instance
}

// Init 初始化数据库
func (s *Store) Init(dbPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}

	// 创建日志表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		level TEXT NOT NULL,
		module TEXT NOT NULL,
		action TEXT NOT NULL,
		message TEXT NOT NULL,
		detail TEXT,
		ip TEXT,
		user_agent TEXT,
		duration INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
	CREATE INDEX IF NOT EXISTS idx_logs_module ON logs(module);
	CREATE INDEX IF NOT EXISTS idx_logs_created_at ON logs(created_at);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return fmt.Errorf("创建表失败: %w", err)
	}

	s.db = db
	return nil
}

// Close 关闭数据库连接
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Insert 插入日志
func (s *Store) Insert(entry *LogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	query := `
	INSERT INTO logs (level, module, action, message, detail, ip, user_agent, duration, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		entry.Level,
		entry.Module,
		entry.Action,
		entry.Message,
		entry.Detail,
		entry.IP,
		entry.UserAgent,
		entry.Duration,
		entry.CreatedAt,
	)

	return err
}

// Query 查询日志
func (s *Store) Query(q *LogQuery) ([]LogEntry, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		return nil, 0, fmt.Errorf("数据库未初始化")
	}

	// 构建查询条件
	var conditions []string
	var args []interface{}

	if q.Level != "" {
		conditions = append(conditions, "level = ?")
		args = append(args, q.Level)
	}
	if q.Module != "" {
		conditions = append(conditions, "module = ?")
		args = append(args, q.Module)
	}
	if q.Action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, q.Action)
	}
	if q.Keyword != "" {
		conditions = append(conditions, "(message LIKE ? OR detail LIKE ?)")
		keyword := "%" + q.Keyword + "%"
		args = append(args, keyword, keyword)
	}
	if q.StartTime != "" {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, q.StartTime)
	}
	if q.EndTime != "" {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, q.EndTime)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM logs %s", whereClause)
	var total int64
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 分页参数
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	offset := (q.Page - 1) * q.PageSize

	// 查询数据
	dataQuery := fmt.Sprintf(`
		SELECT id, level, module, action, message, detail, ip, user_agent, duration, created_at
		FROM logs %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, q.PageSize, offset)
	rows, err := s.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []LogEntry
	for rows.Next() {
		var entry LogEntry
		var detail, ip, userAgent sql.NullString
		var duration sql.NullInt64

		if err := rows.Scan(
			&entry.ID,
			&entry.Level,
			&entry.Module,
			&entry.Action,
			&entry.Message,
			&detail,
			&ip,
			&userAgent,
			&duration,
			&entry.CreatedAt,
		); err != nil {
			return nil, 0, err
		}

		entry.Detail = detail.String
		entry.IP = ip.String
		entry.UserAgent = userAgent.String
		entry.Duration = duration.Int64

		entries = append(entries, entry)
	}

	return entries, total, nil
}

// GetStats 获取统计信息
func (s *Store) GetStats() (*LogStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	stats := &LogStats{
		LevelCounts: make(map[string]int64),
	}

	// 总数
	s.db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&stats.TotalCount)

	// 今日数量（使用本地时间范围查询）
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Format("2006-01-02 15:04:05")
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location()).Format("2006-01-02 15:04:05")
	s.db.QueryRow("SELECT COUNT(*) FROM logs WHERE created_at >= ? AND created_at <= ?", todayStart, todayEnd).Scan(&stats.TodayCount)

	// 错误数量
	s.db.QueryRow("SELECT COUNT(*) FROM logs WHERE level = 'ERROR'").Scan(&stats.ErrorCount)

	// 各级别数量
	rows, err := s.db.Query("SELECT level, COUNT(*) FROM logs GROUP BY level")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var level string
			var count int64
			if rows.Scan(&level, &count) == nil {
				stats.LevelCounts[level] = count
			}
		}
	}

	return stats, nil
}

// CleanOldLogs 清理旧日志
func (s *Store) CleanOldLogs(days int) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db == nil {
		return 0, fmt.Errorf("数据库未初始化")
	}

	cutoff := time.Now().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")
	result, err := s.db.Exec("DELETE FROM logs WHERE created_at < ?", cutoff)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
