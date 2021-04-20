package main

import (
	"context"
	"fmt"
	"github.com/apex/log"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/open-collaboration/server/auth"
	"github.com/open-collaboration/server/migrations"
	"github.com/open-collaboration/server/projects"
	router2 "github.com/open-collaboration/server/router"
	"github.com/open-collaboration/server/users"
	"github.com/open-collaboration/server/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	"os"
)

func main() {

	// Setup logging
	log.SetLevel(log.DebugLevel)
	log.SetHandler(utils.NewTerminalLogger(os.Stdout))

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
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Interface(&utils.GormLogger{}),
	})
	if err != nil {
		log.WithError(err).Error("Failed to connect to database.")
		panic(err)
	}

	db = db.Debug()

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
	usersService := &users.Service{Db: db}

	providers := []interface{}{
		&auth.Service{Db: db, Redis: redisDb, UsersService: usersService},
		usersService,
		&projects.Service{Db: db},
	}

	router := router2.SetupRoutes(providers[:])

	host := utils.GetEnvOrPanic("HOST")
	port := utils.GetEnvOrPanic("PORT")
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, port),
		Handler: router,
	}

	log.Infof("Serving at %s", server.Addr)

	// Start server
	err = server.ListenAndServe()
	if err != nil {
		log.WithError(err).Error("Failed to start the server.")
		panic(err)
	}
}
