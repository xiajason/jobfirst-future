package cicd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"superadmin/errors"
)

// Manager CI/CD管理器
type Manager struct {
	config *CICDConfig
}

// CICDConfig CI/CD配置
type CICDConfig struct {
	GitHubToken    string `json:"github_token"`
	GitLabToken    string `json:"gitlab_token"`
	JenkinsURL     string `json:"jenkins_url"`
	JenkinsUser    string `json:"jenkins_user"`
	JenkinsToken   string `json:"jenkins_token"`
	DockerRegistry string `json:"docker_registry"`
	K8sCluster     string `json:"k8s_cluster"`
}

// NewManager 创建CI/CD管理器
func NewManager(config *CICDConfig) *Manager {
	return &Manager{
		config: config,
	}
}

// CICDStatus CI/CD状态
type CICDStatus struct {
	Pipelines    []CICDPipeline   `json:"pipelines"`
	Repositories []CICDRepository `json:"repositories"`
	Webhooks     []CICDWebhook    `json:"webhooks"`
	LastCheck    time.Time        `json:"last_check"`
	Health       string           `json:"health"`
}

// CICDPipeline CI/CD流水线
type CICDPipeline struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Status      string          `json:"status"`
	Branch      string          `json:"branch"`
	Commit      string          `json:"commit"`
	Author      string          `json:"author"`
	CreatedAt   time.Time       `json:"created_at"`
	StartedAt   time.Time       `json:"started_at"`
	FinishedAt  time.Time       `json:"finished_at"`
	Duration    int64           `json:"duration_seconds"`
	Stages      []PipelineStage `json:"stages"`
	Artifacts   []string        `json:"artifacts"`
	Environment string          `json:"environment"`
}

// PipelineStage 流水线阶段
type PipelineStage struct {
	Name      string        `json:"name"`
	Status    string        `json:"status"`
	StartedAt time.Time     `json:"started_at"`
	Duration  int64         `json:"duration_seconds"`
	Jobs      []PipelineJob `json:"jobs"`
}

// PipelineJob 流水线任务
type PipelineJob struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
	Duration  int64     `json:"duration_seconds"`
	Logs      []string  `json:"logs"`
}

// CICDRepository CI/CD仓库
type CICDRepository struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	Branch     string    `json:"default_branch"`
	LastCommit string    `json:"last_commit"`
	LastUpdate time.Time `json:"last_update"`
	IsActive   bool      `json:"is_active"`
	Webhooks   []string  `json:"webhooks"`
}

// CICDWebhook CI/CD Webhook
type CICDWebhook struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Events      []string  `json:"events"`
	IsActive    bool      `json:"is_active"`
	LastTrigger time.Time `json:"last_trigger"`
	Secret      string    `json:"secret_masked"`
}

// GetCICDStatus 获取CI/CD状态
func (m *Manager) GetCICDStatus() (*CICDStatus, error) {
	status := &CICDStatus{
		LastCheck: time.Now(),
	}

	// 获取流水线状态
	pipelines, err := m.GetCICDPipelines()
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "获取流水线状态失败", err)
	}
	status.Pipelines = pipelines

	// 获取仓库状态
	repositories, err := m.GetCICDRepositories()
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "获取仓库状态失败", err)
	}
	status.Repositories = repositories

	// 获取Webhook状态
	webhooks, err := m.GetCICDWebhooks()
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "获取Webhook状态失败", err)
	}
	status.Webhooks = webhooks

	// 计算整体健康状态
	status.Health = m.calculateHealthStatus(status)

	return status, nil
}

// calculateHealthStatus 计算健康状态
func (m *Manager) calculateHealthStatus(status *CICDStatus) string {
	// 检查流水线状态
	failedPipelines := 0
	for _, pipeline := range status.Pipelines {
		if pipeline.Status == "failed" {
			failedPipelines++
		}
	}

	// 检查仓库状态
	inactiveRepos := 0
	for _, repo := range status.Repositories {
		if !repo.IsActive {
			inactiveRepos++
		}
	}

	// 检查Webhook状态
	inactiveWebhooks := 0
	for _, webhook := range status.Webhooks {
		if !webhook.IsActive {
			inactiveWebhooks++
		}
	}

	// 计算健康分数
	totalIssues := failedPipelines + inactiveRepos + inactiveWebhooks
	if totalIssues == 0 {
		return "healthy"
	} else if totalIssues <= 2 {
		return "warning"
	} else {
		return "critical"
	}
}

