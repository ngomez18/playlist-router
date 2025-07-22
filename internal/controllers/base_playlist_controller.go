package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/services"
	"github.com/pocketbase/pocketbase/apis"
)

type BasePlaylistController struct {
	basePlaylistService services.BasePlaylistServicer
}

func NewBasePlaylistController(bpService services.BasePlaylistServicer) *BasePlaylistController {
	return &BasePlaylistController{
		basePlaylistService: bpService,
	}
}

func (c *BasePlaylistController) Create(ctx echo.Context) error {
	var req models.CreateBasePlaylistRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil {
		return apis.NewBadRequestError("Invalid JSON payload", err)
	}

	newBasePlaylist, err := c.basePlaylistService.CreateBasePlaylist(&req)
	if err != nil {
		return apis.NewInternalServerError("unable to create base playlist", err)
	}

	return ctx.JSON(http.StatusCreated, newBasePlaylist)
}
