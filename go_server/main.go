package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"electric_ai_tool/go_server/config"
	"electric_ai_tool/go_server/domain"
	"electric_ai_tool/go_server/handlers"
	"electric_ai_tool/go_server/middleware"
	"electric_ai_tool/go_server/services"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

func main() {
	execDir, _ := os.Getwd()
	godotenv.Load(filepath.Join(execDir, ".env"))
	godotenv.Load(filepath.Join(execDir, "../web/.env"))

	// Initialize logger with rotation
	logDir := filepath.Join(execDir, "logs")
	if err := config.InitLogger(logDir); err != nil {
		log.Printf("⚠️  Failed to initialize logger: %v", err)
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY not set in environment")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}

	if err := config.InitDatabase(); err != nil {
		log.Printf("⚠️  Database connection failed: %v", err)
		log.Println("Continuing without database...")
	}
	defer config.CloseDatabase()

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	// Initialize services
	aiService := services.NewAIService(client)
	multiModelService := services.NewMultiModelService(client)
	authService := services.NewAuthService()
	emailService := services.NewEmailService()
	taskService := services.NewTaskService()
	taskHistoryService := services.NewTaskHistoryService()
	cdnService := services.NewCDNService()
	copywritingService := services.NewCopywritingService(multiModelService)
	rbacService := services.NewRBACService()
	unifiedTaskService := services.NewUnifiedTaskService()
	
	// 新任务中心相关服务
	taskCenterService := services.NewTaskCenterService()
	copywritingTaskService := services.NewCopywritingTaskService()
	// imageTaskService := services.NewImageTaskService() // TODO: 图片生成也需要迁移

	// Initialize RBAC
	if err := rbacService.InitializeDefaultRolesAndPermissions(); err != nil {
		log.Printf("⚠️  Failed to initialize RBAC: %v", err)
	}

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Health check
	handler := handlers.NewHandler(aiService)
	http.HandleFunc("/api/health", middleware.LoggingMiddleware(middleware.CORS(handler.Health)))

	// Legacy API routes (to be removed later)
	http.HandleFunc("/api/analyze", middleware.LoggingMiddleware(middleware.CORS(handler.Analyze)))
	http.HandleFunc("/api/generate-image", middleware.LoggingMiddleware(middleware.CORS(handler.GenerateImage)))
	http.HandleFunc("/api/edit-image", middleware.LoggingMiddleware(middleware.CORS(handler.EditImage)))
	http.HandleFunc("/api/aplus-content", middleware.LoggingMiddleware(middleware.CORS(handler.APlusContent)))

	// Register domain routes (DDD pattern)
	authDomain := domain.NewAuthDomain(authService, emailService)
	authDomain.RegisterRoutes(authMiddleware)

	taskDomain := domain.NewTaskDomain(multiModelService, taskService, taskHistoryService, cdnService, authService, unifiedTaskService)
	taskDomain.RegisterRoutes(authMiddleware)

	modelDomain := domain.NewModelDomain(multiModelService)
	modelDomain.RegisterRoutes(authMiddleware)

	copywritingDomain := domain.NewCopywritingDomain(copywritingService, authService, unifiedTaskService, taskCenterService, copywritingTaskService)
	copywritingDomain.RegisterRoutes(authMiddleware)

	// 任务中心（新架构）
	taskCenterDomain := domain.NewTaskCenterDomain(taskCenterService, authService)
	taskCenterDomain.RegisterRoutes(authMiddleware)

	// 统一任务管理（旧架构，待废弃）
	unifiedTaskDomain := domain.NewUnifiedTaskDomain(unifiedTaskService, authService)
	unifiedTaskDomain.RegisterRoutes(authMiddleware)

	// Static file serving
	distPath := filepath.Join(execDir, "../web/dist")
	if _, err := os.Stat(distPath); err == nil {
		fs := http.FileServer(http.Dir(distPath))
		http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			if _, err := os.Stat(filepath.Join(distPath, r.URL.Path)); os.IsNotExist(err) {
				http.ServeFile(w, r, filepath.Join(distPath, "index.html"))
				return
			}
			fs.ServeHTTP(w, r)
		}))
		log.Printf("✅ 后端服务已启动（生产模式），监听端口: %s\n", port)
	} else {
		log.Printf("✅ 后端服务已启动（开发模式），监听端口: %s\n", port)
	}

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
