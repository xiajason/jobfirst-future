package config

import (
	"os"
	"strconv"
)

type Config struct {
	Environment string
	Version     string
	Mode        string
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	Upload      UploadConfig
	Logging     LoggingConfig
	Cache       CacheConfig
	Security    SecurityConfig
	Points      PointsConfig
	AI          AIConfig
	Monitoring  MonitoringConfig
	Consul      ConsulConfig
}

type ServerConfig struct {
	Port           string
	Host           string
	ReadTimeout    string
	WriteTimeout   string
	MaxHeaderBytes int
}

type DatabaseConfig struct {
	Driver          string
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	Charset         string
	ParseTime       bool
	Loc             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime string
}

type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  string
	ReadTimeout  string
	WriteTimeout string
}

type JWTConfig struct {
	Secret           string
	ExpiresIn        string
	RefreshExpiresIn string
	Issuer           string
}

type UploadConfig struct {
	MaxSize      int
	AllowedTypes []string
	UploadDir    string
	TempDir      string
}

type LoggingConfig struct {
	Level      string
	Format     string
	Output     string
	File       string
	MaxSize    int
	MaxAge     int
	MaxBackups int
}

type CacheConfig struct {
	DefaultTTL      string
	UserSessionTTL  string
	ResumeDataTTL   string
	TemplateDataTTL string
}

type SecurityConfig struct {
	BcryptCost      int
	RateLimit       int
	RateLimitWindow string
	CORSOrigins     []string
	CORSMethods     []string
	CORSHeaders     []string
}

type PointsConfig struct {
	DefaultBalance     int
	ResumeCreateReward int
	ResumeShareReward  int
	TemplateUseCost    int
	FileUploadCost     int
}

type AIConfig struct {
	Enabled    bool
	ServiceURL string
	APIKey     string
	Timeout    string
	MaxRetries int
}

type MonitoringConfig struct {
	Enabled             bool
	MetricsPort         string
	HealthCheckInterval string
	PrometheusEnabled   bool
}

type ConsulConfig struct {
	Enabled             bool
	Host                string
	Port                string
	Scheme              string
	Datacenter          string
	Token               string
	ServiceName         string
	ServiceID           string
	ServiceTags         []string
	HealthCheckURL      string
	HealthCheckInterval string
	HealthCheckTimeout  string
	DeregisterAfter     string
}

