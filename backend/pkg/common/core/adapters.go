package core

import (
	"time"
)

// DataAdapter 数据适配器接口
type DataAdapter interface {
	// 用户数据适配
	AdaptUserData(oldData interface{}) interface{}
	AdaptUserList(oldData []interface{}) []interface{}

	// 职位数据适配
	AdaptJobData(oldData interface{}) interface{}
	AdaptJobList(oldData []interface{}) []interface{}

	// 企业数据适配
	AdaptCompanyData(oldData interface{}) interface{}
	AdaptCompanyList(oldData []interface{}) []interface{}

	// 轮播图数据适配
	AdaptBannerData(oldData interface{}) interface{}
	AdaptBannerList(oldData []interface{}) []interface{}
}

// V1ToV2Adapter V1到V2数据适配器
type V1ToV2Adapter struct{}

// NewV1ToV2Adapter 创建V1到V2适配器
func NewV1ToV2Adapter() *V1ToV2Adapter {
	return &V1ToV2Adapter{}
}

// getOrDefault 获取值或默认值
func (a *V1ToV2Adapter) getOrDefault(value interface{}, defaultValue interface{}) interface{} {
	if value == nil {
		return defaultValue
	}
	return value
}

// AdaptUserData 适配用户数据
func (a *V1ToV2Adapter) AdaptUserData(oldData interface{}) interface{} {
	// 这里需要根据实际的数据结构进行适配
	// 示例：将旧格式用户数据转换为新格式
	if oldUser, ok := oldData.(map[string]interface{}); ok {
		newUser := map[string]interface{}{
			"id":                   oldUser["id"],
			"username":             oldUser["username"],
			"email":                oldUser["email"],
			"phone":                oldUser["phone"],
			"password_hash":        oldUser["password_hash"],
			"avatar_url":           oldUser["avatar_url"],
			"nickname":             a.getOrDefault(oldUser["nickname"], oldUser["username"]),
			"real_name":            a.getOrDefault(oldUser["real_name"], ""),
			"gender":               a.getOrDefault(oldUser["gender"], "other"),
			"birth_date":           oldUser["birth_date"],
			"location":             a.getOrDefault(oldUser["location"], ""),
			"status":               a.adaptUserStatus(oldUser["status"]),
			"user_type":            "jobseeker", // 默认值
			"certification_status": "pending",   // 默认值
			"last_login_at":        oldUser["last_login_at"],
			"login_count":          0, // 默认值
			"created_at":           oldUser["created_at"],
			"updated_at":           oldUser["updated_at"],
			"deleted_at":           oldUser["deleted_at"],
		}
		return newUser
	}
	return oldData
}

// AdaptUserList 适配用户列表
func (a *V1ToV2Adapter) AdaptUserList(oldData []interface{}) []interface{} {
	var newList []interface{}
	for _, item := range oldData {
		newList = append(newList, a.AdaptUserData(item))
	}
	return newList
}

// AdaptJobData 适配职位数据
func (a *V1ToV2Adapter) AdaptJobData(oldData interface{}) interface{} {
	if oldJob, ok := oldData.(map[string]interface{}); ok {
		newJob := map[string]interface{}{
			"id":                  oldJob["id"],
			"company_id":          oldJob["company_id"],
			"category_id":         a.getOrDefault(oldJob["category_id"], 1), // 默认分类
			"title":               oldJob["title"],
			"job_type":            a.getOrDefault(oldJob["job_type"], "full_time"),
			"location":            oldJob["location"],
			"salary_min":          oldJob["salary_min"],
			"salary_max":          oldJob["salary_max"],
			"salary_type":         a.getOrDefault(oldJob["salary_type"], "monthly"),
			"experience_required": a.getOrDefault(oldJob["experience_required"], "entry"),
			"education_required":  a.getOrDefault(oldJob["education_required"], "bachelor"),
			"description":         oldJob["description"],
			"requirements":        a.getOrDefault(oldJob["requirements"], ""),
			"benefits":            a.getOrDefault(oldJob["benefits"], ""),
			"skills":              a.getOrDefault(oldJob["skills"], "[]"),
			"tags":                a.getOrDefault(oldJob["tags"], "[]"),
			"status":              a.getOrDefault(oldJob["status"], "draft"),
			"priority":            0,
			"view_count":          0,
			"application_count":   0,
			"favorite_count":      0,
			"publish_at":          a.getOrDefault(oldJob["publish_at"], time.Now()),
			"expire_at":           oldJob["expire_at"],
			"created_at":          oldJob["created_at"],
			"updated_at":          oldJob["updated_at"],
			"deleted_at":          oldJob["deleted_at"],
		}
		return newJob
	}
	return oldData
}

// AdaptJobList 适配职位列表
func (a *V1ToV2Adapter) AdaptJobList(oldData []interface{}) []interface{} {
	var newList []interface{}
	for _, item := range oldData {
		newList = append(newList, a.AdaptJobData(item))
	}
	return newList
}

