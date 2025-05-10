package mysql

import (
	"fmt"

	"github.com/Xaytick/chat-zinx/chat-server/conf"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

// User 用户数据结构
type User struct {
	ID       int64  `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
	CreateAt string `db:"create_at"`
	UpdateAt string `db:"update_at"`
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
		username VARCHAR(50) NOT NULL UNIQUE,
		password VARCHAR(50) NOT NULL,
		create_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		update_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_username (username)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`
	_, err := DB.Exec(createUserTable)
	return err
}

// InsertUser 插入新用户
func InsertUser(user *User) (int64, error) {
	sqlStr := "INSERT INTO users(username, password, nickname) VALUES (?, ?, ?)"
	result, err := DB.Exec(sqlStr, user.Username, user.Password)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(username string) (*User, error) {
	var user User
	sqlStr := "SELECT id, username, password, create_at, update_at FROM users WHERE username = ?"
	err := DB.Get(&user, sqlStr, username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 根据ID获取用户
func GetUserByID(id int64) (*User, error) {
	var user User
	sqlStr := "SELECT id, username, password, create_at, update_at FROM users WHERE id = ?"
	err := DB.Get(&user, sqlStr, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
