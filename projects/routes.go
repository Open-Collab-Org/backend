package projects

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RouteCreateProject(c *gin.Context, db *gorm.DB) error {
	newProject := NewProjectDto{}
	err := c.ShouldBind(&newProject)
	if err != nil {
		return err
	}

	createdProject, err := CreateProject(db, newProject)
	if err != nil {
		return err
	}

	c.Header("Location", "/projects/"+string(rune(createdProject.LinkUid)))

	c.JSON(201, createdProject.GetSummary())

	return nil
}

func RouteFetchProjects(c *gin.Context, db *gorm.DB) error {
	projectSummaries := make([]ProjectSummaryDto, 20)
	result := db.
		Model(&Project{}).
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
