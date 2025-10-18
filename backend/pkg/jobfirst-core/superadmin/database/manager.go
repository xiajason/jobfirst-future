package database

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"superadmin/errors"
)

// Manager 数据库管理器
type Manager struct {
	config *DatabaseConfig
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MySQL      MySQLConfig      `json:"mysql"`
	PostgreSQL PostgreSQLConfig `json:"postgresql"`
	Redis      RedisConfig      `json:"redis"`
	Neo4j      Neo4jConfig      `json:"neo4j"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// PostgreSQLConfig PostgreSQL配置
type PostgreSQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Database int    `json:"database"`
}

// Neo4jConfig Neo4j配置
type Neo4jConfig struct {
	URI      string `json:"uri"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// NewManager 创建数据库管理器
func NewManager(config *DatabaseConfig) *Manager {
	return &Manager{
		config: config,
	}
}

// DatabaseStatus 数据库状态
type DatabaseStatus struct {
	MySQL      *DatabaseInfo `json:"mysql"`
	PostgreSQL *DatabaseInfo `json:"postgresql"`
	Redis      *DatabaseInfo `json:"redis"`
	Neo4j      *DatabaseInfo `json:"neo4j"`
}

// DatabaseInfo 数据库信息
type DatabaseInfo struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Version     string    `json:"version"`
	Uptime      string    `json:"uptime"`
	Connections int       `json:"connections"`
	LastCheck   time.Time `json:"last_check"`
	Error       string    `json:"error,omitempty"`
}

// DatabaseInitStatus 数据库初始化状态
type DatabaseInitStatus struct {
	MySQL      *InitStatus `json:"mysql"`
	PostgreSQL *InitStatus `json:"postgresql"`
	Redis      *InitStatus `json:"redis"`
}

// InitStatus 初始化状态
type InitStatus struct {
	Initialized bool      `json:"initialized"`
	Tables      []string  `json:"tables"`
	DataCount   int       `json:"data_count"`
	LastInit    time.Time `json:"last_init"`
	Error       string    `json:"error,omitempty"`
}

// TableInfo 表信息
type TableInfo struct {
	Name      string `json:"name"`
	Rows      int    `json:"rows"`
	Size      string `json:"size"`
	Engine    string `json:"engine"`
	Collation string `json:"collation"`
}

// GetDatabaseStatus 获取数据库状态
func (m *Manager) GetDatabaseStatus() (*DatabaseStatus, error) {
	status := &DatabaseStatus{}

	// 检查MySQL状态
	mysqlStatus, err := m.checkMySQLStatus()
	if err != nil {
		status.MySQL = &DatabaseInfo{
			Name:      "mysql",
			Status:    "error",
			LastCheck: time.Now(),
			Error:     err.Error(),
		}
	} else {
		status.MySQL = mysqlStatus
	}

	// 检查PostgreSQL状态
	postgresStatus, err := m.checkPostgreSQLStatus()
	if err != nil {
		status.PostgreSQL = &DatabaseInfo{
			Name:      "postgresql",
			Status:    "error",
			LastCheck: time.Now(),
			Error:     err.Error(),
		}
	} else {
		status.PostgreSQL = postgresStatus
	}

	// 检查Redis状态
	redisStatus, err := m.checkRedisStatus()
	if err != nil {
		status.Redis = &DatabaseInfo{
			Name:      "redis",
			Status:    "error",
			LastCheck: time.Now(),
			Error:     err.Error(),
		}
	} else {
		status.Redis = redisStatus
	}

	// 检查Neo4j状态
	neo4jStatus, err := m.checkNeo4jStatus()
	if err != nil {
		status.Neo4j = &DatabaseInfo{
			Name:      "neo4j",
			Status:    "error",
			LastCheck: time.Now(),
			Error:     err.Error(),
		}
	} else {
		status.Neo4j = neo4jStatus
	}

	return status, nil
}

