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
	Id               uint           `json:"id" binding:""`
	Name             string         `json:"name" binding:"required"`
	Tags             pq.StringArray `json:"tags" binding:"required" gorm:"type: TEXT[]"`
	ShortDescription string         `json:"shortDescription" binding:"required"`
	Skills           pq.StringArray `json:"skills" binding:"required" gorm:"type: TEXT[]"`
}

type ListProjectsParamsDto struct {
	PageSize   uint     `form:"pageSize"`
	PageOffset uint     `form:"pageOffset"`
	Tags       []string `form:"tags"`
	Skills     []string `form:"skills"`
}
