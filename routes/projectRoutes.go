package routes

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/open-collaboration/server/dtos"
	"github.com/open-collaboration/server/services"
	"net/http"
	"strconv"
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

func RouteListProjects(writer http.ResponseWriter, request *http.Request, projectsService *services.ProjectsService) error {
	queryParams := &dtos.ListProjectsParamsDto{}

	/*err := c.ShouldBindQuery(queryParams)
	if err != nil {
		if errors.Is(err, strconv.ErrSyntax) || errors.Is(err, strconv.ErrRange) {
			c.Status(400)

			return nil
		} else {
			return err
		}
	}*/

	// TODO: move hardcoded maximum and default page size values to
	// 	an env variable
	/*if queryParams.PageSize == 0 || queryParams.PageSize > 20 {
		queryParams.PageSize = 20
	}*/

	projectSummaries, err := projectsService.ListProjects(context.TODO(), queryParams.PageSize, queryParams.PageOffset, []string{}, []string{})

	if err != nil {
		return err
	}

	//c.JSON(200, projectSummaries)

	responseBytes, _ := json.Marshal(projectSummaries)
	/*if err != nil {

	}*/

	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Content-Length", strconv.Itoa(len(responseBytes)))
	_, _ = writer.Write(responseBytes)
	/*if err != nil {

	}*/

	return nil
}
