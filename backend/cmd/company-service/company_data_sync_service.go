package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"gorm.io/gorm"
)

// CompanyDataSyncService 企业数据同步服务
type CompanyDataSyncService struct {
	mysqlDB     *gorm.DB
	postgresDB  *gorm.DB
	neo4jDriver neo4j.Driver
	redisClient *redis.Client
}

// NewCompanyDataSyncService 创建企业数据同步服务
func NewCompanyDataSyncService(mysqlDB *gorm.DB, postgresDB *gorm.DB, neo4jDriver neo4j.Driver, redisClient *redis.Client) *CompanyDataSyncService {
	return &CompanyDataSyncService{
		mysqlDB:     mysqlDB,
		postgresDB:  postgresDB,
		neo4jDriver: neo4jDriver,
		redisClient: redisClient,
	}
}

// SyncCompanyData 同步企业数据到所有数据库
func (s *CompanyDataSyncService) SyncCompanyData(companyID uint) error {
	// 1. 从MySQL获取核心企业数据
	var company EnhancedCompany
	if err := s.mysqlDB.Preload("CompanyUsers").First(&company, companyID).Error; err != nil {
		return fmt.Errorf("获取企业数据失败: %v", err)
	}

	// 2. 同步到PostgreSQL（职位相关数据）
	if err := s.syncToPostgreSQL(company); err != nil {
		log.Printf("同步到PostgreSQL失败: %v", err)
		s.updateSyncStatus(companyID, SyncTargetPostgreSQL, SyncStatusFailed, err.Error())
	} else {
		s.updateSyncStatus(companyID, SyncTargetPostgreSQL, SyncStatusSuccess, "")
	}

	// 3. 同步到Neo4j（地理位置和关系数据）
	if err := s.syncToNeo4j(company); err != nil {
		log.Printf("同步到Neo4j失败: %v", err)
		s.updateSyncStatus(companyID, SyncTargetNeo4j, SyncStatusFailed, err.Error())
	} else {
		s.updateSyncStatus(companyID, SyncTargetNeo4j, SyncStatusSuccess, "")
	}

	// 4. 同步到Redis（缓存数据）
	if err := s.syncToRedis(company); err != nil {
		log.Printf("同步到Redis失败: %v", err)
		s.updateSyncStatus(companyID, SyncTargetRedis, SyncStatusFailed, err.Error())
	} else {
		s.updateSyncStatus(companyID, SyncTargetRedis, SyncStatusSuccess, "")
	}

	return nil
}

