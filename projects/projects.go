package projects

import "gorm.io/gorm"

type NewProjectDto struct {
	Name             string   `json:"name" binding:"required,min=4,max=32"`
	Tags             []string `json:"tags" binding:"required,min=1,max=6,dive,min=1,max=40"`
	LongDescription  string   `json:"longDescription" binding:"required,min=200,max=10000"`
	ShortDescription string   `json:"shortDescription" binding:"required,min=10,max=200"`
	GithubLink       string   `json:"githubLink" binding:"required"`
}

type ProjectSummaryDto struct {
	Name             string   `json:"name" binding:"required"`
	Tags             []string `json:"tags" binding:"required"`
	ShortDescription string   `json:"shortDescription" binding:"required"`
	LinkUid          int      `json:"linkUid" binding:""`
}

func CreateProject(db *gorm.DB, newProject NewProjectDto) (*Project, error) {
	project := Project{
		Name:             newProject.Name,
		Tags:             newProject.Tags,
		LongDescription:  newProject.LongDescription,
		ShortDescription: newProject.ShortDescription,
		GithubLink:       newProject.GithubLink,
	}

	result := db.Create(&project)
	if result.Error != nil {
		return nil, result.Error
	}

	return &project, nil
}
