package infrastructure

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/consul/api"
)

// Infrastructure 基础设施管理器
type Infrastructure struct {
	Logger    Logger
	Config    ConfigManager
	Database  DatabaseManager
	Registry  ServiceRegistry
	Security  *SecurityManager
	Tracing   TracingService
	Messaging MessageQueue
}

// NewInfrastructure 创建基础设施管理器
func NewInfrastructure() *Infrastructure {
	return &Infrastructure{}
}

// Init 初始化基础设施
func (infra *Infrastructure) Init() error {
	// 1. 初始化日志系统
	if err := infra.initLogger(); err != nil {
		return fmt.Errorf("failed to initialize logger: %v", err)
	}

	// 2. 初始化配置管理
	if err := infra.initConfig(); err != nil {
		return fmt.Errorf("failed to initialize config: %v", err)
	}

	// 3. 初始化数据库连接
	if err := infra.initDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	// 4. 初始化服务注册
	if err := infra.initServiceRegistry(); err != nil {
		return fmt.Errorf("failed to initialize service registry: %v", err)
	}

	// 5. 初始化安全管理
	if err := infra.initSecurity(); err != nil {
		return fmt.Errorf("failed to initialize security: %v", err)
	}

	// 6. 初始化分布式追踪
	if err := infra.initTracing(); err != nil {
		return fmt.Errorf("failed to initialize tracing: %v", err)
	}

	// 7. 初始化消息队列（可选）
	if err := infra.initMessaging(); err != nil {
		infra.Logger.Warn("Failed to initialize messaging, continuing without message queue",
			Field{Key: "error", Value: err.Error()},
		)
		// 设置一个空的Noop消息队列
		infra.Messaging = &NoopMessageQueue{}
	}

	infra.Logger.Info("Infrastructure initialized successfully")
	return nil
}

// initLogger 初始化日志系统
func (infra *Infrastructure) initLogger() error {
	// 从环境变量获取日志配置
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	logOutput := os.Getenv("LOG_OUTPUT")
	if logOutput == "" {
		logOutput = "stdout"
	}

	config := &LoggerConfig{
		Level:  LogLevel(logLevel),
		Output: logOutput,
	}

	infra.Logger = NewLogger(config)
	InitGlobalLogger(config)

	infra.Logger.Info("Logger initialized", Field{Key: "level", Value: logLevel})
	return nil
}

// initConfig 初始化配置管理
func (infra *Infrastructure) initConfig() error {
	// 从环境变量获取配置文件路径
	configFile := os.Getenv("CONFIG_FILE")
	envPrefix := os.Getenv("CONFIG_ENV_PREFIX")
	if envPrefix == "" {
		envPrefix = "JOBFIRST_"
	}

	builder := NewConfigBuilder()

	if configFile != "" {
		builder.WithFile(configFile)
	}

	builder.WithEnvPrefix(envPrefix)

	// 设置默认配置
	defaults := map[string]interface{}{
		"app.name":        "JobFirst",
		"app.version":     "1.0.0",
		"app.environment": "development",
	}
	builder.WithDefaults(defaults)

	config, err := builder.Build()
	if err != nil {
		return err
	}

	infra.Config = config
	InitGlobalConfig(config)

	infra.Logger.Info("Config initialized", Field{Key: "config_file", Value: configFile})
	return nil
}

