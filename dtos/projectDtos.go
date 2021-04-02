package dtos

import "github.com/lib/pq"

type NewProjectDto struct {
	Name             string   `json:"name" binding:"required,min=4,max=32"`
	Tags             []string `json:"tags" binding:"required,min=1,max=6,dive,min=1,max=40"`
	LongDescription  string   `json:"longDescription" binding:"required,min=200,max=10000"`
	ShortDescription string   `json:"shortDescription" binding:"required,min=10,max=200"`
	GithubLink       string   `json:"githubLink" binding:"required"`
}

type ProjectSummaryDto struct {
	Name             string         `json:"name" binding:"required"`
	Tags             pq.StringArray `json:"tags" binding:"required" gorm:"type: TEXT[]"`
	ShortDescription string         `json:"shortDescription" binding:"required"`
	LinkUid          int            `json:"linkUid" binding:""`
	Skills           pq.StringArray `json:"skills" binding:"required" gorm:"type: TEXT[]"`
}
