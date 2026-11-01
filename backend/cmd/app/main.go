package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/darkline28/ski-resort/backend/internal/auth"
	"github.com/darkline28/ski-resort/backend/internal/auth/api"
	"github.com/darkline28/ski-resort/backend/internal/database"
	"github.com/darkline28/ski-resort/backend/internal/users"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal(err)
	}
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	port := os.Getenv("POSTGRES_PORT")
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtTTL := os.Getenv("JWT_TTL")

	dsn := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", user, password, port, dbName)
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

	strictHandler := api.NewStrictHandler(authHandler, nil)

	mux := http.NewServeMux()
	handler := api.HandlerFromMux(strictHandler, mux)

	log.Println("PostgreSQL connected successfully")
	log.Println("HTTP server started on :8080")

	log.Fatal(http.ListenAndServe(":8080", handler))

}
