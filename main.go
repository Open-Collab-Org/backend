package main

import (
	"context"
	"fmt"
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/open-collaboration/server/migrations"
	"github.com/open-collaboration/server/routes"
	"github.com/open-collaboration/server/services"
	"github.com/open-collaboration/server/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
)

func main() {

	// Setup logging
	log.SetLevel(log.DebugLevel)
	log.SetHandler(cli.New(os.Stdout))

	// Load env variables
	err := godotenv.Load()
	if err != nil {
		log.WithError(err).Error("Failed to load environment variables.")
		panic(err)
	}

	// Setup db connection
	pgHost := os.Getenv("PG_HOST")
	pgPort := os.Getenv("PG_PORT")
	pgUser := os.Getenv("PG_USER")
	pgPassword := os.Getenv("PG_PASSWORD")
	pgDbName := os.Getenv("PG_DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", pgHost, pgPort, pgUser, pgPassword, pgDbName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.WithError(err).Error("Failed to connect to database.")
		panic(err)
	}

	// Run db migrations
	migration := migrations.GetMigration(db)
	err = migration.Migrate()
	if err != nil {
		log.WithError(err).Error("Failed to run database migrations.")
		panic(err)
	}

	// Setup redisDb connection
	redisHost := utils.GetEnvOrPanic("REDIS_HOST")
	redisPort := utils.GetEnvOrPanic("REDIS_PORT")
	redisDb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
	})

	// Test redisDb connection
	_, err = redisDb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	// Setup server
	providers := []interface{}{
		&services.AuthService{Db: db, Redis: redisDb},
		&services.UsersService{Db: db},
		&services.ProjectsService{Db: db},
	}

	router := routes.SetupRoutes(providers[:])

	host := utils.GetEnvOrPanic("HOST")
	port := utils.GetEnvOrPanic("PORT")
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, port),
		Handler: router,
	}

	// Start server
	err = server.ListenAndServe()
	if err != nil {
		log.WithError(err).Error("Failed to start the server.")
		panic(err)
	}
}
