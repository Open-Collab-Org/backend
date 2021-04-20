package projects

import "github.com/lib/pq"

type NewProjectDto struct {
	Name             string   `json:"name" validate:"required,min=4,max=32"`
	Tags             []string `json:"tags" validate:"required,min=1,max=6,dive,min=1,max=40"`
	LongDescription  string   `json:"longDescription" validate:"required,min=200,max=10000"`
	ShortDescription string   `json:"shortDescription" validate:"required,min=10,max=200"`
	GithubLink       string   `json:"githubLink" validate:"required"`
}

type ProjectSummaryDto struct {
	Id               uint           `json:"id" validate:""`
	Name             string         `json:"name" validate:"required"`
	Tags             pq.StringArray `json:"tags" validate:"required" gorm:"type: TEXT[]" swaggertype:"array,string"`
	ShortDescription string         `json:"shortDescription" validate:"required"`
	Skills           pq.StringArray `json:"skills" validate:"required" gorm:"type: TEXT[]" swaggertype:"array,string"`
}

type ProjectDto struct {
	Id               uint           `json:"id"`
	Name             string         `json:"name"`
	Tags             pq.StringArray `json:"tags" swaggertype:"array,string"`
	ShortDescription string         `json:"shortDescription"`
	LongDescription  string         `json:"fullDescription"`
	GithubLink       string         `json:"githubLink"`
}

type ListProjectsParamsDto struct {
	PageSize   uint     `form:"pageSize"`
	PageOffset uint     `form:"pageOffset"`
	Tags       []string `form:"tags"`
	Skills     []string `form:"skills"`
}