// syncToPostgreSQL 同步到PostgreSQL
func (s *CompanyDataSyncService) syncToPostgreSQL(company EnhancedCompany) error {
	if s.postgresDB == nil {
		return fmt.Errorf("PostgreSQL连接未初始化")
	}

	// 创建或更新企业基础信息
	companyInfo := map[string]interface{}{
		"id":                         company.ID,
		"name":                       company.Name,
		"short_name":                 company.ShortName,
		"industry":                   company.Industry,
		"company_size":               company.CompanySize,
		"location":                   company.Location,
		"website":                    company.Website,
		"description":                company.Description,
		"founded_year":               company.FoundedYear,
		"unified_social_credit_code": company.UnifiedSocialCreditCode,
		"legal_representative":       company.LegalRepresentative,
		"legal_representative_id":    company.LegalRepresentativeID,
		"legal_rep_user_id":          company.LegalRepUserID,
		"status":                     company.Status,
		"verification_level":         company.VerificationLevel,
		"job_count":                  company.JobCount,
		"view_count":                 company.ViewCount,
		"created_by":                 company.CreatedBy,
		"created_at":                 company.CreatedAt,
		"updated_at":                 company.UpdatedAt,
	}

	// 使用UPSERT操作
	query := `
		INSERT INTO companies (
			id, name, short_name, industry, company_size, location, website, description,
			founded_year, unified_social_credit_code, legal_representative, legal_representative_id,
			legal_rep_user_id, status, verification_level, job_count, view_count,
			created_by, created_at, updated_at
		) VALUES (
			@id, @name, @short_name, @industry, @company_size, @location, @website, @description,
			@founded_year, @unified_social_credit_code, @legal_representative, @legal_representative_id,
			@legal_rep_user_id, @status, @verification_level, @job_count, @view_count,
			@created_by, @created_at, @updated_at
		)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			short_name = EXCLUDED.short_name,
			industry = EXCLUDED.industry,
			company_size = EXCLUDED.company_size,
			location = EXCLUDED.location,
			website = EXCLUDED.website,
			description = EXCLUDED.description,
			founded_year = EXCLUDED.founded_year,
			unified_social_credit_code = EXCLUDED.unified_social_credit_code,
			legal_representative = EXCLUDED.legal_representative,
			legal_representative_id = EXCLUDED.legal_representative_id,
			legal_rep_user_id = EXCLUDED.legal_rep_user_id,
			status = EXCLUDED.status,
			verification_level = EXCLUDED.verification_level,
			job_count = EXCLUDED.job_count,
			view_count = EXCLUDED.view_count,
			updated_at = EXCLUDED.updated_at
	`

	if err := s.postgresDB.Exec(query, companyInfo).Error; err != nil {
		return fmt.Errorf("同步企业信息到PostgreSQL失败: %v", err)
	}

	// 同步企业用户关联信息
	for _, companyUser := range company.CompanyUsers {
		userInfo := map[string]interface{}{
			"id":          companyUser.ID,
			"company_id":  companyUser.CompanyID,
			"user_id":     companyUser.UserID,
			"role":        companyUser.Role,
			"status":      companyUser.Status,
			"permissions": companyUser.Permissions,
			"created_at":  companyUser.CreatedAt,
			"updated_at":  companyUser.UpdatedAt,
		}

		userQuery := `
			INSERT INTO company_users (
				id, company_id, user_id, role, status, permissions, created_at, updated_at
			) VALUES (
				@id, @company_id, @user_id, @role, @status, @permissions, @created_at, @updated_at
			)
			ON CONFLICT (id) DO UPDATE SET
				role = EXCLUDED.role,
				status = EXCLUDED.status,
				permissions = EXCLUDED.permissions,
				updated_at = EXCLUDED.updated_at
		`

		if err := s.postgresDB.Exec(userQuery, userInfo).Error; err != nil {
			log.Printf("同步企业用户关联到PostgreSQL失败: %v", err)
		}
	}

	return nil
}

