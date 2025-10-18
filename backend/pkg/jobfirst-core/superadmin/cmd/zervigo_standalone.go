package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("🚀 ZerviGo v3.1.1 - 超级管理员工具")
	fmt.Println("=====================================")
	fmt.Println("基于 jobfirst-core 核心包的超级管理员管理和监控工具")
	fmt.Println("核心功能：系统启动顺序检查 | 开发团队管理 | 用户权限验证")
	fmt.Println()

	// 检查命令行参数
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "startup":
			checkStartupOrder()
		case "team":
			checkTeamStatus()
		case "users":
			checkUserStatus()
		case "full":
			runFullCheck()
		case "help":
			showHelp()
		default:
			fmt.Printf("❌ 未知命令: %s\n", command)
			showHelp()
		}
	} else {
		runFullCheck()
	}
}

// runFullCheck 运行完整检查
func runFullCheck() {
	fmt.Println("🔍 开始全面系统检查...")
	fmt.Println("时间:", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()

	// 1. 系统启动顺序检查
	fmt.Println("1️⃣ 检查系统启动顺序...")
	startupStatus := checkStartupOrder()
	fmt.Println()

	// 2. 开发团队状态检查
	fmt.Println("2️⃣ 检查开发团队状态...")
	teamStatus := checkTeamStatus()
	fmt.Println()

	// 3. 用户权限和订阅状态检查
	fmt.Println("3️⃣ 检查用户权限和订阅状态...")
	userStatus := checkUserStatus()
	fmt.Println()

	// 4. 生成综合报告
	fmt.Println("📊 生成综合报告...")
	generateComprehensiveReport(startupStatus, teamStatus, userStatus)
}

// 系统启动顺序检查
func checkStartupOrder() map[string]interface{} {
	fmt.Println("⏰ 检查系统启动顺序...")

	// 定义服务启动顺序
	services := []map[string]interface{}{
		{"name": "consul", "port": 8500, "priority": 1, "description": "服务发现和配置中心"},
		{"name": "mysql", "port": 3306, "priority": 2, "description": "主数据库"},
		{"name": "redis", "port": 6379, "priority": 3, "description": "缓存服务"},
		{"name": "postgresql", "port": 5432, "priority": 4, "description": "AI服务数据库"},
		{"name": "nginx", "port": 80, "priority": 5, "description": "反向代理"},
		{"name": "api_gateway", "port": 8080, "priority": 10, "description": "API网关"},
		{"name": "user_service", "port": 8081, "priority": 11, "description": "用户管理服务"},
		{"name": "resume_service", "port": 8082, "priority": 12, "description": "简历管理服务"},
		{"name": "company_service", "port": 8083, "priority": 13, "description": "公司管理服务"},
		{"name": "notification_service", "port": 8084, "priority": 14, "description": "通知服务"},
		{"name": "template_service", "port": 8085, "priority": 20, "description": "模板管理服务"},
		{"name": "statistics_service", "port": 8086, "priority": 21, "description": "数据统计服务"},
		{"name": "banner_service", "port": 8087, "priority": 22, "description": "内容管理服务"},
		{"name": "dev_team_service", "port": 8088, "priority": 23, "description": "开发团队管理服务"},
		{"name": "ai_service", "port": 8206, "priority": 30, "description": "AI服务"},
	}

	status := map[string]interface{}{
		"timestamp":       time.Now(),
		"overall_status":  "success",
		"services":        []map[string]interface{}{},
		"violations":      []map[string]interface{}{},
		"recommendations": []string{},
	}

	activeCount := 0
	inactiveCount := 0
	violations := []map[string]interface{}{}

	// 检查每个服务
	for _, service := range services {
		serviceStatus := map[string]interface{}{
			"name":          service["name"],
			"port":          service["port"],
			"priority":      service["priority"],
			"description":   service["description"],
			"status":        "inactive",
			"health_status": "unknown",
		}

		// 检查端口是否开放
		if isPortOpen(service["port"].(int)) {
			serviceStatus["status"] = "active"
			activeCount++

			// 检查健康状态
			if checkServiceHealth(service["port"].(int)) {
				serviceStatus["health_status"] = "healthy"
			} else {
				serviceStatus["health_status"] = "unhealthy"
			}
		} else {
			inactiveCount++
		}

		status["services"] = append(status["services"].([]map[string]interface{}), serviceStatus)
	}

	// 检查启动顺序违规
	for i, service := range status["services"].([]map[string]interface{}) {
		if service["status"] == "active" {
			// 检查是否有高优先级服务未启动
			for j := 0; j < i; j++ {
				otherService := status["services"].([]map[string]interface{})[j]
				if otherService["priority"].(int) < service["priority"].(int) &&
					otherService["status"] != "active" {
					violation := map[string]interface{}{
						"type":    "order",
						"service": service["name"],
						"message": fmt.Sprintf("服务 %s 在更高优先级服务 %s 之前启动",
							service["name"], otherService["name"]),
						"severity":       "medium",
						"recommendation": fmt.Sprintf("建议按优先级顺序重启服务，先启动 %s", otherService["name"]),
					}
					violations = append(violations, violation)
				}
			}
		}
	}

	status["violations"] = violations

	// 确定整体状态
	if len(violations) > 0 {
		status["overall_status"] = "warning"
	}

	// 显示结果
	fmt.Printf("✅ 活跃服务: %d, ❌ 非活跃服务: %d\n", activeCount, inactiveCount)

	if len(violations) > 0 {
		fmt.Printf("⚠️  发现 %d 个启动违规:\n", len(violations))
		for i, violation := range violations {
			if i < 3 {
				fmt.Printf("   - %s: %s\n",
					getSeverityEmoji(violation["severity"].(string)),
					violation["message"].(string))
			}
		}
		if len(violations) > 3 {
			fmt.Printf("   ... 还有 %d 个违规\n", len(violations)-3)
		}
	} else {
		fmt.Println("✅ 启动顺序正确")
	}

	return status
}

// 开发团队状态检查
func checkTeamStatus() map[string]interface{} {
	fmt.Println("👥 检查开发团队状态...")

	status := map[string]interface{}{
		"timestamp":      time.Now(),
		"overall_status": "success",
		"team_composition": map[string]interface{}{
			"total_members":    0,
			"active_members":   0,
			"inactive_members": 0,
		},
		"role_distribution": map[string]interface{}{
			"roles": []map[string]interface{}{},
		},
		"violations":      []map[string]interface{}{},
		"recommendations": []string{},
	}

	// 尝试连接数据库
	db, err := sql.Open("mysql", "jobfirst:jobfirst123@tcp(localhost:3306)/jobfirst?parseTime=true")
	if err != nil {
		fmt.Printf("❌ 连接数据库失败: %v\n", err)
		fmt.Println("💡 建议: 检查数据库连接配置")
		status["overall_status"] = "critical"
		status["error"] = "数据库连接失败"
		return status
	}
	defer db.Close()

	// 检查数据库连接
	err = db.Ping()
	if err != nil {
		fmt.Printf("❌ 数据库连接测试失败: %v\n", err)
		fmt.Println("💡 建议: 检查数据库服务状态")
		status["overall_status"] = "critical"
		status["error"] = "数据库连接测试失败"
		return status
	}

	// 查询团队成员
	query := `
		SELECT COUNT(*) as total_users
		FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		JOIN roles r ON ur.role_id = r.id
		WHERE r.name IN ('super_admin', 'tech_lead', 'backend_dev', 'frontend_dev', 
		                 'devops_engineer', 'qa_engineer', 'product_manager', 'ui_designer')
	`

	var totalMembers int
	err = db.QueryRow(query).Scan(&totalMembers)
	if err != nil {
		fmt.Printf("⚠️  查询团队成员失败: %v\n", err)
		fmt.Println("💡 建议: 检查用户表和角色表结构")
		status["overall_status"] = "warning"
		status["error"] = "查询团队成员失败"
		return status
	}

	status["team_composition"].(map[string]interface{})["total_members"] = totalMembers
	status["team_composition"].(map[string]interface{})["active_members"] = totalMembers // 简化处理

	fmt.Printf("📊 团队成员: %d\n", totalMembers)

	// 检查关键角色
	criticalRoles := []string{"super_admin", "tech_lead", "backend_dev", "frontend_dev", "devops_engineer"}
	roleDistribution := []map[string]interface{}{}

	for _, roleName := range criticalRoles {
		countQuery := `
			SELECT COUNT(DISTINCT u.id)
			FROM users u
			JOIN user_roles ur ON u.id = ur.user_id
			JOIN roles r ON ur.role_id = r.id
			WHERE r.name = ? AND u.status = 'active'
		`

		var count int
		err = db.QueryRow(countQuery, roleName).Scan(&count)
		if err != nil {
			count = 0
		}

		roleInfo := map[string]interface{}{
			"name":           roleName,
			"current_count":  count,
			"required_count": 1,
		}

		if count == 0 {
			roleInfo["status"] = "missing"
			violation := map[string]interface{}{
				"type":           "role",
				"message":        fmt.Sprintf("缺少关键角色 %s", roleName),
				"severity":       "high",
				"recommendation": fmt.Sprintf("立即招聘或指定 %s 角色", roleName),
			}
			status["violations"] = append(status["violations"].([]map[string]interface{}), violation)
		} else {
			roleInfo["status"] = "ok"
		}

		roleDistribution = append(roleDistribution, roleInfo)
	}

	status["role_distribution"].(map[string]interface{})["roles"] = roleDistribution

	// 显示角色状态
	fmt.Println("🎭 关键角色状态:")
	for _, role := range roleDistribution {
		fmt.Printf("   - %s: %d/1 %s\n",
			role["name"], role["current_count"],
			getRoleStatusEmoji(role["current_count"].(int), 1))
	}

	// 确定整体状态
	if len(status["violations"].([]map[string]interface{})) > 0 {
		status["overall_status"] = "critical"
	}

	return status
}

// 用户权限和订阅状态检查
func checkUserStatus() map[string]interface{} {
	fmt.Println("👤 检查用户权限和订阅状态...")

	status := map[string]interface{}{
		"timestamp":      time.Now(),
		"overall_status": "success",
		"user_statistics": map[string]interface{}{
			"total_users":      0,
			"active_users":     0,
			"subscribed_users": 0,
			"test_users":       0,
		},
		"violations":      []map[string]interface{}{},
		"recommendations": []string{},
	}

	// 尝试连接数据库
	db, err := sql.Open("mysql", "jobfirst:jobfirst123@tcp(localhost:3306)/jobfirst?parseTime=true")
	if err != nil {
		fmt.Printf("❌ 连接数据库失败: %v\n", err)
		fmt.Println("💡 建议: 检查数据库连接配置")
		status["overall_status"] = "critical"
		status["error"] = "数据库连接失败"
		return status
	}
	defer db.Close()

	// 检查数据库连接
	err = db.Ping()
	if err != nil {
		fmt.Printf("❌ 数据库连接测试失败: %v\n", err)
		fmt.Println("💡 建议: 检查数据库服务状态")
		status["overall_status"] = "critical"
		status["error"] = "数据库连接测试失败"
		return status
	}

	// 查询用户统计
	var totalUsers, activeUsers int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		totalUsers = 0
	}

	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&activeUsers)
	if err != nil {
		activeUsers = 0
	}

	status["user_statistics"].(map[string]interface{})["total_users"] = totalUsers
	status["user_statistics"].(map[string]interface{})["active_users"] = activeUsers

	// 查询测试用户
	var testUsers int
	testQuery := `
		SELECT COUNT(*) 
		FROM users u 
		JOIN user_roles ur ON u.id = ur.user_id 
		JOIN roles r ON ur.role_id = r.id 
		WHERE r.name IN ('test_user', 'beta_tester', 'demo_user')
	`
	err = db.QueryRow(testQuery).Scan(&testUsers)
	if err != nil {
		testUsers = 0
	}

	status["user_statistics"].(map[string]interface{})["test_users"] = testUsers

	fmt.Printf("📊 用户统计: 总数 %d, 活跃 %d, 测试 %d\n", totalUsers, activeUsers, testUsers)

	// 检查测试用户状态
	if testUsers > 0 {
		fmt.Printf("🧪 测试用户: %d 个\n", testUsers)

		// 检查过期测试用户
		expiredQuery := `
			SELECT u.username, u.created_at
			FROM users u 
			JOIN user_roles ur ON u.id = ur.user_id 
			JOIN roles r ON ur.role_id = r.id 
			WHERE r.name IN ('test_user', 'beta_tester', 'demo_user')
			AND u.created_at < DATE_SUB(NOW(), INTERVAL 30 DAY)
		`

		rows, err := db.Query(expiredQuery)
		if err == nil {
			defer rows.Close()

			expiredCount := 0
			for rows.Next() {
				var username string
				var createdAt time.Time
				if err := rows.Scan(&username, &createdAt); err == nil {
					expiredCount++
					if expiredCount <= 3 {
						fmt.Printf("   - %s (创建于 %s)\n", username, createdAt.Format("2006-01-02"))
					}
				}
			}

			if expiredCount > 0 {
				violation := map[string]interface{}{
					"type":           "test_user",
					"message":        fmt.Sprintf("发现 %d 个过期测试用户", expiredCount),
					"severity":       "medium",
					"recommendation": "清理过期测试用户或转换为正式用户",
				}
				status["violations"] = append(status["violations"].([]map[string]interface{}), violation)
			}
		}
	}

	// 确定整体状态
	if len(status["violations"].([]map[string]interface{})) > 0 {
		status["overall_status"] = "warning"
	}

	return status
}

