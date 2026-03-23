package domain

import (
	"database/sql"
	"electric_ai_tool/go_server/handlers"
	"electric_ai_tool/go_server/middleware"
	"net/http"
)

type AdminDomain struct {
	db *sql.DB
}

func NewAdminDomain(db *sql.DB) *AdminDomain {
	return &AdminDomain{
		db: db,
	}
}

func (d *AdminDomain) RegisterRoutes() {
	adminHandler := handlers.NewAdminHandler(d.db)

	// 创建admin中间件
	adminMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return handlers.AdminMiddleware(d.db, next)
	}

	// 用户管理
	http.HandleFunc("/api/admin/users", 
		middleware.LoggingMiddleware(middleware.CORS(adminMiddleware(adminHandler.GetUsers))))
	http.HandleFunc("/api/admin/users/approve", 
		middleware.LoggingMiddleware(middleware.CORS(adminMiddleware(adminHandler.ApproveUser))))
	http.HandleFunc("/api/admin/users/reject", 
		middleware.LoggingMiddleware(middleware.CORS(adminMiddleware(adminHandler.RejectUser))))

	// 角色管理
	http.HandleFunc("/api/admin/roles", 
		middleware.LoggingMiddleware(middleware.CORS(adminMiddleware(adminHandler.GetRoles))))

	// 权限管理
	http.HandleFunc("/api/admin/permissions", 
		middleware.LoggingMiddleware(middleware.CORS(adminMiddleware(adminHandler.GetPermissions))))

	// 角色权限管理
	http.HandleFunc("/api/admin/role-permissions", 
		middleware.LoggingMiddleware(middleware.CORS(adminMiddleware(adminHandler.GetRolePermissions))))
}
