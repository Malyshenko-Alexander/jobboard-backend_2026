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
	"github.com/study/jobboard/applicant-service/internal/authtoken"
	"github.com/study/jobboard/applicant-service/internal/config"
	"github.com/study/jobboard/applicant-service/internal/events"
	"github.com/study/jobboard/applicant-service/internal/handler"
	"github.com/study/jobboard/applicant-service/internal/repository"
	"github.com/study/jobboard/applicant-service/internal/service"

	_ "github.com/study/jobboard/applicant-service/docs"
)

// @title           Applicant Service API
// @version         1.0
// @description     Applicant cabinet microservice (profile + resume).
// @host            localhost:8082
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

	tokens := authtoken.NewManager(cfg.JWTSecret)
	profileRepo := repository.NewProfileRepository(pool)
	resumeRepo := repository.NewResumeRepository(pool)
	appSvc := service.NewApplicantService(profileRepo, resumeRepo)

	// Stub consumer: no RabbitMQ yet. HTTP hook mimics the message handler.
	consumer := events.NewStubConsumer(appSvc)
	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("events consumer: %v", err)
	}
	defer consumer.Close()

	h := handler.NewApplicantHandler(appSvc, tokens, consumer)

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

	r.Route("/api/v1/applicant", func(r chi.Router) {
		r.Use(h.AuthMiddleware)
		r.Get("/profile", h.GetProfile)
		r.Put("/profile", h.UpdateProfile)
		r.Get("/resume", h.GetResume)
		r.Put("/resume", h.UpsertResume)
	})

	// Temporary stand-in for RabbitMQ until broker is connected.
	r.Post("/api/v1/internal/events/user-created", h.UserCreatedHook)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("applicant-service listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Println("applicant-service stopped")
	os.Exit(0)
}
