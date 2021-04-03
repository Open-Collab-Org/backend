package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/open-collaboration/server/dtos"
	"github.com/open-collaboration/server/models"
	"github.com/open-collaboration/server/services"
)

func RouteCreateProject(c *gin.Context, projectsService *services.ProjectsService) error {
	newProject := dtos.NewProjectDto{}
	err := c.ShouldBind(&newProject)
	if err != nil {
		return err
	}

	createdProject, err := projectsService.CreateProject(newProject)
	if err != nil {
		return err
	}

	c.Header("Location", "/projects/"+string(rune(createdProject.LinkUid)))

	c.JSON(201, projectsService.GetProjectSummary(createdProject))

	return nil
}

func RouteFetchProjects(c *gin.Context, projectsService *services.ProjectsService) error {
	projectSummaries := make([]dtos.ProjectSummaryDto, 20)
	result := projectsService.Db.
		Model(&models.Project{}).
		Select("name", "tags", "short_description", "link_uid").
		Order("created_at desc").
		Limit(20).
		Find(&projectSummaries)

	if result.Error != nil {
		return result.Error
	}

	c.JSON(200, projectSummaries)

	return nil
}