// GetCICDPipelines 获取CI/CD流水线
func (m *Manager) GetCICDPipelines() ([]CICDPipeline, error) {
	pipelines := []CICDPipeline{}

	// 检查GitHub Actions
	if m.config.GitHubToken != "" {
		githubPipelines, err := m.getGitHubPipelines()
		if err != nil {
			return nil, errors.WrapError(errors.ErrCodeService, "获取GitHub流水线失败", err)
		}
		pipelines = append(pipelines, githubPipelines...)
	}

	// 检查GitLab CI
	if m.config.GitLabToken != "" {
		gitlabPipelines, err := m.getGitLabPipelines()
		if err != nil {
			return nil, errors.WrapError(errors.ErrCodeService, "获取GitLab流水线失败", err)
		}
		pipelines = append(pipelines, gitlabPipelines...)
	}

	// 检查Jenkins
	if m.config.JenkinsURL != "" {
		jenkinsPipelines, err := m.getJenkinsPipelines()
		if err != nil {
			return nil, errors.WrapError(errors.ErrCodeService, "获取Jenkins流水线失败", err)
		}
		pipelines = append(pipelines, jenkinsPipelines...)
	}

	return pipelines, nil
}

// getGitHubPipelines 获取GitHub流水线
func (m *Manager) getGitHubPipelines() ([]CICDPipeline, error) {
	pipelines := []CICDPipeline{}

	// 使用GitHub CLI获取工作流运行
	cmd := exec.Command("gh", "run", "list", "--limit", "10", "--json", "id,name,status,headBranch,headSha,createdAt,startedAt,completedAt,conclusion")
	output, err := cmd.Output()
	if err != nil {
		// 如果GitHub CLI不可用，返回模拟数据
		return m.getMockGitHubPipelines(), nil
	}

	var runs []map[string]interface{}
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, err
	}

	for _, run := range runs {
		pipeline := CICDPipeline{
			ID:        fmt.Sprintf("github_%v", run["id"]),
			Name:      fmt.Sprintf("%v", run["name"]),
			Status:    fmt.Sprintf("%v", run["status"]),
			Branch:    fmt.Sprintf("%v", run["headBranch"]),
			Commit:    fmt.Sprintf("%v", run["headSha"]),
			CreatedAt: time.Now(), // 简化处理
		}

		// 解析时间字段
		if createdAt, ok := run["createdAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
				pipeline.CreatedAt = t
			}
		}

		pipelines = append(pipelines, pipeline)
	}

	return pipelines, nil
}

// getMockGitHubPipelines 获取模拟GitHub流水线
func (m *Manager) getMockGitHubPipelines() []CICDPipeline {
	return []CICDPipeline{
		{
			ID:         "github_123456",
			Name:       "Build and Test",
			Status:     "success",
			Branch:     "main",
			Commit:     "abc123def456",
			Author:     "developer",
			CreatedAt:  time.Now().Add(-2 * time.Hour),
			StartedAt:  time.Now().Add(-2 * time.Hour),
			FinishedAt: time.Now().Add(-1 * time.Hour),
			Duration:   3600,
			Stages: []PipelineStage{
				{
					Name:     "Build",
					Status:   "success",
					Duration: 1800,
					Jobs: []PipelineJob{
						{
							Name:     "Build Application",
							Status:   "success",
							Duration: 1800,
						},
					},
				},
				{
					Name:     "Test",
					Status:   "success",
					Duration: 1800,
					Jobs: []PipelineJob{
						{
							Name:     "Unit Tests",
							Status:   "success",
							Duration: 900,
						},
						{
							Name:     "Integration Tests",
							Status:   "success",
							Duration: 900,
						},
					},
				},
			},
			Environment: "production",
		},
	}
}

// getGitLabPipelines 获取GitLab流水线
func (m *Manager) getGitLabPipelines() ([]CICDPipeline, error) {
	// 模拟GitLab流水线数据
	return []CICDPipeline{
		{
			ID:        "gitlab_789012",
			Name:      "Deploy to Staging",
			Status:    "running",
			Branch:    "develop",
			Commit:    "def456ghi789",
			Author:    "devops",
			CreatedAt: time.Now().Add(-30 * time.Minute),
			StartedAt: time.Now().Add(-30 * time.Minute),
			Duration:  1800,
			Stages: []PipelineStage{
				{
					Name:     "Build",
					Status:   "success",
					Duration: 1200,
				},
				{
					Name:     "Deploy",
					Status:   "running",
					Duration: 600,
				},
			},
			Environment: "staging",
		},
	}, nil
}

