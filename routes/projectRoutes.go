package routes

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/open-collaboration/server/dtos"
	"github.com/open-collaboration/server/httpUtils"
	"github.com/open-collaboration/server/services"
	"net/http"
	"strconv"
)

func RouteCreateProject(writer http.ResponseWriter, request *http.Request, projectsService *services.ProjectsService) error {

	dto := dtos.NewProjectDto{}
	err := httpUtils.ReadJson(request, dto)
	if err != nil {
		return err
	}

	createdProject, err := projectsService.CreateProject(dto)
	if err != nil {
		return err
	}

	projectSummary := projectsService.GetProjectSummary(createdProject)
	responseBody, err := json.Marshal(projectSummary)
	if err != nil {
		return err
	}

	writer.Header().Set("Location", "/projects/"+string(rune(createdProject.ID)))

	_, err = writer.Write(responseBody)
	if err != nil {
		return err
	}

	return nil
}

func RouteListProjects(writer http.ResponseWriter, request *http.Request, projectsService *services.ProjectsService) error {
	// TODO: move hardcoded maximum and default page size values to
	// 	an env variable
	pageSize, _ := httpUtils.IntFromQuery(request, "pageSize", 20)
	pageOffset, _ := httpUtils.IntFromQuery(request, "pageOffset", 0)

	if pageSize < 1 || pageSize > 20 {
		pageSize = 20
	}

	if pageOffset < 1 {
		pageOffset = 0
	}

	projectSummaries, err := projectsService.ListProjects(context.TODO(), uint(pageSize), uint(pageOffset), []string{}, []string{})
	if err != nil {
		return err
	}

	err = httpUtils.WriteJson(writer, context.TODO(), projectSummaries)
	if err != nil {
		return err
	}

	return nil
}

func RouteGetProject(writer http.ResponseWriter, request *http.Request, projectsService *services.ProjectsService) error {
	var projectId uint
	vars := mux.Vars(request)
	if idStr, ok := vars["projectId"]; ok {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return err
		}

		projectId = uint(id)
	}

	dto, err := projectsService.GetProject(context.TODO(), projectId)
	if err != nil {
		if errors.Is(err, services.ProjectNotFoundError) {
			writer.WriteHeader(404)
			return nil
		} else {
			return err
		}
	}

	err = httpUtils.WriteJson(writer, context.TODO(), dto)
	if err != nil {
		return err
	}

	return nil
}
