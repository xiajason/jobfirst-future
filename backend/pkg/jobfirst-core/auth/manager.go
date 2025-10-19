package auth

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthManager 认证管理器
type AuthManager struct {
	db     *gorm.DB
	config AuthConfig
}

// NewAuthManager 创建认证管理器
func NewAuthManager(db *gorm.DB, config AuthConfig) *AuthManager {
	return &AuthManager{
		db:     db,
		config: config,
	}
}

// Register 用户注册
func (am *AuthManager) Register(req RegisterRequest) (*RegisterResponse, error) {
	// 检查用户名和邮箱是否已存在
	var existingUser User
	if err := am.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名或邮箱已存在")
	}

	// 生成UUID
	userUUID := uuid.New().String()

	// 哈希密码
	passwordHash, err := am.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码哈希失败: %w", err)
	}

	// 创建用户
	user := User{
		UUID:          userUUID,
		Username:      req.Username,
		Email:         req.Email,
		PasswordHash:  passwordHash,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Phone:         req.Phone,
		Status:        "active",
		EmailVerified: false,
		PhoneVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := am.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return &RegisterResponse{
		Success: true,
		Message: "注册成功",
		User:    user,
	}, nil
}

// Login 用户登录
func (am *AuthManager) Login(req LoginRequest, clientIP, userAgent string) (*LoginResponse, error) {
	// 查找用户
	var user User
	if err := am.db.Where("username = ? AND status = 'active'", req.Username).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在或已被禁用")
	}

	// 验证密码
	if !am.validatePassword(req.Password, user.PasswordHash) {
		am.logLoginAttempt(user.ID, clientIP, userAgent, "failed", "密码错误")
		return nil, errors.New("密码错误")
	}

	// 生成JWT token
	token, expiresAt, err := am.generateToken(user.ID, user.Username, "user")
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}

	// 更新最后登录时间
	now := time.Now()
	am.db.Model(&user).Update("last_login_at", now)

	// 记录登录日志
	am.logLoginAttempt(user.ID, clientIP, userAgent, "success", "登录成功")

	response := &LoginResponse{
		Success:   true,
		Token:     token,
		User:      user,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		Message:   "登录成功",
	}

	// 检查是否为开发团队成员
	var devTeam DevTeamUser
	if err := am.db.Where("user_id = ? AND status = 'active'", user.ID).First(&devTeam).Error; err == nil {
		response.DevTeam = devTeam
	}

	return response, nil
}

// SuperAdminLogin 超级管理员登录
func (am *AuthManager) SuperAdminLogin(req LoginRequest, clientIP, userAgent string) (*LoginResponse, error) {
	// 查找用户
	var user User
	if err := am.db.Where("username = ? AND status = 'active'", req.Username).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在或已被禁用")
	}

	// 验证密码
	if !am.validatePassword(req.Password, user.PasswordHash) {
		am.logLoginAttempt(user.ID, clientIP, userAgent, "failed", "密码错误")
		return nil, errors.New("密码错误")
	}

	// 检查是否为超级管理员
	var devTeam DevTeamUser
	if err := am.db.Where("user_id = ? AND team_role = 'super_admin' AND status = 'active'", user.ID).First(&devTeam).Error; err != nil {
		return nil, errors.New("您不是超级管理员")
	}

	// 生成JWT token
	token, expiresAt, err := am.generateToken(user.ID, user.Username, "super_admin")
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}

	// 更新最后登录时间
	now := time.Now()
	am.db.Model(&user).Update("last_login_at", now)
	am.db.Model(&devTeam).Update("last_login_at", now)

	// 记录登录日志
	am.logLoginAttempt(user.ID, clientIP, userAgent, "success", "超级管理员登录成功")

	return &LoginResponse{
		Success:   true,
		Token:     token,
		User:      user,
		DevTeam:   devTeam,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		Message:   "超级管理员登录成功",
	}, nil
}