// syncToNeo4j 同步到Neo4j
func (s *CompanyDataSyncService) syncToNeo4j(company EnhancedCompany) error {
	if s.neo4jDriver == nil {
		return fmt.Errorf("Neo4j连接未初始化")
	}

	session := s.neo4jDriver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	// 创建或更新企业节点
	query := `
		MERGE (c:Company {id: $id})
		SET c.name = $name,
			c.short_name = $short_name,
			c.industry = $industry,
			c.company_size = $company_size,
			c.location = $location,
			c.website = $website,
			c.description = $description,
			c.founded_year = $founded_year,
			c.unified_social_credit_code = $unified_social_credit_code,
			c.legal_representative = $legal_representative,
			c.legal_representative_id = $legal_representative_id,
			c.legal_rep_user_id = $legal_rep_user_id,
			c.status = $status,
			c.verification_level = $verification_level,
			c.job_count = $job_count,
			c.view_count = $view_count,
			c.created_by = $created_by,
			c.created_at = $created_at,
			c.updated_at = $updated_at
	`

	params := map[string]interface{}{
		"id":                         company.ID,
		"name":                       company.Name,
		"short_name":                 company.ShortName,
		"industry":                   company.Industry,
		"company_size":               company.CompanySize,
		"location":                   company.Location,
		"website":                    company.Website,
		"description":                company.Description,
		"founded_year":               company.FoundedYear,
		"unified_social_credit_code": company.UnifiedSocialCreditCode,
		"legal_representative":       company.LegalRepresentative,
		"legal_representative_id":    company.LegalRepresentativeID,
		"legal_rep_user_id":          company.LegalRepUserID,
		"status":                     company.Status,
		"verification_level":         company.VerificationLevel,
		"job_count":                  company.JobCount,
		"view_count":                 company.ViewCount,
		"created_by":                 company.CreatedBy,
		"created_at":                 company.CreatedAt.Unix(),
		"updated_at":                 company.UpdatedAt.Unix(),
	}

	// 添加地理位置信息
	if company.BDLatitude != nil && company.BDLongitude != nil {
		query += `, c.bd_latitude = $bd_latitude, c.bd_longitude = $bd_longitude`
		params["bd_latitude"] = *company.BDLatitude
		params["bd_longitude"] = *company.BDLongitude

		if company.BDAltitude != nil {
			query += `, c.bd_altitude = $bd_altitude`
			params["bd_altitude"] = *company.BDAltitude
		}
		if company.BDAccuracy != nil {
			query += `, c.bd_accuracy = $bd_accuracy`
			params["bd_accuracy"] = *company.BDAccuracy
		}
		if company.BDTimestamp != nil {
			query += `, c.bd_timestamp = $bd_timestamp`
			params["bd_timestamp"] = *company.BDTimestamp
		}
	}

	// 添加地址信息
	if company.Address != "" {
		query += `, c.address = $address`
		params["address"] = company.Address
	}
	if company.City != "" {
		query += `, c.city = $city`
		params["city"] = company.City
	}
	if company.District != "" {
		query += `, c.district = $district`
		params["district"] = company.District
	}
	if company.Area != "" {
		query += `, c.area = $area`
		params["area"] = company.Area
	}

	_, err := session.Run(query, params)
	if err != nil {
		return fmt.Errorf("同步企业节点到Neo4j失败: %v", err)
	}

	// 创建地理位置关系
	if company.City != "" && company.District != "" {
		locationQuery := `
			MATCH (c:Company {id: $company_id})
			MERGE (city:City {name: $city})
			MERGE (district:District {name: $district})
			MERGE (city)-[:CONTAINS]->(district)
			MERGE (c)-[:LOCATED_IN]->(district)
			MERGE (c)-[:IN_CITY]->(city)
		`

		locationParams := map[string]interface{}{
			"company_id": company.ID,
			"city":       company.City,
			"district":   company.District,
		}

		if company.Area != "" {
			locationQuery += `
				MERGE (area:Area {name: $area})
				MERGE (district)-[:CONTAINS]->(area)
				MERGE (c)-[:LOCATED_IN]->(area)
			`
			locationParams["area"] = company.Area
		}

		_, err := session.Run(locationQuery, locationParams)
		if err != nil {
			log.Printf("创建地理位置关系失败: %v", err)
		}
	}

	// 创建企业用户关系
	for _, companyUser := range company.CompanyUsers {
		userQuery := `
			MATCH (c:Company {id: $company_id})
			MERGE (u:User {id: $user_id})
			MERGE (u)-[:WORKS_FOR {role: $role, status: $status}]->(c)
		`

		userParams := map[string]interface{}{
			"company_id": company.ID,
			"user_id":    companyUser.UserID,
			"role":       companyUser.Role,
			"status":     companyUser.Status,
		}

		_, err := session.Run(userQuery, userParams)
		if err != nil {
			log.Printf("创建企业用户关系失败: %v", err)
		}
	}

	return nil
}

