package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/jobfirst/jobfirst-core/auth"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 配置数据库连接
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "root:@tcp(localhost:3306)/jobfirst?charset=utf8mb4&parseTime=True&loc=Local"
	}

	// 连接数据库
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	// JWT密钥
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "jobfirst-unified-auth-secret-key-2024"
	}

	// 创建统一认证系统
	authSystem := auth.NewUnifiedAuthSystem(db, jwtSecret)

	// 初始化数据库
	log.Println("正在初始化数据库...")
	if err := authSystem.InitializeDatabase(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	log.Println("数据库初始化完成")

	// 创建API服务器
	port := 8207
	if portEnv := os.Getenv("AUTH_SERVICE_PORT"); portEnv != "" {
		fmt.Sscanf(portEnv, "%d", &port)
	}

	api := auth.NewUnifiedAuthAPI(authSystem, port)

	// 启动服务器
	log.Printf("统一认证服务启动在端口 %d", port)
	log.Println("支持的API端点:")
	log.Println("  POST /api/v1/auth/login - 用户登录")
	log.Println("  POST /api/v1/auth/validate - JWT验证")
	log.Println("  GET  /api/v1/auth/permission - 权限检查")
	log.Println("  GET  /api/v1/auth/user - 获取用户信息")
	log.Println("  POST /api/v1/auth/access - 访问验证")
	log.Println("  POST /api/v1/auth/log - 访问日志")
	log.Println("  GET  /api/v1/auth/roles - 获取角色列表")
	log.Println("  GET  /api/v1/auth/permissions - 获取权限列表")
	log.Println("  GET  /health - 健康检查")

	if err := api.Start(); err != nil {
		log.Fatalf("认证服务启动失败: %v", err)
	}
}
