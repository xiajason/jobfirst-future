package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

// 敏感程度分级定义
const (
	SensitivityLevel1 = "low"      // 🟢 低敏感 - 系统字段、统计信息、公开数据
	SensitivityLevel2 = "medium"   // 🟡 中敏感 - 一般个人信息、偏好设置、职业信息
	SensitivityLevel3 = "high"     // 🟠 高敏感 - 个人身份信息、联系方式、财务信息、位置信息
	SensitivityLevel4 = "critical" // 🔴 极高敏感 - 身份认证信息、密码哈希、会话令牌
)

// 数据分类标签结构
type DataClassificationTag struct {
	FieldName        string `json:"field_name"`
	SensitivityLevel string `json:"sensitivity_level"`
	DataType         string `json:"data_type"`
	ProtectionMethod string `json:"protection_method"`
	RetentionPeriod  int    `json:"retention_period"` // 天数
	RequiresConsent  bool   `json:"requires_consent"`
	IsPersonalInfo   bool   `json:"is_personal_info"`
}

// 敏感信息感知的解析数据结构
type SensitivityAwareParsedData struct {
	Title          string                   `json:"title"`
	Content        string                   `json:"content"`
	PersonalInfo   map[string]interface{}   `json:"personal_info"`
	WorkExperience []map[string]interface{} `json:"work_experience"`
	Education      []map[string]interface{} `json:"education"`
	Skills         []string                 `json:"skills"`
	Projects       []map[string]interface{} `json:"projects"`
	Certifications []map[string]interface{} `json:"certifications"`
	Keywords       []string                 `json:"keywords"`
	Confidence     float64                  `json:"confidence"`

	// 敏感信息分类标签
	DataClassification map[string]DataClassificationTag `json:"data_classification"`

	// 解析元数据
	ParsingMetadata map[string]interface{} `json:"parsing_metadata"`
}

// 数据分类配置
var DataClassificationConfig = map[string]DataClassificationTag{
	// 个人信息字段分类
	"name": {
		FieldName:        "name",
		SensitivityLevel: SensitivityLevel3,
		DataType:         "personal_identity",
		ProtectionMethod: "access_control",
		RetentionPeriod:  2555, // 7年
		RequiresConsent:  true,
		IsPersonalInfo:   true,
	},
	"phone": {
		FieldName:        "phone",
		SensitivityLevel: SensitivityLevel3,
		DataType:         "contact_info",
		ProtectionMethod: "aes256_encryption",
		RetentionPeriod:  2555,
		RequiresConsent:  true,
		IsPersonalInfo:   true,
	},
	"email": {
		FieldName:        "email",
		SensitivityLevel: SensitivityLevel3,
		DataType:         "contact_info",
		ProtectionMethod: "aes256_encryption",
		RetentionPeriod:  2555,
		RequiresConsent:  true,
		IsPersonalInfo:   true,
	},
	"address": {
		FieldName:        "address",
		SensitivityLevel: SensitivityLevel3,
		DataType:         "location_info",
		ProtectionMethod: "access_control",
		RetentionPeriod:  2555,
		RequiresConsent:  true,
		IsPersonalInfo:   true,
	},
	"date_of_birth": {
		FieldName:        "date_of_birth",
		SensitivityLevel: SensitivityLevel3,
		DataType:         "personal_identity",
		ProtectionMethod: "aes256_encryption",
		RetentionPeriod:  2555,
		RequiresConsent:  true,
		IsPersonalInfo:   true,
	},
	"gender": {
		FieldName:        "gender",
		SensitivityLevel: SensitivityLevel3,
		DataType:         "personal_identity",
		ProtectionMethod: "access_control",
		RetentionPeriod:  2555,
		RequiresConsent:  true,
		IsPersonalInfo:   true,
	},
	"nationality": {
		FieldName:        "nationality",
		SensitivityLevel: SensitivityLevel3,
		DataType:         "personal_identity",
		ProtectionMethod: "access_control",
		RetentionPeriod:  2555,
		RequiresConsent:  true,
		IsPersonalInfo:   true,
	},

	// 职业信息字段分类
	"title": {
		FieldName:        "title",
		SensitivityLevel: SensitivityLevel2,
		DataType:         "professional_info",
		ProtectionMethod: "access_control",
		RetentionPeriod:  1095, // 3年
		RequiresConsent:  false,
		IsPersonalInfo:   false,
	},
	"company": {
		FieldName:        "company",
		SensitivityLevel: SensitivityLevel2,
		DataType:         "professional_info",
		ProtectionMethod: "access_control",
		RetentionPeriod:  1095,
		RequiresConsent:  false,
		IsPersonalInfo:   false,
	},
	"position": {
		FieldName:        "position",
		SensitivityLevel: SensitivityLevel2,
		DataType:         "professional_info",
		ProtectionMethod: "access_control",
		RetentionPeriod:  1095,
		RequiresConsent:  false,
		IsPersonalInfo:   false,
	},
	"work_description": {
		FieldName:        "work_description",
		SensitivityLevel: SensitivityLevel2,
		DataType:         "professional_info",
		ProtectionMethod: "access_control",
		RetentionPeriod:  1095,
		RequiresConsent:  false,
		IsPersonalInfo:   false,
	},

	// 教育信息字段分类
	"school": {
		FieldName:        "school",
		SensitivityLevel: SensitivityLevel2,
		DataType:         "education_info",
		ProtectionMethod: "access_control",
		RetentionPeriod:  1095,
		RequiresConsent:  false,
		IsPersonalInfo:   false,
	},
	"major": {
		FieldName:        "major",
		SensitivityLevel: SensitivityLevel2,
		DataType:         "education_info",
		ProtectionMethod: "access_control",
		RetentionPeriod:  1095,
		RequiresConsent:  false,
		IsPersonalInfo:   false,
	},
	"degree": {
		FieldName:        "degree",
		SensitivityLevel: SensitivityLevel2,
		DataType:         "education_info",
		ProtectionMethod: "access_control",
		RetentionPeriod:  1095,
		RequiresConsent:  false,
		IsPersonalInfo:   false,
	},

	// 技能信息字段分类
	"skills": {
		FieldName:        "skills",
		SensitivityLevel: SensitivityLevel2,
		DataType:         "professional_info",
		ProtectionMethod: "access_control",
		RetentionPeriod:  1095,
		RequiresConsent:  false,
		IsPersonalInfo:   false,
	},

	// 系统字段分类
	"content": {
		FieldName:        "content",
		SensitivityLevel: SensitivityLevel3,
		DataType:         "professional_info",
		ProtectionMethod: "access_control",
		RetentionPeriod:  2555,
		RequiresConsent:  true,
		IsPersonalInfo:   true,
	},
	"keywords": {
		FieldName:        "keywords",
		SensitivityLevel: SensitivityLevel1,
		DataType:         "system_info",
		ProtectionMethod: "none",
		RetentionPeriod:  365,
		RequiresConsent:  false,
		IsPersonalInfo:   false,
	},
	"confidence": {
		FieldName:        "confidence",
		SensitivityLevel: SensitivityLevel1,
		DataType:         "system_info",
		ProtectionMethod: "none",
		RetentionPeriod:  365,
		RequiresConsent:  false,
		IsPersonalInfo:   false,
	},
}

