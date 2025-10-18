package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// APIServer 认证API服务器
type APIServer struct {
	manager *Manager
	port    int
}

// NewAPIServer 创建认证API服务器
func NewAPIServer(manager *Manager, port int) *APIServer {
	return &APIServer{
		manager: manager,
		port:    port,
	}
}

// Start 启动API服务器
func (s *APIServer) Start() error {
	http.HandleFunc("/api/v1/auth/validate", s.handleValidateJWT)
	http.HandleFunc("/api/v1/auth/permission", s.handleCheckPermission)
	http.HandleFunc("/api/v1/auth/quota", s.handleCheckQuota)
	http.HandleFunc("/api/v1/auth/user", s.handleGetUser)
	http.HandleFunc("/api/v1/auth/access", s.handleValidateAccess)
	http.HandleFunc("/api/v1/auth/log", s.handleLogAccess)
	http.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("认证服务启动在端口 %d\n", s.port)
	return http.ListenAndServe(addr, nil)
}

// handleValidateJWT 处理JWT验证请求
func (s *APIServer) handleValidateJWT(w http.ResponseWriter, r *http.Request) {
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

	result, err := s.manager.ValidateJWT(req.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleCheckPermission 处理权限检查请求
func (s *APIServer) handleCheckPermission(w http.ResponseWriter, r *http.Request) {
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

	hasPermission, err := s.manager.CheckPermission(userID, permission)
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

// handleCheckQuota 处理配额检查请求
func (s *APIServer) handleCheckQuota(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	resourceType := r.URL.Query().Get("resource_type")

	if userIDStr == "" || resourceType == "" {
		http.Error(w, "user_id and resource_type are required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	quota, err := s.manager.CheckQuota(userID, resourceType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quota)
}

// handleGetUser 处理获取用户信息请求
func (s *APIServer) handleGetUser(w http.ResponseWriter, r *http.Request) {
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

	user, err := s.manager.GetUserInfo(userID)
	if err != nil {
		if err.Error() == "用户不存在" {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// handleValidateAccess 处理访问验证请求（用于AI服务）
func (s *APIServer) handleValidateAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID   int    `json:"user_id"`
		Resource string `json:"resource"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.UserID == 0 || req.Resource == "" {
		http.Error(w, "user_id and resource are required", http.StatusBadRequest)
		return
	}

	result, err := s.manager.ValidateUserAccess(req.UserID, req.Resource)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleLogAccess 处理访问日志记录请求
func (s *APIServer) handleLogAccess(w http.ResponseWriter, r *http.Request) {
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

	err := s.manager.LogAccess(req.UserID, req.Action, req.Resource, req.Result, req.IPAddress, req.UserAgent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealth 处理健康检查请求
func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "auth-service",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CORS中间件
func (s *APIServer) enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
}

// 获取客户端IP地址
func getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 检查X-Real-IP头
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用RemoteAddr
	ip := strings.Split(r.RemoteAddr, ":")[0]
	return ip
}

// 获取User-Agent
func getUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}