// checkMySQLStatus 检查MySQL状态
func (m *Manager) checkMySQLStatus() (*DatabaseInfo, error) {
	info := &DatabaseInfo{
		Name:      "mysql",
		LastCheck: time.Now(),
	}

	// 检查端口是否开放
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", m.config.MySQL.Host, m.config.MySQL.Port), 5*time.Second)
	if err != nil {
		info.Status = "down"
		info.Error = err.Error()
		return info, err
	}
	defer conn.Close()

	// 尝试连接数据库
	cmd := exec.Command("mysql", "-h", m.config.MySQL.Host, "-P", fmt.Sprintf("%d", m.config.MySQL.Port), "-u", m.config.MySQL.Username, "-p"+m.config.MySQL.Password, "-e", "SELECT VERSION();")
	output, err := cmd.Output()
	if err != nil {
		info.Status = "error"
		info.Error = err.Error()
		return info, err
	}

	// 解析版本信息
	lines := strings.Split(string(output), "\n")
	if len(lines) > 1 {
		info.Version = strings.TrimSpace(lines[1])
	}

	// 获取连接数
	cmd = exec.Command("mysql", "-h", m.config.MySQL.Host, "-P", fmt.Sprintf("%d", m.config.MySQL.Port), "-u", m.config.MySQL.Username, "-p"+m.config.MySQL.Password, "-e", "SHOW STATUS LIKE 'Threads_connected';")
	output, err = cmd.Output()
	if err == nil {
		_ = string(output)
		// 解析连接数
		info.Connections = 0 // 简化处理
	}

	info.Status = "healthy"
	return info, nil
}

// checkPostgreSQLStatus 检查PostgreSQL状态
func (m *Manager) checkPostgreSQLStatus() (*DatabaseInfo, error) {
	info := &DatabaseInfo{
		Name:      "postgresql",
		LastCheck: time.Now(),
	}

	// 检查端口是否开放
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", m.config.PostgreSQL.Host, m.config.PostgreSQL.Port), 5*time.Second)
	if err != nil {
		info.Status = "down"
		info.Error = err.Error()
		return info, err
	}
	defer conn.Close()

	// 尝试连接数据库
	cmd := exec.Command("psql", "-h", m.config.PostgreSQL.Host, "-p", fmt.Sprintf("%d", m.config.PostgreSQL.Port), "-U", m.config.PostgreSQL.Username, "-d", m.config.PostgreSQL.Database, "-c", "SELECT version();")
	output, err := cmd.Output()
	if err != nil {
		info.Status = "error"
		info.Error = err.Error()
		return info, err
	}

	// 解析版本信息
	lines := strings.Split(string(output), "\n")
	if len(lines) > 2 {
		info.Version = strings.TrimSpace(lines[2])
	}

	info.Status = "healthy"
	return info, nil
}

// checkRedisStatus 检查Redis状态
func (m *Manager) checkRedisStatus() (*DatabaseInfo, error) {
	info := &DatabaseInfo{
		Name:      "redis",
		LastCheck: time.Now(),
	}

	// 检查端口是否开放
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", m.config.Redis.Host, m.config.Redis.Port), 5*time.Second)
	if err != nil {
		info.Status = "down"
		info.Error = err.Error()
		return info, err
	}
	defer conn.Close()

	// 尝试连接Redis
	cmd := exec.Command("redis-cli", "-h", m.config.Redis.Host, "-p", fmt.Sprintf("%d", m.config.Redis.Port), "ping")
	output, err := cmd.Output()
	if err != nil {
		info.Status = "error"
		info.Error = err.Error()
		return info, err
	}

	if strings.Contains(string(output), "PONG") {
		info.Status = "healthy"
	} else {
		info.Status = "error"
		info.Error = "Redis响应异常"
	}

	return info, nil
}

// checkNeo4jStatus 检查Neo4j状态
func (m *Manager) checkNeo4jStatus() (*DatabaseInfo, error) {
	info := &DatabaseInfo{
		Name:      "neo4j",
		LastCheck: time.Now(),
	}

	// 检查端口是否开放
	conn, err := net.DialTimeout("tcp", "localhost:7687", 5*time.Second)
	if err != nil {
		info.Status = "down"
		info.Error = err.Error()
		return info, err
	}
	defer conn.Close()

	// 尝试连接Neo4j
	cmd := exec.Command("cypher-shell", "-u", m.config.Neo4j.Username, "-p", m.config.Neo4j.Password, "RETURN 1;")
	_, err = cmd.Output()
	if err != nil {
		info.Status = "error"
		info.Error = err.Error()
		return info, err
	}

	info.Status = "healthy"
	return info, nil
}