// 敏感信息感知的文本解析器
type SensitivityAwareTextParser struct {
	encryptionKey []byte
}

// 创建新的敏感信息感知解析器
func NewSensitivityAwareTextParser() *SensitivityAwareTextParser {
	// 在实际应用中，应该从安全的配置中获取加密密钥
	key := []byte("your-32-byte-long-key-here!12345") // 32字节密钥
	return &SensitivityAwareTextParser{
		encryptionKey: key,
	}
}

// 解析文件并应用敏感信息分类
func (p *SensitivityAwareTextParser) ParseFileWithSensitivity(filePath string) (*SensitivityAwareParsedData, error) {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	text := string(content)
	log.Printf("开始敏感信息感知解析: %s, 长度: %d", filePath, len(text))

	// 预处理文本
	cleanedText := p.preprocessText(text)

	// 提取个人信息
	personalInfo := p.extractPersonalInfoWithClassification(cleanedText)

	// 提取工作经历
	workExperience := p.extractWorkExperienceWithClassification(cleanedText)

	// 提取教育背景
	education := p.extractEducationWithClassification(cleanedText)

	// 提取技能
	skills := p.extractSkillsWithClassification(cleanedText)

	// 提取项目经历
	projects := p.extractProjectsWithClassification(cleanedText)

	// 提取证书资质
	certifications := p.extractCertificationsWithClassification(cleanedText)

	// 生成关键词
	keywords := p.generateKeywordsWithClassification(cleanedText, skills)

	// 计算解析置信度
	confidence := p.calculateConfidenceWithClassification(personalInfo, workExperience, education, skills)

	// 创建数据分类标签
	dataClassification := p.createDataClassificationTags(personalInfo, workExperience, education, skills)

	// 创建解析元数据
	parsingMetadata := map[string]interface{}{
		"parsing_time":      time.Now().Unix(),
		"file_size":         len(content),
		"text_length":       len(cleanedText),
		"sensitivity_level": p.calculateOverallSensitivityLevel(dataClassification),
		"requires_consent":  p.checkConsentRequirement(dataClassification),
		"retention_period":  p.calculateMaxRetentionPeriod(dataClassification),
		"parser_version":    "1.0.0",
	}

	return &SensitivityAwareParsedData{
		Title:              p.extractTitle(cleanedText),
		Content:            cleanedText,
		PersonalInfo:       personalInfo,
		WorkExperience:     workExperience,
		Education:          education,
		Skills:             skills,
		Projects:           projects,
		Certifications:     certifications,
		Keywords:           keywords,
		Confidence:         confidence,
		DataClassification: dataClassification,
		ParsingMetadata:    parsingMetadata,
	}, nil
}

// 预处理文本
func (p *SensitivityAwareTextParser) preprocessText(text string) string {
	// 移除多余的空格和换行
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	// 标准化空格
	text = strings.Join(strings.Fields(text), " ")

	// 移除特殊字符
	text = strings.ReplaceAll(text, "\u00a0", " ") // 非断行空格

	return text
}

