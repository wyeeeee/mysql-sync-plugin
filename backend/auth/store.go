package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Store 认证存储
type Store struct {
	db *sql.DB
	mu sync.RWMutex
}

var (
	instance *Store
	once     sync.Once
)

// GetStore 获取认证存储单例
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

	// 创建用户表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		display_name TEXT,
		status TEXT NOT NULL DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

	CREATE TABLE IF NOT EXISTS datasources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		host TEXT NOT NULL,
		port INTEGER NOT NULL,
		database_name TEXT NOT NULL,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		created_by INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (created_by) REFERENCES users(id)
	);
	CREATE INDEX IF NOT EXISTS idx_datasources_created_by ON datasources(created_by);

	CREATE TABLE IF NOT EXISTS datasource_tables (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		datasource_id INTEGER NOT NULL,
		table_name TEXT NOT NULL,
		table_alias TEXT,
		query_mode TEXT NOT NULL DEFAULT 'table',
		custom_sql TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (datasource_id) REFERENCES datasources(id) ON DELETE CASCADE,
		UNIQUE(datasource_id, table_name)
	);
	CREATE INDEX IF NOT EXISTS idx_datasource_tables_datasource ON datasource_tables(datasource_id);

	CREATE TABLE IF NOT EXISTS field_mappings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		datasource_table_id INTEGER NOT NULL,
		field_name TEXT NOT NULL,
		field_alias TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (datasource_table_id) REFERENCES datasource_tables(id) ON DELETE CASCADE,
		UNIQUE(datasource_table_id, field_name)
	);
	CREATE INDEX IF NOT EXISTS idx_field_mappings_table ON field_mappings(datasource_table_id);

	CREATE TABLE IF NOT EXISTS user_datasource_permissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		datasource_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (datasource_id) REFERENCES datasources(id) ON DELETE CASCADE,
		UNIQUE(user_id, datasource_id)
	);
	CREATE INDEX IF NOT EXISTS idx_user_datasource_user ON user_datasource_permissions(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_datasource_datasource ON user_datasource_permissions(datasource_id);

	CREATE TABLE IF NOT EXISTS user_table_permissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		datasource_table_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (datasource_table_id) REFERENCES datasource_tables(id) ON DELETE CASCADE,
		UNIQUE(user_id, datasource_table_id)
	);
	CREATE INDEX IF NOT EXISTS idx_user_table_user ON user_table_permissions(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_table_table ON user_table_permissions(datasource_table_id);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return fmt.Errorf("创建表失败: %w", err)
	}

	s.db = db

	// 执行数据库迁移(为已存在的users表添加新字段)
	if err := s.migrateDatabase(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 初始化默认管理员账户
	if err := s.initDefaultAdmin(); err != nil {
		return err
	}

	return nil
}

// migrateDatabase 执行数据库迁移
func (s *Store) migrateDatabase() error {
	// 检查users表是否有role字段
	var hasRole bool
	rows, err := s.db.Query("PRAGMA table_info(users)")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dfltValue interface{}
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		if name == "role" {
			hasRole = true
			break
		}
	}

	// 如果没有role字段,说明是旧数据库,需要迁移
	if !hasRole {
		// 为已存在的users表添加新字段
		alterSQL := `
		ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'admin';
		ALTER TABLE users ADD COLUMN display_name TEXT;
		ALTER TABLE users ADD COLUMN status TEXT NOT NULL DEFAULT 'active';
		`
		if _, err := s.db.Exec(alterSQL); err != nil {
			return fmt.Errorf("添加新字段失败: %w", err)
		}
	}

	return nil
}

// initDefaultAdmin 初始化默认管理员
func (s *Store) initDefaultAdmin() error {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// 创建默认管理员 admin/admin123
		hashedPassword := hashPassword("admin123")
		_, err := s.db.Exec(
			"INSERT INTO users (username, password, role, display_name, status) VALUES (?, ?, ?, ?, ?)",
			"admin", hashedPassword, "admin", "系统管理员", "active",
		)
		if err != nil {
			return fmt.Errorf("创建默认管理员失败: %w", err)
		}
	}

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

// GetDB 获取数据库连接(用于Repository)
func (s *Store) GetDB() *sql.DB {
	return s.db
}

// GetUserByUsername 根据用户名获取用户
func (s *Store) GetUserByUsername(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var user User
	err := s.db.QueryRow(
		"SELECT id, username, password, role, display_name, status, created_at, updated_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.DisplayName, &user.Status, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID 根据ID获取用户
func (s *Store) GetUserByID(id int64) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var user User
	err := s.db.QueryRow(
		"SELECT id, username, password, role, display_name, status, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.DisplayName, &user.Status, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdatePassword 更新密码
func (s *Store) UpdatePassword(userID int64, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	hashedPassword := hashPassword(newPassword)
	_, err := s.db.Exec(
		"UPDATE users SET password = ?, updated_at = ? WHERE id = ?",
		hashedPassword, time.Now(), userID,
	)
	return err
}

// CreateSession 创建会话
func (s *Store) CreateSession(userID int64) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	token := generateToken()
	expiresAt := time.Now().Add(24 * time.Hour) // 24小时过期

	result, err := s.db.Exec(
		"INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)",
		userID, token, expiresAt,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &Session{
		ID:        id,
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}

// GetSessionByToken 根据token获取会话
func (s *Store) GetSessionByToken(token string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var session Session
	err := s.db.QueryRow(
		"SELECT id, user_id, token, expires_at, created_at FROM sessions WHERE token = ? AND expires_at > ?",
		token, time.Now(),
	).Scan(&session.ID, &session.UserID, &session.Token, &session.ExpiresAt, &session.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// DeleteSession 删除会话
func (s *Store) DeleteSession(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

// CleanExpiredSessions 清理过期会话
func (s *Store) CleanExpiredSessions() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	return err
}

// VerifyPassword 验证密码
func VerifyPassword(inputPassword, storedPassword string) bool {
	return hashPassword(inputPassword) == storedPassword
}

// hashPassword 哈希密码
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// generateToken 生成随机token
func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
