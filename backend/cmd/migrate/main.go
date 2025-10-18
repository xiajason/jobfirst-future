// 数据迁移工具 - 从 V1.0 迁移到 V3.0
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/term"
)

// 数据库配置
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// V1.0 数据结构
type ResumeV1 struct {
	ID            int64   `json:"id"`
	UUID          string  `json:"uuid"`
	UserID        int64   `json:"user_id"`
	Title         string  `json:"title"`
	Summary       *string `json:"summary"` // 可能为 NULL
	TemplateID    *int64  `json:"template_id"`
	Content       string  `json:"content"` // JSON 字符串
	Status        string  `json:"status"`
	Visibility    string  `json:"visibility"`
	ViewCount     int     `json:"view_count"`
	DownloadCount int     `json:"download_count"`
	ShareCount    int     `json:"share_count"`
	IsDefault     bool    `json:"is_default"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type UserV1 struct {
	ID        int64  `json:"id"`
	UUID      string `json:"uuid"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	AvatarURL string `json:"avatar_url"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UserProfileV1 struct {
	ID       int64  `json:"id"`
	UserID   int64  `json:"user_id"`
	Bio      string `json:"bio"`
	Location string `json:"location"`
	Website  string `json:"website"`
	Skills   string `json:"skills"` // JSON 字符串
}

// V3.0 数据结构
type ResumeV3 struct {
	ID            int64     `json:"id"`
	UUID          string    `json:"uuid"`
	UserID        int64     `json:"user_id"`
	Title         string    `json:"title"`
	Slug          string    `json:"slug"`
	Summary       string    `json:"summary"`
	TemplateID    *int64    `json:"template_id"`
	Content       string    `json:"content"`        // Markdown 格式
	ContentVector *string   `json:"content_vector"` // AI 向量数据
	Status        string    `json:"status"`
	Visibility    string    `json:"visibility"`
	CanComment    bool      `json:"can_comment"`
	ViewCount     int       `json:"view_count"`
	DownloadCount int       `json:"download_count"`
	ShareCount    int       `json:"share_count"`
	CommentCount  int       `json:"comment_count"`
	LikeCount     int       `json:"like_count"`
	IsDefault     bool      `json:"is_default"`
	PublishedAt   *string   `json:"published_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     *string   `json:"deleted_at"`
}