// initDatabase 初始化数据库连接
func (infra *Infrastructure) initDatabase() error {
	// 从配置获取数据库配置
	dbConfig := CreateDefaultDatabaseConfig()

	// 优先使用环境变量
	if host := os.Getenv("MYSQL_HOST"); host != "" {
		dbConfig.MySQL.Host = host
	}
	if port := os.Getenv("MYSQL_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			dbConfig.MySQL.Port = p
		}
	}
	if database := os.Getenv("MYSQL_DATABASE"); database != "" {
		dbConfig.MySQL.Database = database
	}
	if username := os.Getenv("MYSQL_USER"); username != "" {
		dbConfig.MySQL.Username = username
	}
	if password := os.Getenv("MYSQL_PASSWORD"); password != "" {
		dbConfig.MySQL.Password = password
	}

	if host := os.Getenv("POSTGRESQL_HOST"); host != "" {
		dbConfig.PostgreSQL.Host = host
	}
	if port := os.Getenv("POSTGRESQL_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			dbConfig.PostgreSQL.Port = p
		}
	}
	if database := os.Getenv("POSTGRESQL_DATABASE"); database != "" {
		dbConfig.PostgreSQL.Database = database
	}
	if username := os.Getenv("POSTGRESQL_USER"); username != "" {
		dbConfig.PostgreSQL.Username = username
	}
	if password := os.Getenv("POSTGRESQL_PASSWORD"); password != "" {
		dbConfig.PostgreSQL.Password = password
	}

	if host := os.Getenv("NEO4J_HOST"); host != "" {
		dbConfig.Neo4j.Host = host
	}
	if port := os.Getenv("NEO4J_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			dbConfig.Neo4j.Port = p
		}
	}
	if username := os.Getenv("NEO4J_USER"); username != "" {
		dbConfig.Neo4j.Username = username
	}
	if password := os.Getenv("NEO4J_PASSWORD"); password != "" {
		dbConfig.Neo4j.Password = password
	}

	if host := os.Getenv("REDIS_HOST"); host != "" {
		dbConfig.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			dbConfig.Redis.Port = p
		}
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		dbConfig.Redis.Password = password
	}
	if db := os.Getenv("REDIS_DB"); db != "" {
		if d, err := strconv.Atoi(db); err == nil {
			dbConfig.Redis.DB = d
		}
	}

	// 如果配置中有数据库配置，则使用配置中的值
	if infra.Config != nil {
		// MySQL配置
		if host := infra.Config.GetString("database.mysql.host"); host != "" {
			dbConfig.MySQL.Host = host
		}
		if port := infra.Config.GetInt("database.mysql.port"); port > 0 {
			dbConfig.MySQL.Port = port
		}
		if database := infra.Config.GetString("database.mysql.database"); database != "" {
			dbConfig.MySQL.Database = database
		}
		if username := infra.Config.GetString("database.mysql.username"); username != "" {
			dbConfig.MySQL.Username = username
		}
		if password := infra.Config.GetString("database.mysql.password"); password != "" {
			dbConfig.MySQL.Password = password
		}

		// PostgreSQL配置
		if host := infra.Config.GetString("database.postgresql.host"); host != "" {
			dbConfig.PostgreSQL.Host = host
		}
		if port := infra.Config.GetInt("database.postgresql.port"); port > 0 {
			dbConfig.PostgreSQL.Port = port
		}
		if database := infra.Config.GetString("database.postgresql.database"); database != "" {
			dbConfig.PostgreSQL.Database = database
		}
		if username := infra.Config.GetString("database.postgresql.username"); username != "" {
			dbConfig.PostgreSQL.Username = username
		}
		if password := infra.Config.GetString("database.postgresql.password"); password != "" {
			dbConfig.PostgreSQL.Password = password
		}

		// Neo4j配置
		if host := infra.Config.GetString("database.neo4j.host"); host != "" {
			dbConfig.Neo4j.Host = host
		}
		if port := infra.Config.GetInt("database.neo4j.port"); port > 0 {
			dbConfig.Neo4j.Port = port
		}
		if username := infra.Config.GetString("database.neo4j.username"); username != "" {
			dbConfig.Neo4j.Username = username
		}
		if password := infra.Config.GetString("database.neo4j.password"); password != "" {
			dbConfig.Neo4j.Password = password
		}

		// Redis配置
		if host := infra.Config.GetString("database.redis.host"); host != "" {
			dbConfig.Redis.Host = host
		}
		if port := infra.Config.GetInt("database.redis.port"); port > 0 {
			dbConfig.Redis.Port = port
		}
		if password := infra.Config.GetString("database.redis.password"); password != "" {
			dbConfig.Redis.Password = password
		}
	}

	// 创建数据库管理器
	manager := NewDatabaseManager(dbConfig)

	// 连接数据库
	if err := manager.Connect(); err != nil {
		return err
	}

	infra.Database = manager
	InitGlobalDatabaseManager(dbConfig)

	infra.Logger.Info("Database connections initialized")
	return nil
}

