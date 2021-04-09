package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model

	Name             string
	Tags             pq.StringArray `gorm:"type: TEXT[]"`
	LongDescription  string
	ShortDescription string
	GithubLink       string
}