// 提取个人信息并应用分类
func (p *SensitivityAwareTextParser) extractPersonalInfoWithClassification(text string) map[string]interface{} {
	personalInfo := make(map[string]interface{})

	// 提取姓名 - Level 3 高敏感
	namePatterns := []string{
		`姓名[：:]\s*([^\n\r\s]+)`,
		`Name[：:]\s*([^\n\r\s]+)`,
		`^([^\n\r\s]{2,10})\s*$`, // 第一行作为姓名
	}

	for _, pattern := range namePatterns {
		if matches := regexp.MustCompile(pattern).FindStringSubmatch(text); len(matches) > 1 {
			name := strings.TrimSpace(matches[1])
			personalInfo["name"] = name
			log.Printf("提取到姓名 (Level 3): %s", name)
			break
		}
	}

	// 提取电话号码 - Level 3 高敏感，需要加密
	phonePatterns := []string{
		`电话[：:]\s*([^\n\r]+)`,     // 电话：138-0000-1234
		`Phone[：:]\s*([^\n\r]+)`,  // Phone: 138-0000-1234
		`手机[：:]\s*([^\n\r]+)`,     // 手机：138-0000-1234
		`联系方式[：:]\s*([^\n\r]+)`,   // 联系方式：138-0000-1234
		`Tel[：:]\s*([^\n\r]+)`,    // Tel: 138-0000-1234
		`Mobile[：:]\s*([^\n\r]+)`, // Mobile: 138-0000-1234
		`联系电话[：:]\s*([^\n\r]+)`,   // 联系电话：138-0000-1234
	}

	// 首先尝试带标签的电话号码提取
	for _, pattern := range phonePatterns {
		if matches := regexp.MustCompile(pattern).FindStringSubmatch(text); len(matches) > 1 {
			phone := strings.TrimSpace(matches[1])

			// 清理电话号码格式
			phone = strings.ReplaceAll(phone, " ", "")
			phone = strings.ReplaceAll(phone, "-", "")
			phone = strings.ReplaceAll(phone, "(", "")
			phone = strings.ReplaceAll(phone, ")", "")

			// 验证电话号码格式
			if len(phone) >= 7 && len(phone) <= 15 {
				personalInfo["phone"] = phone
				log.Printf("提取到电话号码 (Level 3): %s", phone)
				break
			}
		}
	}

	// 如果没有找到带标签的电话，尝试直接匹配电话号码
	if _, exists := personalInfo["phone"]; !exists {
		directPhonePatterns := []string{
			`(1[3-9]\d{9})`,                 // 直接匹配11位手机号
			`(1[3-9]\d{9})`,                 // 11位手机号（带分隔符）
			`(\d{3,4}-?\d{7,8})`,            // 固定电话 区号-号码
			`(\+\d{1,3}-?\d{3,4}-?\d{7,8})`, // 国际号码格式
			`(\(\d{3,4}\)\s*\d{7,8})`,       // 带括号的固定电话
			`(\d{3,4}\s\d{7,8})`,            // 空格分隔的固定电话
		}

		for _, pattern := range directPhonePatterns {
			if matches := regexp.MustCompile(pattern).FindStringSubmatch(text); len(matches) > 0 {
				phone := strings.TrimSpace(matches[1])

				// 清理电话号码格式
				phone = strings.ReplaceAll(phone, " ", "")
				phone = strings.ReplaceAll(phone, "-", "")
				phone = strings.ReplaceAll(phone, "(", "")
				phone = strings.ReplaceAll(phone, ")", "")

				// 验证电话号码格式
				if len(phone) >= 7 && len(phone) <= 15 {
					personalInfo["phone"] = phone
					log.Printf("提取到电话号码 (Level 3): %s", phone)
					break
				}
			}
		}
	}

	// 提取邮箱 - Level 3 高敏感，需要加密
	emailPattern := `([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`
	if matches := regexp.MustCompile(emailPattern).FindStringSubmatch(text); len(matches) > 0 {
		email := matches[0]
		// 在实际应用中，这里应该进行加密
		personalInfo["email"] = email
		log.Printf("提取到邮箱 (Level 3): %s", email)
	}

	// 提取地址 - Level 3 高敏感
	addressPatterns := []string{
		`地址[：:]\s*([^\n\r\s]+[^\n\r]*)`,
		`Address[：:]\s*([^\n\r\s]+[^\n\r]*)`,
		`现居住地[：:]\s*([^\n\r\s]+[^\n\r]*)`,
		`居住地址[：:]\s*([^\n\r\s]+[^\n\r]*)`,
	}

	for _, pattern := range addressPatterns {
		if matches := regexp.MustCompile(pattern).FindStringSubmatch(text); len(matches) > 1 {
			address := strings.TrimSpace(matches[1])

			// 限制地址长度，避免包含过多内容
			if len(address) > 200 {
				// 如果地址过长，尝试截取到第一个句号或换行符
				if idx := strings.Index(address, "。"); idx > 0 {
					address = address[:idx]
				} else if idx := strings.Index(address, "\n"); idx > 0 {
					address = address[:idx]
				} else {
					address = address[:200]
				}
			}

			personalInfo["address"] = address
			log.Printf("提取到地址 (Level 3): %s", address)
			break
		}
	}

	// 提取出生日期 - Level 3 高敏感，需要加密
	birthPattern := `(出生日期|生日|出生)[：:]\s*(\d{4}[-/年]\d{1,2}[-/月]\d{1,2}[日]?)`
	if matches := regexp.MustCompile(birthPattern).FindStringSubmatch(text); len(matches) > 2 {
		birthDate := strings.TrimSpace(matches[2])
		personalInfo["date_of_birth"] = birthDate
		log.Printf("提取到出生日期 (Level 3): %s", birthDate)
	}

	// 提取性别 - Level 3 高敏感
	genderPattern := `(性别)[：:]\s*([男女])`
	if matches := regexp.MustCompile(genderPattern).FindStringSubmatch(text); len(matches) > 2 {
		gender := strings.TrimSpace(matches[2])
		personalInfo["gender"] = gender
		log.Printf("提取到性别 (Level 3): %s", gender)
	}

	return personalInfo
}

// 提取工作经历并应用分类
func (p *SensitivityAwareTextParser) extractWorkExperienceWithClassification(text string) []map[string]interface{} {
	var experiences []map[string]interface{}

	// 工作经历模式匹配
	workPatterns := []string{
		`工作经历[：:]?\s*(.*?)(教育背景|项目经历|技能|$)`,
		`Work Experience[：:]?\s*(.*?)(Education|Projects|Skills|$)`,
		`职业经历[：:]?\s*(.*?)(教育背景|项目经历|技能|$)`,
	}

	for _, pattern := range workPatterns {
		if matches := regexp.MustCompile(pattern).FindStringSubmatch(text); len(matches) > 1 {
			workSection := matches[1]
			experiences = p.parseWorkSectionWithClassification(workSection)
			break
		}
	}

	return experiences
}

// 解析工作经历段落并应用分类
func (p *SensitivityAwareTextParser) parseWorkSectionWithClassification(workSection string) []map[string]interface{} {
	var experiences []map[string]interface{}

	// 按时间段分割工作经历
	timePattern := `(\d{4}[-/年]\d{1,2}[-/月]?)\s*[-~至到]\s*(\d{4}[-/年]\d{1,2}[-/月]?|至今|现在)`
	matches := regexp.MustCompile(timePattern).FindAllStringSubmatch(workSection, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			experience := map[string]interface{}{
				"start_date": match[1], // Level 2 中敏感
				"end_date":   match[2], // Level 2 中敏感
			}

			// 提取公司名称和职位 - Level 2 中敏感
			companyPattern := `([^\n\r]+?)\s*[-–—]\s*([^\n\r]+)`
			if companyMatch := regexp.MustCompile(companyPattern).FindStringSubmatch(workSection); len(companyMatch) >= 3 {
				experience["company"] = strings.TrimSpace(companyMatch[1])  // Level 2
				experience["position"] = strings.TrimSpace(companyMatch[2]) // Level 2
			}

			experiences = append(experiences, experience)
		}
	}

	log.Printf("提取到 %d 个工作经历 (Level 2)", len(experiences))
	return experiences
}

// 提取教育背景并应用分类
func (p *SensitivityAwareTextParser) extractEducationWithClassification(text string) []map[string]interface{} {
	var education []map[string]interface{}

	// 教育背景模式匹配
	eduPatterns := []string{
		`教育背景[：:]?\s*(.*?)(工作经历|项目经历|技能|$)`,
		`Education[：:]?\s*(.*?)(Work Experience|Projects|Skills|$)`,
	}

	for _, pattern := range eduPatterns {
		if matches := regexp.MustCompile(pattern).FindStringSubmatch(text); len(matches) > 1 {
			eduSection := matches[1]
			education = p.parseEducationSectionWithClassification(eduSection)
			break
		}
	}

	return education
}