// getMySQLTables 获取MySQL表信息
func (m *Manager) getMySQLTables() ([]TableInfo, error) {
	tables := []TableInfo{}

	cmd := exec.Command("mysql", "-h", m.config.MySQL.Host, "-P", fmt.Sprintf("%d", m.config.MySQL.Port), "-u", m.config.MySQL.Username, "-p"+m.config.MySQL.Password, "-e", "SHOW TABLE STATUS;")
	output, err := cmd.Output()
	if err != nil {
		return tables, err
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 { // 跳过标题行
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 {
			table := TableInfo{
				Name:   fields[0],
				Engine: fields[1],
				// 其他字段需要根据实际输出格式解析
			}
			tables = append(tables, table)
		}
	}

	return tables, nil
}

// getMySQLDataCounts 获取MySQL数据统计
func (m *Manager) getMySQLDataCounts() (map[string]int, error) {
	counts := make(map[string]int)

	tables, err := m.getMySQLTables()
	if err != nil {
		return counts, err
	}

	for _, table := range tables {
		cmd := exec.Command("mysql", "-h", m.config.MySQL.Host, "-P", fmt.Sprintf("%d", m.config.MySQL.Port), "-u", m.config.MySQL.Username, "-p"+m.config.MySQL.Password, "-e", fmt.Sprintf("SELECT COUNT(*) FROM %s;", table.Name))
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) > 1 {
				// 解析行数
				counts[table.Name] = 0 // 简化处理
			}
		}
	}

	return counts, nil
}

// GetDatabaseInitStatus 获取数据库初始化状态
func (m *Manager) GetDatabaseInitStatus() (*DatabaseInitStatus, error) {
	status := &DatabaseInitStatus{}

	// 检查MySQL初始化状态
	mysqlInit, err := m.checkMySQLInitStatus()
	if err != nil {
		status.MySQL = &InitStatus{
			Initialized: false,
			Error:       err.Error(),
		}
	} else {
		status.MySQL = mysqlInit
	}

	// 检查PostgreSQL初始化状态
	postgresInit, err := m.checkPostgreSQLInitStatus()
	if err != nil {
		status.PostgreSQL = &InitStatus{
			Initialized: false,
			Error:       err.Error(),
		}
	} else {
		status.PostgreSQL = postgresInit
	}

	// 检查Redis初始化状态
	redisInit, err := m.checkRedisInitStatus()
	if err != nil {
		status.Redis = &InitStatus{
			Initialized: false,
			Error:       err.Error(),
		}
	} else {
		status.Redis = redisInit
	}

	return status, nil
}

// checkMySQLInitStatus 检查MySQL初始化状态
func (m *Manager) checkMySQLInitStatus() (*InitStatus, error) {
	status := &InitStatus{}

	// 检查数据库是否存在
	cmd := exec.Command("mysql", "-h", m.config.MySQL.Host, "-P", fmt.Sprintf("%d", m.config.MySQL.Port), "-u", m.config.MySQL.Username, "-p"+m.config.MySQL.Password, "-e", "SHOW DATABASES;")
	output, err := cmd.Output()
	if err != nil {
		status.Error = err.Error()
		return status, err
	}

	// 检查是否包含项目数据库
	if strings.Contains(string(output), m.config.MySQL.Database) {
		status.Initialized = true

		// 获取表信息
		tables, err := m.getMySQLTables()
		if err == nil {
			status.Tables = make([]string, len(tables))
			for i, table := range tables {
				status.Tables[i] = table.Name
			}
		}
	} else {
		status.Initialized = false
	}

	return status, nil
}

// checkPostgreSQLInitStatus 检查PostgreSQL初始化状态
func (m *Manager) checkPostgreSQLInitStatus() (*InitStatus, error) {
	status := &InitStatus{}

	// 检查数据库是否存在
	cmd := exec.Command("psql", "-h", m.config.PostgreSQL.Host, "-p", fmt.Sprintf("%d", m.config.PostgreSQL.Port), "-U", m.config.PostgreSQL.Username, "-l")
	output, err := cmd.Output()
	if err != nil {
		status.Error = err.Error()
		return status, err
	}

	// 检查是否包含项目数据库
	if strings.Contains(string(output), m.config.PostgreSQL.Database) {
		status.Initialized = true
	} else {
		status.Initialized = false
	}

	return status, nil
}

// checkRedisInitStatus 检查Redis初始化状态
func (m *Manager) checkRedisInitStatus() (*InitStatus, error) {
	status := &InitStatus{}

	// 检查Redis连接
	cmd := exec.Command("redis-cli", "-h", m.config.Redis.Host, "-p", fmt.Sprintf("%d", m.config.Redis.Port), "ping")
	output, err := cmd.Output()
	if err != nil {
		status.Error = err.Error()
		return status, err
	}

	if strings.Contains(string(output), "PONG") {
		status.Initialized = true
	} else {
		status.Initialized = false
	}

	return status, nil
}