// syncToRedis 同步到Redis
func (s *CompanyDataSyncService) syncToRedis(company EnhancedCompany) error {
	if s.redisClient == nil {
		return fmt.Errorf("Redis连接未初始化")
	}

	// 缓存企业基础信息
	companyKey := fmt.Sprintf("company:%d", company.ID)
	companyData, err := json.Marshal(company)
	if err != nil {
		return fmt.Errorf("序列化企业数据失败: %v", err)
	}

	if err := s.redisClient.Set(context.Background(), companyKey, companyData, time.Hour).Err(); err != nil {
		return fmt.Errorf("缓存企业数据失败: %v", err)
	}

	// 缓存企业权限信息
	permissionsKey := fmt.Sprintf("company_permissions:%d", company.ID)
	var permissions []CompanyPermissionInfo

	// 获取企业创建者权限
	permissions = append(permissions, CompanyPermissionInfo{
		UserID:                   company.CreatedBy,
		CompanyID:                company.ID,
		EffectivePermissionLevel: PermissionCompanyOwner,
		Role:                     "company_owner",
		Status:                   "active",
		Permissions:              []string{"*"},
	})

	// 获取法定代表人权限
	if company.LegalRepUserID > 0 {
		permissions = append(permissions, CompanyPermissionInfo{
			UserID:                   company.LegalRepUserID,
			CompanyID:                company.ID,
			EffectivePermissionLevel: PermissionLegalRepresentative,
			Role:                     string(RoleLegalRepresentative),
			Status:                   "active",
			Permissions:              []string{"read", "write", "manage_users"},
		})
	}

	// 获取授权用户权限
	for _, companyUser := range company.CompanyUsers {
		permissions = append(permissions, CompanyPermissionInfo{
			UserID:                   companyUser.UserID,
			CompanyID:                company.ID,
			EffectivePermissionLevel: PermissionAuthorizedUser,
			Role:                     companyUser.Role,
			Status:                   companyUser.Status,
			Permissions:              companyUser.GetPermissions(),
		})
	}

	permissionsData, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("序列化权限数据失败: %v", err)
	}

	if err := s.redisClient.Set(context.Background(), permissionsKey, permissionsData, time.Hour).Err(); err != nil {
		return fmt.Errorf("缓存权限数据失败: %v", err)
	}

	// 缓存地理位置信息
	if company.BDLatitude != nil && company.BDLongitude != nil {
		locationKey := fmt.Sprintf("company_location:%d", company.ID)
		locationData := map[string]interface{}{
			"latitude":      *company.BDLatitude,
			"longitude":     *company.BDLongitude,
			"altitude":      getFloat64Value(company.BDAltitude),
			"accuracy":      getFloat64Value(company.BDAccuracy),
			"timestamp":     getInt64Value(company.BDTimestamp),
			"address":       company.Address,
			"city":          company.City,
			"district":      company.District,
			"area":          company.Area,
			"postal_code":   company.PostalCode,
			"city_code":     company.CityCode,
			"district_code": company.DistrictCode,
			"area_code":     company.AreaCode,
		}

		locationJSON, err := json.Marshal(locationData)
		if err != nil {
			return fmt.Errorf("序列化地理位置数据失败: %v", err)
		}

		if err := s.redisClient.Set(context.Background(), locationKey, locationJSON, time.Hour).Err(); err != nil {
			return fmt.Errorf("缓存地理位置数据失败: %v", err)
		}
	}

	return nil
}

// updateSyncStatus 更新同步状态
func (s *CompanyDataSyncService) updateSyncStatus(companyID uint, target SyncTarget, status SyncStatus, errorMsg string) {
	syncStatus := CompanyDataSyncStatus{
		CompanyID:    companyID,
		SyncTarget:   string(target),
		SyncStatus:   string(status),
		LastSyncTime: &[]time.Time{time.Now()}[0],
		SyncError:    errorMsg,
		UpdatedAt:    time.Now(),
	}

	if status == SyncStatusFailed {
		// 获取当前重试次数
		var existingStatus CompanyDataSyncStatus
		if err := s.mysqlDB.Where("company_id = ? AND sync_target = ?", companyID, target).First(&existingStatus).Error; err == nil {
			syncStatus.RetryCount = existingStatus.RetryCount + 1
		}
	}

	// 使用UPSERT操作
	s.mysqlDB.Where("company_id = ? AND sync_target = ?", companyID, target).
		Assign(syncStatus).
		FirstOrCreate(&syncStatus)
}

