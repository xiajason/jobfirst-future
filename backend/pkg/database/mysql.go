package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xiajason/zervi-basic/basic/backend/pkg/config"
)

func InitMySQL(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.Charset,
		cfg.ParseTime,
		cfg.Loc,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	// 解析连接最大生命周期
	if lifetime, err := time.ParseDuration(cfg.ConnMaxLifetime); err == nil {
		db.SetConnMaxLifetime(lifetime)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	log.Printf("Connected to MySQL database: %s:%s/%s", cfg.Host, cfg.Port, cfg.Name)
	return db, nil
}

func CloseMySQL(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("Error closing MySQL connection: %v", err)
		} else {
			log.Println("MySQL connection closed")
		}
	}
}
