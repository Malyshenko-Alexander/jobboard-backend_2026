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
	"github.com/study/jobboard/employer-service/internal/authtoken"
	"github.com/study/jobboard/employer-service/internal/config"
	"github.com/study/jobboard/employer-service/internal/events"
	"github.com/study/jobboard/employer-service/internal/handler"
	"github.com/study/jobboard/employer-service/internal/repository"
	"github.com/study/jobboard/employer-service/internal/service"

	_ "github.com/study/jobboard/employer-service/docs"
)

// @title           Employer Service API
// @version         1.0
// @description     Employer cabinet and public company info microservice.
// @host            localhost:8083
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
	publisher, err := events.NewRabbitPublisher(cfg.RabbitURL)
	if err != nil {
		log.Fatalf("rabbit publisher: %v", err)
	}
	defer publisher.Close()

	profileRepo := repository.NewProfileRepository(pool)
	empSvc := service.NewEmployerService(profileRepo, publisher)

	consumer, err := events.NewRabbitConsumer(cfg.RabbitURL, empSvc)
	if err != nil {
		log.Fatalf("rabbit consumer: %v", err)
	}
	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("events consumer: %v", err)
	}
	defer consumer.Close()

	h := handler.NewEmployerHandler(empSvc, tokens)

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

	r.Get("/api/v1/employer/companies/{id}", h.GetCompany)

	r.Route("/api/v1/employer", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(h.AuthMiddleware)
			r.Get("/profile", h.GetProfile)
			r.Put("/profile", h.UpdateProfile)
		})
	})

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("employer-service listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Println("employer-service stopped")
	os.Exit(0)
}