// CheckDataConsistency 检查数据一致性
func (s *CompanyDataSyncService) CheckDataConsistency(companyID uint) error {
	// 1. 检查MySQL核心数据
	var company EnhancedCompany
	if err := s.mysqlDB.First(&company, companyID).Error; err != nil {
		return fmt.Errorf("MySQL数据缺失: %v", err)
	}

	// 2. 检查PostgreSQL数据
	if s.postgresDB != nil {
		var count int64
		s.postgresDB.Table("companies").Where("id = ?", companyID).Count(&count)
		if count == 0 {
			return fmt.Errorf("PostgreSQL数据缺失")
		}
	}

	// 3. 检查Neo4j数据
	if s.neo4jDriver != nil {
		session := s.neo4jDriver.NewSession(neo4j.SessionConfig{})
		defer session.Close()

		result, err := session.Run("MATCH (c:Company {id: $id}) RETURN c", map[string]interface{}{"id": companyID})
		if err != nil {
			return fmt.Errorf("Neo4j数据检查失败: %v", err)
		}

		if !result.Next() {
			return fmt.Errorf("Neo4j数据缺失")
		}
	}

	// 4. 检查Redis缓存
	if s.redisClient != nil {
		companyKey := fmt.Sprintf("company:%d", companyID)
		if _, err := s.redisClient.Get(context.Background(), companyKey).Result(); err != nil {
			return fmt.Errorf("Redis缓存缺失: %v", err)
		}
	}

	return nil
}

// GetSyncStatus 获取同步状态
func (s *CompanyDataSyncService) GetSyncStatus(companyID uint) ([]CompanySyncInfo, error) {
	var syncStatuses []CompanyDataSyncStatus
	if err := s.mysqlDB.Where("company_id = ?", companyID).Find(&syncStatuses).Error; err != nil {
		return nil, fmt.Errorf("获取同步状态失败: %v", err)
	}

	var syncInfos []CompanySyncInfo
	for _, status := range syncStatuses {
		syncInfos = append(syncInfos, CompanySyncInfo{
			CompanyID:    status.CompanyID,
			SyncTarget:   SyncTarget(status.SyncTarget),
			SyncStatus:   SyncStatus(status.SyncStatus),
			LastSyncTime: status.LastSyncTime,
			SyncError:    status.SyncError,
			RetryCount:   status.RetryCount,
		})
	}

	return syncInfos, nil
}

// GetCompanyRelationships 获取企业关系
func (s *CompanyDataSyncService) GetCompanyRelationships(companyID uint) ([]CompanyRelationship, error) {
	if s.neo4jDriver == nil {
		return nil, fmt.Errorf("Neo4j未连接，无法获取关系")
	}

	session := s.neo4jDriver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.Run(`
		MATCH (source:Company {id: $companyID})-[r:RELATED_TO]->(target:Company)
		RETURN target.id as target_id, r.type as relationship, r.weight as weight
		ORDER BY r.weight DESC
	`, map[string]interface{}{
		"companyID": companyID,
	})

	if err != nil {
		return nil, err
	}

	var relationships []CompanyRelationship
	for result.Next() {
		record := result.Record()
		rel := CompanyRelationship{
			CompanyID:          companyID,
			RelatedCompanyName: fmt.Sprintf("Company_%d", record.Values[0].(int64)),
			RelationshipType:   record.Values[1].(string),
			InvestmentAmount:   record.Values[2].(float64),
		}
		relationships = append(relationships, rel)
	}

	return relationships, nil
}

// CreateCompanyRelationship 创建企业关系
func (s *CompanyDataSyncService) CreateCompanyRelationship(sourceID, targetID uint, relationship string, weight float64) error {
	if s.neo4jDriver == nil {
		return fmt.Errorf("Neo4j未连接，无法创建关系")
	}

	session := s.neo4jDriver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	// 创建关系
	_, err := session.Run(`
		MATCH (source:Company {id: $sourceID})
		MATCH (target:Company {id: $targetID})
		MERGE (source)-[r:RELATED_TO {type: $relationship}]->(target)
		SET r.weight = $weight, r.created_at = datetime()
	`, map[string]interface{}{
		"sourceID":     sourceID,
		"targetID":     targetID,
		"relationship": relationship,
		"weight":       weight,
	})

	return err
}

// GetCachedCompany 从Redis获取缓存的企业
func (s *CompanyDataSyncService) GetCachedCompany(companyID uint) (*EnhancedCompany, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("Redis未连接")
	}

	ctx := context.Background()
	key := fmt.Sprintf("company:%d", companyID)

	data, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var company EnhancedCompany
	if err := json.Unmarshal([]byte(data), &company); err != nil {
		return nil, err
	}

	return &company, nil
}

