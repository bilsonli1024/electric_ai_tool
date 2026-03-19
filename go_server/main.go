package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"electric_ai_tool/go_server/config"
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

	aiService := services.NewAIService(client)
	authService := services.NewAuthService()
	taskService := services.NewTaskService()
	taskHistoryService := services.NewTaskHistoryService()
	cdnService := services.NewCDNService()

	handler := handlers.NewHandler(aiService)
	authHandler := handlers.NewAuthHandler(authService)
	taskHandler := handlers.NewTaskHandler(aiService, taskService, taskHistoryService, cdnService, authService)

	authMiddleware := middleware.NewAuthMiddleware(authService)

	http.HandleFunc("/api/health", middleware.CORS(handler.Health))

	http.HandleFunc("/api/auth/register", middleware.CORS(authHandler.Register))
	http.HandleFunc("/api/auth/login", middleware.CORS(authHandler.Login))
	http.HandleFunc("/api/auth/logout", middleware.CORS(authMiddleware.RequireAuth(authHandler.Logout)))
	http.HandleFunc("/api/auth/me", middleware.CORS(authMiddleware.RequireAuth(authHandler.Me)))

	http.HandleFunc("/api/analyze", middleware.CORS(handler.Analyze))
	http.HandleFunc("/api/generate-image", middleware.CORS(handler.GenerateImage))
	http.HandleFunc("/api/edit-image", middleware.CORS(handler.EditImage))
	http.HandleFunc("/api/aplus-content", middleware.CORS(handler.APlusContent))

	http.HandleFunc("/api/tasks/analyze", middleware.CORS(authMiddleware.RequireAuth(taskHandler.AnalyzeWithTask)))
	http.HandleFunc("/api/tasks/generate-image", middleware.CORS(authMiddleware.RequireAuth(taskHandler.GenerateImageWithTask)))
	http.HandleFunc("/api/tasks", middleware.CORS(authMiddleware.RequireAuth(taskHandler.GetTasks)))
	http.HandleFunc("/api/tasks/all", middleware.CORS(authMiddleware.RequireAuth(taskHandler.GetAllTasks)))
	http.HandleFunc("/api/tasks/history", middleware.CORS(authMiddleware.RequireAuth(taskHandler.GetTaskHistory)))

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