// InitializeDatabase 初始化数据库
func (m *Manager) InitializeDatabase(dbType string) error {
	switch dbType {
	case "mysql":
		return m.initializeMySQL()
	case "postgresql":
		return m.initializePostgreSQL()
	case "redis":
		return m.initializeRedis()
	default:
		return errors.NewError(errors.ErrCodeValidation, "不支持的数据库类型")
	}
}

// initializeMySQL 初始化MySQL
func (m *Manager) initializeMySQL() error {
	// 创建数据库
	cmd := exec.Command("mysql", "-h", m.config.MySQL.Host, "-P", fmt.Sprintf("%d", m.config.MySQL.Port), "-u", m.config.MySQL.Username, "-p"+m.config.MySQL.Password, "-e", fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", m.config.MySQL.Database))
	if err := cmd.Run(); err != nil {
		return errors.WrapError(errors.ErrCodeDatabase, "创建MySQL数据库失败", err)
	}

	// 创建基础表
	createTablesSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		role VARCHAR(20) DEFAULT 'user',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS resumes (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		title VARCHAR(200) NOT NULL,
		content TEXT,
		status VARCHAR(20) DEFAULT 'draft',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS companies (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(200) NOT NULL,
		description TEXT,
		website VARCHAR(200),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS jobs (
		id INT AUTO_INCREMENT PRIMARY KEY,
		company_id INT NOT NULL,
		title VARCHAR(200) NOT NULL,
		description TEXT,
		requirements TEXT,
		location VARCHAR(100),
		salary_range VARCHAR(50),
		status VARCHAR(20) DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (company_id) REFERENCES companies(id)
	);
	`

	cmd = exec.Command("mysql", "-h", m.config.MySQL.Host, "-P", fmt.Sprintf("%d", m.config.MySQL.Port), "-u", m.config.MySQL.Username, "-p"+m.config.MySQL.Password, m.config.MySQL.Database, "-e", createTablesSQL)
	if err := cmd.Run(); err != nil {
		return errors.WrapError(errors.ErrCodeDatabase, "创建MySQL表失败", err)
	}

	return nil
}

// initializePostgreSQL 初始化PostgreSQL
func (m *Manager) initializePostgreSQL() error {
	// 创建数据库
	cmd := exec.Command("createdb", "-h", m.config.PostgreSQL.Host, "-p", fmt.Sprintf("%d", m.config.PostgreSQL.Port), "-U", m.config.PostgreSQL.Username, m.config.PostgreSQL.Database)
	if err := cmd.Run(); err != nil {
		// 数据库可能已存在，继续执行
	}

	// 创建向量扩展
	cmd = exec.Command("psql", "-h", m.config.PostgreSQL.Host, "-p", fmt.Sprintf("%d", m.config.PostgreSQL.Port), "-U", m.config.PostgreSQL.Username, "-d", m.config.PostgreSQL.Database, "-c", "CREATE EXTENSION IF NOT EXISTS vector;")
	if err := cmd.Run(); err != nil {
		return errors.WrapError(errors.ErrCodeDatabase, "创建PostgreSQL向量扩展失败", err)
	}

	// 创建基础表
	createTablesSQL := `
	CREATE TABLE IF NOT EXISTS vector_documents (
		id SERIAL PRIMARY KEY,
		content TEXT NOT NULL,
		embedding VECTOR(1536),
		metadata JSONB,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS vector_documents_embedding_idx 
	ON vector_documents USING ivfflat (embedding vector_cosine_ops);
	`

	cmd = exec.Command("psql", "-h", m.config.PostgreSQL.Host, "-p", fmt.Sprintf("%d", m.config.PostgreSQL.Port), "-U", m.config.PostgreSQL.Username, "-d", m.config.PostgreSQL.Database, "-c", createTablesSQL)
	if err := cmd.Run(); err != nil {
		return errors.WrapError(errors.ErrCodeDatabase, "创建PostgreSQL表失败", err)
	}

	return nil
}

// initializeRedis 初始化Redis
func (m *Manager) initializeRedis() error {
	// Redis不需要特殊的初始化，只需要确保连接正常
	cmd := exec.Command("redis-cli", "-h", m.config.Redis.Host, "-p", fmt.Sprintf("%d", m.config.Redis.Port), "ping")
	output, err := cmd.Output()
	if err != nil {
		return errors.WrapError(errors.ErrCodeDatabase, "Redis连接失败", err)
	}

	if !strings.Contains(string(output), "PONG") {
		return errors.NewError(errors.ErrCodeDatabase, "Redis响应异常")
	}

	return nil
}
