package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

var usersTable = gormigrate.Migration{
	ID: "0",
	Migrate: func(db *gorm.DB) error {
		type User struct {
			gorm.Model
			Username     string
			Email        string
			PasswordHash string
		}

		return db.AutoMigrate(&User{})
	},
	Rollback: func(db *gorm.DB) error {
		return db.Migrator().DropTable("users")
	},
}

var projectsTable = gormigrate.Migration{
	ID: "1",
	Migrate: func(db *gorm.DB) error {
		type Project struct {
			gorm.Model

			Name             string         `gorm:"type: VARCHAR(32)"`
			Tags             pq.StringArray `gorm:"type: TEXT[]"`
			LongDescription  string         `gorm:"type: VARCHAR(10000)"`
			ShortDescription string         `gorm:"type: VARCHAR(200)"`
			GithubLink       string
			LinkUid          int `gorm:"autoIncrement"`
		}

		return db.AutoMigrate(&Project{})
	},
	Rollback: func(db *gorm.DB) error {
		return db.Migrator().DropTable("projects")
	},
}

func GetMigration(db *gorm.DB) *gormigrate.Gormigrate {
	return gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		&usersTable,
		&projectsTable,
	})
}
