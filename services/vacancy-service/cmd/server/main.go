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
	"github.com/study/jobboard/vacancy-service/internal/authtoken"
	"github.com/study/jobboard/vacancy-service/internal/client"
	"github.com/study/jobboard/vacancy-service/internal/config"
	"github.com/study/jobboard/vacancy-service/internal/events"
	"github.com/study/jobboard/vacancy-service/internal/handler"
	"github.com/study/jobboard/vacancy-service/internal/repository"
	"github.com/study/jobboard/vacancy-service/internal/service"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/study/jobboard/vacancy-service/docs"
)

// @title           Vacancy Service API
// @version         1.0
// @description     Vacancy search, details and employer CRUD microservice.
// @host            localhost:8084
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

	repo := repository.NewVacancyRepository(pool)
	employerClient := client.NewEmployerClient(cfg.EmployerServiceURL)
	vacSvc := service.NewVacancyService(repo, employerClient, publisher)

	consumer, err := events.NewRabbitConsumer(cfg.RabbitURL, vacSvc)
	if err != nil {
		log.Fatalf("rabbit consumer: %v", err)
	}
	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("events consumer: %v", err)
	}
	defer consumer.Close()

	h := handler.NewVacancyHandler(vacSvc, tokens)

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

	r.Get("/api/v1/vacancies", h.ListVacancies)

	r.Group(func(r chi.Router) {
		r.Use(h.AuthMiddleware)
		r.Get("/api/v1/vacancies/my", h.ListMyVacancies)
		r.Post("/api/v1/vacancies", h.CreateVacancy)
		r.Put("/api/v1/vacancies/{id}", h.UpdateVacancy)
		r.Delete("/api/v1/vacancies/{id}", h.DeleteVacancy)
	})

	r.Get("/api/v1/vacancies/{id}", h.GetVacancy)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("vacancy-service listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Println("vacancy-service stopped")
	os.Exit(0)
}
