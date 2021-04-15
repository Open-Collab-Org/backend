package services

import (
	"context"
	"errors"
	"github.com/apex/log"
	"github.com/lib/pq"
	"github.com/open-collaboration/server/dtos"
	"github.com/open-collaboration/server/models"
	"gorm.io/gorm"
)

type ProjectsService struct {
	Db *gorm.DB
}

var ProjectNotFoundError = errors.New("project not found")

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

// Get the given project's summary
func (s *ProjectsService) GetProjectSummary(project *models.Project) dtos.ProjectSummaryDto {
	return dtos.ProjectSummaryDto{
		Id:               project.ID,
		Name:             project.Name,
		Tags:             project.Tags,
		ShortDescription: project.ShortDescription,
	}
}

// Get a project by id.
// Returns ProjectNotFoundError if the project can't be found.
func (s *ProjectsService) GetProject(ctx context.Context, projectId uint) (dtos.ProjectDto, error) {
	logger := log.FromContext(ctx)

	logger.Debugf("Querying for project of id %d", projectId)

	project := models.Project{}
	result := s.Db.First(&project, projectId)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Debugf("Project of id %d was not found", projectId)
			return dtos.ProjectDto{}, ProjectNotFoundError
		} else {
			logger.WithError(result.Error).Errorf("Failed to query for project of id %d", projectId)
			return dtos.ProjectDto{}, result.Error
		}
	}

	logger.Debugf("Project of id %d was found", projectId)

	return dtos.ProjectDto{
		Id:               project.ID,
		Name:             project.Name,
		Tags:             project.Tags,
		ShortDescription: project.ShortDescription,
		LongDescription:  project.LongDescription,
		GithubLink:       project.GithubLink,
	}, nil
}

// List all projects ordered by creation date, newest to oldest.
//
// Results are returned in "pages". A page is determined by the pageSize and
// pageOffset parameters. pageSize determines the maximum amount of projects
// that can be returned, and page offset determines how many pages (i.e. projects)
// to skip. For example: if pageSize is 20 and pageOffset is 3, a maximum of 20
// projects will be returned and 60 (3x20) projects will be skipped.
//
// You can also filter the results by tags and skills. If tags is specified
// (non-nil and non-empty), any projects that have at least one of the specified
// tags will be returned. If skills is specified (non-nil and non-empty), any projects
// that have at least one role that require at least one of the specified skills will
// be returned.
func (s *ProjectsService) ListProjects(
	ctx context.Context,
	pageSize uint,
	pageOffset uint,
	tags []string,
	skills []string,
) ([]dtos.ProjectSummaryDto, error) {
	logger := log.FromContext(ctx)

	logger.WithFields(log.Fields{
		"page_size":   pageSize,
		"page_offset": pageOffset,
		"tags":        tags,
		"skills":      skills,
	}).
		Debug("Listing projects")

	if tags == nil {
		tags = []string{}
	}

	if skills == nil {
		skills = []string{}
	}

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
