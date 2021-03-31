package projects

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
	LinkUid          int `gorm:"autoIncrement"`
}

func (project *Project) GetSummary() ProjectSummaryDto {
	return ProjectSummaryDto{
		Name:             project.Name,
		Tags:             project.Tags,
		ShortDescription: project.ShortDescription,
		LinkUid:          project.LinkUid,
	}
}