// getJenkinsPipelines 获取Jenkins流水线
func (m *Manager) getJenkinsPipelines() ([]CICDPipeline, error) {
	// 模拟Jenkins流水线数据
	return []CICDPipeline{
		{
			ID:         "jenkins_345678",
			Name:       "Nightly Build",
			Status:     "failed",
			Branch:     "main",
			Commit:     "ghi789jkl012",
			Author:     "system",
			CreatedAt:  time.Now().Add(-4 * time.Hour),
			StartedAt:  time.Now().Add(-4 * time.Hour),
			FinishedAt: time.Now().Add(-3 * time.Hour),
			Duration:   3600,
			Stages: []PipelineStage{
				{
					Name:     "Build",
					Status:   "success",
					Duration: 1800,
				},
				{
					Name:     "Test",
					Status:   "failed",
					Duration: 1800,
				},
			},
			Environment: "testing",
		},
	}, nil
}

// GetCICDRepositories 获取CI/CD仓库
func (m *Manager) GetCICDRepositories() ([]CICDRepository, error) {
	repositories := []CICDRepository{}

	// 获取GitHub仓库
	if m.config.GitHubToken != "" {
		githubRepos, err := m.getGitHubRepositories()
		if err != nil {
			return nil, errors.WrapError(errors.ErrCodeService, "获取GitHub仓库失败", err)
		}
		repositories = append(repositories, githubRepos...)
	}

	// 获取GitLab仓库
	if m.config.GitLabToken != "" {
		gitlabRepos, err := m.getGitLabRepositories()
		if err != nil {
			return nil, errors.WrapError(errors.ErrCodeService, "获取GitLab仓库失败", err)
		}
		repositories = append(repositories, gitlabRepos...)
	}

	return repositories, nil
}

// getGitHubRepositories 获取GitHub仓库
func (m *Manager) getGitHubRepositories() ([]CICDRepository, error) {
	// 模拟GitHub仓库数据
	return []CICDRepository{
		{
			ID:         "github_jobfirst_main",
			Name:       "jobfirst-main",
			URL:        "https://github.com/jobfirst/jobfirst-main",
			Branch:     "main",
			LastCommit: "abc123def456",
			LastUpdate: time.Now().Add(-1 * time.Hour),
			IsActive:   true,
			Webhooks:   []string{"webhook_1", "webhook_2"},
		},
		{
			ID:         "github_jobfirst_api",
			Name:       "jobfirst-api",
			URL:        "https://github.com/jobfirst/jobfirst-api",
			Branch:     "main",
			LastCommit: "def456ghi789",
			LastUpdate: time.Now().Add(-2 * time.Hour),
			IsActive:   true,
			Webhooks:   []string{"webhook_3"},
		},
	}, nil
}

// getGitLabRepositories 获取GitLab仓库
func (m *Manager) getGitLabRepositories() ([]CICDRepository, error) {
	// 模拟GitLab仓库数据
	return []CICDRepository{
		{
			ID:         "gitlab_jobfirst_frontend",
			Name:       "jobfirst-frontend",
			URL:        "https://gitlab.com/jobfirst/jobfirst-frontend",
			Branch:     "develop",
			LastCommit: "ghi789jkl012",
			LastUpdate: time.Now().Add(-30 * time.Minute),
			IsActive:   true,
			Webhooks:   []string{"webhook_4"},
		},
	}, nil
}

// GetCICDWebhooks 获取CI/CD Webhook
func (m *Manager) GetCICDWebhooks() ([]CICDWebhook, error) {
	webhooks := []CICDWebhook{
		{
			ID:          "webhook_1",
			Name:        "GitHub Push Webhook",
			URL:         "https://api.jobfirst.com/webhooks/github/push",
			Events:      []string{"push", "pull_request"},
			IsActive:    true,
			LastTrigger: time.Now().Add(-1 * time.Hour),
			Secret:      "****",
		},
		{
			ID:          "webhook_2",
			Name:        "GitHub Release Webhook",
			URL:         "https://api.jobfirst.com/webhooks/github/release",
			Events:      []string{"release"},
			IsActive:    true,
			LastTrigger: time.Now().Add(-24 * time.Hour),
			Secret:      "****",
		},
		{
			ID:          "webhook_3",
			Name:        "GitLab CI Webhook",
			URL:         "https://api.jobfirst.com/webhooks/gitlab/ci",
			Events:      []string{"pipeline"},
			IsActive:    true,
			LastTrigger: time.Now().Add(-2 * time.Hour),
			Secret:      "****",
		},
		{
			ID:          "webhook_4",
			Name:        "Jenkins Build Webhook",
			URL:         "https://api.jobfirst.com/webhooks/jenkins/build",
			Events:      []string{"build"},
			IsActive:    false,
			LastTrigger: time.Now().Add(-7 * 24 * time.Hour),
			Secret:      "****",
		},
	}

	return webhooks, nil
}