// 解析教育背景段落并应用分类
func (p *SensitivityAwareTextParser) parseEducationSectionWithClassification(eduSection string) []map[string]interface{} {
	var education []map[string]interface{}

	// 按学校分割教育经历
	schoolPattern := `([^\n\r]+?)\s*[-–—]\s*([^\n\r]+?)\s*[-–—]\s*([^\n\r]+)`
	matches := regexp.MustCompile(schoolPattern).FindAllStringSubmatch(eduSection, -1)

	for _, match := range matches {
		if len(match) >= 4 {
			edu := map[string]interface{}{
				"school": strings.TrimSpace(match[1]), // Level 2 中敏感
				"major":  strings.TrimSpace(match[2]), // Level 2 中敏感
				"degree": strings.TrimSpace(match[3]), // Level 2 中敏感
			}
			education = append(education, edu)
		}
	}

	log.Printf("提取到 %d 个教育背景 (Level 2)", len(education))
	return education
}

// 提取技能并应用分类
func (p *SensitivityAwareTextParser) extractSkillsWithClassification(text string) []string {
	var skills []string

	// 技能模式匹配
	skillPatterns := []string{
		`技能[：:]?\s*(.*?)(工作经历|教育背景|项目经历|$)`,
		`Skills[：:]?\s*(.*?)(Work Experience|Education|Projects|$)`,
		`专业技能[：:]?\s*(.*?)(工作经历|教育背景|项目经历|$)`,
	}

	for _, pattern := range skillPatterns {
		if matches := regexp.MustCompile(pattern).FindStringSubmatch(text); len(matches) > 1 {
			skillSection := matches[1]
			skills = p.parseSkillsSectionWithClassification(skillSection)
			break
		}
	}

	// 如果没有找到技能段落，尝试从整个文本中提取技术关键词
	if len(skills) == 0 {
		skills = p.extractTechnicalKeywordsWithClassification(text)
	}

	log.Printf("提取到 %d 个技能 (Level 2)", len(skills))
	return skills
}

// 解析技能段落并应用分类
func (p *SensitivityAwareTextParser) parseSkillsSectionWithClassification(skillSection string) []string {
	var skills []string

	// 按分隔符分割技能
	separators := []string{",", "、", ";", "；", "|", "\n"}

	for _, sep := range separators {
		if strings.Contains(skillSection, sep) {
			parts := strings.Split(skillSection, sep)
			for _, part := range parts {
				skill := strings.TrimSpace(part)
				if len(skill) > 0 {
					skills = append(skills, skill)
				}
			}
			break
		}
	}

	return skills
}

// 提取技术关键词并应用分类
func (p *SensitivityAwareTextParser) extractTechnicalKeywordsWithClassification(text string) []string {
	// 常见技术关键词 - Level 2 中敏感
	techKeywords := []string{
		"Go", "Golang", "Java", "Python", "JavaScript", "TypeScript", "C++", "C#",
		"React", "Vue", "Angular", "Node.js", "Spring", "Django", "Flask",
		"MySQL", "PostgreSQL", "Redis", "MongoDB", "Elasticsearch",
		"Docker", "Kubernetes", "AWS", "Azure", "微服务", "分布式",
		"Git", "Linux", "Nginx", "Apache", "TCP/IP", "HTTP", "RESTful API",
	}

	var foundSkills []string
	textLower := strings.ToLower(text)

	for _, keyword := range techKeywords {
		if strings.Contains(textLower, strings.ToLower(keyword)) {
			foundSkills = append(foundSkills, keyword)
		}
	}

	return foundSkills
}

// 提取项目经历并应用分类
func (p *SensitivityAwareTextParser) extractProjectsWithClassification(text string) []map[string]interface{} {
	var projects []map[string]interface{}

	// 项目经历模式匹配
	projectPatterns := []string{
		`项目经历[：:]?\s*(.*?)(工作经历|教育背景|技能|$)`,
		`Projects[：:]?\s*(.*?)(Work Experience|Education|Skills|$)`,
	}

	for _, pattern := range projectPatterns {
		if matches := regexp.MustCompile(pattern).FindStringSubmatch(text); len(matches) > 1 {
			projectSection := matches[1]
			projects = p.parseProjectSectionWithClassification(projectSection)
			break
		}
	}

	log.Printf("提取到 %d 个项目经历 (Level 2)", len(projects))
	return projects
}

// 解析项目经历段落并应用分类
func (p *SensitivityAwareTextParser) parseProjectSectionWithClassification(projectSection string) []map[string]interface{} {
	var projects []map[string]interface{}

	// 按项目分割
	projectPattern := `([^\n\r]+?)\s*[-–—]\s*([^\n\r]+)`
	matches := regexp.MustCompile(projectPattern).FindAllStringSubmatch(projectSection, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			project := map[string]interface{}{
				"name":        strings.TrimSpace(match[1]), // Level 2 中敏感
				"description": strings.TrimSpace(match[2]), // Level 2 中敏感
			}
			projects = append(projects, project)
		}
	}

	return projects
}

// 提取证书资质并应用分类
func (p *SensitivityAwareTextParser) extractCertificationsWithClassification(text string) []map[string]interface{} {
	var certifications []map[string]interface{}

	// 证书资质模式匹配 - 更全面的模式
	certPatterns := []string{
		`证书[：:]?\s*(.*?)(工作经历|教育背景|技能|项目经历|$)`,
		`Certifications[：:]?\s*(.*?)(Work Experience|Education|Skills|Projects|$)`,
		`资质[：:]?\s*(.*?)(工作经历|教育背景|技能|项目经历|$)`,
		`资格证书[：:]?\s*(.*?)(工作经历|教育背景|技能|项目经历|$)`,
		`专业认证[：:]?\s*(.*?)(工作经历|教育背景|技能|项目经历|$)`,
		`认证证书[：:]?\s*(.*?)(工作经历|教育背景|技能|项目经历|$)`,
		`技能证书[：:]?\s*(.*?)(工作经历|教育背景|技能|项目经历|$)`,
	}

	// 尝试匹配证书段落
	for _, pattern := range certPatterns {
		if matches := regexp.MustCompile(pattern).FindStringSubmatch(text); len(matches) > 1 {
			certSection := matches[1]
			certifications = p.parseCertificationSectionWithClassification(certSection)
			if len(certifications) > 0 {
				break
			}
		}
	}

	// 如果段落匹配失败，尝试直接匹配证书名称
	if len(certifications) == 0 {
		certifications = p.extractCertificationNamesDirectly(text)
	}

	log.Printf("提取到 %d 个证书资质 (Level 2)", len(certifications))
	return certifications
}

