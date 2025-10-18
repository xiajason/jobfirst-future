package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jobfirst/jobfirst-core/config"
	"github.com/jobfirst/jobfirst-core/logging"
)

func main() {
	fmt.Println("=== JobFirst Core Config & Logging Integration Test ===")

	// 1. 测试配置管理
	fmt.Println("\n1. Testing Configuration Management...")

	// 创建配置管理器
	configManager, err := config.NewSimpleManager("./test-configs")
	if err != nil {
		log.Fatalf("Failed to create config manager: %v", err)
	}
	defer configManager.Close()

	// 测试获取配置
	value, err := configManager.Get("database.mysql.host")
	if err != nil {
		fmt.Printf("   ❌ Failed to get config: %v\n", err)
	} else {
		fmt.Printf("   ✅ Got config value: %v\n", value)
	}

	// 测试设置配置
	err = configManager.Set("database.mysql.host", "192.168.1.100")
	if err != nil {
		fmt.Printf("   ❌ Failed to set config: %v\n", err)
	} else {
		fmt.Printf("   ✅ Set config value successfully\n")
	}

	// 测试按类型获取配置
	dbConfigs, err := configManager.GetByType(config.ConfigTypeDatabase)
	if err != nil {
		fmt.Printf("   ❌ Failed to get database configs: %v\n", err)
	} else {
		fmt.Printf("   ✅ Got %d database configs\n", len(dbConfigs))
	}

	// 测试创建版本
	version, err := configManager.CreateVersion("Test configuration update", "test-user")
	if err != nil {
		fmt.Printf("   ❌ Failed to create version: %v\n", err)
	} else {
		fmt.Printf("   ✅ Created version: %s\n", version.Version)
	}

	// 2. 测试日志系统
	fmt.Println("\n2. Testing Logging System...")

	// 创建日志配置
	logConfig := &logging.LoggerConfig{
		Level:      logging.LogLevelInfo,
		Format:     logging.LogFormatText,
		Service:    "config-test",
		Module:     "integration",
		Output:     []string{"stdout"},
		Caller:     true,
		Stacktrace: false,
	}

	// 创建日志记录器
	logger, err := logging.NewStandardLogger(logConfig)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// 测试不同级别的日志
	logger.Debug("This is a debug message", map[string]interface{}{
		"component": "config-manager",
		"action":    "test",
	})

	logger.Info("Configuration management test started", map[string]interface{}{
		"test_type": "integration",
		"timestamp": time.Now().Unix(),
	})

	logger.Warn("This is a warning message", map[string]interface{}{
		"warning_type": "test_warning",
	})

	logger.Error("This is an error message", fmt.Errorf("test error"), map[string]interface{}{
		"error_code": "TEST_ERROR",
		"component":  "config-manager",
	})

	// 测试带字段的日志记录器
	serviceLogger := logger.WithService("database-service").WithModule("connection")
	serviceLogger.Info("Database connection established", map[string]interface{}{
		"host":     "localhost",
		"port":     3306,
		"database": "jobfirst",
	})

	// 测试跟踪ID
	traceLogger := logger.WithTraceID("trace-12345").WithUserID("user-67890")
	traceLogger.Info("User action performed", map[string]interface{}{
		"action": "config_update",
		"key":    "database.mysql.host",
		"value":  "192.168.1.100",
	})

	// 3. 测试配置变更监听
	fmt.Println("\n3. Testing Configuration Change Monitoring...")

	// 添加配置变更监听器
	err = configManager.WatchChanges(func(change *config.ConfigChange) {
		logger.Info("Configuration changed", map[string]interface{}{
			"change_id":   change.ID,
			"change_type": change.Type,
			"description": change.Description,
			"timestamp":   change.Timestamp,
		})
	})
	if err != nil {
		fmt.Printf("   ❌ Failed to add change watcher: %v\n", err)
	} else {
		fmt.Printf("   ✅ Added configuration change watcher\n")
	}

	// 触发配置变更
	err = configManager.Set("system.log_level", "debug")
	if err != nil {
		fmt.Printf("   ❌ Failed to trigger config change: %v\n", err)
	} else {
		fmt.Printf("   ✅ Triggered configuration change\n")
	}

	// 4. 测试配置验证
	fmt.Println("\n4. Testing Configuration Validation...")

	// 创建验证器
	portValidator := config.NewPortValidator()
	hostValidator := config.NewHostValidator()
	logLevelValidator := config.NewLogLevelValidator()

	// 测试端口验证
	err = portValidator.Validate(8080)
	if err != nil {
		fmt.Printf("   ❌ Port validation failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Port validation passed\n")
	}

	// 测试主机验证
	err = hostValidator.Validate("localhost")
	if err != nil {
		fmt.Printf("   ❌ Host validation failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Host validation passed\n")
	}

	// 测试日志级别验证
	err = logLevelValidator.Validate("info")
	if err != nil {
		fmt.Printf("   ❌ Log level validation failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Log level validation passed\n")
	}

	// 5. 测试日志指标
	fmt.Println("\n5. Testing Logging Metrics...")

	// 获取日志指标
	metrics := logger.GetMetrics()
	fmt.Printf("   Total logs: %d\n", metrics.TotalLogs)
	fmt.Printf("   Error rate: %.2f%%\n", metrics.ErrorRate)
	fmt.Printf("   Last log time: %s\n", metrics.LastLogTime.Format("2006-01-02 15:04:05"))

	// 显示按级别的日志统计
	fmt.Printf("   Logs by level:\n")
	for level, count := range metrics.LogsByLevel {
		fmt.Printf("     %s: %d\n", level, count)
	}

	// 显示按服务的日志统计
	fmt.Printf("   Logs by service:\n")
	for service, count := range metrics.LogsByService {
		fmt.Printf("     %s: %d\n", service, count)
	}

	// 6. 测试配置热更新
	fmt.Println("\n6. Testing Configuration Hot Reload...")

	// 创建热更新器
	hotReloader, err := config.NewHotReloader(configManager, "./test-configs/current.json")
	if err != nil {
		fmt.Printf("   ❌ Failed to create hot reloader: %v\n", err)
	} else {
		fmt.Printf("   ✅ Created hot reloader\n")

		// 添加热更新回调
		hotReloader.AddCallback(func(change *config.ConfigChange) {
			logger.Info("Hot reload detected", map[string]interface{}{
				"change_id":   change.ID,
				"description": change.Description,
				"type":        change.Type,
			})
		})

		// 启动热更新
		err = hotReloader.Start()
		if err != nil {
			fmt.Printf("   ❌ Failed to start hot reloader: %v\n", err)
		} else {
			fmt.Printf("   ✅ Started hot reloader\n")
			fmt.Printf("   ✅ Hot reloader is enabled: %t\n", hotReloader.IsEnabled())
		}

		// 停止热更新
		err = hotReloader.Stop()
		if err != nil {
			fmt.Printf("   ❌ Failed to stop hot reloader: %v\n", err)
		} else {
			fmt.Printf("   ✅ Stopped hot reloader\n")
		}
	}

	// 7. 最终测试结果
	fmt.Println("\n7. Final Test Results...")

	// 获取所有配置
	allConfigs, err := configManager.GetAll()
	if err != nil {
		fmt.Printf("   ❌ Failed to get all configs: %v\n", err)
	} else {
		fmt.Printf("   ✅ Total configurations: %d\n", len(allConfigs))
	}

	// 最终日志
	logger.Info("Integration test completed successfully", map[string]interface{}{
		"test_duration": "completed",
		"config_count":  len(allConfigs),
		"log_count":     metrics.TotalLogs,
	})

	fmt.Println("\n=== Config & Logging Integration Test Completed ===")
}
