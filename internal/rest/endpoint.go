package rest

import "github.com/labstack/echo/v4"

type RestEndpoint interface {
	Register(e *echo.Echo) error
}