type SkillV3 struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	IsPopular   bool      `json:"is_popular"`
	SearchCount int       `json:"search_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ResumeSkillV3 struct {
	ID                int64     `json:"id"`
	ResumeID          int64     `json:"resume_id"`
	SkillID           int64     `json:"skill_id"`
	ProficiencyLevel  string    `json:"proficiency_level"`
	YearsOfExperience float64   `json:"years_of_experience"`
	IsHighlighted     bool      `json:"is_highlighted"`
	CreatedAt         time.Time `json:"created_at"`
}

type WorkExperienceV3 struct {
	ID           int64     `json:"id"`
	ResumeID     int64     `json:"resume_id"`
	CompanyID    *int64    `json:"company_id"`
	PositionID   *int64    `json:"position_id"`
	Title        string    `json:"title"`
	StartDate    string    `json:"start_date"`
	EndDate      *string   `json:"end_date"`
	IsCurrent    bool      `json:"is_current"`
	Location     string    `json:"location"`
	Description  string    `json:"description"`
	Achievements string    `json:"achievements"`
	Technologies string    `json:"technologies"`
	SalaryRange  string    `json:"salary_range"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type EducationV3 struct {
	ID            int64     `json:"id"`
	ResumeID      int64     `json:"resume_id"`
	School        string    `json:"school"`
	Degree        string    `json:"degree"`
	Major         string    `json:"major"`
	StartDate     *string   `json:"start_date"`
	EndDate       *string   `json:"end_date"`
	GPA           *float64  `json:"gpa"`
	Location      string    `json:"location"`
	Description   string    `json:"description"`
	IsHighlighted bool      `json:"is_highlighted"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ProjectV3 struct {
	ID              int64     `json:"id"`
	ResumeID        int64     `json:"resume_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	StartDate       *string   `json:"start_date"`
	EndDate         *string   `json:"end_date"`
	Status          string    `json:"status"`
	TechnologyStack string    `json:"technology_stack"`
	ProjectURL      string    `json:"project_url"`
	GithubURL       string    `json:"github_url"`
	DemoURL         string    `json:"demo_url"`
	CompanyID       *int64    `json:"company_id"`
	IsHighlighted   bool      `json:"is_highlighted"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CertificationV3 struct {
	ID            int64     `json:"id"`
	ResumeID      int64     `json:"resume_id"`
	Name          string    `json:"name"`
	Issuer        string    `json:"issuer"`
	IssueDate     string    `json:"issue_date"`
	ExpiryDate    *string   `json:"expiry_date"`
	CredentialID  string    `json:"credential_id"`
	CredentialURL string    `json:"credential_url"`
	IsHighlighted bool      `json:"is_highlighted"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// 迁移器
type Migrator struct {
	sourceDB *sql.DB
	targetDB *sql.DB
}

// 创建迁移器
func NewMigrator(sourceConfig, targetConfig DBConfig) (*Migrator, error) {
	// 连接源数据库 (V1.0)
	sourceDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		sourceConfig.User, sourceConfig.Password, sourceConfig.Host, sourceConfig.Port, sourceConfig.Database)
	sourceDB, err := sql.Open("mysql", sourceDSN)
	if err != nil {
		return nil, fmt.Errorf("连接源数据库失败: %v", err)
	}

	// 连接目标数据库 (V3.0)
	targetDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		targetConfig.User, targetConfig.Password, targetConfig.Host, targetConfig.Port, targetConfig.Database)
	targetDB, err := sql.Open("mysql", targetDSN)
	if err != nil {
		return nil, fmt.Errorf("连接目标数据库失败: %v", err)
	}

	return &Migrator{
		sourceDB: sourceDB,
		targetDB: targetDB,
	}, nil
}

// 关闭数据库连接
func (m *Migrator) Close() {
	if m.sourceDB != nil {
		m.sourceDB.Close()
	}
	if m.targetDB != nil {
		m.targetDB.Close()
	}
}

