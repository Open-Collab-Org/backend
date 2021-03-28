package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
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

func GetMigration(db *gorm.DB) *gormigrate.Gormigrate {
	return gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		&usersTable,
	})
}