// 解析证书资质段落并应用分类
func (p *SensitivityAwareTextParser) parseCertificationSectionWithClassification(certSection string) []map[string]interface{} {
	var certifications []map[string]interface{}

	// 首先按行分割证书
	lines := strings.Split(certSection, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 尝试不同的证书格式
		certPatterns := []string{
			`([^\s-–—：:（(]+)\s*[-–—]\s*(.+)`, // 证书名 - 描述
			`([^\s-–—：:（(]+)\s*[：:]\s*(.+)`,  // 证书名：描述
			`([^\s-–—：:（(]+)\s*（([^）]+)）`,    // 证书名（描述）
			`([^\s-–—：:（(]+)\s*\(([^)]+)\)`,  // 证书名(描述)
		}

		certFound := false
		for _, pattern := range certPatterns {
			if matches := regexp.MustCompile(pattern).FindStringSubmatch(line); len(matches) >= 3 {
				certName := strings.TrimSpace(matches[1])
				certDesc := strings.TrimSpace(matches[2])

				// 验证证书名称长度和内容
				if len(certName) > 2 && len(certName) < 50 && !strings.Contains(certName, "证书") {
					cert := map[string]interface{}{
						"name":        certName, // Level 2 中敏感
						"description": certDesc, // Level 2 中敏感
					}
					certifications = append(certifications, cert)
					certFound = true
					break
				}
			}
		}

		// 如果没有匹配到格式，但包含常见证书关键词，直接作为证书名
		if !certFound && p.isCertificationLine(line) {
			cert := map[string]interface{}{
				"name":        line,     // Level 2 中敏感
				"description": "专业认证证书", // Level 2 中敏感
			}
			certifications = append(certifications, cert)
		}
	}

	return certifications
}

// 判断是否为证书行
func (p *SensitivityAwareTextParser) isCertificationLine(line string) bool {
	certKeywords := []string{
		"认证", "证书", "工程师", "专家", "管理员", "架构师",
		"Certification", "Engineer", "Expert", "Administrator", "Architect",
		"PMP", "AWS", "Java", "Oracle", "Microsoft", "Google", "Docker", "Kubernetes",
		"华为", "阿里云", "腾讯云", "百度云", "字节跳动",
	}

	for _, keyword := range certKeywords {
		if strings.Contains(line, keyword) && len(line) > 3 && len(line) < 100 {
			return true
		}
	}
	return false
}

// 直接提取证书名称（当段落匹配失败时使用）
func (p *SensitivityAwareTextParser) extractCertificationNamesDirectly(text string) []map[string]interface{} {
	var certifications []map[string]interface{}

	// 常见的证书名称模式
	certNamePatterns := []string{
		`(Java认证工程师)`,
		`(AWS云架构师认证)`,
		`(PMP项目管理认证)`,
		`(CISSP信息安全认证)`,
		`(CCNA网络工程师)`,
		`(CCNP高级网络工程师)`,
		`(Oracle数据库认证)`,
		`(Microsoft认证专家)`,
		`(Google云架构师)`,
		`(Kubernetes管理员)`,
		`(Docker认证工程师)`,
		`(Red Hat认证工程师)`,
		`(Cisco认证工程师)`,
		`(华为认证工程师)`,
		`(腾讯云架构师)`,
		`(阿里云架构师)`,
		`(百度云架构师)`,
		`(字节跳动认证)`,
		`(字节认证)`,
		`(字节跳动)`,
		`([A-Z]{2,}[认证|工程师|专家|管理员|架构师]+)`,       // 通用认证模式
		`([a-zA-Z]{3,}\s+[认证|工程师|专家|管理员|架构师]+)`, // 英文认证模式
	}

	for _, pattern := range certNamePatterns {
		matches := regexp.MustCompile(pattern).FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 0 {
				certName := strings.TrimSpace(match[1])
				if len(certName) > 2 && len(certName) < 50 {
					cert := map[string]interface{}{
						"name":        certName, // Level 2 中敏感
						"description": "专业认证证书", // Level 2 中敏感
					}
					certifications = append(certifications, cert)
				}
			}
		}
	}

	// 去重
	uniqueCerts := make(map[string]bool)
	var result []map[string]interface{}
	for _, cert := range certifications {
		name := cert["name"].(string)
		if !uniqueCerts[name] {
			uniqueCerts[name] = true
			result = append(result, cert)
		}
	}

	return result
}

// 生成关键词并应用分类
func (p *SensitivityAwareTextParser) generateKeywordsWithClassification(text string, skills []string) []string {
	keywords := make([]string, 0)

	// 添加技能作为关键词 - Level 1 低敏感
	keywords = append(keywords, skills...)

	// 从文本中提取其他关键词 - Level 1 低敏感
	words := strings.Fields(text)
	wordCount := make(map[string]int)

	for _, word := range words {
		if len(word) > 2 && len(word) < 10 {
			wordCount[word]++
		}
	}

	// 选择出现频率较高的词作为关键词
	for word, count := range wordCount {
		if count > 2 && len(keywords) < 10 {
			keywords = append(keywords, word)
		}
	}

	log.Printf("生成 %d 个关键词 (Level 1)", len(keywords))
	return keywords
}

// 计算解析置信度并应用分类
func (p *SensitivityAwareTextParser) calculateConfidenceWithClassification(personalInfo map[string]interface{}, workExperience []map[string]interface{}, education []map[string]interface{}, skills []string) float64 {
	confidence := 0.0

	// 个人信息权重 30% (Level 3 高敏感)
	if name, ok := personalInfo["name"]; ok && name != "" {
		confidence += 0.1
	}
	if phone, ok := personalInfo["phone"]; ok && phone != "" {
		confidence += 0.1
	}
	if email, ok := personalInfo["email"]; ok && email != "" {
		confidence += 0.1
	}

	// 工作经历权重 40% (Level 2 中敏感)
	if len(workExperience) > 0 {
		confidence += 0.2
		if len(workExperience) > 1 {
			confidence += 0.2
		}
	}

	// 教育背景权重 20% (Level 2 中敏感)
	if len(education) > 0 {
		confidence += 0.2
	}

	// 技能权重 10% (Level 2 中敏感)
	if len(skills) > 0 {
		confidence += 0.1
	}

	log.Printf("计算解析置信度: %.2f", confidence)
	return confidence
}

