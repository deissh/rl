package api

import (
	"context"
	"github.com/deissh/rl/ayako/app"
	myctx "github.com/deissh/rl/ayako/ctx"
	"github.com/labstack/echo/v4"
	"net/http"
)

type MeHandlers struct {
	*app.App
}

func (h *MeHandlers) Me(c echo.Context) error {
	ctx, _ := c.Get("context").(context.Context)

	userId, err := myctx.GetUserID(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	mode := c.Param("mode")

	user, err := h.Store().User().Get(ctx, userId, mode)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	return c.JSON(200, user)
}
