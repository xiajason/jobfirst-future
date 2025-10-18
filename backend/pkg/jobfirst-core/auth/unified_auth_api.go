package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// UnifiedAuthAPI 统一认证API服务器
type UnifiedAuthAPI struct {
	authSystem *UnifiedAuthSystem
	port       int
}

// NewUnifiedAuthAPI 创建统一认证API服务器
func NewUnifiedAuthAPI(authSystem *UnifiedAuthSystem, port int) *UnifiedAuthAPI {
	return &UnifiedAuthAPI{
		authSystem: authSystem,
		port:       port,
	}
}

// Start 启动API服务器
func (api *UnifiedAuthAPI) Start() error {
	// 设置路由
	http.HandleFunc("/api/v1/auth/login", api.handleLogin)
	http.HandleFunc("/api/v1/auth/validate", api.handleValidateJWT)
	http.HandleFunc("/api/v1/auth/permission", api.handleCheckPermission)
	http.HandleFunc("/api/v1/auth/user", api.handleGetUser)
	http.HandleFunc("/api/v1/auth/access", api.handleValidateAccess)
	http.HandleFunc("/api/v1/auth/log", api.handleLogAccess)
	http.HandleFunc("/api/v1/auth/roles", api.handleGetRoles)
	http.HandleFunc("/api/v1/auth/permissions", api.handleGetPermissions)
	http.HandleFunc("/health", api.handleHealth)

	// 启动服务器
	addr := fmt.Sprintf(":%d", api.port)
	fmt.Printf("统一认证服务启动在端口 %d\n", api.port)
	return http.ListenAndServe(addr, nil)
}

// handleLogin 处理登录请求
func (api *UnifiedAuthAPI) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	result, err := api.authSystem.Authenticate(req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 记录访问日志
	api.authSystem.logAccess(0, "login", "auth",
		map[bool]string{true: "success", false: "failed"}[result.Success],
		getClientIP(r), getUserAgent(r))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleValidateJWT 处理JWT验证请求
func (api *UnifiedAuthAPI) handleValidateJWT(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	result, err := api.authSystem.ValidateJWT(req.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleCheckPermission 处理权限检查请求
func (api *UnifiedAuthAPI) handleCheckPermission(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	permission := r.URL.Query().Get("permission")

	if userIDStr == "" || permission == "" {
		http.Error(w, "user_id and permission are required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	hasPermission, err := api.authSystem.CheckPermission(userID, permission)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user_id":        userID,
		"permission":     permission,
		"has_permission": hasPermission,
		"timestamp":      time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetUser 处理获取用户信息请求
func (api *UnifiedAuthAPI) handleGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	user, err := api.authSystem.getUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// handleValidateAccess 处理访问验证请求
func (api *UnifiedAuthAPI) handleValidateAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID   int    `json:"user_id"`
		Resource string `json:"resource"`
		Action   string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.UserID == 0 || req.Resource == "" || req.Action == "" {
		http.Error(w, "user_id, resource, and action are required", http.StatusBadRequest)
		return
	}

	// 构建权限字符串
	permission := fmt.Sprintf("%s:%s", req.Action, req.Resource)
	hasPermission, err := api.authSystem.CheckPermission(req.UserID, permission)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user_id":        req.UserID,
		"resource":       req.Resource,
		"action":         req.Action,
		"permission":     permission,
		"has_permission": hasPermission,
		"timestamp":      time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleLogAccess 处理访问日志记录请求
func (api *UnifiedAuthAPI) handleLogAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID    int    `json:"user_id"`
		Action    string `json:"action"`
		Resource  string `json:"resource"`
		Result    string `json:"result"`
		IPAddress string `json:"ip_address"`
		UserAgent string `json:"user_agent"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.UserID == 0 || req.Action == "" || req.Resource == "" || req.Result == "" {
		http.Error(w, "user_id, action, resource, and result are required", http.StatusBadRequest)
		return
	}

	api.authSystem.logAccess(req.UserID, req.Action, req.Resource, req.Result, req.IPAddress, req.UserAgent)

	response := map[string]interface{}{
		"success":   true,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetRoles 处理获取角色列表请求
func (api *UnifiedAuthAPI) handleGetRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.authSystem.roleConfig.Roles)
}

// handleGetPermissions 处理获取权限列表请求
func (api *UnifiedAuthAPI) handleGetPermissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role := r.URL.Query().Get("role")
	if role == "" {
		http.Error(w, "role is required", http.StatusBadRequest)
		return
	}

	permissions, err := api.authSystem.getUserPermissions(role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"role":        role,
		"permissions": permissions,
		"timestamp":   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealth 处理健康检查请求
func (api *UnifiedAuthAPI) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "unified-auth-service",
		"timestamp": time.Now(),
		"version":   "2.0.0",
		"features": []string{
			"unified_role_system",
			"complete_jwt_validation",
			"permission_management",
			"access_logging",
			"database_optimization",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 辅助函数
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

func getUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}