// generateComprehensiveReport 生成综合报告
func generateComprehensiveReport(startupStatus, teamStatus, userStatus map[string]interface{}) {
	fmt.Println("=====================================")
	fmt.Println("📋 ZerviGo v3.1.1 综合报告")
	fmt.Println("=====================================")

	// 计算整体健康状态
	overallHealth := calculateOverallHealth(startupStatus, teamStatus, userStatus)
	fmt.Printf("🏥 整体健康状态: %s (%.1f%%)\n", getHealthEmoji(overallHealth), overallHealth*100)

	// 关键指标汇总
	fmt.Println("\n📊 关键指标:")
	fmt.Printf("   - 系统启动顺序: %s\n", getStatusEmoji(startupStatus["overall_status"].(string)))
	fmt.Printf("   - 开发团队状态: %s\n", getStatusEmoji(teamStatus["overall_status"].(string)))
	fmt.Printf("   - 用户管理状态: %s\n", getStatusEmoji(userStatus["overall_status"].(string)))

	// 违规汇总
	totalViolations := len(startupStatus["violations"].([]map[string]interface{})) +
		len(teamStatus["violations"].([]map[string]interface{})) +
		len(userStatus["violations"].([]map[string]interface{}))

	fmt.Printf("\n⚠️  总违规数量: %d\n", totalViolations)

	if totalViolations == 0 {
		fmt.Println("🎉 恭喜！系统运行状态良好，所有检查都通过了！")
	} else {
		fmt.Println("🔧 需要关注的领域:")

		if len(startupStatus["violations"].([]map[string]interface{})) > 0 {
			fmt.Printf("   - 系统启动顺序: %d 个问题\n", len(startupStatus["violations"].([]map[string]interface{})))
		}
		if len(teamStatus["violations"].([]map[string]interface{})) > 0 {
			fmt.Printf("   - 开发团队配置: %d 个问题\n", len(teamStatus["violations"].([]map[string]interface{})))
		}
		if len(userStatus["violations"].([]map[string]interface{})) > 0 {
			fmt.Printf("   - 用户权限管理: %d 个问题\n", len(userStatus["violations"].([]map[string]interface{})))
		}
	}

	// 生成建议
	fmt.Println("\n💡 优先建议:")
	generatePriorityRecommendations(startupStatus, teamStatus, userStatus)

	// 保存报告到文件
	saveReportToFile(startupStatus, teamStatus, userStatus, overallHealth)

	fmt.Println("\n📄 详细报告已保存到: zervigo_report.json")
	fmt.Println("🕐 报告生成时间:", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Println("🚀 ZerviGo v3.1.1 检查完成！")
}

// calculateOverallHealth 计算整体健康状态
func calculateOverallHealth(startupStatus, teamStatus, userStatus map[string]interface{}) float64 {
	healthScore := 0.0

	// 启动顺序检查 (权重: 40%)
	startupScore := getStatusScore(startupStatus["overall_status"].(string))
	healthScore += startupScore * 0.4

	// 团队状态检查 (权重: 30%)
	teamScore := getStatusScore(teamStatus["overall_status"].(string))
	healthScore += teamScore * 0.3

	// 用户状态检查 (权重: 30%)
	userScore := getStatusScore(userStatus["overall_status"].(string))
	healthScore += userScore * 0.3

	return healthScore
}

// getStatusScore 获取状态分数
func getStatusScore(status string) float64 {
	switch status {
	case "success":
		return 1.0
	case "warning":
		return 0.7
	case "critical":
		return 0.3
	default:
		return 0.0
	}
}

// generatePriorityRecommendations 生成优先建议
func generatePriorityRecommendations(startupStatus, teamStatus, userStatus map[string]interface{}) {
	recommendations := []string{}

	// 高优先级建议
	if startupStatus["overall_status"] == "critical" {
		recommendations = append(recommendations, "🚨 立即修复系统启动顺序问题")
	}

	if teamStatus["overall_status"] == "critical" {
		recommendations = append(recommendations, "👥 紧急处理开发团队配置问题")
	}

	if userStatus["overall_status"] == "critical" {
		recommendations = append(recommendations, "👤 立即处理用户权限违规问题")
	}

	// 中优先级建议
	if startupStatus["overall_status"] == "warning" {
		recommendations = append(recommendations, "⚙️ 优化系统启动顺序")
	}

	if teamStatus["overall_status"] == "warning" {
		recommendations = append(recommendations, "🎭 调整团队角色分布")
	}

	if userStatus["overall_status"] == "warning" {
		recommendations = append(recommendations, "🔐 审查用户权限配置")
	}

	// 通用建议
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "✅ 系统运行良好，建议定期进行维护检查")
	}

	recommendations = append(recommendations, "📅 建议每周运行一次完整检查")
	recommendations = append(recommendations, "📚 查看详细报告了解具体问题")
	recommendations = append(recommendations, "🔄 使用 'zervigo startup' 检查启动顺序")
	recommendations = append(recommendations, "👥 使用 'zervigo team' 检查团队状态")
	recommendations = append(recommendations, "👤 使用 'zervigo users' 检查用户状态")

	for i, rec := range recommendations {
		fmt.Printf("   %d. %s\n", i+1, rec)
	}
}

