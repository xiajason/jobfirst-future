package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"superadmin/system"
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

// checkStartupOrder 检查系统启动顺序
func checkStartupOrder() *system.StartupStatus {
	checker := system.NewStartupChecker(nil)
	status := checker.CheckStartupOrder()

	fmt.Printf("⏰ 启动顺序检查结果: %s\n", getStatusEmoji(status.OverallStatus))
	fmt.Printf("📊 检查服务数量: %d\n", len(status.Services))

	// 显示服务状态概览
	activeCount := 0
	inactiveCount := 0
	for _, service := range status.Services {
		if service.Status == "active" {
			activeCount++
		} else {
			inactiveCount++
		}
	}

	fmt.Printf("✅ 活跃服务: %d, ❌ 非活跃服务: %d\n", activeCount, inactiveCount)

	// 显示违规情况
	if len(status.Violations) > 0 {
		fmt.Printf("⚠️  发现 %d 个启动违规:\n", len(status.Violations))
		for i, violation := range status.Violations {
			if i < 3 { // 只显示前3个
				fmt.Printf("   - %s: %s\n", getSeverityEmoji(violation.Severity), violation.Message)
			}
		}
		if len(status.Violations) > 3 {
			fmt.Printf("   ... 还有 %d 个违规\n", len(status.Violations)-3)
		}
	} else {
		fmt.Println("✅ 启动顺序正确")
	}

	// 显示关键建议
	if len(status.Recommendations) > 0 {
		fmt.Println("💡 关键建议:")
		for i, rec := range status.Recommendations {
			if i < 2 { // 只显示前2个建议
				fmt.Printf("   - %s\n", rec)
			}
		}
	}

	return status
}

// checkTeamStatus 检查团队状态
func checkTeamStatus() *system.TeamStatus {
	checker, err := system.NewTeamChecker(nil)
	if err != nil {
		fmt.Printf("❌ 创建团队检查器失败: %v\n", err)
		fmt.Println("💡 建议: 检查数据库连接配置")
		return nil
	}
	defer checker.Close()

	status, err := checker.CheckTeamStatus()
	if err != nil {
		fmt.Printf("❌ 检查团队状态失败: %v\n", err)
		fmt.Println("💡 建议: 检查数据库表结构和权限")
		return nil
	}

	fmt.Printf("👥 团队状态: %s\n", getStatusEmoji(status.OverallStatus))
	fmt.Printf("📊 团队成员: %d (活跃: %d, 非活跃: %d)\n",
		status.TeamComposition.TotalMembers,
		status.TeamComposition.ActiveMembers,
		status.TeamComposition.InactiveMembers)

	// 显示关键角色状态
	fmt.Println("🎭 关键角色状态:")
	criticalRoles := []string{"super_admin", "tech_lead", "backend_dev", "frontend_dev", "devops_engineer"}
	for _, role := range status.RoleDistribution.Roles {
		for _, criticalRole := range criticalRoles {
			if role.Name == criticalRole {
				fmt.Printf("   - %s: %d/%d %s\n",
					role.Name, role.CurrentCount, role.RequiredCount,
					getRoleStatusEmoji(role.CurrentCount, role.RequiredCount))
				break
			}
		}
	}

	// 显示违规情况
	if len(status.Violations) > 0 {
		fmt.Printf("⚠️  发现 %d 个团队违规:\n", len(status.Violations))
		for i, violation := range status.Violations {
			if i < 3 {
				fmt.Printf("   - %s: %s\n", getSeverityEmoji(violation.Severity), violation.Message)
			}
		}
		if len(status.Violations) > 3 {
			fmt.Printf("   ... 还有 %d 个违规\n", len(status.Violations)-3)
		}
	} else {
		fmt.Println("✅ 团队配置正确")
	}

	return status
}

