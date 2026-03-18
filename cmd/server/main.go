package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/blueprintautomation/blueprint-chat/internal/api"
	"github.com/blueprintautomation/blueprint-chat/internal/api/handlers"
	"github.com/blueprintautomation/blueprint-chat/internal/billing"
	mw "github.com/blueprintautomation/blueprint-chat/internal/api/middleware"
	"github.com/blueprintautomation/blueprint-chat/internal/chat"
	"github.com/blueprintautomation/blueprint-chat/internal/config"
	"github.com/blueprintautomation/blueprint-chat/internal/db"
	"github.com/blueprintautomation/blueprint-chat/internal/knowledge"
	"github.com/blueprintautomation/blueprint-chat/internal/leads"
	"github.com/blueprintautomation/blueprint-chat/internal/notification"
	"github.com/blueprintautomation/blueprint-chat/internal/tenant"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Handle migrate subcommand: ./blueprint-chat migrate up|down
	if len(os.Args) >= 3 && os.Args[1] == "migrate" {
		runMigrations(os.Args[2])
		return
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := newLogger(cfg.LogLevel)
	defer log.Sync()

	log.Info("starting Blueprint Chat",
		zap.String("env", cfg.Env),
		zap.String("port", cfg.Port),
	)

	// Background context for startup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to PostgreSQL
	pgPool, err := db.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("connect to postgres", zap.Error(err))
	}
	defer pgPool.Close()
	log.Info("connected to postgres")

	// Connect to Redis
	redisClient, err := db.NewRedisClient(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatal("connect to redis", zap.Error(err))
	}
	defer redisClient.Close()
	log.Info("connected to redis")

	// Initialize Anthropic client
	anthropicClient := anthropic.NewClient(
		option.WithAPIKey(cfg.AnthropicAPIKey),
	)

	// Initialize services
	sessionMgr := chat.NewSessionManager(redisClient, cfg.RedisSessionTTL)
	assembler := knowledge.NewAssembler(redisClient, cfg.RedisConfigCacheTTL)
	chatEngine := chat.NewEngine(
		anthropicClient,
		sessionMgr,
		assembler,
		cfg.AnthropicModel,
		cfg.AnthropicMaxTokens,
		log,
	)

	tenantSvc := tenant.NewService(pgPool, cfg.EncryptionKey, log)
	tenantProv := tenant.NewProvisioner(tenantSvc, cfg.BaseURL)
	billingClient := billing.NewTenantLimitClient(cfg.PortalsURL, cfg.BPAInternalKey)

	emailSender := notification.NewEmailSender(
		cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass,
		cfg.NotificationFromEmail, cfg.NotificationFromName, log,
	)
	discordSender := notification.NewDiscordSender(log)

	leadNotifier := leads.NewLeadNotifier(log, emailSender, discordSender)
	leadSvc := leads.NewService(pgPool, leadNotifier, log)

	rateLimiter := mw.NewRateLimiter(redisClient, log)

	// Initialize handlers
	tenantsHandler := handlers.NewTenantsHandler(tenantSvc, tenantProv, log, billingClient)
	chatHandler := handlers.NewChatHandler(chatEngine, cfg.EncryptionKey, log)
	leadsHandler := handlers.NewLeadsHandler(leadSvc, log)
	bookingHandler := handlers.NewBookingHandler(cfg.CalComAPIBase, cfg.CalendlyAPIBase, cfg.EncryptionKey, log)
	widgetHandler := handlers.NewWidgetHandler(cfg.WidgetBundlePath, log)
	healthHandler := handlers.NewHealthHandler(pgPool, redisClient)

	// Build router
	router := api.NewRouter(api.RouterDeps{
		TenantService:   tenantSvc,
		TenantHandler:   tenantsHandler,
		ChatHandler:     chatHandler,
		LeadsHandler:    leadsHandler,
		BookingHandler:  bookingHandler,
		WidgetHandler:   widgetHandler,
		HealthHandler:   healthHandler,
		RateLimiter:     rateLimiter,
		AdminKey:        cfg.AdminKey,
		RateLimitPerMin: cfg.RateLimitPerMin,
		Log:             log,
	})

	// Start HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second, // Long for SSE streams
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	<-quit
	log.Info("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown error", zap.Error(err))
	}

	log.Info("server stopped")
}

func runMigrations(direction string) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL not set")
		os.Exit(1)
	}

	m, err := migrate.New("file://internal/db/migrations", dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migration setup error: %v\n", err)
		os.Exit(1)
	}
	defer m.Close()

	switch direction {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			fmt.Fprintf(os.Stderr, "migration up error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Migrations applied")
	case "down":
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			fmt.Fprintf(os.Stderr, "migration down error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Migration rolled back")
	default:
		fmt.Fprintf(os.Stderr, "unknown direction: %s (use 'up' or 'down')\n", direction)
		os.Exit(1)
	}
}

func newLogger(level string) *zap.Logger {
	lvl := zapcore.InfoLevel
	switch level {
	case "debug":
		lvl = zapcore.DebugLevel
	case "warn":
		lvl = zapcore.WarnLevel
	case "error":
		lvl = zapcore.ErrorLevel
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	log, _ := cfg.Build()
	return log
}
