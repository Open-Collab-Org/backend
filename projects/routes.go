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

	c.JSON(201, createdProject.GetSummary())

	return nil
}