// initServiceRegistry 初始化服务注册
func (infra *Infrastructure) initServiceRegistry() error {
	// 从配置获取服务注册配置
	registryType := infra.Config.GetString("service.registry.type")
	if registryType == "" {
		registryType = "memory" // 默认使用内存注册器
	}

	switch registryType {
	case "consul":
		// 创建Consul配置
		consulConfig := &api.Config{
			Address: infra.Config.GetString("service.registry.consul.address"),
		}
		if consulConfig.Address == "" {
			consulConfig.Address = "localhost:8500"
		}

		registry, err := NewConsulRegistry(consulConfig)
		if err != nil {
			return err
		}
		infra.Registry = registry
		InitGlobalServiceRegistry(registry)

	case "memory":
		fallthrough
	default:
		// 使用内存注册器
		registry := NewInMemoryRegistry()
		infra.Registry = registry
		InitGlobalServiceRegistry(registry)
	}

	infra.Logger.Info("Service registry initialized",
		Field{Key: "type", Value: registryType},
	)
	return nil
}

// initSecurity 初始化安全管理
func (infra *Infrastructure) initSecurity() error {
	// 创建安全配置
	securityConfig := CreateDefaultSecurityConfig()

	// 从配置覆盖默认值
	if secret := infra.Config.GetString("security.jwt.secret"); secret != "" {
		securityConfig.JWTSecret = secret
	}
	if expiresIn := infra.Config.GetDuration("security.jwt.expires_in"); expiresIn > 0 {
		securityConfig.JWTExpiresIn = expiresIn
	}
	if refreshExpiresIn := infra.Config.GetDuration("security.jwt.refresh_expires_in"); refreshExpiresIn > 0 {
		securityConfig.JWTRefreshExpiresIn = refreshExpiresIn
	}

	// 创建缓存实例（用于限流）
	var cache Cache
	if infra.Database != nil {
		redisClient := infra.Database.GetRedisConnection()
		if redisClient != nil {
			cache = NewRedisCache(&RedisClientWrapper{client: redisClient})
		}
	}

	// 创建安全管理器
	securityManager := NewSecurityManager(securityConfig, cache)
	infra.Security = securityManager
	InitGlobalSecurityManager(securityConfig, cache)

	infra.Logger.Info("Security manager initialized")
	return nil
}

// initTracing 初始化分布式追踪
func (infra *Infrastructure) initTracing() error {
	// 创建追踪配置
	tracingConfig := CreateDefaultTracingConfig()

	// 从配置覆盖默认值
	if serviceName := infra.Config.GetString("tracing.service_name"); serviceName != "" {
		tracingConfig.ServiceName = serviceName
	}
	if serviceVersion := infra.Config.GetString("tracing.service_version"); serviceVersion != "" {
		tracingConfig.ServiceVersion = serviceVersion
	}
	if environment := infra.Config.GetString("tracing.environment"); environment != "" {
		tracingConfig.Environment = environment
	}
	if jaegerEndpoint := infra.Config.GetString("tracing.jaeger.endpoint"); jaegerEndpoint != "" {
		tracingConfig.JaegerEndpoint = jaegerEndpoint
	}
	if sampleRate := infra.Config.GetFloat("tracing.sample_rate"); sampleRate > 0 {
		tracingConfig.SampleRate = sampleRate
	}
	if enabled := infra.Config.GetBool("tracing.enabled"); !enabled {
		tracingConfig.Enabled = false
	}

	// 创建追踪服务
	var tracingService TracingService
	if tracingConfig.Enabled {
		tracingService = NewSimpleTracing(tracingConfig)
	} else {
		tracingService = &NoopTracing{}
	}

	infra.Tracing = tracingService
	InitGlobalTracingService(tracingService)

	infra.Logger.Info("Tracing service initialized",
		Field{Key: "service_name", Value: tracingConfig.ServiceName},
		Field{Key: "enabled", Value: tracingConfig.Enabled},
	)
	return nil
}