// ValidateToken 验证JWT token（支持跨云量子认证）
func (am *AuthManager) ValidateToken(tokenString string) (*Claims, error) {
	if len(tokenString) > 50 {
		log.Printf("DEBUG: 开始验证JWT token: %s...", tokenString[:50])
	} else {
		log.Printf("DEBUG: 开始验证JWT token: %s", tokenString)
	}
	log.Printf("DEBUG: 基础JWT secret: %s", am.config.JWTSecret)

	// 第一步：预解析Token（不验证签名）判断是否为量子Token
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	unverifiedToken, _, err := parser.ParseUnverified(tokenString, &Claims{})
	if err != nil {
		log.Printf("DEBUG: Token预解析失败: %v", err)
		return nil, fmt.Errorf("token格式错误: %w", err)
	}

	unverifiedClaims := unverifiedToken.Claims.(*Claims)

	// 第二步：根据Token类型选择密钥
	var signingKey string
	if unverifiedClaims.Quantum && unverifiedClaims.QSeed != "" {
		// 量子Token - 使用密钥增强
		enhancedKey := am.config.JWTSecret + unverifiedClaims.QSeed
		if len(enhancedKey) > 64 {
			signingKey = enhancedKey[:64]
		} else {
			signingKey = enhancedKey
		}
		log.Printf("✅ [Quantum Auth] 检测到量子Token, qseed=%s...", unverifiedClaims.QSeed[:8])
		log.Printf("🔑 [Quantum Auth] 使用增强密钥（长度: %d）", len(signingKey))
	} else {
		// 传统Token - 标准密钥（向后兼容）
		signingKey = am.config.JWTSecret
		log.Printf("🔑 [Standard Auth] 使用标准密钥")
	}

	// 第三步：使用正确的密钥验证Token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		log.Printf("DEBUG: JWT解析 - 算法: %v", token.Header["alg"])
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("DEBUG: JWT解析失败 - 不支持的签名方法: %v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(signingKey), nil
	})

	if err != nil {
		log.Printf("DEBUG: JWT验证失败: %v", err)
		return nil, fmt.Errorf("token验证失败: %w", err)
	}

	if !token.Valid {
		log.Printf("DEBUG: JWT token无效")
		return nil, errors.New("无效的token")
	}

	// 检查token是否过期（使用jwt.RegisteredClaims的ExpiresAt）
	if claims.ExpiresAt != nil {
		currentTime := time.Now()
		expiryTime := claims.ExpiresAt.Time
		log.Printf("DEBUG: 当前时间: %v, Token过期时间: %v", currentTime, expiryTime)
		if currentTime.After(expiryTime) {
			log.Printf("DEBUG: Token已过期")
			return nil, errors.New("token已过期")
		}
	}

	if claims.Quantum {
		log.Printf("✅ [Quantum Auth] 量子Token验证成功: %s (role: %s)", claims.Username, claims.Role)
	} else {
		log.Printf("✅ [Standard Auth] 标准Token验证成功: %s (role: %s)", claims.Username, claims.Role)
	}

	return claims, nil
}

// GetUserByID 根据ID获取用户
func (am *AuthManager) GetUserByID(userID uint) (*User, error) {
	var user User
	if err := am.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}
	return &user, nil
}

// GetDevTeamUser 获取开发团队成员信息
func (am *AuthManager) GetDevTeamUser(userID uint) (*DevTeamUser, error) {
	var devTeam DevTeamUser
	if err := am.db.Preload("User").Where("user_id = ? AND status = 'active'", userID).First(&devTeam).Error; err != nil {
		return nil, fmt.Errorf("不是开发团队成员: %w", err)
	}
	return &devTeam, nil
}

// CheckPermission 检查用户权限
func (am *AuthManager) CheckPermission(userID uint, requiredRole string) (bool, error) {
	var devTeam DevTeamUser
	if err := am.db.Where("user_id = ? AND status = 'active'", userID).First(&devTeam).Error; err != nil {
		return false, errors.New("用户不是开发团队成员")
	}

	// 超级管理员拥有所有权限
	if devTeam.TeamRole == "super_admin" {
		return true, nil
	}

	// 检查角色权限
	return devTeam.TeamRole == requiredRole, nil
}

// 辅助方法

// hashPassword 哈希密码
func (am *AuthManager) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// validatePassword 验证密码
func (am *AuthManager) validatePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateToken 生成JWT token（支持量子认证格式）
func (am *AuthManager) generateToken(userID uint, username, role string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(am.config.TokenExpiry)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		Quantum:  false, // 本地生成的是标准Token
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(am.config.JWTSecret))

	return tokenString, expiresAt, err
}

// logLoginAttempt 记录登录尝试
func (am *AuthManager) logLoginAttempt(userID uint, ipAddress, userAgent, status, message string) {
	log := DevOperationLog{
		UserID:           userID,
		OperationType:    "login_attempt",
		OperationTarget:  "auth",
		OperationDetails: fmt.Sprintf(`{"status": "%s", "message": "%s"}`, status, message),
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		Status:           status,
		CreatedAt:        time.Now(),
	}
	am.db.Create(&log)
}
