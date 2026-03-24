package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"gpilot/internal/adapter/datasource/kubernetes"
	"gpilot/internal/adapter/datasource/loki"
	llmadapter "gpilot/internal/adapter/llm"
	"gpilot/internal/adapter/handler"
	"gpilot/internal/adapter/repository/postgres"
	rediscache "gpilot/internal/adapter/repository/redis"
	"gpilot/internal/app"
	"gpilot/internal/domain/alert"
	"gpilot/internal/infra/cache"
	"gpilot/internal/infra/config"
	"gpilot/internal/infra/database"
	"gpilot/internal/infra/logger"
	ws "gpilot/internal/infra/websocket"
)

func main() {
	// Determine config path
	configPath := os.Getenv("GPILOT_CONFIG")
	if configPath == "" {
		exe, _ := os.Executable()
		configPath = filepath.Join(filepath.Dir(exe), "..", "..", "config", "config.yaml")
		// Fallback to relative path for dev
		if _, err := os.Stat(configPath); err != nil {
			configPath = "config/config.yaml"
		}
	}

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Init logger
	logger.Init(cfg.Server.Mode)
	defer logger.Sync()

	logger.L.Info("starting gPilot server...")

	// Init PostgreSQL
	pool, err := database.NewPostgresPool(cfg.Database)
	if err != nil {
		logger.L.Fatalw("failed to connect to database", "error", err)
	}
	defer pool.Close()
	logger.L.Info("connected to PostgreSQL")

	// Run migrations
	migrationsDir := os.Getenv("GPILOT_MIGRATIONS")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	if err := postgres.RunMigrations(pool, migrationsDir); err != nil {
		logger.L.Fatalw("failed to run migrations", "error", err)
	}
	logger.L.Info("database migrations completed")

	// Init Redis
	redisClient, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		logger.L.Warnw("redis not available, running without cache", "error", err)
	} else {
		logger.L.Info("connected to Redis")
	}

	// Init repositories
	alertRepo := postgres.NewAlertRepo(pool)
	groupRepo := postgres.NewAlertGroupRepo(pool)
	analysisRepo := postgres.NewAnalysisRepo(pool)
	eventRepo := postgres.NewEventRepo(pool)

	// Init cache
	var alertCache alert.Cache
	if redisClient != nil {
		alertCache = rediscache.NewAlertCache(redisClient)
	}

	// Init WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// Init LLM client
	llmClient := llmadapter.NewClient(cfg.LLM)

	// Init datasource clients
	lokiClient := loki.NewClient(cfg.Datasources.Loki.URL)

	// Init K8s client (non-fatal if unavailable)
	k8sClient, _ := kubernetes.NewClient(cfg.Datasources.Kubernetes.Kubeconfig, eventRepo)
	if k8sClient != nil && k8sClient.IsConnected() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go k8sClient.WatchEvents(ctx)
		logger.L.Info("K8s event watcher started")
	}

	// Build alert pipeline
	var processors []alert.Processor
	if alertCache != nil {
		processors = append(processors, alert.NewDeduplicateProcessor(alertCache, alertRepo))
	}
	processors = append(processors,
		alert.NewCorrelateProcessor(groupRepo),
		alert.NewPersistProcessor(alertRepo, alertCache),
		alert.NewNotifyProcessor(func(alerts []*alert.Alert) {
			// Notification is handled in AlertApp
		}),
	)
	pipeline := alert.NewPipeline(processors...)

	// Init application services
	alertApp := app.NewAlertApp(pipeline, alertRepo, groupRepo, analysisRepo, llmClient, hub)
	analysisApp := app.NewAnalysisApp(analysisRepo, alertRepo, groupRepo, eventRepo, llmClient, hub)
	logApp := app.NewLogApp(lokiClient, llmClient)
	dashboardApp := app.NewDashboardApp(alertRepo, groupRepo, analysisRepo)

	// Setup HTTP router
	if cfg.Server.Mode == "release" {
		// gin.SetMode(gin.ReleaseMode) - handled in router
	}
	router := handler.NewRouter(alertApp, analysisApp, logApp, dashboardApp, hub)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		logger.L.Infow("gPilot API server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L.Fatalw("server error", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.L.Info("shutting down gPilot server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.L.Fatalw("server forced shutdown", "error", err)
	}

	logger.L.Info("gPilot server stopped")
}