// 创建数据分类标签
func (p *SensitivityAwareTextParser) createDataClassificationTags(personalInfo map[string]interface{}, workExperience []map[string]interface{}, education []map[string]interface{}, skills []string) map[string]DataClassificationTag {
	classification := make(map[string]DataClassificationTag)

	// 为个人信息字段添加分类标签
	for field := range personalInfo {
		if tag, exists := DataClassificationConfig[field]; exists {
			classification[field] = tag
		}
	}

	// 为工作经历字段添加分类标签
	for _, exp := range workExperience {
		for field := range exp {
			if tag, exists := DataClassificationConfig[field]; exists {
				classification[field] = tag
			}
		}
	}

	// 为教育背景字段添加分类标签
	for _, edu := range education {
		for field := range edu {
			if tag, exists := DataClassificationConfig[field]; exists {
				classification[field] = tag
			}
		}
	}

	// 为技能字段添加分类标签
	if len(skills) > 0 {
		if tag, exists := DataClassificationConfig["skills"]; exists {
			classification["skills"] = tag
		}
	}

	// 添加系统字段分类标签
	classification["content"] = DataClassificationConfig["content"]
	classification["keywords"] = DataClassificationConfig["keywords"]
	classification["confidence"] = DataClassificationConfig["confidence"]

	log.Printf("创建 %d 个数据分类标签", len(classification))
	return classification
}

// 计算整体敏感程度
func (p *SensitivityAwareTextParser) calculateOverallSensitivityLevel(classification map[string]DataClassificationTag) string {
	levelCount := map[string]int{
		SensitivityLevel1: 0,
		SensitivityLevel2: 0,
		SensitivityLevel3: 0,
		SensitivityLevel4: 0,
	}

	for _, tag := range classification {
		levelCount[tag.SensitivityLevel]++
	}

	// 返回最高敏感程度
	if levelCount[SensitivityLevel4] > 0 {
		return SensitivityLevel4
	}
	if levelCount[SensitivityLevel3] > 0 {
		return SensitivityLevel3
	}
	if levelCount[SensitivityLevel2] > 0 {
		return SensitivityLevel2
	}
	return SensitivityLevel1
}

// 检查是否需要用户同意
func (p *SensitivityAwareTextParser) checkConsentRequirement(classification map[string]DataClassificationTag) bool {
	for _, tag := range classification {
		if tag.RequiresConsent {
			return true
		}
	}
	return false
}

// 计算最大保留期限
func (p *SensitivityAwareTextParser) calculateMaxRetentionPeriod(classification map[string]DataClassificationTag) int {
	maxRetention := 0
	for _, tag := range classification {
		if tag.RetentionPeriod > maxRetention {
			maxRetention = tag.RetentionPeriod
		}
	}
	return maxRetention
}

// 提取标题
func (p *SensitivityAwareTextParser) extractTitle(text string) string {
	// 取第一行作为标题，或者从个人信息中提取姓名
	lines := strings.Split(text, "\n")
	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		if len(firstLine) > 0 && len(firstLine) < 50 {
			return firstLine
		}
	}

	return "解析的简历"
}

// SensitivityAwareParsedDataForStorage 用于存储的敏感信息感知解析数据结构
type SensitivityAwareParsedDataForStorage struct {
	PersonalInfoJSON   string  `json:"personal_info"`
	WorkExperienceJSON string  `json:"work_experience"`
	EducationJSON      string  `json:"education"`
	SkillsJSON         string  `json:"skills"`
	ProjectsJSON       string  `json:"projects"`
	CertificationsJSON string  `json:"certifications"`
	KeywordsJSON       string  `json:"keywords"`
	Confidence         float64 `json:"confidence"`
	SensitivityLevel   string  `json:"sensitivity_level"`
	DataClassification string  `json:"data_classification"`
}

// NewSensitivityAwareParser 创建新的敏感信息感知解析器
func NewSensitivityAwareParser() *SensitivityAwareTextParser {
	return &SensitivityAwareTextParser{
		encryptionKey: []byte("default-encryption-key-32-bytes-long"),
	}
}

// ParseMinerUResult 解析MinerU返回的结果
func (p *SensitivityAwareTextParser) ParseMinerUResult(data map[string]interface{}) (*SensitivityAwareParsedData, error) {
	// 从MinerU结果中提取内容
	content, ok := data["content"].(string)
	if !ok {
		return nil, fmt.Errorf("MinerU结果中缺少content字段")
	}

	// 创建基础解析数据结构
	parsedData := &SensitivityAwareParsedData{
		Title:              "解析的简历",
		Content:            content,
		PersonalInfo:       make(map[string]interface{}),
		WorkExperience:     []map[string]interface{}{},
		Education:          []map[string]interface{}{},
		Skills:             []string{},
		Projects:           []map[string]interface{}{},
		Certifications:     []map[string]interface{}{},
		Keywords:           []string{},
		Confidence:         0.95, // 默认置信度
		DataClassification: make(map[string]DataClassificationTag),
		ParsingMetadata:    make(map[string]interface{}),
	}

	// 设置置信度
	if confidence, exists := data["confidence"]; exists {
		if conf, ok := confidence.(float64); ok {
			parsedData.Confidence = conf
		}
	}

	// 从MinerU结果中提取文件信息
	if fileInfo, exists := data["file_info"]; exists {
		if fileInfoMap, ok := fileInfo.(map[string]interface{}); ok {
			parsedData.ParsingMetadata["file_info"] = fileInfoMap
		}
	}

	// 从MinerU结果中提取元数据
	if metadata, exists := data["metadata"]; exists {
		if metadataMap, ok := metadata.(map[string]interface{}); ok {
			parsedData.ParsingMetadata["metadata"] = metadataMap
		}
	}

	// 从MinerU结果中提取结构信息
	if structure, exists := data["structure"]; exists {
		if structureMap, ok := structure.(map[string]interface{}); ok {
			parsedData.ParsingMetadata["structure"] = structureMap
			// 尝试从结构中提取标题
			if title, ok := structureMap["title"].(string); ok && title != "" {
				parsedData.Title = title
			}
		}
	}

	// 使用改进的文本解析逻辑
	err := p.parseBasicResumeContent(content, parsedData)
	if err != nil {
		log.Printf("警告: 基础内容解析失败: %v", err)
		// 即使解析失败，仍然返回基础数据结构
	}

	return parsedData, nil
}