// AdaptCompanyData 适配企业数据
func (a *V1ToV2Adapter) AdaptCompanyData(oldData interface{}) interface{} {
	if oldCompany, ok := oldData.(map[string]interface{}); ok {
		newCompany := map[string]interface{}{
			"id":                 oldCompany["id"],
			"name":               oldCompany["name"],
			"short_name":         a.getOrDefault(oldCompany["short_name"], oldCompany["name"]),
			"logo_url":           a.getOrDefault(oldCompany["logo_url"], ""),
			"industry":           a.getOrDefault(oldCompany["industry"], ""),
			"company_size":       a.getOrDefault(oldCompany["company_size"], "medium"),
			"location":           a.getOrDefault(oldCompany["location"], ""),
			"website":            a.getOrDefault(oldCompany["website"], ""),
			"description":        a.getOrDefault(oldCompany["description"], ""),
			"founded_year":       oldCompany["founded_year"],
			"business_license":   a.getOrDefault(oldCompany["business_license"], ""),
			"status":             a.getOrDefault(oldCompany["status"], "pending"),
			"verification_level": a.getOrDefault(oldCompany["verification_level"], "basic"),
			"job_count":          0,
			"view_count":         0,
			"created_at":         oldCompany["created_at"],
			"updated_at":         oldCompany["updated_at"],
			"deleted_at":         oldCompany["deleted_at"],
		}
		return newCompany
	}
	return oldData
}

// AdaptCompanyList 适配企业列表
func (a *V1ToV2Adapter) AdaptCompanyList(oldData []interface{}) []interface{} {
	var newList []interface{}
	for _, item := range oldData {
		newList = append(newList, a.AdaptCompanyData(item))
	}
	return newList
}

// AdaptBannerData 适配轮播图数据
func (a *V1ToV2Adapter) AdaptBannerData(oldData interface{}) interface{} {
	if oldBanner, ok := oldData.(map[string]interface{}); ok {
		newBanner := map[string]interface{}{
			"id":          oldBanner["id"],
			"title":       oldBanner["title"],
			"image_url":   oldBanner["image_url"],
			"link_url":    a.getOrDefault(oldBanner["link_url"], ""),
			"link_type":   a.getOrDefault(oldBanner["link_type"], "internal"),
			"sort_order":  a.getOrDefault(oldBanner["sort_order"], 0),
			"status":      a.getOrDefault(oldBanner["status"], "active"),
			"start_time":  oldBanner["start_time"],
			"end_time":    oldBanner["end_time"],
			"view_count":  0,
			"click_count": 0,
			"created_at":  oldBanner["created_at"],
			"updated_at":  oldBanner["updated_at"],
		}
		return newBanner
	}
	return oldData
}

// AdaptBannerList 适配轮播图列表
func (a *V1ToV2Adapter) AdaptBannerList(oldData []interface{}) []interface{} {
	var newList []interface{}
	for _, item := range oldData {
		newList = append(newList, a.AdaptBannerData(item))
	}
	return newList
}

// adaptUserStatus 适配用户状态
func (a *V1ToV2Adapter) adaptUserStatus(status interface{}) string {
	if status == nil {
		return "inactive"
	}

	switch status.(string) {
	case "active":
		return "active"
	case "inactive":
		return "inactive"
	case "banned":
		return "suspended"
	default:
		return "inactive"
	}
}

// APIVersionManager API版本管理器
type APIVersionManager struct {
	adapter DataAdapter
}

// NewAPIVersionManager 创建API版本管理器
func NewAPIVersionManager() *APIVersionManager {
	return &APIVersionManager{
		adapter: NewV1ToV2Adapter(),
	}
}

// GetAdapter 获取数据适配器
func (m *APIVersionManager) GetAdapter() DataAdapter {
	return m.adapter
}

// ShouldUseNewAPI 判断是否使用新API
func (m *APIVersionManager) ShouldUseNewAPI(version string) bool {
	return version == "v2"
}

// AdaptResponse 适配响应数据
func (m *APIVersionManager) AdaptResponse(version string, data interface{}, dataType string) interface{} {
	if !m.ShouldUseNewAPI(version) {
		return data
	}

	switch dataType {
	case "user":
		return m.adapter.AdaptUserData(data)
	case "user_list":
		if list, ok := data.([]interface{}); ok {
			return m.adapter.AdaptUserList(list)
		}
	case "job":
		return m.adapter.AdaptJobData(data)
	case "job_list":
		if list, ok := data.([]interface{}); ok {
			return m.adapter.AdaptJobList(list)
		}
	case "company":
		return m.adapter.AdaptCompanyData(data)
	case "company_list":
		if list, ok := data.([]interface{}); ok {
			return m.adapter.AdaptCompanyList(list)
		}
	case "banner":
		return m.adapter.AdaptBannerData(data)
	case "banner_list":
		if list, ok := data.([]interface{}); ok {
			return m.adapter.AdaptBannerList(list)
		}
	}

	return data
}