// checkUserStatus 检查用户状态
func checkUserStatus() *system.UserStatus {
	checker, err := system.NewUserChecker(nil)
	if err != nil {
		fmt.Printf("❌ 创建用户检查器失败: %v\n", err)
		fmt.Println("💡 建议: 检查数据库连接配置")
		return nil
	}
	defer checker.Close()

	status, err := checker.CheckUserStatus()
	if err != nil {
		fmt.Printf("❌ 检查用户状态失败: %v\n", err)
		fmt.Println("💡 建议: 检查用户表和权限表结构")
		return nil
	}

	fmt.Printf("👤 用户状态: %s\n", getStatusEmoji(status.OverallStatus))
	fmt.Printf("📊 用户统计: 总数 %d, 活跃 %d, 订阅 %d, 测试 %d\n",
		status.UserStatistics.TotalUsers,
		status.UserStatistics.ActiveUsers,
		status.UserStatistics.SubscribedUsers,
		status.UserStatistics.TestUsers)

	// 显示订阅收入（如果有）
	if status.SubscriptionStatus.SubscriptionRevenue.ActiveSubscriptions > 0 {
		fmt.Printf("💰 订阅收入: 月度 $%.2f, 年度 $%.2f, 流失率 %.1f%%\n",
			status.SubscriptionStatus.SubscriptionRevenue.MonthlyRevenue,
			status.SubscriptionStatus.SubscriptionRevenue.YearlyRevenue,
			status.SubscriptionStatus.SubscriptionRevenue.ChurnRate)
	}

	// 显示测试用户状态
	if len(status.TestUserStatus.TestUsers) > 0 {
		activeTestUsers := 0
		expiredTestUsers := 0
		for _, testUser := range status.TestUserStatus.TestUsers {
			if testUser.Status == "active" {
				activeTestUsers++
			} else {
				expiredTestUsers++
			}
		}
		fmt.Printf("🧪 测试用户: 活跃 %d, 过期 %d\n", activeTestUsers, expiredTestUsers)
	}

	// 显示违规情况
	if len(status.Violations) > 0 {
		fmt.Printf("⚠️  发现 %d 个用户违规:\n", len(status.Violations))
		for i, violation := range status.Violations {
			if i < 3 {
				fmt.Printf("   - %s: %s\n", getSeverityEmoji(violation.Severity), violation.Message)
			}
		}
		if len(status.Violations) > 3 {
			fmt.Printf("   ... 还有 %d 个违规\n", len(status.Violations)-3)
		}
	} else {
		fmt.Println("✅ 用户权限配置正确")
	}

	return status
}