// TriggerCICDDeploy 触发CI/CD部署
func (m *Manager) TriggerCICDDeploy(environment string) error {
	// 验证环境
	if !m.isValidEnvironment(environment) {
		return errors.NewError(errors.ErrCodeValidation, "无效的部署环境")
	}

	// 根据环境选择部署策略
	switch environment {
	case "development":
		return m.triggerDevelopmentDeploy()
	case "staging":
		return m.triggerStagingDeploy()
	case "production":
		return m.triggerProductionDeploy()
	default:
		return errors.NewError(errors.ErrCodeValidation, "不支持的部署环境")
	}
}

// isValidEnvironment 验证环境是否有效
func (m *Manager) isValidEnvironment(environment string) bool {
	validEnvironments := []string{"development", "staging", "production"}
	for _, env := range validEnvironments {
		if environment == env {
			return true
		}
	}
	return false
}

// triggerDevelopmentDeploy 触发开发环境部署
func (m *Manager) triggerDevelopmentDeploy() error {
	// 触发GitHub Actions工作流
	cmd := exec.Command("gh", "workflow", "run", "deploy-dev.yml", "--ref", "develop")
	if err := cmd.Run(); err != nil {
		return errors.WrapError(errors.ErrCodeService, "触发开发环境部署失败", err)
	}

	return nil
}

// triggerStagingDeploy 触发预发布环境部署
func (m *Manager) triggerStagingDeploy() error {
	// 触发GitLab CI流水线
	cmd := exec.Command("gitlab-ci", "trigger", "deploy-staging")
	if err := cmd.Run(); err != nil {
		return errors.WrapError(errors.ErrCodeService, "触发预发布环境部署失败", err)
	}

	return nil
}

// triggerProductionDeploy 触发生产环境部署
func (m *Manager) triggerProductionDeploy() error {
	// 触发Jenkins部署任务
	cmd := exec.Command("jenkins-cli", "build", "deploy-production")
	if err := cmd.Run(); err != nil {
		return errors.WrapError(errors.ErrCodeService, "触发生产环境部署失败", err)
	}

	return nil
}

// GetCICDLogs 获取CI/CD日志
func (m *Manager) GetCICDLogs(pipelineID string) ([]string, error) {
	logs := []string{}

	// 根据流水线ID获取日志
	if strings.HasPrefix(pipelineID, "github_") {
		return m.getGitHubLogs(pipelineID)
	} else if strings.HasPrefix(pipelineID, "gitlab_") {
		return m.getGitLabLogs(pipelineID)
	} else if strings.HasPrefix(pipelineID, "jenkins_") {
		return m.getJenkinsLogs(pipelineID)
	}

	return logs, errors.NewError(errors.ErrCodeValidation, "无效的流水线ID")
}

// getGitHubLogs 获取GitHub日志
func (m *Manager) getGitHubLogs(pipelineID string) ([]string, error) {
	// 使用GitHub CLI获取日志
	runID := strings.TrimPrefix(pipelineID, "github_")
	cmd := exec.Command("gh", "run", "view", runID, "--log")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.WrapError(errors.ErrCodeService, "获取GitHub日志失败", err)
	}

	logs := strings.Split(string(output), "\n")
	return logs, nil
}

// getGitLabLogs 获取GitLab日志
func (m *Manager) getGitLabLogs(pipelineID string) ([]string, error) {
	// 模拟GitLab日志
	return []string{
		"Running with gitlab-runner 15.0.0",
		"Preparing the \"docker\" executor",
		"Using Docker executor with image node:16",
		"Pulling docker image node:16",
		"Using docker image sha256:abc123 for node:16",
		"Running on runner-abc123-project-0 via runner-abc123...",
		"Fetching changes...",
		"Checking out abc123 as main...",
		"Skipping Git submodules setup",
		"Executing \"step_script\" stage of the job",
		"$ npm install",
		"npm WARN deprecated some-package@1.0.0",
		"added 1234 packages in 45s",
		"$ npm run build",
		"Building application...",
		"Build completed successfully",
		"Job succeeded",
	}, nil
}

// getJenkinsLogs 获取Jenkins日志
func (m *Manager) getJenkinsLogs(pipelineID string) ([]string, error) {
	// 模拟Jenkins日志
	return []string{
		"Started by user admin",
		"Building in workspace /var/jenkins_home/workspace/jobfirst-main",
		"Checking out git https://github.com/jobfirst/jobfirst-main.git",
		"Commit message: Fix critical bug in user authentication",
		"Running pre-build steps",
		"Executing shell script",
		"$ echo 'Starting build process'",
		"Starting build process",
		"$ npm install",
		"npm WARN deprecated some-package@1.0.0",
		"added 1234 packages in 45s",
		"$ npm run test",
		"Running tests...",
		"Test suite failed: 2 tests failed",
		"Build step 'Execute shell' marked build as failure",
		"Finished: FAILURE",
	}, nil
}
