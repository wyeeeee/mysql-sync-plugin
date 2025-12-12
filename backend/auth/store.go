package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
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
func (s *Store) Init(dsn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 创建用户表
	createUsersSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		username VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		role VARCHAR(50) NOT NULL DEFAULT 'user',
		display_name VARCHAR(255),
		status VARCHAR(50) NOT NULL DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	createSessionsSQL := `
	CREATE TABLE IF NOT EXISTS sessions (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		user_id BIGINT NOT NULL,
		token VARCHAR(255) UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		INDEX idx_sessions_token (token),
		INDEX idx_sessions_expires (expires_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	createDatasourcesSQL := `
	CREATE TABLE IF NOT EXISTS datasources (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		host VARCHAR(255) NOT NULL,
		port INT NOT NULL,
		database_name VARCHAR(255) NOT NULL,
		username VARCHAR(255) NOT NULL,
		password TEXT NOT NULL,
		created_by BIGINT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (created_by) REFERENCES users(id),
		INDEX idx_datasources_created_by (created_by)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	createDatasourceTablesSQL := `
	CREATE TABLE IF NOT EXISTS datasource_tables (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		datasource_id BIGINT NOT NULL,
		table_name VARCHAR(255) NOT NULL,
		table_alias VARCHAR(255),
		query_mode VARCHAR(50) NOT NULL DEFAULT 'table',
		custom_sql TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (datasource_id) REFERENCES datasources(id) ON DELETE CASCADE,
		UNIQUE KEY unique_datasource_table (datasource_id, table_name),
		INDEX idx_datasource_tables_datasource (datasource_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	createFieldMappingsSQL := `
	CREATE TABLE IF NOT EXISTS field_mappings (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		datasource_table_id BIGINT NOT NULL,
		field_name VARCHAR(255) NOT NULL,
		field_alias VARCHAR(255) NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (datasource_table_id) REFERENCES datasource_tables(id) ON DELETE CASCADE,
		UNIQUE KEY unique_table_field (datasource_table_id, field_name),
		INDEX idx_field_mappings_table (datasource_table_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	createUserDatasourcePermissionsSQL := `
	CREATE TABLE IF NOT EXISTS user_datasource_permissions (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		user_id BIGINT NOT NULL,
		datasource_id BIGINT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (datasource_id) REFERENCES datasources(id) ON DELETE CASCADE,
		UNIQUE KEY unique_user_datasource (user_id, datasource_id),
		INDEX idx_user_datasource_user (user_id),
		INDEX idx_user_datasource_datasource (datasource_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	createUserTablePermissionsSQL := `
	CREATE TABLE IF NOT EXISTS user_table_permissions (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		user_id BIGINT NOT NULL,
		datasource_table_id BIGINT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (datasource_table_id) REFERENCES datasource_tables(id) ON DELETE CASCADE,
		UNIQUE KEY unique_user_table (user_id, datasource_table_id),
		INDEX idx_user_table_user (user_id),
		INDEX idx_user_table_table (datasource_table_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	// 执行所有建表语句
	tables := []string{
		createUsersSQL,
		createSessionsSQL,
		createDatasourcesSQL,
		createDatasourceTablesSQL,
		createFieldMappingsSQL,
		createUserDatasourcePermissionsSQL,
		createUserTablePermissionsSQL,
	}

	for _, sql := range tables {
		if _, err := db.Exec(sql); err != nil {
			db.Close()
			return fmt.Errorf("创建表失败: %w", err)
		}
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