// 迁移技能数据
func (m *Migrator) MigrateSkills() error {
	log.Println("开始迁移技能数据...")

	// 从用户资料中提取技能
	query := `
		SELECT DISTINCT JSON_UNQUOTE(JSON_EXTRACT(skills, CONCAT('$[', numbers.n, ']'))) as skill_name
		FROM user_profiles up
		CROSS JOIN (
			SELECT 0 as n UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION
			SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9
		) numbers
		WHERE JSON_EXTRACT(skills, CONCAT('$[', numbers.n, ']')) IS NOT NULL
		AND JSON_UNQUOTE(JSON_EXTRACT(skills, CONCAT('$[', numbers.n, ']'))) != ''
	`

	rows, err := m.sourceDB.Query(query)
	if err != nil {
		return fmt.Errorf("查询技能数据失败: %v", err)
	}
	defer rows.Close()

	skillMap := make(map[string]bool)
	for rows.Next() {
		var skillName string
		if err := rows.Scan(&skillName); err != nil {
			continue
		}
		if strings.TrimSpace(skillName) != "" {
			skillMap[strings.TrimSpace(skillName)] = true
		}
	}

	// 插入技能到 V3.0 数据库
	for skillName := range skillMap {
		// 确定技能分类
		category := m.determineSkillCategory(skillName)

		skill := SkillV3{
			Name:        skillName,
			Category:    category,
			Description: fmt.Sprintf("%s 相关技能", skillName),
			IsPopular:   m.isPopularSkill(skillName),
			SearchCount: 0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		_, err := m.targetDB.Exec(`
			INSERT INTO skills (name, category, description, is_popular, search_count, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
			updated_at = VALUES(updated_at)
		`, skill.Name, skill.Category, skill.Description, skill.IsPopular, skill.SearchCount, skill.CreatedAt, skill.UpdatedAt)

		if err != nil {
			log.Printf("插入技能失败 %s: %v", skillName, err)
		}
	}

	log.Printf("技能数据迁移完成，共迁移 %d 个技能", len(skillMap))
	return nil
}

// 确定技能分类
func (m *Migrator) determineSkillCategory(skillName string) string {
	skillName = strings.ToLower(skillName)

	// 前端技能
	frontendSkills := []string{"react", "vue", "angular", "javascript", "typescript", "html", "css", "sass", "less", "webpack", "vite", "next.js", "nuxt.js"}
	for _, skill := range frontendSkills {
		if strings.Contains(skillName, skill) {
			return "前端开发"
		}
	}

	// 后端技能
	backendSkills := []string{"go", "golang", "java", "python", "node.js", "php", "ruby", "c#", "spring", "django", "flask", "express", "gin", "echo"}
	for _, skill := range backendSkills {
		if strings.Contains(skillName, skill) {
			return "后端开发"
		}
	}

	// 数据库技能
	dbSkills := []string{"mysql", "postgresql", "mongodb", "redis", "elasticsearch", "sql", "nosql", "oracle", "sqlite"}
	for _, skill := range dbSkills {
		if strings.Contains(skillName, skill) {
			return "数据库"
		}
	}

	// 运维技能
	devopsSkills := []string{"docker", "kubernetes", "aws", "azure", "gcp", "jenkins", "gitlab", "ci/cd", "linux", "nginx", "apache"}
	for _, skill := range devopsSkills {
		if strings.Contains(skillName, skill) {
			return "运维部署"
		}
	}

	// 设计技能
	designSkills := []string{"photoshop", "illustrator", "figma", "sketch", "ui", "ux", "design", "adobe"}
	for _, skill := range designSkills {
		if strings.Contains(skillName, skill) {
			return "设计"
		}
	}

	return "其他"
}

// 判断是否为热门技能
func (m *Migrator) isPopularSkill(skillName string) bool {
	popularSkills := []string{
		"react", "vue", "javascript", "typescript", "python", "java", "go", "golang",
		"mysql", "postgresql", "redis", "docker", "kubernetes", "aws", "git",
		"html", "css", "node.js", "spring", "django", "flask", "mongodb",
	}

	skillName = strings.ToLower(skillName)
	for _, skill := range popularSkills {
		if strings.Contains(skillName, skill) {
			return true
		}
	}
	return false
}

// 迁移简历数据
func (m *Migrator) MigrateResumes() error {
	log.Println("开始迁移简历数据...")

	// 查询 V1.0 简历数据
	query := `
		SELECT id, uuid, user_id, title, summary, template_id, content, status, 
		       visibility, view_count, download_count, share_count, is_default,
		       created_at, updated_at
		FROM resumes
		WHERE deleted_at IS NULL
	`

	rows, err := m.sourceDB.Query(query)
	if err != nil {
		return fmt.Errorf("查询简历数据失败: %v", err)
	}
	defer rows.Close()

	var resumes []ResumeV1
	for rows.Next() {
		var resume ResumeV1
		err := rows.Scan(
			&resume.ID, &resume.UUID, &resume.UserID, &resume.Title, &resume.Summary,
			&resume.TemplateID, &resume.Content, &resume.Status, &resume.Visibility,
			&resume.ViewCount, &resume.DownloadCount, &resume.ShareCount, &resume.IsDefault,
			&resume.CreatedAt, &resume.UpdatedAt,
		)
		if err != nil {
			log.Printf("扫描简历数据失败: %v", err)
			continue
		}
		resumes = append(resumes, resume)
	}

	// 迁移每个简历
	for _, resumeV1 := range resumes {
		if err := m.migrateSingleResume(resumeV1); err != nil {
			log.Printf("迁移简历 %d 失败: %v", resumeV1.ID, err)
			continue
		}
	}

	log.Printf("简历数据迁移完成，共迁移 %d 个简历", len(resumes))
	return nil
}

// 迁移单个简历
func (m *Migrator) migrateSingleResume(resumeV1 ResumeV1) error {
	// 解析 V1.0 的 JSON 内容
	var contentData map[string]interface{}
	if err := json.Unmarshal([]byte(resumeV1.Content), &contentData); err != nil {
		log.Printf("解析简历内容失败: %v", err)
		contentData = make(map[string]interface{})
	}

	// 转换为 Markdown 格式
	markdownContent := m.convertToMarkdown(contentData)

	// 创建 V3.0 简历
	createdAt, err := time.Parse("2006-01-02 15:04:05", resumeV1.CreatedAt)
	if err != nil {
		createdAt = time.Now()
	}
	updatedAt, err := time.Parse("2006-01-02 15:04:05", resumeV1.UpdatedAt)
	if err != nil {
		updatedAt = time.Now()
	}

	var publishedAt *string
	if resumeV1.Status == "published" && !updatedAt.IsZero() {
		publishedAtStr := updatedAt.Format("2006-01-02 15:04:05")
		publishedAt = &publishedAtStr
	}

	// 处理 summary 字段
	summary := ""
	if resumeV1.Summary != nil {
		summary = *resumeV1.Summary
	}

	resumeV3 := ResumeV3{
		UUID:          resumeV1.UUID,
		UserID:        resumeV1.UserID,
		Title:         resumeV1.Title,
		Slug:          m.generateSlug(resumeV1.Title),
		Summary:       summary,
		TemplateID:    resumeV1.TemplateID,
		Content:       markdownContent,
		Status:        resumeV1.Status,
		Visibility:    resumeV1.Visibility,
		CanComment:    true,
		ViewCount:     resumeV1.ViewCount,
		DownloadCount: resumeV1.DownloadCount,
		ShareCount:    resumeV1.ShareCount,
		CommentCount:  0,
		LikeCount:     0,
		IsDefault:     resumeV1.IsDefault,
		PublishedAt:   publishedAt,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}

	// 插入 V3.0 简历
	result, err := m.targetDB.Exec(`
		INSERT INTO resumes (uuid, user_id, title, slug, summary, template_id, content, status, 
		                    visibility, can_comment, view_count, download_count, share_count, 
		                    comment_count, like_count, is_default, published_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, resumeV3.UUID, resumeV3.UserID, resumeV3.Title, resumeV3.Slug, resumeV3.Summary,
		resumeV3.TemplateID, resumeV3.Content, resumeV3.Status, resumeV3.Visibility,
		resumeV3.CanComment, resumeV3.ViewCount, resumeV3.DownloadCount, resumeV3.ShareCount,
		resumeV3.CommentCount, resumeV3.LikeCount, resumeV3.IsDefault, resumeV3.PublishedAt,
		resumeV3.CreatedAt, resumeV3.UpdatedAt)

	if err != nil {
		return fmt.Errorf("插入简历失败: %v", err)
	}

	newResumeID, _ := result.LastInsertId()

	// 迁移简历技能
	if err := m.migrateResumeSkills(newResumeID, contentData); err != nil {
		log.Printf("迁移简历技能失败: %v", err)
	}

	// 迁移工作经历
	if err := m.migrateWorkExperiences(newResumeID, contentData); err != nil {
		log.Printf("迁移工作经历失败: %v", err)
	}

	// 迁移教育背景
	if err := m.migrateEducations(newResumeID, contentData); err != nil {
		log.Printf("迁移教育背景失败: %v", err)
	}

	// 迁移项目经验
	if err := m.migrateProjects(newResumeID, contentData); err != nil {
		log.Printf("迁移项目经验失败: %v", err)
	}

	// 迁移证书
	if err := m.migrateCertifications(newResumeID, contentData); err != nil {
		log.Printf("迁移证书失败: %v", err)
	}

	return nil
}

// 转换为 Markdown 格式
func (m *Migrator) convertToMarkdown(contentData map[string]interface{}) string {
	var markdown strings.Builder

	// 个人信息
	if basicInfo, ok := contentData["basic_info"].(map[string]interface{}); ok {
		markdown.WriteString("# 个人信息\n\n")
		if name, ok := basicInfo["name"].(string); ok {
			markdown.WriteString(fmt.Sprintf("**姓名：** %s\n", name))
		}
		if email, ok := basicInfo["email"].(string); ok {
			markdown.WriteString(fmt.Sprintf("**邮箱：** %s\n", email))
		}
		if phone, ok := basicInfo["phone"].(string); ok {
			markdown.WriteString(fmt.Sprintf("**电话：** %s\n", phone))
		}
		if location, ok := basicInfo["location"].(string); ok {
			markdown.WriteString(fmt.Sprintf("**地址：** %s\n", location))
		}
		markdown.WriteString("\n")
	}

	// 工作经历
	if experiences, ok := contentData["experience"].([]interface{}); ok {
		markdown.WriteString("## 工作经历\n\n")
		for _, exp := range experiences {
			if expMap, ok := exp.(map[string]interface{}); ok {
				markdown.WriteString("### ")
				if title, ok := expMap["title"].(string); ok {
					markdown.WriteString(title)
				}
				markdown.WriteString("\n\n")

				if company, ok := expMap["company"].(string); ok {
					markdown.WriteString(fmt.Sprintf("**公司：** %s\n", company))
				}
				if duration, ok := expMap["duration"].(string); ok {
					markdown.WriteString(fmt.Sprintf("**时间：** %s\n", duration))
				}
				if description, ok := expMap["description"].(string); ok {
					markdown.WriteString(fmt.Sprintf("**描述：** %s\n", description))
				}
				markdown.WriteString("\n")
			}
		}
	}

	// 教育背景
	if educations, ok := contentData["education"].([]interface{}); ok {
		markdown.WriteString("## 教育背景\n\n")
		for _, edu := range educations {
			if eduMap, ok := edu.(map[string]interface{}); ok {
				if school, ok := eduMap["school"].(string); ok {
					markdown.WriteString(fmt.Sprintf("- **%s**", school))
				}
				if degree, ok := eduMap["degree"].(string); ok {
					markdown.WriteString(fmt.Sprintf(" - %s", degree))
				}
				if major, ok := eduMap["major"].(string); ok {
					markdown.WriteString(fmt.Sprintf(" - %s", major))
				}
				markdown.WriteString("\n")
			}
		}
		markdown.WriteString("\n")
	}

	// 技能
	if skills, ok := contentData["skills"].([]interface{}); ok {
		markdown.WriteString("## 技能\n\n")
		for _, skill := range skills {
			if skillMap, ok := skill.(map[string]interface{}); ok {
				if name, ok := skillMap["name"].(string); ok {
					markdown.WriteString(fmt.Sprintf("- %s", name))
					if level, ok := skillMap["level"].(string); ok {
						markdown.WriteString(fmt.Sprintf(" (%s)", level))
					}
					markdown.WriteString("\n")
				}
			}
		}
		markdown.WriteString("\n")
	}

	return markdown.String()
}

// 生成 URL 友好的 slug
func (m *Migrator) generateSlug(title string) string {
	// 简单的 slug 生成逻辑
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "简历", "resume")
	slug = strings.ReplaceAll(slug, "工程师", "engineer")
	slug = strings.ReplaceAll(slug, "开发", "developer")
	return slug
}

// 迁移简历技能
func (m *Migrator) migrateResumeSkills(resumeID int64, contentData map[string]interface{}) error {
	if skills, ok := contentData["skills"].([]interface{}); ok {
		for _, skill := range skills {
			if skillMap, ok := skill.(map[string]interface{}); ok {
				if name, ok := skillMap["name"].(string); ok {
					// 查找技能 ID
					var skillID int64
					err := m.targetDB.QueryRow("SELECT id FROM skills WHERE name = ?", name).Scan(&skillID)
					if err != nil {
						// 如果技能不存在，创建一个
						result, err := m.targetDB.Exec(`
							INSERT INTO skills (name, category, description, is_popular, search_count, created_at, updated_at)
							VALUES (?, ?, ?, ?, ?, ?, ?)
						`, name, "其他", fmt.Sprintf("%s 相关技能", name), false, 0, time.Now(), time.Now())
						if err != nil {
							continue
						}
						skillID, _ = result.LastInsertId()
					}

					// 确定熟练度
					level := "intermediate"
					if levelStr, ok := skillMap["level"].(string); ok {
						switch strings.ToLower(levelStr) {
						case "beginner", "初级":
							level = "beginner"
						case "intermediate", "中级":
							level = "intermediate"
						case "advanced", "高级":
							level = "advanced"
						case "expert", "专家":
							level = "expert"
						}
					}

					// 插入简历技能关联
					_, err = m.targetDB.Exec(`
						INSERT INTO resume_skills (resume_id, skill_id, proficiency_level, years_of_experience, is_highlighted, created_at)
						VALUES (?, ?, ?, ?, ?, ?)
					`, resumeID, skillID, level, 1.0, false, time.Now())
					if err != nil {
						log.Printf("插入简历技能失败: %v", err)
					}
				}
			}
		}
	}
	return nil
}

// 迁移工作经历
func (m *Migrator) migrateWorkExperiences(resumeID int64, contentData map[string]interface{}) error {
	if experiences, ok := contentData["experience"].([]interface{}); ok {
		for _, exp := range experiences {
			if expMap, ok := exp.(map[string]interface{}); ok {
				title := ""
				if titleStr, ok := expMap["title"].(string); ok {
					title = titleStr
				}

				// company := ""
				// if companyStr, ok := expMap["company"].(string); ok {
				// 	company = companyStr
				// }

				description := ""
				if descStr, ok := expMap["description"].(string); ok {
					description = descStr
				}

				duration := ""
				if durationStr, ok := expMap["duration"].(string); ok {
					duration = durationStr
				}

				// 解析时间范围
				startDate, endDate := m.parseDuration(duration)

				// 插入工作经历
				_, err := m.targetDB.Exec(`
					INSERT INTO work_experiences (resume_id, title, start_date, end_date, is_current, description, created_at, updated_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?)
				`, resumeID, title, startDate, endDate, endDate == "", description, time.Now(), time.Now())
				if err != nil {
					log.Printf("插入工作经历失败: %v", err)
				}
			}
		}
	}
	return nil
}

// 迁移教育背景
func (m *Migrator) migrateEducations(resumeID int64, contentData map[string]interface{}) error {
	if educations, ok := contentData["education"].([]interface{}); ok {
		for _, edu := range educations {
			if eduMap, ok := edu.(map[string]interface{}); ok {
				school := ""
				if schoolStr, ok := eduMap["school"].(string); ok {
					school = schoolStr
				}

				degree := ""
				if degreeStr, ok := eduMap["degree"].(string); ok {
					degree = degreeStr
				}

				major := ""
				if majorStr, ok := eduMap["major"].(string); ok {
					major = majorStr
				}

				// 插入教育背景
				_, err := m.targetDB.Exec(`
					INSERT INTO educations (resume_id, school, degree, major, is_highlighted, created_at, updated_at)
					VALUES (?, ?, ?, ?, ?, ?, ?)
				`, resumeID, school, degree, major, false, time.Now(), time.Now())
				if err != nil {
					log.Printf("插入教育背景失败: %v", err)
				}
			}
		}
	}
	return nil
}

// 迁移项目经验
func (m *Migrator) migrateProjects(resumeID int64, contentData map[string]interface{}) error {
	if projects, ok := contentData["projects"].([]interface{}); ok {
		for _, proj := range projects {
			if projMap, ok := proj.(map[string]interface{}); ok {
				name := ""
				if nameStr, ok := projMap["name"].(string); ok {
					name = nameStr
				}

				description := ""
				if descStr, ok := projMap["description"].(string); ok {
					description = descStr
				}

				technologies := ""
				if techStr, ok := projMap["technologies"].(string); ok {
					technologies = techStr
				}

				// 插入项目
				_, err := m.targetDB.Exec(`
					INSERT INTO projects (resume_id, name, description, technology_stack, status, is_highlighted, created_at, updated_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?)
				`, resumeID, name, description, technologies, "completed", false, time.Now(), time.Now())
				if err != nil {
					log.Printf("插入项目失败: %v", err)
				}
			}
		}
	}
	return nil
}

// 迁移证书
func (m *Migrator) migrateCertifications(resumeID int64, contentData map[string]interface{}) error {
	if certifications, ok := contentData["certifications"].([]interface{}); ok {
		for _, cert := range certifications {
			if certMap, ok := cert.(map[string]interface{}); ok {
				name := ""
				if nameStr, ok := certMap["name"].(string); ok {
					name = nameStr
				}

				issuer := ""
				if issuerStr, ok := certMap["issuer"].(string); ok {
					issuer = issuerStr
				}

				date := ""
				if dateStr, ok := certMap["date"].(string); ok {
					date = dateStr
				}

				// 插入证书
				_, err := m.targetDB.Exec(`
					INSERT INTO certifications (resume_id, name, issuer, issue_date, is_highlighted, created_at, updated_at)
					VALUES (?, ?, ?, ?, ?, ?, ?)
				`, resumeID, name, issuer, date, false, time.Now(), time.Now())
				if err != nil {
					log.Printf("插入证书失败: %v", err)
				}
			}
		}
	}
	return nil
}

// 解析时间范围
func (m *Migrator) parseDuration(duration string) (string, string) {
	// 简单的持续时间解析逻辑
	// 例如: "2020-2024", "2020年1月-2024年12月"
	if strings.Contains(duration, "-") {
		parts := strings.Split(duration, "-")
		if len(parts) == 2 {
			start := strings.TrimSpace(parts[0])
			end := strings.TrimSpace(parts[1])

			// 转换为标准日期格式
			startDate := m.normalizeDate(start)
			endDate := m.normalizeDate(end)

			return startDate, endDate
		}
	}
	return "", ""
}

// 标准化日期格式
func (m *Migrator) normalizeDate(dateStr string) string {
	// 简单的日期标准化逻辑
	// 例如: "2020" -> "2020-01-01", "2020年1月" -> "2020-01-01"
	dateStr = strings.TrimSpace(dateStr)

	// 移除中文
	dateStr = strings.ReplaceAll(dateStr, "年", "-")
	dateStr = strings.ReplaceAll(dateStr, "月", "")
	dateStr = strings.ReplaceAll(dateStr, "日", "")

	// 如果只有年份，添加月份和日期
	if len(dateStr) == 4 {
		return dateStr + "-01-01"
	}

	return dateStr
}

// 安全获取密码
func getPassword(prompt string) string {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("读取密码失败: %v", err)
	}
	fmt.Println() // 换行
	return strings.TrimSpace(string(bytePassword))
}

// 主函数
func main() {
	log.Println("开始数据迁移...")

	// 从环境变量获取密码，如果没有则交互式获取
	sourcePassword := os.Getenv("SOURCE_DB_PASSWORD")
	if sourcePassword == "" {
		sourcePassword = getPassword("请输入源数据库密码: ")
	}

	targetPassword := os.Getenv("TARGET_DB_PASSWORD")
	if targetPassword == "" {
		targetPassword = getPassword("请输入目标数据库密码: ")
	}

	// 数据库配置
	sourceConfig := DBConfig{
		Host:     "localhost",
		Port:     "3306",
		User:     "root",
		Password: sourcePassword,
		Database: "jobfirst",
	}

	targetConfig := DBConfig{
		Host:     "localhost",
		Port:     "3306",
		User:     "root",
		Password: targetPassword,
		Database: "jobfirst_v3",
	}

	// 创建迁移器
	migrator, err := NewMigrator(sourceConfig, targetConfig)
	if err != nil {
		log.Fatalf("创建迁移器失败: %v", err)
	}
	defer migrator.Close()

	// 执行迁移
	if err := migrator.MigrateSkills(); err != nil {
		log.Printf("迁移技能数据失败: %v", err)
	}

	if err := migrator.MigrateResumes(); err != nil {
		log.Printf("迁移简历数据失败: %v", err)
	}

	log.Println("数据迁移完成！")
}
