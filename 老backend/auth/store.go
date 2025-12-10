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
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return fmt.Errorf("创建表失败: %w", err)
	}

	s.db = db

	// 初始化默认管理员账户
	if err := s.initDefaultAdmin(); err != nil {
		return err
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
			"INSERT INTO users (username, password) VALUES (?, ?)",
			"admin", hashedPassword,
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

// GetUserByUsername 根据用户名获取用户
func (s *Store) GetUserByUsername(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var user User
	err := s.db.QueryRow(
		"SELECT id, username, password, created_at, updated_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt)

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
		"SELECT id, username, password, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt)

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
