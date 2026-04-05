package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"github.com/developer-space/api/internal/config"
	"github.com/developer-space/api/internal/database"
	"github.com/developer-space/api/internal/handler"
	"github.com/developer-space/api/internal/middleware"
	"github.com/developer-space/api/internal/repository"
	"github.com/developer-space/api/internal/response"
	"github.com/developer-space/api/internal/service"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logLevel := parseLogLevel(cfg.LogLevel)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	// Handle CLI subcommands
	if len(os.Args) > 1 {
		runCLI(cfg, os.Args[1:])
		return
	}

	ctx := context.Background()

	// Run migrations before starting the server
	if err := database.RunMigrations(cfg.DatabaseURL, "migrations"); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	r := chi.NewRouter()

	r.Use(middleware.Recovery(logger))
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging(logger))
	r.Use(middleware.CORS(cfg.FrontendURL))
	r.Use(middleware.ContentType)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Wire up dependencies
	memberRepo := repository.NewMemberRepository(pool)
	tokenRepo := repository.NewMagicTokenRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	seriesRepo := repository.NewSeriesRepository(pool)
	rsvpRepo := repository.NewRSVPRepository(pool)

	emailSender := service.NewResendEmailSender(cfg.ResendAPIKey, cfg.ResendFromEmail)
	// Notifications disabled for now
	var notifier service.Notifier = &service.NoopNotifier{}
	slog.Info("all notifications disabled")
	memberSvc := service.NewMemberService(memberRepo, emailSender, cfg.FrontendURL)
	authSvc := service.NewAuthService(tokenRepo, memberRepo, emailSender, cfg.SessionSecret, cfg.FrontendURL, cfg.IsSecure())
	sessionSvc := service.NewSessionService(sessionRepo, notifier)
	sessionSvc.SetSeriesRepo(seriesRepo)
	// sessionSvc.SetEmailNotifier(emailSender, rsvpRepo) // notifications disabled for now
	rsvpSvc := service.NewRSVPService(rsvpRepo, memberRepo, notifier)

	// Ensure uploads directory exists
	uploadsDir := "uploads/sessions"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		logger.Error("failed to create uploads directory", "error", err)
		os.Exit(1)
	}

	memberHandler := handler.NewMemberHandler(memberSvc)
	authHandler := handler.NewAuthHandler(authSvc)
	sessionHandler := handler.NewSessionHandler(sessionSvc)
	rsvpHandler := handler.NewRSVPHandler(rsvpSvc)
	profileHandler := handler.NewProfileHandler(authSvc)
	imageHandler := handler.NewImageHandler(sessionSvc, uploadsDir)
	skillsHandler := handler.NewSkillsHandler(memberSvc)

	handler.RegisterRoutes(r, memberHandler, authHandler, sessionHandler, rsvpHandler, profileHandler, imageHandler, skillsHandler, authSvc, memberRepo)

	// Serve uploaded files
	fileServer := http.FileServer(http.Dir("."))
	r.Handle("/uploads/*", fileServer)

	// Background token cleanup every hour
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			authSvc.CleanExpiredTokens(context.Background())
		}
	}()

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	logger.Info("server shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}

	logger.Info("server stopped")
}

func runCLI(cfg *config.Config, args []string) {
	switch args[0] {
	case "migrate":
		runMigrate(cfg, args[1:])
	case "seed-admin":
		runSeedAdmin(cfg, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		fmt.Fprintln(os.Stderr, "usage: api [migrate up|migrate down N|seed-admin --email EMAIL --name NAME]")
		os.Exit(1)
	}
}

func runMigrate(cfg *config.Config, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: api migrate [up|down N]")
		os.Exit(1)
	}

	switch args[0] {
	case "up":
		if err := database.RunMigrations(cfg.DatabaseURL, "migrations"); err != nil {
			slog.Error("migration up failed", "error", err)
			os.Exit(1)
		}
		fmt.Println("migrations applied successfully")
	case "down":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: api migrate down N")
			os.Exit(1)
		}
		steps, err := strconv.Atoi(args[1])
		if err != nil || steps < 1 {
			fmt.Fprintln(os.Stderr, "N must be a positive integer")
			os.Exit(1)
		}
		if err := database.MigrateDown(cfg.DatabaseURL, "migrations", steps); err != nil {
			slog.Error("migration down failed", "error", err)
			os.Exit(1)
		}
		fmt.Printf("rolled back %d migration(s) successfully\n", steps)
	default:
		fmt.Fprintf(os.Stderr, "unknown migrate command: %s\n", args[0])
		os.Exit(1)
	}
}

func runSeedAdmin(cfg *config.Config, args []string) {
	var email, name string
	for i := 0; i < len(args)-1; i++ {
		switch args[i] {
		case "--email":
			email = args[i+1]
			i++
		case "--name":
			name = args[i+1]
			i++
		}
	}

	if email == "" || name == "" {
		fmt.Fprintln(os.Stderr, "usage: api seed-admin --email EMAIL --name NAME")
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx,
		`INSERT INTO members (email, name, is_admin, is_active)
		 VALUES ($1, $2, true, true)
		 ON CONFLICT (email) DO UPDATE SET is_admin = true, updated_at = now()`,
		email, name,
	)
	if err != nil {
		slog.Error("failed to seed admin", "error", err)
		os.Exit(1)
	}

	fmt.Printf("admin member seeded: %s <%s>\n", name, email)
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
