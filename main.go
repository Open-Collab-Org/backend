package main

import (
	"fmt"
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/joho/godotenv"
	"github.com/open-collaboration/server/migrations"
	"github.com/open-collaboration/server/routes"
	"github.com/open-collaboration/server/services"
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

	// Setup server
	providers := []interface{}{
		&services.UsersService{Db: db},
		&services.ProjectsService{Db: db},
	}

	router := routes.SetupRoutes(providers[:])

	addr := os.Getenv("HOST")
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	/*
		server.Use(gin.Recovery())
		server.Use(logging.LoggerMiddleware)
		server.Use(cors.Default())
	*/

	// Start server
	err = server.ListenAndServe()
	if err != nil {
		log.WithError(err).Error("Failed to start the server.")
		panic(err)
	}
}
