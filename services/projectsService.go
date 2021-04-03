package services

import (
	"github.com/open-collaboration/server/dtos"
	"github.com/open-collaboration/server/models"
	"gorm.io/gorm"
)

type ProjectsService struct {
	Db *gorm.DB
}

func (s *ProjectsService) CreateProject(newProject dtos.NewProjectDto) (*models.Project, error) {
	project := models.Project{
		Name:             newProject.Name,
		Tags:             newProject.Tags,
		LongDescription:  newProject.LongDescription,
		ShortDescription: newProject.ShortDescription,
		GithubLink:       newProject.GithubLink,
	}

	result := s.Db.Create(&project)
	if result.Error != nil {
		return nil, result.Error
	}

	return &project, nil
}

func (s *ProjectsService) GetProjectSummary(project *models.Project) dtos.ProjectSummaryDto {
	return dtos.ProjectSummaryDto{
		Name:             project.Name,
		Tags:             project.Tags,
		ShortDescription: project.ShortDescription,
		LinkUid:          project.LinkUid,
	}
}