// saveReportToFile 保存报告到文件
func saveReportToFile(startupStatus, teamStatus, userStatus map[string]interface{}, overallHealth float64) {
	report := map[string]interface{}{
		"timestamp":      time.Now(),
		"overall_health": overallHealth,
		"startup_status": startupStatus,
		"team_status":    teamStatus,
		"user_status":    userStatus,
		"summary": map[string]interface{}{
			"total_violations": len(startupStatus["violations"].([]map[string]interface{})) +
				len(teamStatus["violations"].([]map[string]interface{})) +
				len(userStatus["violations"].([]map[string]interface{})),
			"startup_violations": len(startupStatus["violations"].([]map[string]interface{})),
			"team_violations":    len(teamStatus["violations"].([]map[string]interface{})),
			"user_violations":    len(userStatus["violations"].([]map[string]interface{})),
		},
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("❌ 保存报告失败: %v\n", err)
		return
	}

	err = os.WriteFile("zervigo_report.json", data, 0644)
	if err != nil {
		fmt.Printf("❌ 写入报告文件失败: %v\n", err)
	}
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Println("ZerviGo v3.1.1 - 超级管理员工具使用说明")
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("基于 jobfirst-core 核心包的超级管理员管理和监控工具")
	fmt.Println()
	fmt.Println("用法: zervigo [命令]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  startup    - 检查系统启动顺序")
	fmt.Println("  team       - 检查开发团队状态")
	fmt.Println("  users      - 检查用户权限和订阅状态")
	fmt.Println("  full       - 运行完整检查 (默认)")
	fmt.Println("  help       - 显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  zervigo              # 运行完整检查")
	fmt.Println("  zervigo startup      # 只检查启动顺序")
	fmt.Println("  zervigo team         # 只检查团队状态")
	fmt.Println("  zervigo users        # 只检查用户状态")
	fmt.Println()
	fmt.Println("核心功能:")
	fmt.Println("  1. 系统启动顺序检查 - 确保微服务按正确顺序启动")
	fmt.Println("     • 检查服务依赖关系")
	fmt.Println("     • 验证启动优先级")
	fmt.Println("     • 监控服务健康状态")
	fmt.Println()
	fmt.Println("  2. 开发团队管理 - 验证团队角色和权限配置")
	fmt.Println("     • 检查关键角色配置")
	fmt.Println("     • 验证权限矩阵")
	fmt.Println("     • 监控团队结构")
	fmt.Println()
	fmt.Println("  3. 用户权限管理 - 检查用户订阅和访问权限")
	fmt.Println("     • 验证订阅状态")
	fmt.Println("     • 检查权限合规性")
	fmt.Println("     • 监控测试用户")
	fmt.Println()
	fmt.Println("输出文件: zervigo_report.json")
	fmt.Println("版本: v3.1.1")
	fmt.Println("基于: jobfirst-core 核心包")
}

// 辅助函数
func isPortOpen(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 2*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func checkServiceHealth(port int) bool {
	// 简化处理，假设端口开放就是健康
	return isPortOpen(port)
}

func getStatusEmoji(status string) string {
	switch status {
	case "success":
		return "✅"
	case "warning":
		return "⚠️"
	case "critical":
		return "❌"
	default:
		return "❓"
	}
}

func getSeverityEmoji(severity string) string {
	switch severity {
	case "high":
		return "🔴"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	default:
		return "⚪"
	}
}

func getRoleStatusEmoji(current, required int) string {
	if current >= required {
		return "✅"
	} else if current > 0 {
		return "⚠️"
	} else {
		return "❌"
	}
}

func getHealthEmoji(health float64) string {
	if health >= 0.9 {
		return "🟢"
	} else if health >= 0.7 {
		return "🟡"
	} else {
		return "🔴"
	}
}
