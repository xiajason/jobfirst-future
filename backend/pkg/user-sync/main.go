package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	usersync "github.com/zervigo/backend/pkg/user-sync"
)

// 数据库配置
type DatabaseConfig struct {
	MySQL      MySQLConfig      `json:"mysql"`
	PostgreSQL PostgreSQLConfig `json:"postgresql"`
	Redis      RedisConfig      `json:"redis"`
}

type MySQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type PostgreSQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

func main() {
	// 命令行参数
	var configFile = flag.String("config", "", "配置文件路径")
	flag.Parse()

	// 默认配置
	config := &usersync.SyncConfig{
		QueueSize:     1000,
		Workers:       5,
		Timeout:       30 * time.Second,
		RetryInterval: 5 * time.Second,
		MaxRetries:    3,
	}

	// 数据库配置
	dbConfig := &DatabaseConfig{
		MySQL: MySQLConfig{
			Host:     "localhost",
			Port:     3306,
			Database: "jobfirst",
			Username: "root",
			Password: "",
		},
		PostgreSQL: PostgreSQLConfig{
			Host:     "localhost",
			Port:     5434,
			Database: "looma_independent",
			Username: "looma_user",
			Password: "looma_password",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
	}

	// 连接MySQL数据库
	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConfig.MySQL.Username, dbConfig.MySQL.Password, dbConfig.MySQL.Host, dbConfig.MySQL.Port, dbConfig.MySQL.Database)

	mysqlDB, err := gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接MySQL数据库失败: %v", err)
	}

	// 连接PostgreSQL数据库
	postgresDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.PostgreSQL.Host, dbConfig.PostgreSQL.Port, dbConfig.PostgreSQL.Username, dbConfig.PostgreSQL.Password, dbConfig.PostgreSQL.Database)

	postgresDB, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接PostgreSQL数据库失败: %v", err)
	}

	// 连接Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", dbConfig.Redis.Host, dbConfig.Redis.Port),
		Password: dbConfig.Redis.Password,
		DB:       dbConfig.Redis.DB,
	})

	// 测试Redis连接
	ctx := context.Background()
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("连接Redis失败: %v", err)
	}

	// 创建用户同步服务
	syncService := usersync.NewUserSyncService(config, redisClient)

	// 启动同步服务
	log.Println("启动用户数据同步服务...")
	syncService.Start()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动HTTP服务器用于健康检查
	go func() {
		// 简单的HTTP服务器
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		log.Println("健康检查服务器启动在 :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	log.Println("用户数据同步服务已启动，按 Ctrl+C 停止...")

	// 等待信号
	<-sigChan
	log.Println("收到停止信号，正在关闭服务...")

	// 停止同步服务
	syncService.Stop()
	log.Println("用户数据同步服务已停止")
}
