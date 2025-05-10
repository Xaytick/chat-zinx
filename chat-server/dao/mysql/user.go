package mysql

import (
	"fmt"
	"time"

	"github.com/Xaytick/chat-zinx/chat-server/conf"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

// User 用户数据结构
type User struct {
	ID           int64     `db:"id"`
	UserID       string    `db:"user_id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	Salt         string    `db:"salt"`
	Email        string    `db:"email"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	LastLogin    time.Time `db:"last_login"`
}

// InitMySQL 初始化MySQL连接
func InitMySQL(cfg *conf.MySQLConfig) (err error) {
	// 构建DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	// 连接数据库
	DB, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		return err
	}

	// 设置最大连接数和最大空闲连接数
	DB.SetMaxOpenConns(cfg.MaxOpenConns)
	DB.SetMaxIdleConns(cfg.MaxIdleConns)

	// 测试连接
	err = DB.Ping()
	if err != nil {
		return err
	}

	// 初始化表结构
	err = initTables()
	if err != nil {
		return err
	}

	return nil
}

// initTables 初始化表结构
func initTables() error {
	// 用户表
	createUserTable := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		user_id VARCHAR(36) NOT NULL UNIQUE,
		username VARCHAR(50) NOT NULL UNIQUE,
		password_hash VARCHAR(100) NOT NULL,
		salt VARCHAR(50) DEFAULT '',
		email VARCHAR(100) DEFAULT '',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		last_login TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_username (username),
		INDEX idx_user_id (user_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`
	_, err := DB.Exec(createUserTable)
	return err
}

// InsertUser 插入新用户
func InsertUser(user *User) (string, error) {
	sqlStr := "INSERT INTO users(user_id, username, password_hash, salt, email, created_at, updated_at, last_login) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := DB.Exec(sqlStr, user.UserID, user.Username, user.PasswordHash, user.Salt, user.Email, user.CreatedAt, user.UpdatedAt, user.LastLogin)
	if err != nil {
		return "", err
	}
	return user.UserID, nil
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(username string) (*User, error) {
	var user User
	sqlStr := "SELECT id, user_id, username, password_hash, salt, email, created_at, updated_at, last_login FROM users WHERE username = ?"
	err := DB.Get(&user, sqlStr, username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 根据ID获取用户
func GetUserByID(userID string) (*User, error) {
	var user User
	sqlStr := "SELECT id, user_id, username, password_hash, salt, email, created_at, updated_at, last_login FROM users WHERE user_id = ?"
	err := DB.Get(&user, sqlStr, userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateLastLogin 更新用户最后登录时间
func UpdateLastLogin(userID string) error {
	sqlStr := "UPDATE users SET last_login = ? WHERE user_id = ?"
	_, err := DB.Exec(sqlStr, time.Now(), userID)
	return err
}