// initMessaging 初始化消息队列
func (infra *Infrastructure) initMessaging() error {
	// 创建消息队列配置
	messagingConfig := CreateDefaultMessagingConfig()

	// 从配置覆盖默认值
	if redisAddr := infra.Config.GetString("messaging.redis.addr"); redisAddr != "" {
		messagingConfig.RedisAddr = redisAddr
	}
	if redisPassword := infra.Config.GetString("messaging.redis.password"); redisPassword != "" {
		messagingConfig.RedisPassword = redisPassword
	}
	if redisDB := infra.Config.GetInt("messaging.redis.db"); redisDB >= 0 {
		messagingConfig.RedisDB = redisDB
	}
	if maxRetries := infra.Config.GetInt("messaging.max_retries"); maxRetries > 0 {
		messagingConfig.MaxRetries = maxRetries
	}
	if retryDelay := infra.Config.GetDuration("messaging.retry_delay"); retryDelay > 0 {
		messagingConfig.RetryDelay = retryDelay
	}
	if batchSize := infra.Config.GetInt("messaging.batch_size"); batchSize > 0 {
		messagingConfig.BatchSize = batchSize
	}
	if consumerGroup := infra.Config.GetString("messaging.consumer_group"); consumerGroup != "" {
		messagingConfig.ConsumerGroup = consumerGroup
	}

	// 创建消息队列
	messageQueue, err := NewRedisStreamsQueue(messagingConfig)
	if err != nil {
		return err
	}

	infra.Messaging = messageQueue
	InitGlobalMessageQueue(messageQueue)

	infra.Logger.Info("Message queue initialized",
		Field{Key: "redis_addr", Value: messagingConfig.RedisAddr},
		Field{Key: "consumer_group", Value: messagingConfig.ConsumerGroup},
	)
	return nil
}

// Close 关闭基础设施
func (infra *Infrastructure) Close() error {
	var errors []error

	// 关闭数据库连接
	if infra.Database != nil {
		if err := infra.Database.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close database: %v", err))
		}
	}

	// 关闭服务注册器
	if registry, ok := infra.Registry.(*ConsulRegistry); ok {
		if err := registry.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close service registry: %v", err))
		}
	}

	// 关闭追踪服务
	if infra.Tracing != nil {
		if err := infra.Tracing.Shutdown(context.Background()); err != nil {
			errors = append(errors, fmt.Errorf("failed to close tracing service: %v", err))
		}
	}

	// 关闭消息队列
	if infra.Messaging != nil {
		if err := infra.Messaging.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close message queue: %v", err))
		}
	}

	// 记录关闭信息
	if infra.Logger != nil {
		infra.Logger.Info("Infrastructure closed")
	}

	// 返回第一个错误
	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

// HealthCheck 健康检查
func (infra *Infrastructure) HealthCheck() map[string]interface{} {
	health := make(map[string]interface{})

	// 日志系统健康状态
	health["logger"] = infra.Logger != nil

	// 配置系统健康状态
	health["config"] = infra.Config != nil

	// 数据库健康状态
	if infra.Database != nil {
		health["database"] = infra.Database.HealthCheck()
	} else {
		health["database"] = map[string]bool{
			"mysql":      false,
			"postgresql": false,
			"neo4j":      false,
			"redis":      false,
		}
	}

	// 服务注册健康状态
	health["registry"] = infra.Registry != nil

	// 安全管理健康状态
	health["security"] = infra.Security != nil

	// 追踪服务健康状态
	health["tracing"] = infra.Tracing != nil

	// 消息队列健康状态
	health["messaging"] = infra.Messaging != nil

	return health
}

// 全局基础设施实例
var globalInfrastructure *Infrastructure

// InitGlobalInfrastructure 初始化全局基础设施
func InitGlobalInfrastructure() error {
	infra := NewInfrastructure()
	err := infra.Init()
	if err != nil {
		return err
	}
	globalInfrastructure = infra
	return nil
}

// GetInfrastructure 获取全局基础设施
func GetInfrastructure() *Infrastructure {
	return globalInfrastructure
}

// CloseGlobalInfrastructure 关闭全局基础设施
func CloseGlobalInfrastructure() error {
	if globalInfrastructure != nil {
		return globalInfrastructure.Close()
	}
	return nil
}
