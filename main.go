package main

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/gorilla/mux"
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
	dsn := "host=localhost user=root password=changeme dbname=opencollab port=5432 sslmode=disable"
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

	router := mux.NewRouter()

	routes.SetupRoutes(router, providers[:])

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
