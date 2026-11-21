package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/darkline28/ski-resort/backend/internal/admin"
	adminapi "github.com/darkline28/ski-resort/backend/internal/admin/api"
	"github.com/darkline28/ski-resort/backend/internal/auth"
	authapi "github.com/darkline28/ski-resort/backend/internal/auth/api"
	"github.com/darkline28/ski-resort/backend/internal/bookings"
	bookingsapi "github.com/darkline28/ski-resort/backend/internal/bookings/api"
	"github.com/darkline28/ski-resort/backend/internal/catalog"
	catalogapi "github.com/darkline28/ski-resort/backend/internal/catalog/api"
	"github.com/darkline28/ski-resort/backend/internal/database"
	"github.com/darkline28/ski-resort/backend/internal/instructors"
	instructorsapi "github.com/darkline28/ski-resort/backend/internal/instructors/api"
	"github.com/darkline28/ski-resort/backend/internal/middleware"
	"github.com/darkline28/ski-resort/backend/internal/rental"
	rentalapi "github.com/darkline28/ski-resort/backend/internal/rental/api"
	"github.com/darkline28/ski-resort/backend/internal/services"
	servicesapi "github.com/darkline28/ski-resort/backend/internal/services/api"
	"github.com/darkline28/ski-resort/backend/internal/users"
	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load("../.env")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	port := os.Getenv("POSTGRES_PORT")
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	jwtTTL := os.Getenv("JWT_TTL")

	// dsn := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", user, password, port, dbName)
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbName,
	)
	ctx := context.Background()
	pool, err := database.NewPool(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()

	tokenTTL, err := time.ParseDuration(jwtTTL)
	if err != nil {
		log.Fatal("invalid JWT_TTL: ", err)
	}
	if tokenTTL <= 0 {
		log.Fatal("JWT_TTL must be positive")
	}
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 bytes")
	}

	userRepo := users.NewPostgresRepository(pool)
	authService := auth.NewService(userRepo, []byte(jwtSecret), tokenTTL)
	authHandler := auth.NewHandler(authService)

	catalogRepo := catalog.NewPostgresRepository(pool)
	catalogService := catalog.NewEquipmentService(catalogRepo)
	catalogHandler := catalog.NewHandler(catalogService)

	instructorRepo := instructors.NewPostgresRepository(pool)
	instructorsService := instructors.NewInstructorService(instructorRepo)
	instructorsHandler := instructors.NewHandler(instructorsService)

	serviceRepo := services.NewPostgresRepository(pool)
	serviceService := services.NewServiceService(serviceRepo)
	serviceHandler := services.NewHandler(serviceService)

	bookingsRepo := bookings.NewPostgresRepository(pool)
	bookingsService := bookings.NewService(bookingsRepo)
	bookingsHandler := bookings.NewHandler(bookingsService)

	rentalRepo := rental.NewPostgresRepository(pool)
	rentalService := rental.NewService(rentalRepo)
	rentalHandler := rental.NewHandler(rentalService)

	adminRepo := admin.NewPostgresRepository(pool)
	adminService := admin.NewService(adminRepo)
	adminHandler := admin.NewHandler(adminService)

	authStrictHandler := authapi.NewStrictHandler(authHandler, nil)
	catalogStrictHandler := catalogapi.NewStrictHandler(catalogHandler, nil)
	instructorsStrictHandelr := instructorsapi.NewStrictHandler(instructorsHandler, nil)
	servicesStrictHandler := servicesapi.NewStrictHandler(serviceHandler, nil)
	bookingsStrictHandler := bookingsapi.NewStrictHandler(bookingsHandler, nil)
	rentalStrictHandler := rentalapi.NewStrictHandler(rentalHandler, nil)
	adminStrictHandler := adminapi.NewStrictHandler(adminHandler, nil)

	mux := http.NewServeMux()

	authapi.HandlerFromMux(authStrictHandler, mux)
	catalogapi.HandlerFromMux(catalogStrictHandler, mux)
	instructorsapi.HandlerFromMux(instructorsStrictHandelr, mux)
	servicesapi.HandlerFromMux(servicesStrictHandler, mux)
	bookingsapi.HandlerFromMux(bookingsStrictHandler, mux)
	rentalapi.HandlerFromMux(rentalStrictHandler, mux)
	adminapi.HandlerFromMux(adminStrictHandler, mux)

	handler := middleware.AuthMiddleware([]byte(jwtSecret), mux)

	log.Println("PostgreSQL connected successfully")
	log.Println("HTTP server started on :8080")

	log.Fatal(http.ListenAndServe(":8080", handler))

}
