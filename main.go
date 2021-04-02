package main

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/open-collaboration/server/migrations"
	"github.com/open-collaboration/server/routes"
	"github.com/open-collaboration/server/services"
	apex_gin "github.com/thedanielforum/apex-gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

func main() {

	// Setup logger
	log.SetHandler(cli.New(os.Stdout))
	log.SetLevel(log.DebugLevel)

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
	server := gin.New()

	server.Use(cors.Default())
	server.Use(apex_gin.Handler("request"))

	providers := make([]interface{}, 0)
	providers = append(providers, &services.UsersService{Db: db})
	providers = append(providers, &services.ProjectsService{Db: db})

	routes.SetupRoutes(server, providers)

	// Start server
	addr := os.Getenv("HOST")
	err = server.Run(addr)
	if err != nil {
		log.WithError(err).Error("Failed to start the server.")
		panic(err)
	}
}