// parseBasicResumeContent 基础的简历内容解析
func (p *SensitivityAwareTextParser) parseBasicResumeContent(content string, data *SensitivityAwareParsedData) error {
	// 这里实现基础的文本解析逻辑
	// 可以根据需要添加更复杂的解析规则

	// 提取关键词
	keywords := p.extractKeywords(content)
	data.Keywords = keywords

	// 提取技能（简单实现）
	skills := p.extractSkills(content)
	data.Skills = skills

	// 尝试从内容中提取个人信息
	p.extractPersonalInfo(content, data)

	// 尝试从内容中提取工作经历
	p.extractWorkExperience(content, data)

	// 尝试从内容中提取教育背景
	p.extractEducation(content, data)

	// 尝试从内容中提取项目经验
	p.extractProjects(content, data)

	// 尝试从内容中提取证书认证
	p.extractCertifications(content, data)

	// 设置解析元数据
	if data.ParsingMetadata == nil {
		data.ParsingMetadata = make(map[string]interface{})
	}
	data.ParsingMetadata["parser_version"] = "basic-v1.0"
	data.ParsingMetadata["content_length"] = len(content)
	data.ParsingMetadata["parsed_at"] = time.Now().Format(time.RFC3339)

	return nil
}

// extractKeywords 提取关键词
func (p *SensitivityAwareTextParser) extractKeywords(content string) []string {
	// 简单的关键词提取逻辑
	keywords := []string{}

	// 常见的简历关键词
	commonKeywords := []string{
		"工作经验", "教育背景", "技能", "项目经验", "证书", "获奖",
		"工作经验", "教育背景", "技能", "项目经验", "证书", "获奖",
		"工作经验", "教育背景", "技能", "项目经验", "证书", "获奖",
	}

	for _, keyword := range commonKeywords {
		if strings.Contains(content, keyword) {
			keywords = append(keywords, keyword)
		}
	}

	return keywords
}

// extractSkills 提取技能
func (p *SensitivityAwareTextParser) extractSkills(content string) []string {
	// 简单的技能提取逻辑
	skills := []string{}

	// 常见的技能关键词
	commonSkills := []string{
		"Python", "Java", "JavaScript", "Go", "C++", "C#",
		"React", "Vue", "Angular", "Node.js", "Spring",
		"MySQL", "PostgreSQL", "MongoDB", "Redis",
		"Docker", "Kubernetes", "AWS", "Azure",
	}

	for _, skill := range commonSkills {
		if strings.Contains(content, skill) {
			skills = append(skills, skill)
		}
	}

	return skills
}

// ClassifySensitiveData 对解析数据进行敏感信息分类
func (p *SensitivityAwareTextParser) ClassifySensitiveData(data *SensitivityAwareParsedData) (map[string]DataClassificationTag, error) {
	classification := make(map[string]DataClassificationTag)

	// 分类个人信息
	if data.PersonalInfo != nil {
		classification["personal_info"] = DataClassificationTag{
			FieldName:        "personal_info",
			SensitivityLevel: SensitivityLevel3, // 高敏感
			DataType:         "personal_identification",
			ProtectionMethod: "encryption",
			RetentionPeriod:  365,
			RequiresConsent:  true,
			IsPersonalInfo:   true,
		}
	}

	// 分类工作经历
	if len(data.WorkExperience) > 0 {
		classification["work_experience"] = DataClassificationTag{
			FieldName:        "work_experience",
			SensitivityLevel: SensitivityLevel2, // 中敏感
			DataType:         "professional_history",
			ProtectionMethod: "access_control",
			RetentionPeriod:  730,
			RequiresConsent:  true,
			IsPersonalInfo:   false,
		}
	}

	// 分类教育背景
	if len(data.Education) > 0 {
		classification["education"] = DataClassificationTag{
			FieldName:        "education",
			SensitivityLevel: SensitivityLevel2, // 中敏感
			DataType:         "educational_background",
			ProtectionMethod: "access_control",
			RetentionPeriod:  1095,
			RequiresConsent:  true,
			IsPersonalInfo:   false,
		}
	}

	// 分类技能
	if len(data.Skills) > 0 {
		classification["skills"] = DataClassificationTag{
			FieldName:        "skills",
			SensitivityLevel: SensitivityLevel1, // 低敏感
			DataType:         "professional_skills",
			ProtectionMethod: "none",
			RetentionPeriod:  365,
			RequiresConsent:  false,
			IsPersonalInfo:   false,
		}
	}

	// 分类项目经验
	if len(data.Projects) > 0 {
		classification["projects"] = DataClassificationTag{
			FieldName:        "projects",
			SensitivityLevel: SensitivityLevel2, // 中敏感
			DataType:         "project_experience",
			ProtectionMethod: "access_control",
			RetentionPeriod:  730,
			RequiresConsent:  true,
			IsPersonalInfo:   false,
		}
	}

	// 分类证书认证
	if len(data.Certifications) > 0 {
		classification["certifications"] = DataClassificationTag{
			FieldName:        "certifications",
			SensitivityLevel: SensitivityLevel2, // 中敏感
			DataType:         "professional_certifications",
			ProtectionMethod: "access_control",
			RetentionPeriod:  1095,
			RequiresConsent:  true,
			IsPersonalInfo:   false,
		}
	}

	// 分类关键词
	if len(data.Keywords) > 0 {
		classification["keywords"] = DataClassificationTag{
			FieldName:        "keywords",
			SensitivityLevel: SensitivityLevel1, // 低敏感
			DataType:         "search_keywords",
			ProtectionMethod: "none",
			RetentionPeriod:  180,
			RequiresConsent:  false,
			IsPersonalInfo:   false,
		}
	}

	return classification, nil
}