// AnalyzeCompanyData 分析企业数据
func (s *CompanyDataSyncService) AnalyzeCompanyData(companyID uint) (map[string]interface{}, error) {
	analysis := make(map[string]interface{})

	// 从MySQL获取基础统计
	var company EnhancedCompany
	if err := s.mysqlDB.First(&company, companyID).Error; err != nil {
		return nil, err
	}

	analysis["basic_stats"] = map[string]interface{}{
		"name":       company.Name,
		"industry":   company.Industry,
		"location":   company.Location,
		"status":     company.Status,
		"view_count": company.ViewCount,
	}

	// 从Neo4j获取关系分析
	if s.neo4jDriver != nil {
		relationships, err := s.GetCompanyRelationships(companyID)
		if err == nil {
			analysis["relationships"] = relationships
			analysis["relationship_count"] = len(relationships)
		}
	}

	// 从PostgreSQL获取职位数据
	if s.postgresDB != nil {
		var jobCount int64
		s.postgresDB.Model(&JobData{}).Where("company_id = ?", companyID).Count(&jobCount)
		analysis["job_count"] = jobCount
	}

	return analysis, nil
}

// GetLocationBasedRecommendations 获取基于地理位置的推荐
func (s *CompanyDataSyncService) GetLocationBasedRecommendations(companyID uint, radius float64, limit int) ([]EnhancedCompany, error) {
	// 获取目标企业的地理位置
	var targetCompany EnhancedCompany
	if err := s.mysqlDB.First(&targetCompany, companyID).Error; err != nil {
		return nil, fmt.Errorf("企业不存在: %v", err)
	}

	if targetCompany.BDLatitude == nil || targetCompany.BDLongitude == nil {
		return nil, fmt.Errorf("企业地理位置信息不完整")
	}

	// 查找附近的企业
	var companies []EnhancedCompany
	err := s.mysqlDB.Where(`
		bd_latitude IS NOT NULL 
		AND bd_longitude IS NOT NULL 
		AND id != ? 
		AND status = 'active'
	`, companyID).Find(&companies).Error

	if err != nil {
		return nil, err
	}

	// 计算距离并筛选
	var nearbyCompanies []EnhancedCompany
	for _, company := range companies {
		if company.BDLatitude != nil && company.BDLongitude != nil {
			distance := calculateDistance(
				*targetCompany.BDLatitude, *targetCompany.BDLongitude,
				*company.BDLatitude, *company.BDLongitude,
			)
			if distance <= radius {
				nearbyCompanies = append(nearbyCompanies, company)
			}
		}
	}

	// 限制返回数量
	if len(nearbyCompanies) > limit {
		nearbyCompanies = nearbyCompanies[:limit]
	}

	return nearbyCompanies, nil
}

// GetIndustryBasedRecommendations 获取基于行业关系的推荐
func (s *CompanyDataSyncService) GetIndustryBasedRecommendations(companyID uint, limit int) ([]EnhancedCompany, error) {
	// 获取目标企业的行业信息
	var targetCompany EnhancedCompany
	if err := s.mysqlDB.First(&targetCompany, companyID).Error; err != nil {
		return nil, fmt.Errorf("企业不存在: %v", err)
	}

	// 查找同行业的企业
	var companies []EnhancedCompany
	err := s.mysqlDB.Where(`
		industry = ? 
		AND id != ? 
		AND status = 'active'
	`, targetCompany.Industry, companyID).Limit(limit).Find(&companies).Error

	if err != nil {
		return nil, err
	}

	return companies, nil
}

// calculateDistance 计算两点间距离（公里）
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // 地球半径（公里）

	dLat := (lat2 - lat1) * 3.14159265359 / 180
	dLon := (lon2 - lon1) * 3.14159265359 / 180

	a := 0.5 - 0.5*math.Cos(dLat) +
		0.5*math.Cos(lat1*3.14159265359/180)*math.Cos(lat2*3.14159265359/180)*
			(1-math.Cos(dLon))

	return R * math.Asin(math.Sqrt(a))
}

// 辅助函数
func getFloat64Value(ptr *float64) float64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func getInt64Value(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}