func Load() *Config {
	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Version:     getEnv("VERSION", "1.0.0"),
		Mode:        getEnv("MODE", "basic"),
		Server: ServerConfig{
			Port:           getEnv("SERVER_PORT", "8601"),
			Host:           getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:    getEnv("SERVER_READ_TIMEOUT", "30s"),
			WriteTimeout:   getEnv("SERVER_WRITE_TIMEOUT", "30s"),
			MaxHeaderBytes: getEnvAsInt("SERVER_MAX_HEADER_BYTES", 1048576),
		},
		Database: DatabaseConfig{
			Driver:          getEnv("DB_DRIVER", "mysql"),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "3306"),
			Name:            getEnv("DB_NAME", "jobfirst"),
			User:            getEnv("DB_USER", "root"),
			Password:        getEnv("DB_PASSWORD", ""),
			Charset:         getEnv("DB_CHARSET", "utf8mb4"),
			ParseTime:       getEnvAsBool("DB_PARSE_TIME", true),
			Loc:             getEnv("DB_LOC", "Local"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnv("DB_CONN_MAX_LIFETIME", "3600s"),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvAsInt("REDIS_DB", 0),
			PoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
			MaxRetries:   getEnvAsInt("REDIS_MAX_RETRIES", 3),
			DialTimeout:  getEnv("REDIS_DIAL_TIMEOUT", "5s"),
			ReadTimeout:  getEnv("REDIS_READ_TIMEOUT", "3s"),
			WriteTimeout: getEnv("REDIS_WRITE_TIMEOUT", "3s"),
		},
		JWT: JWTConfig{
			Secret:           getEnv("JWT_SECRET", "jobfirst-basic-secret-key-2024"),
			ExpiresIn:        getEnv("JWT_EXPIRES_IN", "24h"),
			RefreshExpiresIn: getEnv("JWT_REFRESH_EXPIRES_IN", "168h"),
			Issuer:           getEnv("JWT_ISSUER", "jobfirst-basic"),
		},
		Upload: UploadConfig{
			MaxSize:      getEnvAsInt("UPLOAD_MAX_SIZE", 10485760),
			AllowedTypes: []string{"pdf", "doc", "docx", "jpg", "jpeg", "png"},
			UploadDir:    getEnv("UPLOAD_DIR", "./uploads"),
			TempDir:      getEnv("TEMP_DIR", "./temp"),
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOGGING_LEVEL", "info"),
			Format:     getEnv("LOGGING_FORMAT", "json"),
			Output:     getEnv("LOGGING_OUTPUT", "stdout"),
			File:       getEnv("LOGGING_FILE", "./logs/basic-server.log"),
			MaxSize:    getEnvAsInt("LOGGING_MAX_SIZE", 100),
			MaxAge:     getEnvAsInt("LOGGING_MAX_AGE", 30),
			MaxBackups: getEnvAsInt("LOGGING_MAX_BACKUPS", 10),
		},
		Cache: CacheConfig{
			DefaultTTL:      getEnv("CACHE_DEFAULT_TTL", "3600s"),
			UserSessionTTL:  getEnv("CACHE_USER_SESSION_TTL", "7200s"),
			ResumeDataTTL:   getEnv("CACHE_RESUME_DATA_TTL", "1800s"),
			TemplateDataTTL: getEnv("CACHE_TEMPLATE_DATA_TTL", "7200s"),
		},
		Security: SecurityConfig{
			BcryptCost:      getEnvAsInt("SECURITY_BCRYPT_COST", 12),
			RateLimit:       getEnvAsInt("SECURITY_RATE_LIMIT", 100),
			RateLimitWindow: getEnv("SECURITY_RATE_LIMIT_WINDOW", "1m"),
			CORSOrigins:     []string{"*"},
			CORSMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			CORSHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		},
		Points: PointsConfig{
			DefaultBalance:     getEnvAsInt("POINTS_DEFAULT_BALANCE", 100),
			ResumeCreateReward: getEnvAsInt("POINTS_RESUME_CREATE_REWARD", 10),
			ResumeShareReward:  getEnvAsInt("POINTS_RESUME_SHARE_REWARD", 5),
			TemplateUseCost:    getEnvAsInt("POINTS_TEMPLATE_USE_COST", 2),
			FileUploadCost:     getEnvAsInt("POINTS_FILE_UPLOAD_COST", 1),
		},
		AI: AIConfig{
			Enabled:    getEnvAsBool("AI_ENABLED", true),
			ServiceURL: getEnv("AI_SERVICE_URL", "http://localhost:8206"),
			APIKey:     getEnv("AI_API_KEY", ""),
			Timeout:    getEnv("AI_TIMEOUT", "30s"),
			MaxRetries: getEnvAsInt("AI_MAX_RETRIES", 3),
		},
		Monitoring: MonitoringConfig{
			Enabled:             getEnvAsBool("MONITORING_ENABLED", true),
			MetricsPort:         getEnv("MONITORING_METRICS_PORT", "9090"),
			HealthCheckInterval: getEnv("MONITORING_HEALTH_CHECK_INTERVAL", "30s"),
			PrometheusEnabled:   getEnvAsBool("MONITORING_PROMETHEUS_ENABLED", true),
		},
		Consul: ConsulConfig{
			Enabled:             getEnvAsBool("CONSUL_ENABLED", false),
			Host:                getEnv("CONSUL_HOST", "localhost"),
			Port:                getEnv("CONSUL_PORT", "8500"),
			Scheme:              getEnv("CONSUL_SCHEME", "http"),
			Datacenter:          getEnv("CONSUL_DATACENTER", "dc1"),
			Token:               getEnv("CONSUL_TOKEN", ""),
			ServiceName:         getEnv("CONSUL_SERVICE_NAME", "basic-server"),
			ServiceID:           getEnv("CONSUL_SERVICE_ID", "basic-server-1"),
			ServiceTags:         []string{"api", "gateway", "basic"},
			HealthCheckURL:      getEnv("CONSUL_HEALTH_CHECK_URL", "/health"),
			HealthCheckInterval: getEnv("CONSUL_HEALTH_CHECK_INTERVAL", "10s"),
			HealthCheckTimeout:  getEnv("CONSUL_HEALTH_CHECK_TIMEOUT", "5s"),
			DeregisterAfter:     getEnv("CONSUL_DEREGISTER_AFTER", "30s"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
