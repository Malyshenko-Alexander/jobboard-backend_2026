package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/study/jobboard/auth-service/internal/authtoken"
	"github.com/study/jobboard/auth-service/internal/config"
	"github.com/study/jobboard/auth-service/internal/events"
	"github.com/study/jobboard/auth-service/internal/handler"
	"github.com/study/jobboard/auth-service/internal/repository"
	"github.com/study/jobboard/auth-service/internal/service"

	_ "github.com/study/jobboard/auth-service/docs"
)

// @title           Auth Service API
// @version         1.0
// @description     Authentication and authorization microservice for job board.
// @host            localhost:8081
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT.
func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	tokens := authtoken.NewManager(cfg.JWTSecret, cfg.TokenTTL())
	// Stub for now: logs user.created instead of publishing to RabbitMQ.
	publisher := events.NewStubPublisher()
	defer publisher.Close()

	userRepo := repository.NewUserRepository(pool)
	authSvc := service.NewAuthService(userRepo, tokens, publisher)
	authHandler := handler.NewAuthHandler(authSvc, tokens)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Group(func(r chi.Router) {
			r.Use(authHandler.AuthMiddleware)
			r.Get("/me", authHandler.Me)
		})
	})

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("auth-service listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Println("auth-service stopped")
	os.Exit(0)
}