// generateComprehensiveReport 生成综合报告
func generateComprehensiveReport(startupStatus *system.StartupStatus, teamStatus *system.TeamStatus, userStatus *system.UserStatus) {
	fmt.Println("=====================================")
	fmt.Println("📋 ZerviGo v3.1.1 综合报告")
	fmt.Println("=====================================")

	// 计算整体健康状态
	overallHealth := calculateOverallHealth(startupStatus, teamStatus, userStatus)
	fmt.Printf("🏥 整体健康状态: %s (%.1f%%)\n", getHealthEmoji(overallHealth), overallHealth*100)

	// 关键指标汇总
	fmt.Println("\n📊 关键指标:")
	fmt.Printf("   - 系统启动顺序: %s\n", getStatusEmoji(startupStatus.OverallStatus))
	if teamStatus != nil {
		fmt.Printf("   - 开发团队状态: %s\n", getStatusEmoji(teamStatus.OverallStatus))
	} else {
		fmt.Printf("   - 开发团队状态: ❓ (检查失败)\n")
	}
	if userStatus != nil {
		fmt.Printf("   - 用户管理状态: %s\n", getStatusEmoji(userStatus.OverallStatus))
	} else {
		fmt.Printf("   - 用户管理状态: ❓ (检查失败)\n")
	}

	// 违规汇总
	totalViolations := len(startupStatus.Violations)
	if teamStatus != nil {
		totalViolations += len(teamStatus.Violations)
	}
	if userStatus != nil {
		totalViolations += len(userStatus.Violations)
	}

	fmt.Printf("\n⚠️  总违规数量: %d\n", totalViolations)

	if totalViolations == 0 {
		fmt.Println("🎉 恭喜！系统运行状态良好，所有检查都通过了！")
	} else {
		fmt.Println("🔧 需要关注的领域:")

		if len(startupStatus.Violations) > 0 {
			fmt.Printf("   - 系统启动顺序: %d 个问题\n", len(startupStatus.Violations))
		}
		if teamStatus != nil && len(teamStatus.Violations) > 0 {
			fmt.Printf("   - 开发团队配置: %d 个问题\n", len(teamStatus.Violations))
		}
		if userStatus != nil && len(userStatus.Violations) > 0 {
			fmt.Printf("   - 用户权限管理: %d 个问题\n", len(userStatus.Violations))
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
func calculateOverallHealth(startupStatus *system.StartupStatus, teamStatus *system.TeamStatus, userStatus *system.UserStatus) float64 {
	healthScore := 0.0
	totalChecks := 0

	// 启动顺序检查 (权重: 40%)
	startupScore := getStatusScore(startupStatus.OverallStatus)
	healthScore += startupScore * 0.4
	totalChecks++

	// 团队状态检查 (权重: 30%)
	if teamStatus != nil {
		teamScore := getStatusScore(teamStatus.OverallStatus)
		healthScore += teamScore * 0.3
		totalChecks++
	}

	// 用户状态检查 (权重: 30%)
	if userStatus != nil {
		userScore := getStatusScore(userStatus.OverallStatus)
		healthScore += userScore * 0.3
		totalChecks++
	}

	// 如果某些检查失败，调整权重
	if totalChecks < 3 {
		healthScore = healthScore / float64(totalChecks) * 3.0
	}

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
func generatePriorityRecommendations(startupStatus *system.StartupStatus, teamStatus *system.TeamStatus, userStatus *system.UserStatus) {
	recommendations := []string{}

	// 高优先级建议
	if startupStatus.OverallStatus == "critical" {
		recommendations = append(recommendations, "🚨 立即修复系统启动顺序问题")
	}

	if teamStatus != nil && teamStatus.OverallStatus == "critical" {
		recommendations = append(recommendations, "👥 紧急处理开发团队配置问题")
	}

	if userStatus != nil && userStatus.OverallStatus == "critical" {
		recommendations = append(recommendations, "👤 立即处理用户权限违规问题")
	}

	// 中优先级建议
	if startupStatus.OverallStatus == "warning" {
		recommendations = append(recommendations, "⚙️ 优化系统启动顺序")
	}

	if teamStatus != nil && teamStatus.OverallStatus == "warning" {
		recommendations = append(recommendations, "🎭 调整团队角色分布")
	}

	if userStatus != nil && userStatus.OverallStatus == "warning" {
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
func saveReportToFile(startupStatus *system.StartupStatus, teamStatus *system.TeamStatus, userStatus *system.UserStatus, overallHealth float64) {
	report := map[string]interface{}{
		"timestamp":      time.Now(),
		"overall_health": overallHealth,
		"startup_status": startupStatus,
		"team_status":    teamStatus,
		"user_status":    userStatus,
		"summary": map[string]interface{}{
			"total_violations": len(startupStatus.Violations) + func() int {
				total := 0
				if teamStatus != nil {
					total += len(teamStatus.Violations)
				}
				if userStatus != nil {
					total += len(userStatus.Violations)
				}
				return total
			}(),
			"startup_violations": len(startupStatus.Violations),
			"team_violations": func() int {
				if teamStatus != nil {
					return len(teamStatus.Violations)
				}
				return 0
			}(),
			"user_violations": func() int {
				if userStatus != nil {
					return len(userStatus.Violations)
				}
				return 0
			}(),
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