// ProtectSensitiveData 根据敏感度级别保护数据
func (p *SensitivityAwareTextParser) ProtectSensitiveData(data *SensitivityAwareParsedData, classification map[string]DataClassificationTag) (*SensitivityAwareParsedDataForStorage, error) {
	result := &SensitivityAwareParsedDataForStorage{
		Confidence: data.Confidence,
	}

	// 处理个人信息（高敏感 - 加密）
	if personalInfo, exists := classification["personal_info"]; exists {
		if personalInfo.SensitivityLevel == SensitivityLevel3 {
			// 对高敏感信息进行加密
			encryptedData, err := p.encryptSensitiveData(data.PersonalInfo)
			if err != nil {
				return nil, fmt.Errorf("加密个人信息失败: %v", err)
			}
			result.PersonalInfoJSON = encryptedData
		} else {
			// 其他级别直接序列化
			jsonData, _ := json.Marshal(data.PersonalInfo)
			result.PersonalInfoJSON = string(jsonData)
		}
	} else {
		result.PersonalInfoJSON = "{}"
	}

	// 处理工作经历（中敏感 - 访问控制）
	if workExp, exists := classification["work_experience"]; exists {
		if workExp.SensitivityLevel == SensitivityLevel2 {
			// 中敏感信息添加访问控制标记
			jsonData, _ := json.Marshal(data.WorkExperience)
			result.WorkExperienceJSON = string(jsonData)
		} else {
			jsonData, _ := json.Marshal(data.WorkExperience)
			result.WorkExperienceJSON = string(jsonData)
		}
	} else {
		result.WorkExperienceJSON = "[]"
	}

	// 处理教育背景（中敏感 - 访问控制）
	if _, exists := classification["education"]; exists {
		jsonData, _ := json.Marshal(data.Education)
		result.EducationJSON = string(jsonData)
	} else {
		result.EducationJSON = "[]"
	}

	// 处理技能（低敏感 - 无保护）
	if _, exists := classification["skills"]; exists {
		jsonData, _ := json.Marshal(data.Skills)
		result.SkillsJSON = string(jsonData)
	} else {
		result.SkillsJSON = "[]"
	}

	// 处理项目经验（中敏感 - 访问控制）
	if _, exists := classification["projects"]; exists {
		jsonData, _ := json.Marshal(data.Projects)
		result.ProjectsJSON = string(jsonData)
	} else {
		result.ProjectsJSON = "[]"
	}

	// 处理证书认证（中敏感 - 访问控制）
	if _, exists := classification["certifications"]; exists {
		jsonData, _ := json.Marshal(data.Certifications)
		result.CertificationsJSON = string(jsonData)
	} else {
		result.CertificationsJSON = "[]"
	}

	// 处理关键词（低敏感 - 无保护）
	if _, exists := classification["keywords"]; exists {
		jsonData, _ := json.Marshal(data.Keywords)
		result.KeywordsJSON = string(jsonData)
	} else {
		result.KeywordsJSON = "[]"
	}

	// 设置整体敏感度级别
	result.SensitivityLevel = p.determineOverallSensitivityLevel(classification)

	// 序列化分类信息
	classificationJSON, _ := json.Marshal(classification)
	result.DataClassification = string(classificationJSON)

	return result, nil
}

// determineOverallSensitivityLevel 确定整体敏感度级别
func (p *SensitivityAwareTextParser) determineOverallSensitivityLevel(classification map[string]DataClassificationTag) string {
	// 检查是否有极高敏感信息
	for _, tag := range classification {
		if tag.SensitivityLevel == SensitivityLevel4 {
			return SensitivityLevel4
		}
	}

	// 检查是否有高敏感信息
	for _, tag := range classification {
		if tag.SensitivityLevel == SensitivityLevel3 {
			return SensitivityLevel3
		}
	}

	// 检查是否有中敏感信息
	for _, tag := range classification {
		if tag.SensitivityLevel == SensitivityLevel2 {
			return SensitivityLevel2
		}
	}

	// 默认为低敏感
	return SensitivityLevel1
}

// encryptSensitiveData 加密敏感数据
func (p *SensitivityAwareTextParser) encryptSensitiveData(data interface{}) (string, error) {
	// 将数据序列化为JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("序列化数据失败: %v", err)
	}

	// 这里应该使用真正的加密算法，现在只是简单编码
	// 在实际生产环境中，应该使用AES等强加密算法
	encoded := base64.StdEncoding.EncodeToString(jsonData)

	return fmt.Sprintf("ENCRYPTED:%s", encoded), nil
}

// extractPersonalInfo 从内容中提取个人信息
func (p *SensitivityAwareTextParser) extractPersonalInfo(content string, data *SensitivityAwareParsedData) {
	// 简单的个人信息提取逻辑
	personalInfo := make(map[string]interface{})

	// 提取姓名（简单实现）
	if name := p.extractName(content); name != "" {
		personalInfo["name"] = name
	}

	// 提取邮箱
	if email := p.extractEmail(content); email != "" {
		personalInfo["email"] = email
	}

	// 提取电话
	if phone := p.extractPhone(content); phone != "" {
		personalInfo["phone"] = phone
	}

	// 提取地址
	if address := p.extractAddress(content); address != "" {
		personalInfo["address"] = address
	}

	if len(personalInfo) > 0 {
		data.PersonalInfo = personalInfo
	}
}

// extractWorkExperience 从内容中提取工作经历
func (p *SensitivityAwareTextParser) extractWorkExperience(content string, data *SensitivityAwareParsedData) {
	// 简单的工作经历提取逻辑
	workExperiences := []map[string]interface{}{}

	// 这里可以实现更复杂的工作经历提取逻辑
	// 目前返回空数组，表示没有提取到工作经历
	data.WorkExperience = workExperiences
}

// extractEducation 从内容中提取教育背景
func (p *SensitivityAwareTextParser) extractEducation(content string, data *SensitivityAwareParsedData) {
	// 简单的教育背景提取逻辑
	education := []map[string]interface{}{}

	// 这里可以实现更复杂的教育背景提取逻辑
	// 目前返回空数组，表示没有提取到教育背景
	data.Education = education
}

// extractProjects 从内容中提取项目经验
func (p *SensitivityAwareTextParser) extractProjects(content string, data *SensitivityAwareParsedData) {
	// 简单的项目经验提取逻辑
	projects := []map[string]interface{}{}

	// 这里可以实现更复杂的项目经验提取逻辑
	// 目前返回空数组，表示没有提取到项目经验
	data.Projects = projects
}

// extractCertifications 从内容中提取证书认证
func (p *SensitivityAwareTextParser) extractCertifications(content string, data *SensitivityAwareParsedData) {
	// 简单的证书认证提取逻辑
	certifications := []map[string]interface{}{}

	// 这里可以实现更复杂的证书认证提取逻辑
	// 目前返回空数组，表示没有提取到证书认证
	data.Certifications = certifications
}

// extractName 提取姓名
func (p *SensitivityAwareTextParser) extractName(content string) string {
	// 简单的姓名提取逻辑
	// 这里可以实现更复杂的姓名识别算法
	return ""
}

// extractEmail 提取邮箱
func (p *SensitivityAwareTextParser) extractEmail(content string) string {
	// 使用正则表达式提取邮箱
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	matches := emailRegex.FindStringSubmatch(content)
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

// extractPhone 提取电话
func (p *SensitivityAwareTextParser) extractPhone(content string) string {
	// 使用正则表达式提取电话
	phoneRegex := regexp.MustCompile(`1[3-9]\d{9}`)
	matches := phoneRegex.FindStringSubmatch(content)
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

// extractAddress 提取地址
func (p *SensitivityAwareTextParser) extractAddress(content string) string {
	// 简单的地址提取逻辑
	// 这里可以实现更复杂的地址识别算法
	return ""
}
