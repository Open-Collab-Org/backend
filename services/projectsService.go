package services

import (
	"context"
	"github.com/apex/log"
	"github.com/lib/pq"
	"github.com/open-collaboration/server/dtos"
	"github.com/open-collaboration/server/logging"
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
		Id:               project.ID,
		Name:             project.Name,
		Tags:             project.Tags,
		ShortDescription: project.ShortDescription,
	}
}

func (s *ProjectsService) ListProjects(ctx context.Context, pageSize uint, pageOffset uint, tags []string, skills []string) ([]dtos.ProjectSummaryDto, error) {
	logger := logging.LoggerFromCtx(ctx)

	logger.WithFields(log.Fields{
		"page_size":   pageSize,
		"page_offset": pageOffset,
		"tags":        tags,
		"skills":      skills,
	}).
		Debug("Listing projects")

	projectSummaries := make([]dtos.ProjectSummaryDto, pageSize)
	result := s.Db.
		Model(&models.Project{}).
		Select("name", "tags", "short_description", "id").
		Where("cardinality(?::TEXT[]) < 1 OR tags && ?", pq.StringArray(tags), pq.StringArray(tags)).
		Order("created_at desc").
		Limit(int(pageSize)).
		Offset(int(pageOffset * pageSize)).
		Find(&projectSummaries)

	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to list projects")

		return nil, result.Error
	}

	logger.Debugf("Found %d projects", result.RowsAffected)

	return projectSummaries[:result.RowsAffected], nil
}
